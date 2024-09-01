package streamplayer

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/facebookincubator/go-belt/tool/experimental/errmon"
	"github.com/facebookincubator/go-belt/tool/logger"
	"github.com/hashicorp/go-multierror"
	"github.com/xaionaro-go/streamctl/pkg/observability"
	"github.com/xaionaro-go/streamctl/pkg/player"
	"github.com/xaionaro-go/streamctl/pkg/player/types"
	"github.com/xaionaro-go/streamctl/pkg/streamd/api"
	"github.com/xaionaro-go/streamctl/pkg/xsync"
)

type StreamPortServer struct {
	Addr string
	Type api.StreamServerType
}

type StreamServer interface {
	GetPortServers(context.Context) ([]StreamPortServer, error)
	WaitPublisher(context.Context, api.StreamID) (<-chan struct{}, error)
}

type StreamPlayers struct {
	StreamPlayersLocker xsync.RWMutex
	StreamPlayers       map[api.StreamID]*StreamPlayer

	StreamServer   StreamServer
	PlayerManager  *player.Manager
	DefaultOptions Options
}

type PlayerManager interface {
	SupportedBackends() []player.Backend
	NewPlayer(
		title string,
		backend player.Backend,
	) (player.Player, error)
}

func New(
	streamServer StreamServer,
	playerManager *player.Manager,
	defaultOptions ...Option,
) *StreamPlayers {
	return &StreamPlayers{
		StreamPlayers:  map[api.StreamID]*StreamPlayer{},
		StreamServer:   streamServer,
		PlayerManager:  playerManager,
		DefaultOptions: defaultOptions,
	}
}

type StreamPlayer struct {
	PlayerLocker xsync.Mutex
	Parent       *StreamPlayers
	Player       player.Player
	StreamID     api.StreamID

	Cancel context.CancelFunc
	Config Config
}

func (sp *StreamPlayers) Create(
	ctx context.Context,
	streamID api.StreamID,
	opts ...Option,
) (_ret *StreamPlayer, _err error) {
	logger.Debugf(ctx, "StreamPlayers.Create(ctx, '%s', %#+v)", streamID, opts)
	defer func() {
		logger.Debugf(ctx, "/StreamPlayers.Create(ctx, '%s', %#+v): (%v, %v)", streamID, opts, _ret, _err)
	}()
	ctx, cancel := context.WithCancel(ctx)

	resultingOpts := make(Options, 0, len(sp.DefaultOptions)+len(opts))
	resultingOpts = append(resultingOpts, sp.DefaultOptions...)
	resultingOpts = append(resultingOpts, opts...)

	p := &StreamPlayer{
		Parent:   sp,
		Cancel:   cancel,
		Config:   resultingOpts.Config(),
		StreamID: streamID,
	}

	if p.Config.CatchupMaxSpeedFactor <= 1 {
		return nil, fmt.Errorf("MaxCatchupSpeedFactor should be higher than 1, but it is %v", p.Config.CatchupMaxSpeedFactor)
	}

	if p.Config.MaxCatchupAtLag <= p.Config.JitterBufDuration {
		return nil, fmt.Errorf("MaxCatchupAtLag (%v) should be higher than JitterBufDuration (%v)", p.Config.MaxCatchupAtLag, p.Config.JitterBufDuration)
	}

	if err := p.startU(ctx); err != nil {
		return nil, fmt.Errorf("unable to start the player: %w", err)
	}

	return xsync.DoR2(ctx, &sp.StreamPlayersLocker, func() (*StreamPlayer, error) {
		sp.StreamPlayers[streamID] = p
		return p, nil
	})
}

func (sp *StreamPlayers) Remove(
	ctx context.Context,
	streamID api.StreamID,
) error {
	logger.Debugf(ctx, "StreamPlayers.Remove(ctx, '%s')", streamID)
	defer logger.Debugf(ctx, "/StreamPlayers.Remove(ctx, '%s')", streamID)
	return xsync.DoR1(ctx, &sp.StreamPlayersLocker, func() error {
		p, ok := sp.StreamPlayers[streamID]
		if !ok {
			return nil
		}
		errmon.ObserveErrorCtx(ctx, p.Close())
		delete(sp.StreamPlayers, streamID)
		return nil
	})
}

func (sp *StreamPlayers) Get(streamID api.StreamID) *StreamPlayer {
	ctx := context.TODO()
	return xsync.DoR1(ctx, &sp.StreamPlayersLocker, func() *StreamPlayer {
		return sp.StreamPlayers[streamID]
	})
}

func (sp *StreamPlayers) GetAll() map[api.StreamID]*StreamPlayer {
	ctx := context.TODO()
	return xsync.DoR1(ctx, &sp.StreamPlayersLocker, func() map[api.StreamID]*StreamPlayer {
		r := map[api.StreamID]*StreamPlayer{}
		for k, v := range sp.StreamPlayers {
			r[k] = v
		}
		return r
	})
}

const (
	processTitlePrefix = "streampanel-player-"
)

func StreamID2Title(streamID api.StreamID) string {
	return fmt.Sprintf("%s%s", processTitlePrefix, streamID)
}

func Title2StreamID(title string) api.StreamID {
	if !strings.HasPrefix(title, processTitlePrefix) {
		return ""
	}
	return api.StreamID(title[len(processTitlePrefix):])
}

func (p *StreamPlayer) startU(ctx context.Context) error {
	logger.Debugf(ctx, "StreamPlayers.startU(ctx): '%s'", p.StreamID)
	defer logger.Debugf(ctx, "/StreamPlayers.startU(ctx): '%s'", p.StreamID)

	instanceCtx, cancelFn := context.WithCancel(ctx)

	playerType := p.Parent.PlayerManager.SupportedBackends()[0]
	player, err := p.Parent.PlayerManager.NewPlayer(
		instanceCtx,
		StreamID2Title(p.StreamID),
		playerType,
	)
	if err != nil {
		errmon.ObserveErrorCtx(ctx, p.Close())
		cancelFn()
		return fmt.Errorf("unable to run a video player '%s': %w", playerType, err)
	}
	p.Player = player
	logger.Debugf(ctx, "initialized player #%+v", player)

	observability.Go(ctx, func() { p.controllerLoop(ctx, cancelFn) })
	return nil
}

func (p *StreamPlayer) stopU(ctx context.Context) error {
	logger.Debugf(ctx, "StreamPlayers.stopU(ctx): '%s'", p.StreamID)
	defer logger.Debugf(ctx, "/StreamPlayers.stopU(ctx): '%s'", p.StreamID)

	err := p.Player.Close(ctx)
	if err != nil {
		errmon.ObserveErrorCtx(ctx, p.Close())
		return fmt.Errorf("unable to close the player: %w", err)
	}
	return nil
}

func (p *StreamPlayer) restartU(ctx context.Context) error {
	logger.Debugf(ctx, "StreamPlayers.restartU(ctx): '%s'", p.StreamID)
	defer logger.Debugf(ctx, "/StreamPlayers.restartU(ctx): '%s'", p.StreamID)

	if err := p.stopU(ctx); err != nil {
		return fmt.Errorf("unable to stop the stream player: %w", err)
	}
	if err := p.startU(ctx); err != nil {
		return fmt.Errorf("unable to start the stream player: %w", err)
	}
	return nil
}

func (p *StreamPlayer) getURL(ctx context.Context) (*url.URL, error) {
	if p.Config.OverrideURL != "" {
		logger.Debugf(ctx, "override URL is '%s'", p.Config.OverrideURL)
		return p.getOverriddenURL(ctx)
	} else {
		logger.Debugf(ctx, "no override URL")
		return p.getInternalURL(ctx)
	}
}

func (p *StreamPlayer) getOverriddenURL(context.Context) (*url.URL, error) {
	return url.Parse(p.Config.OverrideURL)
}

func (p *StreamPlayer) getInternalURL(ctx context.Context) (*url.URL, error) {
	portSrvs, err := p.Parent.StreamServer.GetPortServers(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get the list of stream server ports: %w", err)
	}
	if len(portSrvs) == 0 {
		return nil, fmt.Errorf("there are no open server ports")
	}
	portSrv := portSrvs[0]

	var u url.URL
	u.Scheme = portSrv.Type.String()
	u.Host = portSrv.Addr
	u.Path = string(p.StreamID)
	return &u, nil
}

func (p *StreamPlayer) openStream(ctx context.Context) (_err error) {
	logger.Debugf(ctx, "openStream")
	defer func() { logger.Debugf(ctx, "/openStream: %v", _err) }()

	u, err := p.getURL(ctx)
	if err != nil {
		return fmt.Errorf("unable to get URL: %w", err)
	}
	logger.Debugf(ctx, "opening '%s'", u.String())
	err = p.withPlayer(ctx, func(ctx context.Context, player types.Player) {
		ctx, cancelFn := context.WithTimeout(ctx, 1*time.Second)
		defer cancelFn()
		var once sync.Once
		observability.Go(ctx, func() {
			<-ctx.Done()
			once.Do(func() {
				err := player.Close(ctx)
				logger.Debugf(ctx, "closing player error: %v", err)
			})
		})
		err = player.OpenURL(ctx, u.String())
		once.Do(func() {})
		if err != nil {
			err = fmt.Errorf("unable to open the URL: %w", err)
		}
	})
	logger.Debugf(ctx, "opened '%s': %v", u.String(), err)
	if err != nil {
		return fmt.Errorf("unable to open '%s' in the player: %w", u.String(), err)
	}

	return nil
}

func (p *StreamPlayer) Resetup(opts ...Option) {
	for _, opt := range opts {
		opt.Apply(&p.Config)
	}
}

func (p *StreamPlayer) notifyStart(ctx context.Context) {
	logger.Debugf(ctx, "notifyStart")
	defer logger.Debugf(ctx, "/notifyStart")

	for _, f := range p.Config.NotifierStart {
		func(f FuncNotifyStart) {
			defer func() {
				r := recover()
				if r != nil {
					logger.Error(ctx, "got panic during notification about a start: %v", r)
				}
			}()

			f(ctx, p.StreamID)
		}(f)
	}
}

func (p *StreamPlayer) controllerLoop(
	ctx context.Context,
	cancelPlayerInstance context.CancelFunc,
) {
	defer cancelPlayerInstance() // this is not necessary, but exists, just in case to reduce risks of a bad cleanup

	logger.Debugf(ctx, "StreamPlayer[%s].controllerLoop", p.StreamID)
	defer logger.Debugf(ctx, "/StreamPlayer[%s].controllerLoop", p.StreamID)

	instanceCtx, cancelFn := context.WithCancel(ctx)

	isClosed := false
	restart := func() {
		isClosed = true
		observability.Go(ctx, func() {
			err := p.restartU(ctx)
			errmon.ObserveErrorCtx(ctx, err)
			if err != nil {
				err := p.Parent.Remove(ctx, p.StreamID)
				errmon.ObserveErrorCtx(ctx, err)
			}
		})
	}

	// wait for video to start:
	{
		var ch <-chan struct{}
		_ch := make(chan struct{})
		close(_ch)
		ch = _ch

		for func() bool {
			if p.Config.OverrideURL == "" {
				waitPublisherCtx, waitPublisherCancel := context.WithCancel(ctx)
				defer waitPublisherCancel()

				var err error
				ch, err = p.Parent.StreamServer.WaitPublisher(waitPublisherCtx, p.StreamID)
				logger.Debugf(ctx, "got a waiter from WaitPublisher for '%s'; %v", p.StreamID, err)
				errmon.ObserveErrorCtx(ctx, err)

				logger.Debugf(ctx, "waiting for stream '%s'", p.StreamID)
				select {
				case <-instanceCtx.Done():
					logger.Debugf(ctx, "the instance was cancelled")
					errmon.ObserveErrorCtx(ctx, p.Close())
					return false
				case <-ch:
					logger.Debugf(ctx, "a stream started, let's open it in the player")
				}
				logger.Debugf(ctx, "opening the stream")
				err = p.openStream(ctx)
				logger.Debugf(ctx, "opened the stream: %v", err)
				errmon.ObserveErrorCtx(ctx, err)
			} else {
				t := time.NewTicker(1 * time.Second)
				defer t.Stop()
				for {
					select {
					case <-instanceCtx.Done():
						return false
					case <-t.C:
					}
					logger.Debugf(ctx, "opening the external stream")
					err := p.openStream(ctx)
					logger.Debugf(ctx, "opened the external stream: %v", err)
					if err != nil {
						logger.Debugf(ctx, "unable to open the stream: %v", err)
						continue
					}
					deadline := time.Now().Add(30 * time.Second)
					for {
						select {
						case <-instanceCtx.Done():
							return false
						case <-t.C:
						}
						logger.Debugf(ctx, "checking if we get get the position")
						err = p.withPlayer(ctx, func(ctx context.Context, player types.Player) {
							var pos time.Duration
							pos, err = player.GetPosition(ctx)
							logger.Debugf(ctx, "result of getting the position: %v %v", pos, err)
							if err != nil {
								err = fmt.Errorf("unable to get the position: %w", err)
							}
						})
						if errors.As(err, &ErrNilPlayer{}) {
							logger.Debugf(ctx, "player is nil, finishing")
							cancelFn()
							return false
						}
						if err == nil {
							break
						}
						now := time.Now()
						logger.Debugf(ctx, "checking if deadline reached: %v %v", now, deadline)
						if now.After(deadline) {
							break
						}
					}
					if err == nil {
						break
					}
				}
				logger.Debugf(ctx, "we opened the external stream and the player started to play it")
			}

			triedReopeningStream := false
			startedWaitingForBuffering := time.Now()
			for time.Since(startedWaitingForBuffering) <= p.Config.StartTimeout {
				var (
					pos time.Duration
					err error
				)
				err = p.withPlayer(ctx, func(ctx context.Context, player types.Player) {
					if err := player.SetPause(ctx, false); err != nil {
						logger.Errorf(ctx, "unable to unpause: %v", err)
					}
					pos, err = player.GetPosition(ctx)
					if err != nil {
						err = fmt.Errorf("unable to get position: %w", err)
					}
				})
				if err != nil {
					logger.Tracef(ctx, "StreamPlayer[%s].controllerLoop: unable to get the current position: %v", p.StreamID, err)
					time.Sleep(100 * time.Millisecond)
					continue
				}
				logger.Tracef(ctx, "StreamPlayer[%s].controllerLoop: pos == %v", p.StreamID, pos)
				var l time.Duration
				err = p.withPlayer(ctx, func(ctx context.Context, player types.Player) {
					l, err = player.GetLength(ctx)
					if err != nil {
						err = fmt.Errorf("unable to get length: %w", err)
					}
				})
				logger.Tracef(ctx, "StreamPlayer[%s].controllerLoop: length == %v", p.StreamID, l)
				if l < 0 {
					logger.Debugf(ctx, "StreamPlayer[%s].controllerLoop: negative length, restarting", p.StreamID)
					restart()
					return false
				}
				if l > time.Hour {
					logger.Debugf(ctx, "StreamPlayer[%s].controllerLoop: the length is more than an hour: %v (we expect only like a second, not an hour)", l, p.StreamID)
					if triedReopeningStream {
						logger.Debugf(ctx, "StreamPlayer[%s].controllerLoop: already tried reopening the stream, did not help, so restarting")
						restart()
						return false
					}
					if err := p.openStream(ctx); err != nil {
						logger.Error(ctx, "unable to re-open the stream: %v", err)
						restart()
						return false
					}
					triedReopeningStream = true
					startedWaitingForBuffering = time.Now()
					continue
				}
				if pos != 0 {
					return false
				}
				time.Sleep(100 * time.Millisecond)
			}

			logger.Errorf(ctx, "StreamPlayer[%s].controllerLoop: timed out on waiting until the player would start up; restarting", p.StreamID)
			restart()
			return false
		}() {
		}
	}

	if isClosed {
		return
	}

	select {
	case <-instanceCtx.Done():
		return
	default:
	}

	err := p.withPlayer(ctx, func(ctx context.Context, player types.Player) {
		err := player.SetupForStreaming(ctx)
		if err != nil {
			logger.Errorf(ctx, "unable to setup the player for streaming: %v", err)
		}
	})
	if err != nil {
		logger.Error(ctx, "unable to access the player for setting it up for streaming: %v", err)
	}

	observability.Go(ctx, func() {
		time.Sleep(time.Second) // TODO: delete this ugly racy hack
		p.notifyStart(context.WithValue(ctx, CtxKeyStreamPlayer, p))
	})

	getRestartChan := context.Background().Done()
	if fn := p.Config.GetRestartChanFunc; fn != nil {
		getRestartChan = fn()
	}

	logger.Debugf(ctx, "finished waiting for a publisher at '%s'", p.StreamID)

	t := time.NewTicker(100 * time.Millisecond)
	defer t.Stop()

	// now monitoring if everything is OK:
	var prevPos time.Duration
	var prevLength time.Duration
	posUpdatedAt := time.Now()
	curSpeed := float64(1)
	for {
		if isClosed {
			logger.Debug(ctx, "the player is closed, so closing the controllerLoop")
			return
		}
		select {
		case <-instanceCtx.Done():
			errmon.ObserveErrorCtx(ctx, p.Close())
			return
		case <-getRestartChan:
			logger.Debugf(ctx, "received a notification that the player should be restarted immediately")
			restart()
			return
		case <-t.C:
		}

		err := p.withPlayer(ctx, func(ctx context.Context, player types.Player) {
			now := time.Now()
			l, err := player.GetLength(ctx)
			if err != nil {
				logger.Errorf(ctx, "StreamPlayer[%s].controllerLoop: unable to get the current length: %v", p.StreamID, err)
				if prevLength != 0 {
					logger.Debugf(ctx, "previously GetLength worked, so it seems like the player died or something, restarting")
					restart()
					return
				}
				time.Sleep(time.Second)
				return
			}
			prevLength = l

			pos, err := player.GetPosition(ctx)
			if err != nil {
				logger.Errorf(ctx, "StreamPlayer[%s].controllerLoop: unable to get the current position: %v", p.StreamID, err)
				time.Sleep(time.Second)
				return
			}
			logger.Tracef(ctx, "StreamPlayer[%s].controllerLoop: now == %v, posUpdatedAt == %v, len == %v; pos == %v; readTimeout == %v", p.StreamID, now, posUpdatedAt, l, pos, p.Config.ReadTimeout)

			if pos < 0 {
				logger.Debugf(ctx, "negative position: %v", pos)
				restart()
				return
			}

			if l < 0 {
				logger.Debugf(ctx, "negative length: %v", l)
				restart()
				return
			}

			if pos != prevPos {
				posUpdatedAt = now
				prevPos = pos
			} else {
				if now.Sub(posUpdatedAt) > p.Config.ReadTimeout {
					logger.Debugf(ctx, "StreamPlayer[%s].controllerLoop: now == %v, posUpdatedAt == %v, len == %v; pos == %v; readTimeout == %v, restarting", p.StreamID, now, posUpdatedAt, l, pos, p.Config.ReadTimeout)
					restart()
					return
				}
			}

			lag := l - pos
			logger.Tracef(ctx, "StreamPlayer[%s].controllerLoop: lag == %v", p.StreamID, lag)
			if lag <= p.Config.JitterBufDuration {
				if curSpeed == 1 {
					return
				}
				logger.Debugf(ctx, "StreamPlayer[%s].controllerLoop: resetting the speed to 1", p.StreamID)
				err := player.SetSpeed(ctx, 1)
				if err != nil {
					logger.Errorf(ctx, "unable to reset the speed to 1: %v", err)
					return
				}
				curSpeed = 1
				return
			}

			speed := float64(1) +
				(p.Config.CatchupMaxSpeedFactor-float64(1))*
					(lag.Seconds()-p.Config.JitterBufDuration.Seconds())/
					(p.Config.MaxCatchupAtLag.Seconds()-p.Config.JitterBufDuration.Seconds())

			if speed > p.Config.CatchupMaxSpeedFactor {
				logger.Warnf(
					ctx,
					"speed is calculated higher than the maximum: %v > %v: (%v-1)*(%v-%v)/(%v-%v); lag calculation: %v - %v",
					speed, p.Config.CatchupMaxSpeedFactor,
					p.Config.CatchupMaxSpeedFactor,
					lag.Seconds(), p.Config.JitterBufDuration.Seconds(),
					p.Config.MaxCatchupAtLag.Seconds(), p.Config.JitterBufDuration.Seconds(),
					l, pos,
				)
				speed = p.Config.CatchupMaxSpeedFactor
			}

			if speed != curSpeed {
				logger.Debugf(ctx, "StreamPlayer[%s].controllerLoop: setting the speed to %v", p.StreamID, speed)
				err = player.SetSpeed(ctx, speed)
				if err != nil {
					logger.Errorf(ctx, "unable to set the speed to %v: %v", speed, err)
					return
				}
				curSpeed = speed
			}
		})
		if err != nil {
			logger.Error(ctx, "unable to get the player: %v", err)
			return
		}
	}
}

type ErrNilPlayer struct{}

func (e ErrNilPlayer) Error() string {
	return "p.Player is nil"
}

func (p *StreamPlayer) withPlayer(
	ctx context.Context,
	fn func(context.Context, types.Player),
) error {
	return xsync.DoR1(ctx, &p.PlayerLocker, func() error {
		if p.Player == nil {
			return ErrNilPlayer{}
		}
		fn(ctx, p.Player)
		return nil
	})
}

func (p *StreamPlayer) Close() error {
	ctx := context.TODO()
	return xsync.DoR1(ctx, &p.PlayerLocker, func() error {
		var err *multierror.Error
		if p.Cancel != nil {
			p.Cancel()
		}

		if p.Player != nil {
			err = multierror.Append(err, p.Player.Close(ctx))
			p.Player = nil
		}
		return err.ErrorOrNil()
	})
}
