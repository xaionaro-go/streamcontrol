package streamserver

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync/atomic"

	"github.com/facebookincubator/go-belt"
	"github.com/facebookincubator/go-belt/tool/logger"
	"github.com/hashicorp/go-multierror"
	"github.com/xaionaro-go/streamctl/pkg/streamserver/types"
	"github.com/xaionaro-go/streamctl/pkg/xlogger"
	flvtag "github.com/yutopp/go-flv/tag"
	"github.com/yutopp/go-rtmp"
	rtmpmsg "github.com/yutopp/go-rtmp/message"
)

const (
	chunkSize = 128
)

type ActiveStreamForwarding struct {
	StreamID      types.StreamID
	DestinationID types.DestinationID
	Client        *rtmp.ClientConn
	OutStream     *rtmp.Stream
	Sub           *Sub
	CancelFunc    context.CancelFunc
	ReadCount     atomic.Uint64
	WriteCount    atomic.Uint64
}

func newActiveStreamForward(
	ctx context.Context,
	streamID types.StreamID,
	dstID types.DestinationID,
	urlString string,
	relayService *RelayService,
) (*ActiveStreamForwarding, error) {
	ctx, cancelFn := context.WithCancel(ctx)
	fwd := &ActiveStreamForwarding{
		StreamID:      streamID,
		DestinationID: dstID,
		CancelFunc:    cancelFn,
	}

	urlParsed, err := url.Parse(urlString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse URL '%s': %w", urlString, err)
	}

	go func() {
		for {
			err := fwd.waitForPublisherAndStart(
				ctx,
				relayService,
				urlParsed,
			)
			select {
			case <-ctx.Done():
				return
			default:
			}
			if err != nil {
				logger.Errorf(ctx, "%s", err)
			}
		}
	}()

	return fwd, nil
}

func (fwd *ActiveStreamForwarding) waitForPublisherAndStart(
	ctx context.Context,
	relayService *RelayService,
	urlParsed *url.URL,
) (_ret error) {
	defer func() {
		if _ret == nil {
			return
		}
		logger.Errorf(ctx, "%v", _ret)
		fwd.Close()
	}()

	pathParts := strings.SplitN(urlParsed.Path, "/", -2)
	remoteAppName := "live"
	apiKey := pathParts[len(pathParts)-1]
	if len(pathParts) >= 2 {
		remoteAppName = strings.Trim(strings.Join(pathParts[:len(pathParts)-1], "/"), "/")
	}
	streamID := fwd.StreamID
	streamIDParts := strings.Split(string(streamID), "/")
	localAppName := string(streamID)
	if len(streamIDParts) == 2 {
		localAppName = streamIDParts[1]
	}

	ctx = belt.WithField(belt.WithField(ctx, "appNameLocal", localAppName), "appNameRemote", apiKey)

	logger.Tracef(ctx, "wait for stream '%s'", streamID)
	pubSub := relayService.WaitPubsub(ctx, localAppName)
	logger.Tracef(ctx, "wait for stream '%s' result: %#+v", streamID, pubSub)
	if pubSub == nil {
		return fmt.Errorf(
			"unable to find stream ID '%s', available stream IDs: %s",
			streamID,
			strings.Join(relayService.PubsubNames(), ", "),
		)
	}

	logger.Tracef(ctx, "connecting to '%s'", urlParsed.String())
	if urlParsed.Port() == "" {
		urlParsed.Host += ":1935"
	}
	client, err := rtmp.Dial("rtmp", urlParsed.Host, &rtmp.ConnConfig{
		Logger: xlogger.LogrusFieldLoggerFromCtx(ctx),
	})
	if err != nil {
		return fmt.Errorf("unable to connect to '%s': %w", urlParsed.String(), err)
	}
	fwd.Client = client

	logger.Tracef(ctx, "connected to '%s'", urlParsed.String())

	tcURL := *urlParsed
	tcURL.Path = "/" + remoteAppName
	if tcURL.Port() == "1935" {
		tcURL.Host = tcURL.Hostname()
	}

	if err := client.Connect(&rtmpmsg.NetConnectionConnect{
		Command: rtmpmsg.NetConnectionConnectCommand{
			App:      remoteAppName,
			Type:     "nonprivate",
			FlashVer: "StreamPanel",
			TCURL:    tcURL.String(),
		},
	}); err != nil {
		return fmt.Errorf("unable to connect the stream to '%s': %w", urlParsed.String(), err)
	}
	logger.Tracef(ctx, "connected the stream to '%s'", urlParsed.String())

	fwd.OutStream, err = client.CreateStream(&rtmpmsg.NetConnectionCreateStream{}, chunkSize)
	if err != nil {
		return fmt.Errorf("unable to create a stream to '%s': %w", urlParsed.String(), err)
	}

	logger.Tracef(ctx, "calling Publish at '%s'", urlParsed.String())
	if err := fwd.OutStream.Publish(&rtmpmsg.NetStreamPublish{
		PublishingName: apiKey,
		PublishingType: "live",
	}); err != nil {
		return fmt.Errorf("unable to send the Publish message to '%s': %w", urlParsed.String(), err)
	}

	logger.Tracef(ctx, "starting publishing to '%s'", urlParsed.String())

	eventCallback := func(flv *flvtag.FlvTag) error {
		var buf bytes.Buffer

		switch d := flv.Data.(type) {
		case *flvtag.AudioData:
			// Consume flv payloads (d)
			if err := flvtag.EncodeAudioData(&buf, d); err != nil {
				return err
			}

			fwd.WriteCount.Add(uint64(buf.Len()))

			// TODO: Fix these values
			chunkStreamID := 5
			return fwd.OutStream.Write(chunkStreamID, flv.Timestamp, &rtmpmsg.AudioMessage{
				Payload: &buf,
			})

		case *flvtag.VideoData:
			// Consume flv payloads (d)
			if err := flvtag.EncodeVideoData(&buf, d); err != nil {
				return err
			}

			fwd.WriteCount.Add(uint64(buf.Len()))

			// TODO: Fix these values
			chunkStreamID := 6
			return fwd.OutStream.Write(chunkStreamID, flv.Timestamp, &rtmpmsg.VideoMessage{
				Payload: &buf,
			})

		case *flvtag.ScriptData:
			// Consume flv payloads (d)
			if err := flvtag.EncodeScriptData(&buf, d); err != nil {
				return err
			}

			fwd.WriteCount.Add(uint64(buf.Len()))

			// TODO: hide these implementation
			amdBuf := new(bytes.Buffer)
			amfEnc := rtmpmsg.NewAMFEncoder(amdBuf, rtmpmsg.EncodingTypeAMF0)
			if err := rtmpmsg.EncodeBodyAnyValues(amfEnc, &rtmpmsg.NetStreamSetDataFrame{
				Payload: buf.Bytes(),
			}); err != nil {
				return err
			}

			// TODO: Fix these values
			chunkStreamID := 8
			return fwd.OutStream.Write(chunkStreamID, flv.Timestamp, &rtmpmsg.DataMessage{
				Name:     "@setDataFrame", // TODO: fix
				Encoding: rtmpmsg.EncodingTypeAMF0,
				Body:     amdBuf,
			})

		default:
			panic("unreachable")
		}
	}

	fwd.Sub = pubSub.Sub()
	fwd.Sub.pubSub.m.Lock()
	fwd.Sub.eventCallback = eventCallback
	fwd.Sub.pubSub.m.Unlock()

	<-ctx.Done()
	return nil
}

func (fwd *ActiveStreamForwarding) Close() error {
	var result *multierror.Error
	fwd.CancelFunc()
	if fwd.Sub != nil {
		result = multierror.Append(result, fwd.Sub.Close())
	}
	if fwd.Client != nil {
		result = multierror.Append(result, fwd.Client.Close())
	}
	return result.ErrorOrNil()
}
