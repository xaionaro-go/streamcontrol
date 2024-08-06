package client

import (
	"bytes"
	"context"
	"crypto"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/andreykaipov/goobs/api/requests/scenes"
	"github.com/andreykaipov/goobs/api/typedefs"
	"github.com/facebookincubator/go-belt"
	"github.com/facebookincubator/go-belt/tool/logger"
	"github.com/goccy/go-yaml"
	"github.com/hashicorp/go-multierror"
	"github.com/xaionaro-go/streamctl/pkg/observability"
	"github.com/xaionaro-go/streamctl/pkg/player"
	"github.com/xaionaro-go/streamctl/pkg/player/protobuf/go/player_grpc"
	"github.com/xaionaro-go/streamctl/pkg/streamcontrol"
	obs "github.com/xaionaro-go/streamctl/pkg/streamcontrol/obs/types"
	twitch "github.com/xaionaro-go/streamctl/pkg/streamcontrol/twitch/types"
	youtube "github.com/xaionaro-go/streamctl/pkg/streamcontrol/youtube/types"
	"github.com/xaionaro-go/streamctl/pkg/streamd/api"
	streamdconfig "github.com/xaionaro-go/streamctl/pkg/streamd/config"
	"github.com/xaionaro-go/streamctl/pkg/streamd/grpc/go/streamd_grpc"
	"github.com/xaionaro-go/streamctl/pkg/streamd/grpc/goconv"
	"github.com/xaionaro-go/streamctl/pkg/streampanel/consts"
	sptypes "github.com/xaionaro-go/streamctl/pkg/streamplayer/types"
	"github.com/xaionaro-go/streamctl/pkg/streamserver/types"
	"github.com/xaionaro-go/streamctl/pkg/streamtypes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type Client struct {
	Target string
	Config Config

	PersistentConnectionLocker sync.Mutex
	PersistentConnection       *grpc.ClientConn
	PersistentClient           streamd_grpc.StreamDClient
}

var _ api.StreamD = (*Client)(nil)

func New(
	ctx context.Context,
	target string,
	opts ...Option,
) (*Client, error) {
	c := &Client{
		Target: target,
		Config: Options(opts).Config(ctx),
	}
	if err := c.init(ctx); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) init(ctx context.Context) error {
	var result *multierror.Error
	if c.Config.UsePersistentConnection {
		result = multierror.Append(result, c.initPersistentConnection(ctx))
	}
	return result.ErrorOrNil()
}

func (c *Client) initPersistentConnection(ctx context.Context) error {
	conn, err := c.connect(ctx)
	if err != nil {
		return err
	}
	c.PersistentConnection = conn
	c.PersistentClient = streamd_grpc.NewStreamDClient(conn)
	return nil
}

func (c *Client) connect(ctx context.Context) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{
			Backoff: backoff.Config{
				BaseDelay:  c.Config.Reconnect.InitialInterval,
				Multiplier: c.Config.Reconnect.IntervalMultiplier,
				Jitter:     0.2,
				MaxDelay:   c.Config.Reconnect.MaximalInterval,
			},
			MinConnectTimeout: c.Config.Reconnect.InitialInterval,
		}),
	}
	wrapper := c.Config.ConnectWrapper
	if wrapper == nil {
		return c.doConnect(ctx, opts...)
	}

	return wrapper(
		ctx,
		func(ctx context.Context, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
			return c.doConnect(ctx, opts...)
		},
		opts...,
	)
}

func (c *Client) doConnect(ctx context.Context, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	logger.Debugf(ctx, "doConnect(ctx, %#+v): Config: %#+v", opts, c.Config)
	delay := c.Config.Reconnect.InitialInterval
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		logger.Debugf(ctx, "trying to (re-)connect to %s", c.Target)

		dialCtx, cancelFn := context.WithTimeout(ctx, time.Second)
		conn, err := grpc.DialContext(dialCtx, c.Target, opts...)
		cancelFn()
		if err == nil {
			logger.Debugf(ctx, "successfully (re-)connected to %s", c.Target)
			return conn, nil
		}
		logger.Debugf(ctx, "(re-)connection failed to %s: %v; sleeping %v before the next try", c.Target, err, delay)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}
		delay = time.Duration(float64(delay) * c.Config.Reconnect.IntervalMultiplier)
		if delay > c.Config.Reconnect.MaximalInterval {
			delay = c.Config.Reconnect.MaximalInterval
		}
	}
}

func (c *Client) grpcClient(ctx context.Context) (streamd_grpc.StreamDClient, io.Closer, error) {
	if c.Config.UsePersistentConnection {
		return c.grpcPersistentClient(ctx)
	} else {
		return c.grpcNewClient(ctx)
	}
}

func (c *Client) grpcNewClient(ctx context.Context) (streamd_grpc.StreamDClient, *grpc.ClientConn, error) {
	conn, err := c.connect(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to initialize a gRPC client: %w", err)
	}

	client := streamd_grpc.NewStreamDClient(conn)
	return client, conn, nil
}

func (c *Client) grpcPersistentClient(context.Context) (streamd_grpc.StreamDClient, dummyCloser, error) {
	c.PersistentConnectionLocker.Lock()
	defer c.PersistentConnectionLocker.Unlock()
	return c.PersistentClient, dummyCloser{}, nil
}

func (c *Client) Run(ctx context.Context) error {
	return nil
}

func callWrapper[REQ any, REPLY any](
	ctx context.Context,
	c *Client,
	fn func(context.Context, *REQ, ...grpc.CallOption) (REPLY, error),
	req *REQ,
	opts ...grpc.CallOption,
) (REPLY, error) {

	var reply REPLY
	callFn := func(ctx context.Context, opts ...grpc.CallOption) error {
		var err error
		delay := c.Config.Reconnect.InitialInterval
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			reply, err = fn(ctx, req, opts...)
			if err == nil {
				return nil
			}
			err = c.processError(ctx, err)
			if err != nil {
				return err
			}
			logger.Debugf(ctx, "retrying; sleeping %v for the retry", delay)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
			delay = time.Duration(float64(delay) * c.Config.Reconnect.IntervalMultiplier)
			if delay > c.Config.Reconnect.MaximalInterval {
				delay = c.Config.Reconnect.MaximalInterval
			}
		}
	}

	wrapper := c.Config.CallWrapper
	if wrapper == nil {
		err := callFn(ctx, opts...)
		return reply, err
	}

	err := wrapper(ctx, req, callFn, opts...)
	return reply, err
}

func withClient[REPLY any](
	ctx context.Context,
	c *Client,
	fn func(context.Context, streamd_grpc.StreamDClient, io.Closer) (*REPLY, error),
) (*REPLY, error) {
	pc, _, _, _ := runtime.Caller(1)
	caller := runtime.FuncForPC(pc)
	ctx = belt.WithField(ctx, "caller_func", caller.Name())

	client, conn, err := c.grpcClient(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	if client == nil {
		return nil, fmt.Errorf("internal error: client is nil")
	}
	return fn(ctx, client, conn)
}

type receiver[T any] interface {
	grpc.ClientStream

	Recv() (*T, error)
}

func unwrapChan[E any, R any, S receiver[R]](
	ctx context.Context,
	c *Client,
	fn func(ctx context.Context, client streamd_grpc.StreamDClient) (S, error),
	parse func(ctx context.Context, event *R) E,
) (<-chan E, error) {

	pc, _, _, _ := runtime.Caller(1)
	caller := runtime.FuncForPC(pc)
	ctx = belt.WithField(ctx, "caller_func", caller.Name())

	ctx, cancelFn := context.WithCancel(ctx)
	getSub := func() (S, io.Closer, error) {
		client, closer, err := c.grpcClient(ctx)
		if err != nil {
			var emptyS S
			return emptyS, nil, err
		}
		sub, err := fn(ctx, client)
		if err != nil {
			var emptyS S
			return emptyS, nil, fmt.Errorf("unable to subscribe: %w", err)
		}
		return sub, closer, nil
	}

	sub, closer, err := getSub()
	if err != nil {
		cancelFn()
		return nil, err
	}

	r := make(chan E)
	observability.Go(ctx, func() {
		defer closer.Close()
		defer cancelFn()
		for {
			event, err := sub.Recv()
			select {
			case <-ctx.Done():
				return
			default:
			}
			if err != nil {
				switch {
				case errors.Is(err, io.EOF):
					logger.Debugf(ctx, "the receiver is closed: %v", err)
					return
				case strings.Contains(err.Error(), grpc.ErrClientConnClosing.Error()):
					logger.Debugf(ctx, "apparently we are closing the client: %v", err)
					return
				case strings.Contains(err.Error(), context.Canceled.Error()):
					logger.Debugf(ctx, "subscription was cancelled: %v", err)
					return
				default:
					for {
						err = c.processError(ctx, err)
						if err != nil {
							logger.Errorf(ctx, "unable to read data: %v", err)
							return
						}
						closer.Close()
						sub, closer, err = getSub()
						if err != nil {
							logger.Errorf(ctx, "unable to resubscribe: %v", err)
							continue
						}
						break
					}
					continue
				}
			}

			r <- parse(ctx, event)
		}
	})
	return r, nil
}

func (c *Client) processError(
	ctx context.Context,
	err error,
) error {
	logger.Debugf(ctx, "processError(ctx, '%v'): %T", err, err)
	if s, ok := status.FromError(err); ok {
		logger.Debugf(ctx, "processError(ctx, '%v'): code == %#+v; msg == %#+v", err, s.Code(), s.Message())
		switch s.Code() {
		case codes.Unavailable:
			logger.Debugf(ctx, "suppressed the error (forcing a retry)")
			return nil
		}
	}
	return err
}

func (c *Client) Ping(
	ctx context.Context,
	beforeSend func(context.Context, *streamd_grpc.PingRequest),
) error {
	_, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.PingReply, error) {
		req := &streamd_grpc.PingRequest{}
		beforeSend(ctx, req)
		return callWrapper(ctx, c, client.Ping, req)
	})
	return err
}

func (c *Client) FetchConfig(ctx context.Context) error {
	_, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.OBSOLETE_FetchConfigReply, error) {
		return callWrapper(ctx, c, client.OBSOLETE_FetchConfig, &streamd_grpc.OBSOLETE_FetchConfigRequest{})
	})
	return err
}

func (c *Client) InitCache(ctx context.Context) error {
	_, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.InitCacheReply, error) {
		return callWrapper(ctx, c, client.InitCache, &streamd_grpc.InitCacheRequest{})
	})
	return err
}

func (c *Client) SaveConfig(ctx context.Context) error {
	_, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.SaveConfigReply, error) {
		return callWrapper(ctx, c, client.SaveConfig, &streamd_grpc.SaveConfigRequest{})
	})
	return err
}

func (c *Client) ResetCache(ctx context.Context) error {
	_, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.ResetCacheReply, error) {
		return callWrapper(ctx, c, client.ResetCache, &streamd_grpc.ResetCacheRequest{})
	})
	return err
}

func (c *Client) GetConfig(ctx context.Context) (*streamdconfig.Config, error) {
	reply, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.GetConfigReply, error) {
		return callWrapper(ctx, c, client.GetConfig, &streamd_grpc.GetConfigRequest{})
	})
	if err != nil {
		return nil, err
	}

	var result streamdconfig.Config
	_, err = result.Read([]byte(reply.Config))
	if err != nil {
		return nil, fmt.Errorf("unable to unserialize the received config: %w", err)
	}
	return &result, nil
}

func (c *Client) SetConfig(ctx context.Context, cfg *streamdconfig.Config) error {
	var buf bytes.Buffer
	_, err := cfg.WriteTo(&buf)
	if err != nil {
		return fmt.Errorf("unable to serialize the config: %w", err)
	}

	_, err = withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.SetConfigReply, error) {
		return callWrapper(ctx, c, client.SetConfig, &streamd_grpc.SetConfigRequest{
			Config: buf.String(),
		})
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) IsBackendEnabled(ctx context.Context, id streamcontrol.PlatformName) (bool, error) {
	reply, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.IsBackendEnabledReply, error) {
		return callWrapper(ctx, c, client.IsBackendEnabled, &streamd_grpc.IsBackendEnabledRequest{
			PlatID: string(id),
		})
	})
	if err != nil {
		return false, err
	}
	return reply.IsInitialized, nil
}

func (c *Client) OBSOLETE_IsGITInitialized(ctx context.Context) (bool, error) {
	reply, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.OBSOLETE_GetGitInfoReply, error) {
		return callWrapper(ctx, c, client.OBSOLETE_GitInfo, &streamd_grpc.OBSOLETE_GetGitInfoRequest{})
	})
	if err != nil {
		return false, err
	}
	return reply.IsInitialized, nil
}

func (c *Client) StartStream(
	ctx context.Context,
	platID streamcontrol.PlatformName,
	title string, description string,
	profile streamcontrol.AbstractStreamProfile,
	customArgs ...any,
) error {
	b, err := yaml.Marshal(profile)
	if err != nil {
		return fmt.Errorf("unable to serialize the profile: %w", err)
	}
	logger.Debugf(ctx, "serialized profile: '%s'", profile)
	_, err = withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.StartStreamReply, error) {
		return callWrapper(ctx, c, client.StartStream, &streamd_grpc.StartStreamRequest{
			PlatID:      string(platID),
			Title:       title,
			Description: description,
			Profile:     string(b),
		})
	})
	return err
}
func (c *Client) EndStream(ctx context.Context, platID streamcontrol.PlatformName) error {
	_, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.EndStreamReply, error) {
		return callWrapper(ctx, c, client.EndStream, &streamd_grpc.EndStreamRequest{PlatID: string(platID)})
	})
	return err
}

func (c *Client) OBSOLETE_GitRelogin(ctx context.Context) error {
	_, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.OBSOLETE_GitReloginReply, error) {
		return callWrapper(ctx, c, client.OBSOLETE_GitRelogin, &streamd_grpc.OBSOLETE_GitReloginRequest{})
	})
	return err
}

func (c *Client) GetBackendData(ctx context.Context, platID streamcontrol.PlatformName) (any, error) {
	reply, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.GetBackendInfoReply, error) {
		return callWrapper(ctx, c, client.GetBackendInfo, &streamd_grpc.GetBackendInfoRequest{
			PlatID: string(platID),
		})
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get backend info: %w", err)
	}

	var data any
	switch platID {
	case obs.ID:
		_data := api.BackendDataOBS{}
		err = json.Unmarshal([]byte(reply.GetData()), &_data)
		data = _data
	case twitch.ID:
		_data := api.BackendDataTwitch{}
		err = json.Unmarshal([]byte(reply.GetData()), &_data)
		data = _data
	case youtube.ID:
		_data := api.BackendDataYouTube{}
		err = json.Unmarshal([]byte(reply.GetData()), &_data)
		data = _data
	default:
		return nil, fmt.Errorf("unknown platform: '%s'", platID)
	}

	if err != nil {
		return nil, fmt.Errorf("unable to deserialize data: %w", err)
	}
	return data, nil
}

func (c *Client) Restart(ctx context.Context) error {
	_, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.RestartReply, error) {
		return callWrapper(ctx, c, client.Restart, &streamd_grpc.RestartRequest{})
	})
	return err
}

func (c *Client) EXPERIMENTAL_ReinitStreamControllers(ctx context.Context) error {
	_, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.EXPERIMENTAL_ReinitStreamControllersReply, error) {
		return callWrapper(ctx, c, client.EXPERIMENTAL_ReinitStreamControllers, &streamd_grpc.EXPERIMENTAL_ReinitStreamControllersRequest{})
	})
	return err
}

func (c *Client) GetStreamStatus(
	ctx context.Context,
	platID streamcontrol.PlatformName,
) (*streamcontrol.StreamStatus, error) {
	streamStatus, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.GetStreamStatusReply, error) {
		return callWrapper(ctx, c, client.GetStreamStatus, &streamd_grpc.GetStreamStatusRequest{
			PlatID: string(platID),
		})
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get the stream status of '%s': %w", platID, err)
	}

	var startedAt *time.Time
	if streamStatus != nil && streamStatus.StartedAt != nil {
		v := *streamStatus.StartedAt
		startedAt = ptr(time.Unix(v/1000000000, v%1000000000))
	}

	var customData any
	switch platID {
	case youtube.ID:
		d := youtube.StreamStatusCustomData{}
		err := json.Unmarshal([]byte(streamStatus.GetCustomData()), &d)
		if err != nil {
			return nil, fmt.Errorf("unable to unserialize the custom data: %w", err)
		}
		customData = d
	}

	return &streamcontrol.StreamStatus{
		IsActive:   streamStatus.GetIsActive(),
		StartedAt:  startedAt,
		CustomData: customData,
	}, nil
}

func (c *Client) SetTitle(
	ctx context.Context,
	platID streamcontrol.PlatformName,
	title string,
) error {
	_, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.SetTitleReply, error) {
		return callWrapper(ctx, c, client.SetTitle, &streamd_grpc.SetTitleRequest{
			PlatID: string(platID),
			Title:  title,
		})
	})
	return err
}
func (c *Client) SetDescription(
	ctx context.Context,
	platID streamcontrol.PlatformName,
	description string,
) error {
	_, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.SetDescriptionReply, error) {
		return callWrapper(ctx, c, client.SetDescription, &streamd_grpc.SetDescriptionRequest{
			PlatID:      string(platID),
			Description: description,
		})
	})
	return err
}
func (c *Client) ApplyProfile(
	ctx context.Context,
	platID streamcontrol.PlatformName,
	profile streamcontrol.AbstractStreamProfile,
	customArgs ...any,
) error {
	b, err := yaml.Marshal(profile)
	if err != nil {
		return fmt.Errorf("unable to serialize the profile: %w", err)
	}
	logger.Debugf(ctx, "serialized profile: '%s'", profile)

	_, err = withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.ApplyProfileReply, error) {
		return callWrapper(ctx, c, client.ApplyProfile, &streamd_grpc.ApplyProfileRequest{
			PlatID:  string(platID),
			Profile: string(b),
		})
	})
	return err
}

func (c *Client) UpdateStream(
	ctx context.Context,
	platID streamcontrol.PlatformName,
	title string, description string,
	profile streamcontrol.AbstractStreamProfile,
	customArgs ...any,
) error {
	b, err := yaml.Marshal(profile)
	if err != nil {
		return fmt.Errorf("unable to serialize the profile: %w", err)
	}
	logger.Debugf(ctx, "serialized profile: '%s'", profile)

	_, err = withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.UpdateStreamReply, error) {
		return callWrapper(ctx, c, client.UpdateStream, &streamd_grpc.UpdateStreamRequest{
			PlatID:      string(platID),
			Title:       title,
			Description: description,
			Profile:     string(b),
		})
	})
	return err
}

func (c *Client) SubscribeToOAuthURLs(
	ctx context.Context,
	listenPort uint16,
) (<-chan *streamd_grpc.OAuthRequest, error) {
	return unwrapChan(
		ctx,
		c,
		func(
			ctx context.Context,
			client streamd_grpc.StreamDClient,
		) (streamd_grpc.StreamD_SubscribeToOAuthRequestsClient, error) {
			return callWrapper(
				ctx,
				c,
				client.SubscribeToOAuthRequests,
				&streamd_grpc.SubscribeToOAuthRequestsRequest{
					ListenPort: int32(listenPort),
				},
			)
		},
		func(
			ctx context.Context,
			event *streamd_grpc.OAuthRequest,
		) *streamd_grpc.OAuthRequest {
			return event
		},
	)
}

func (c *Client) GetVariable(
	ctx context.Context,
	key consts.VarKey,
) ([]byte, error) {
	reply, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.GetVariableReply, error) {
		return callWrapper(ctx, c, client.GetVariable, &streamd_grpc.GetVariableRequest{Key: string(key)})
	})

	if err != nil {
		return nil, fmt.Errorf("unable to get the variable '%s' value: %w", key, err)
	}

	b := reply.GetValue()
	logger.Tracef(ctx, "downloaded variable value of size %d", len(b))
	return b, nil
}

func (c *Client) GetVariableHash(
	ctx context.Context,
	key consts.VarKey,
	hashType crypto.Hash,
) ([]byte, error) {
	var hashTypeArg streamd_grpc.HashType
	switch hashType {
	case crypto.SHA1:
		hashTypeArg = streamd_grpc.HashType_HASH_SHA1
	default:
		return nil, fmt.Errorf("unsupported hash type: %s", hashType)
	}

	reply, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.GetVariableHashReply, error) {
		return callWrapper(ctx, c, client.GetVariableHash, &streamd_grpc.GetVariableHashRequest{
			Key:      string(key),
			HashType: hashTypeArg,
		})
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get the variable '%s' hash: %w", key, err)
	}

	b := reply.GetHash()
	logger.Tracef(ctx, "the downloaded hash of the variable '%s' is %X", key, b)
	return b, nil
}

func (c *Client) SetVariable(
	ctx context.Context,
	key consts.VarKey,
	value []byte,
) error {
	_, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.SetVariableReply, error) {
		return callWrapper(ctx, c, client.SetVariable, &streamd_grpc.SetVariableRequest{
			Key:   string(key),
			Value: value,
		})
	})
	return err
}

func (c *Client) OBSGetSceneList(
	ctx context.Context,
) (*scenes.GetSceneListResponse, error) {
	resp, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.OBSGetSceneListReply, error) {
		return callWrapper(ctx, c, client.OBSGetSceneList, &streamd_grpc.OBSGetSceneListRequest{})
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get the list of OBS scenes: %w", err)
	}

	result := &scenes.GetSceneListResponse{
		CurrentPreviewSceneName: resp.CurrentPreviewSceneName,
		CurrentPreviewSceneUuid: resp.CurrentPreviewSceneUUID,
		CurrentProgramSceneName: resp.CurrentProgramSceneName,
		CurrentProgramSceneUuid: resp.CurrentProgramSceneUUID,
	}
	for _, scene := range resp.Scenes {
		result.Scenes = append(result.Scenes, &typedefs.Scene{
			SceneUuid:  scene.Uuid,
			SceneIndex: int(scene.Index),
			SceneName:  scene.Name,
		})
	}

	return result, nil
}
func (c *Client) OBSSetCurrentProgramScene(
	ctx context.Context,
	in *scenes.SetCurrentProgramSceneParams,
) error {
	req := &streamd_grpc.OBSSetCurrentProgramSceneRequest{}
	switch {
	case in.SceneUuid != nil:
		req.OBSSceneID = &streamd_grpc.OBSSetCurrentProgramSceneRequest_SceneUUID{
			SceneUUID: *in.SceneUuid,
		}
	case in.SceneName != nil:
		req.OBSSceneID = &streamd_grpc.OBSSetCurrentProgramSceneRequest_SceneName{
			SceneName: *in.SceneName,
		}
	}
	_, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.OBSSetCurrentProgramSceneReply, error) {
		return callWrapper(ctx, c, client.OBSSetCurrentProgramScene, req)
	})
	return err
}

func ptr[T any](in T) *T {
	return &in
}

func (c *Client) SubmitOAuthCode(
	ctx context.Context,
	req *streamd_grpc.SubmitOAuthCodeRequest,
) (*streamd_grpc.SubmitOAuthCodeReply, error) {
	return withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.SubmitOAuthCodeReply, error) {
		return callWrapper(ctx, c, client.SubmitOAuthCode, req)
	})
}

func (c *Client) ListStreamServers(
	ctx context.Context,
) ([]api.StreamServer, error) {
	reply, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.ListStreamServersReply, error) {
		return callWrapper(ctx, c, client.ListStreamServers, &streamd_grpc.ListStreamServersRequest{})
	})
	if err != nil {
		return nil, fmt.Errorf("unable to request to list of the stream servers: %w", err)
	}
	var result []api.StreamServer
	for _, server := range reply.GetStreamServers() {
		t, err := goconv.StreamServerTypeGRPC2Go(server.Config.GetServerType())
		if err != nil {
			return nil, fmt.Errorf("unable to convert the server type value: %w", err)
		}
		result = append(result, api.StreamServer{
			Type:                  t,
			ListenAddr:            server.Config.GetListenAddr(),
			NumBytesConsumerWrote: uint64(server.GetStatistics().GetNumBytesConsumerWrote()),
			NumBytesProducerRead:  uint64(server.GetStatistics().GetNumBytesProducerRead()),
		})
	}
	return result, nil
}

func (c *Client) StartStreamServer(
	ctx context.Context,
	serverType api.StreamServerType,
	listenAddr string,
) error {
	t, err := goconv.StreamServerTypeGo2GRPC(serverType)
	if err != nil {
		return fmt.Errorf("unable to convert the server type: %w", err)
	}

	_, err = withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.StartStreamServerReply, error) {
		return callWrapper(ctx, c, client.StartStreamServer, &streamd_grpc.StartStreamServerRequest{
			Config: &streamd_grpc.StreamServer{
				ServerType: t,
				ListenAddr: listenAddr,
			},
		})
	})
	if err != nil {
		return fmt.Errorf("unable to request to start the stream server: %w", err)
	}
	return nil
}

func (c *Client) StopStreamServer(
	ctx context.Context,
	listenAddr string,
) error {
	_, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.StopStreamServerReply, error) {
		return callWrapper(ctx, c, client.StopStreamServer, &streamd_grpc.StopStreamServerRequest{
			ListenAddr: listenAddr,
		})
	})
	return err
}

func (c *Client) AddIncomingStream(
	ctx context.Context,
	streamID api.StreamID,
) error {
	_, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.AddIncomingStreamReply, error) {
		return callWrapper(ctx, c, client.AddIncomingStream, &streamd_grpc.AddIncomingStreamRequest{
			StreamID: string(streamID),
		})
	})
	return err
}

func (c *Client) RemoveIncomingStream(
	ctx context.Context,
	streamID api.StreamID,
) error {
	_, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.RemoveIncomingStreamReply, error) {
		return callWrapper(ctx, c, client.RemoveIncomingStream, &streamd_grpc.RemoveIncomingStreamRequest{
			StreamID: string(streamID),
		})
	})
	return err
}

func (c *Client) ListIncomingStreams(
	ctx context.Context,
) ([]api.IncomingStream, error) {
	reply, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.ListIncomingStreamsReply, error) {
		return callWrapper(ctx, c, client.ListIncomingStreams, &streamd_grpc.ListIncomingStreamsRequest{})
	})
	if err != nil {
		return nil, fmt.Errorf("unable to request to list the incoming streams: %w", err)
	}

	var result []api.IncomingStream
	for _, stream := range reply.GetIncomingStreams() {
		result = append(result, api.IncomingStream{
			StreamID: api.StreamID(stream.GetStreamID()),
		})
	}
	return result, nil
}

func (c *Client) ListStreamDestinations(
	ctx context.Context,
) ([]api.StreamDestination, error) {
	reply, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.ListStreamDestinationsReply, error) {
		return callWrapper(ctx, c, client.ListStreamDestinations, &streamd_grpc.ListStreamDestinationsRequest{})
	})
	if err != nil {
		return nil, fmt.Errorf("unable to request to list the stream destinations: %w", err)
	}

	var result []api.StreamDestination
	for _, dst := range reply.GetStreamDestinations() {
		result = append(result, api.StreamDestination{
			ID:  api.DestinationID(dst.GetDestinationID()),
			URL: dst.GetUrl(),
		})
	}
	return result, nil
}

func (c *Client) AddStreamDestination(
	ctx context.Context,
	destinationID api.DestinationID,
	url string,
) error {
	_, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.AddStreamDestinationReply, error) {
		return callWrapper(ctx, c, client.AddStreamDestination, &streamd_grpc.AddStreamDestinationRequest{
			Config: &streamd_grpc.StreamDestination{
				DestinationID: string(destinationID),
				Url:           url,
			},
		})
	})
	return err
}

func (c *Client) RemoveStreamDestination(
	ctx context.Context,
	destinationID api.DestinationID,
) error {
	_, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.RemoveStreamDestinationReply, error) {
		return callWrapper(ctx, c, client.RemoveStreamDestination, &streamd_grpc.RemoveStreamDestinationRequest{
			DestinationID: string(destinationID),
		})
	})
	return err
}

func (c *Client) ListStreamForwards(
	ctx context.Context,
) ([]api.StreamForward, error) {
	reply, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.ListStreamForwardsReply, error) {
		return callWrapper(ctx, c, client.ListStreamForwards, &streamd_grpc.ListStreamForwardsRequest{})
	})
	if err != nil {
		return nil, fmt.Errorf("unable to request to list the stream forwards: %w", err)
	}

	var result []api.StreamForward
	for _, forward := range reply.GetStreamForwards() {
		item := api.StreamForward{
			Enabled:       forward.Config.Enabled,
			StreamID:      api.StreamID(forward.Config.GetStreamID()),
			DestinationID: api.DestinationID(forward.Config.GetDestinationID()),
			NumBytesWrote: uint64(forward.Statistics.NumBytesWrote),
			NumBytesRead:  uint64(forward.Statistics.NumBytesRead),
		}
		restartUntilYoutubeRecognizesStream := forward.GetConfig().GetQuirks().GetRestartUntilYoutubeRecognizesStream()
		if restartUntilYoutubeRecognizesStream != nil {
			item.Quirks = api.StreamForwardingQuirks{
				RestartUntilYoutubeRecognizesStream: types.RestartUntilYoutubeRecognizesStream{
					Enabled:        restartUntilYoutubeRecognizesStream.Enabled,
					StartTimeout:   time.Duration(float64(time.Second) * restartUntilYoutubeRecognizesStream.StartTimeout),
					StopStartDelay: time.Duration(float64(time.Second) * restartUntilYoutubeRecognizesStream.StopStartDelay),
				},
			}
		}
		result = append(result, item)
	}
	return result, nil
}

func (c *Client) AddStreamForward(
	ctx context.Context,
	streamID api.StreamID,
	destinationID api.DestinationID,
	enabled bool,
	quirks api.StreamForwardingQuirks,
) error {
	_, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.AddStreamForwardReply, error) {
		return callWrapper(ctx, c, client.AddStreamForward, &streamd_grpc.AddStreamForwardRequest{
			Config: &streamd_grpc.StreamForward{
				StreamID:      string(streamID),
				DestinationID: string(destinationID),
				Enabled:       enabled,
				Quirks: &streamd_grpc.StreamForwardQuirks{
					RestartUntilYoutubeRecognizesStream: &streamd_grpc.RestartUntilYoutubeRecognizesStream{
						Enabled:        quirks.RestartUntilYoutubeRecognizesStream.Enabled,
						StartTimeout:   quirks.RestartUntilYoutubeRecognizesStream.StartTimeout.Seconds(),
						StopStartDelay: quirks.RestartUntilYoutubeRecognizesStream.StopStartDelay.Seconds(),
					},
				},
			},
		})
	})
	return err
}

func (c *Client) UpdateStreamForward(
	ctx context.Context,
	streamID api.StreamID,
	destinationID api.DestinationID,
	enabled bool,
	quirks api.StreamForwardingQuirks,
) error {
	_, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.UpdateStreamForwardReply, error) {
		return callWrapper(ctx, c, client.UpdateStreamForward, &streamd_grpc.UpdateStreamForwardRequest{
			Config: &streamd_grpc.StreamForward{
				StreamID:      string(streamID),
				DestinationID: string(destinationID),
				Enabled:       enabled,
				Quirks: &streamd_grpc.StreamForwardQuirks{
					RestartUntilYoutubeRecognizesStream: &streamd_grpc.RestartUntilYoutubeRecognizesStream{
						Enabled:        quirks.RestartUntilYoutubeRecognizesStream.Enabled,
						StartTimeout:   quirks.RestartUntilYoutubeRecognizesStream.StartTimeout.Seconds(),
						StopStartDelay: quirks.RestartUntilYoutubeRecognizesStream.StopStartDelay.Seconds(),
					},
				},
			},
		})
	})
	return err
}

func (c *Client) RemoveStreamForward(
	ctx context.Context,
	streamID api.StreamID,
	destinationID api.DestinationID,
) error {
	_, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.RemoveStreamForwardReply, error) {
		return callWrapper(ctx, c, client.RemoveStreamForward, &streamd_grpc.RemoveStreamForwardRequest{
			Config: &streamd_grpc.StreamForward{
				StreamID:      string(streamID),
				DestinationID: string(destinationID),
			},
		})
	})
	return err
}

func (c *Client) WaitForStreamPublisher(
	ctx context.Context,
	streamID api.StreamID,
) (<-chan struct{}, error) {
	return unwrapChan(
		ctx,
		c,
		func(
			ctx context.Context,
			client streamd_grpc.StreamDClient,
		) (streamd_grpc.StreamD_WaitForStreamPublisherClient, error) {
			return callWrapper(
				ctx,
				c,
				client.WaitForStreamPublisher,
				&streamd_grpc.WaitForStreamPublisherRequest{
					StreamID: ptr(string(streamID)),
				},
			)
		},
		func(
			ctx context.Context,
			event *streamd_grpc.StreamPublisher,
		) struct{} {
			return struct{}{}
		},
	)
}

func (c *Client) AddStreamPlayer(
	ctx context.Context,
	streamID streamtypes.StreamID,
	playerType player.Backend,
	disabled bool,
	streamPlaybackConfig sptypes.Config,
) error {
	_, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.AddStreamPlayerReply, error) {
		return callWrapper(ctx, c, client.AddStreamPlayer, &streamd_grpc.AddStreamPlayerRequest{
			Config: &streamd_grpc.StreamPlayerConfig{
				StreamID:             string(streamID),
				PlayerType:           goconv.StreamPlayerTypeGo2GRPC(playerType),
				Disabled:             disabled,
				StreamPlaybackConfig: goconv.StreamPlaybackConfigGo2GRPC(&streamPlaybackConfig),
			},
		})
	})
	return err
}

func (c *Client) UpdateStreamPlayer(
	ctx context.Context,
	streamID streamtypes.StreamID,
	playerType player.Backend,
	disabled bool,
	streamPlaybackConfig sptypes.Config,
) error {
	_, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.UpdateStreamPlayerReply, error) {
		return callWrapper(ctx, c, client.UpdateStreamPlayer, &streamd_grpc.UpdateStreamPlayerRequest{
			Config: &streamd_grpc.StreamPlayerConfig{
				StreamID:             string(streamID),
				PlayerType:           goconv.StreamPlayerTypeGo2GRPC(playerType),
				Disabled:             disabled,
				StreamPlaybackConfig: goconv.StreamPlaybackConfigGo2GRPC(&streamPlaybackConfig),
			},
		})
	})
	return err
}

func (c *Client) RemoveStreamPlayer(
	ctx context.Context,
	streamID streamtypes.StreamID,
) error {
	_, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.RemoveStreamPlayerReply, error) {
		return callWrapper(ctx, c, client.RemoveStreamPlayer, &streamd_grpc.RemoveStreamPlayerRequest{
			StreamID: string(streamID),
		})
	})
	return err
}

func (c *Client) ListStreamPlayers(
	ctx context.Context,
) ([]api.StreamPlayer, error) {
	resp, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.ListStreamPlayersReply, error) {
		return callWrapper(ctx, c, client.ListStreamPlayers, &streamd_grpc.ListStreamPlayersRequest{})
	})
	if err != nil {
		return nil, fmt.Errorf("unable to query: %w", err)
	}

	result := make([]api.StreamPlayer, 0, len(resp.GetPlayers()))
	for _, player := range resp.GetPlayers() {
		result = append(result, api.StreamPlayer{
			StreamID:             streamtypes.StreamID(player.GetStreamID()),
			PlayerType:           goconv.StreamPlayerTypeGRPC2Go(player.PlayerType),
			Disabled:             player.GetDisabled(),
			StreamPlaybackConfig: goconv.StreamPlaybackConfigGRPC2Go(player.GetStreamPlaybackConfig()),
		})
	}
	return result, nil
}

func (c *Client) GetStreamPlayer(
	ctx context.Context,
	streamID streamtypes.StreamID,
) (*api.StreamPlayer, error) {
	resp, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.GetStreamPlayerReply, error) {
		return callWrapper(ctx, c, client.GetStreamPlayer, &streamd_grpc.GetStreamPlayerRequest{
			StreamID: string(streamID),
		})
	})
	if err != nil {
		return nil, fmt.Errorf("unable to query: %w", err)
	}

	cfg := resp.GetConfig()
	return &api.StreamPlayer{
		StreamID:             streamID,
		PlayerType:           goconv.StreamPlayerTypeGRPC2Go(cfg.PlayerType),
		Disabled:             cfg.GetDisabled(),
		StreamPlaybackConfig: goconv.StreamPlaybackConfigGRPC2Go(cfg.StreamPlaybackConfig),
	}, nil
}

func (c *Client) StreamPlayerProcessTitle(
	ctx context.Context,
	streamID streamtypes.StreamID,
) (string, error) {
	resp, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.StreamPlayerProcessTitleReply, error) {
		return callWrapper(ctx, c, client.StreamPlayerProcessTitle, &streamd_grpc.StreamPlayerProcessTitleRequest{
			StreamID: string(streamID),
		})
	})
	if err != nil {
		return "", fmt.Errorf("unable to query: %w", err)
	}
	return resp.Reply.GetTitle(), nil
}

func (c *Client) StreamPlayerOpenURL(
	ctx context.Context,
	streamID streamtypes.StreamID,
	link string,
) error {
	_, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.StreamPlayerOpenReply, error) {
		return callWrapper(ctx, c, client.StreamPlayerOpen, &streamd_grpc.StreamPlayerOpenRequest{
			StreamID: string(streamID),
			Request:  &player_grpc.OpenRequest{},
		})
	})
	return err
}

func (c *Client) StreamPlayerGetLink(
	ctx context.Context,
	streamID streamtypes.StreamID,
) (string, error) {
	resp, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.StreamPlayerGetLinkReply, error) {
		return callWrapper(ctx, c, client.StreamPlayerGetLink, &streamd_grpc.StreamPlayerGetLinkRequest{
			StreamID: string(streamID),
			Request:  &player_grpc.GetLinkRequest{},
		})
	})
	if err != nil {
		return "", fmt.Errorf("unable to query: %w", err)
	}
	return resp.GetReply().Link, nil
}

func (c *Client) StreamPlayerEndChan(
	ctx context.Context,
	streamID streamtypes.StreamID,
) (<-chan struct{}, error) {
	return unwrapChan(
		ctx,
		c,
		func(
			ctx context.Context,
			client streamd_grpc.StreamDClient,
		) (streamd_grpc.StreamD_StreamPlayerEndChanClient, error) {
			return callWrapper(
				ctx,
				c,
				client.StreamPlayerEndChan,
				&streamd_grpc.StreamPlayerEndChanRequest{
					StreamID: string(streamID),
					Request:  &player_grpc.EndChanRequest{},
				},
			)
		},
		func(
			ctx context.Context,
			event *streamd_grpc.StreamPlayerEndChanReply,
		) struct{} {
			return struct{}{}
		},
	)
}

func (c *Client) StreamPlayerIsEnded(
	ctx context.Context,
	streamID streamtypes.StreamID,
) (bool, error) {
	resp, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.StreamPlayerIsEndedReply, error) {
		return callWrapper(ctx, c, client.StreamPlayerIsEnded, &streamd_grpc.StreamPlayerIsEndedRequest{
			StreamID: string(streamID),
			Request:  &player_grpc.IsEndedRequest{},
		})
	})
	if err != nil {
		return false, fmt.Errorf("unable to query: %w", err)
	}
	return resp.GetReply().IsEnded, nil
}

func (c *Client) StreamPlayerGetPosition(
	ctx context.Context,
	streamID streamtypes.StreamID,
) (time.Duration, error) {
	resp, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.StreamPlayerGetPositionReply, error) {
		return callWrapper(ctx, c, client.StreamPlayerGetPosition, &streamd_grpc.StreamPlayerGetPositionRequest{
			StreamID: string(streamID),
			Request:  &player_grpc.GetPositionRequest{},
		})
	})
	if err != nil {
		return 0, fmt.Errorf("unable to query: %w", err)
	}
	return time.Duration(float64(time.Second) * resp.GetReply().GetPositionSecs()), nil
}

func (c *Client) StreamPlayerGetLength(
	ctx context.Context,
	streamID streamtypes.StreamID,
) (time.Duration, error) {
	resp, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.StreamPlayerGetLengthReply, error) {
		return callWrapper(ctx, c, client.StreamPlayerGetLength, &streamd_grpc.StreamPlayerGetLengthRequest{
			StreamID: string(streamID),
			Request:  &player_grpc.GetLengthRequest{},
		})
	})
	if err != nil {
		return 0, fmt.Errorf("unable to query: %w", err)
	}
	return time.Duration(float64(time.Second) * resp.GetReply().GetLengthSecs()), nil
}

func (c *Client) StreamPlayerSetSpeed(
	ctx context.Context,
	streamID streamtypes.StreamID,
	speed float64,
) error {
	_, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.StreamPlayerSetSpeedReply, error) {
		return callWrapper(ctx, c, client.StreamPlayerSetSpeed, &streamd_grpc.StreamPlayerSetSpeedRequest{
			StreamID: string(streamID),
			Request: &player_grpc.SetSpeedRequest{
				Speed: speed,
			},
		})
	})
	return err
}

func (c *Client) StreamPlayerSetPause(
	ctx context.Context,
	streamID streamtypes.StreamID,
	pause bool,
) error {
	_, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.StreamPlayerSetPauseReply, error) {
		return callWrapper(ctx, c, client.StreamPlayerSetPause, &streamd_grpc.StreamPlayerSetPauseRequest{
			StreamID: string(streamID),
			Request: &player_grpc.SetPauseRequest{
				SetPaused: pause,
			},
		})
	})
	return err
}

func (c *Client) StreamPlayerStop(
	ctx context.Context,
	streamID streamtypes.StreamID,
) error {
	_, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.StreamPlayerStopReply, error) {
		return callWrapper(ctx, c, client.StreamPlayerStop, &streamd_grpc.StreamPlayerStopRequest{
			StreamID: string(streamID),
			Request:  &player_grpc.StopRequest{},
		})
	})
	return err
}

func (c *Client) StreamPlayerClose(
	ctx context.Context,
	streamID streamtypes.StreamID,
) error {
	_, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.StreamPlayerCloseReply, error) {
		return callWrapper(ctx, c, client.StreamPlayerClose, &streamd_grpc.StreamPlayerCloseRequest{
			StreamID: string(streamID),
			Request:  &player_grpc.CloseRequest{},
		})
	})
	return err
}

func (c *Client) SubscribeToConfigChanges(
	ctx context.Context,
) (<-chan api.DiffConfig, error) {
	return unwrapChan(
		ctx,
		c,
		func(
			ctx context.Context,
			client streamd_grpc.StreamDClient,
		) (streamd_grpc.StreamD_SubscribeToConfigChangesClient, error) {
			return callWrapper(
				ctx,
				c,
				client.SubscribeToConfigChanges,
				&streamd_grpc.SubscribeToConfigChangesRequest{},
			)
		},
		func(
			ctx context.Context,
			event *streamd_grpc.ConfigChange,
		) api.DiffConfig {
			return api.DiffConfig{}
		},
	)
}

func (c *Client) SubscribeToStreamsChanges(
	ctx context.Context,
) (<-chan api.DiffStreams, error) {
	return unwrapChan(
		ctx,
		c,
		func(
			ctx context.Context,
			client streamd_grpc.StreamDClient,
		) (streamd_grpc.StreamD_SubscribeToStreamsChangesClient, error) {
			return callWrapper(
				ctx,
				c,
				client.SubscribeToStreamsChanges,
				&streamd_grpc.SubscribeToStreamsChangesRequest{},
			)
		},
		func(
			ctx context.Context,
			event *streamd_grpc.StreamsChange,
		) api.DiffStreams {
			return api.DiffStreams{}
		},
	)
}

func (c *Client) SubscribeToStreamServersChanges(
	ctx context.Context,
) (<-chan api.DiffStreamServers, error) {
	return unwrapChan(
		ctx,
		c,
		func(
			ctx context.Context,
			client streamd_grpc.StreamDClient,
		) (streamd_grpc.StreamD_SubscribeToStreamServersChangesClient, error) {
			return callWrapper(
				ctx,
				c,
				client.SubscribeToStreamServersChanges,
				&streamd_grpc.SubscribeToStreamServersChangesRequest{},
			)
		},
		func(
			ctx context.Context,
			event *streamd_grpc.StreamServersChange,
		) api.DiffStreamServers {
			return api.DiffStreamServers{}
		},
	)
}

func (c *Client) SubscribeToStreamDestinationsChanges(
	ctx context.Context,
) (<-chan api.DiffStreamDestinations, error) {
	return unwrapChan(
		ctx,
		c,
		func(
			ctx context.Context,
			client streamd_grpc.StreamDClient,
		) (streamd_grpc.StreamD_SubscribeToStreamDestinationsChangesClient, error) {
			return callWrapper(
				ctx,
				c,
				client.SubscribeToStreamDestinationsChanges,
				&streamd_grpc.SubscribeToStreamDestinationsChangesRequest{},
			)
		},
		func(
			ctx context.Context,
			event *streamd_grpc.StreamDestinationsChange,
		) api.DiffStreamDestinations {
			return api.DiffStreamDestinations{}
		},
	)
}

func (c *Client) SubscribeToIncomingStreamsChanges(
	ctx context.Context,
) (<-chan api.DiffIncomingStreams, error) {
	return unwrapChan(
		ctx,
		c,
		func(
			ctx context.Context,
			client streamd_grpc.StreamDClient,
		) (streamd_grpc.StreamD_SubscribeToIncomingStreamsChangesClient, error) {
			return callWrapper(
				ctx,
				c,
				client.SubscribeToIncomingStreamsChanges,
				&streamd_grpc.SubscribeToIncomingStreamsChangesRequest{},
			)
		},
		func(
			ctx context.Context,
			event *streamd_grpc.IncomingStreamsChange,
		) api.DiffIncomingStreams {
			return api.DiffIncomingStreams{}
		},
	)
}

func (c *Client) SubscribeToStreamForwardsChanges(
	ctx context.Context,
) (<-chan api.DiffStreamForwards, error) {
	return unwrapChan(
		ctx,
		c,
		func(
			ctx context.Context,
			client streamd_grpc.StreamDClient,
		) (streamd_grpc.StreamD_SubscribeToStreamForwardsChangesClient, error) {
			return callWrapper(
				ctx,
				c,
				client.SubscribeToStreamForwardsChanges,
				&streamd_grpc.SubscribeToStreamForwardsChangesRequest{},
			)
		},
		func(
			ctx context.Context,
			event *streamd_grpc.StreamForwardsChange,
		) api.DiffStreamForwards {
			return api.DiffStreamForwards{}
		},
	)
}

func (c *Client) SubscribeToStreamPlayersChanges(
	ctx context.Context,
) (<-chan api.DiffStreamPlayers, error) {
	return unwrapChan(
		ctx,
		c,
		func(
			ctx context.Context,
			client streamd_grpc.StreamDClient,
		) (streamd_grpc.StreamD_SubscribeToStreamPlayersChangesClient, error) {
			return callWrapper(
				ctx,
				c,
				client.SubscribeToStreamPlayersChanges,
				&streamd_grpc.SubscribeToStreamPlayersChangesRequest{},
			)
		},
		func(
			ctx context.Context,
			event *streamd_grpc.StreamPlayersChange,
		) api.DiffStreamPlayers {
			return api.DiffStreamPlayers{}
		},
	)
}

func (c *Client) SetLoggingLevel(ctx context.Context, level logger.Level) error {
	_, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.SetLoggingLevelReply, error) {
		return callWrapper(ctx, c, client.SetLoggingLevel, &streamd_grpc.SetLoggingLevelRequest{
			LoggingLevel: goconv.LoggingLevelGo2GRPC(level),
		})
	})
	return err
}

func (c *Client) GetLoggingLevel(ctx context.Context) (logger.Level, error) {
	reply, err := withClient(ctx, c, func(
		ctx context.Context,
		client streamd_grpc.StreamDClient,
		conn io.Closer,
	) (*streamd_grpc.GetLoggingLevelReply, error) {
		return callWrapper(ctx, c, client.GetLoggingLevel, &streamd_grpc.GetLoggingLevelRequest{})
	})
	if err != nil {
		return logger.LevelUndefined, fmt.Errorf("unable to get the logging level: %w", err)
	}

	return goconv.LoggingLevelGRPC2Go(reply.GetLoggingLevel()), nil
}
