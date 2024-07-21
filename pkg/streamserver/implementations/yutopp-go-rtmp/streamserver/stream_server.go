package streamserver

import (
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"

	"github.com/facebookincubator/go-belt"
	"github.com/facebookincubator/go-belt/tool/logger"
	"github.com/xaionaro-go/streamctl/pkg/streamserver/types"
	"github.com/xaionaro-go/streamctl/pkg/xlogger"
	"github.com/yutopp/go-rtmp"
)

type StreamServer struct {
	sync.Mutex
	Config                  *types.Config
	RelayServer             *RelayService
	ServerHandlers          []types.PortServer
	StreamDestinations      []types.StreamDestination
	ActiveStreamForwardings map[types.DestinationID]*ActiveStreamForwarding
}

func New(cfg *types.Config) *StreamServer {
	return &StreamServer{
		RelayServer: NewRelayService(),
		Config:      cfg,

		ActiveStreamForwardings: map[types.DestinationID]*ActiveStreamForwarding{},
	}
}

func (s *StreamServer) Init(ctx context.Context) error {
	s.Lock()
	defer s.Unlock()

	cfg := s.Config
	logger.Debugf(ctx, "config == %#+v", *cfg)

	for _, srv := range cfg.Servers {
		err := s.startServer(ctx, srv.Type, srv.Listen)
		if err != nil {
			return fmt.Errorf("unable to initialize %s server at %s: %w", srv.Type, srv.Listen, err)
		}
	}

	for dstID, dstCfg := range cfg.Destinations {
		err := s.addStreamDestination(ctx, dstID, dstCfg.URL)
		if err != nil {
			return fmt.Errorf("unable to initialize stream destination '%s' to %#+v: %w", dstID, dstCfg, err)
		}
	}

	for streamID, streamCfg := range cfg.Streams {
		err := s.addIncomingStream(ctx, streamID)
		if err != nil {
			return fmt.Errorf("unable to initialize stream '%s': %w", streamID, err)
		}

		for dstID, fwd := range streamCfg.Forwardings {
			if !fwd.Disabled {
				err := s.addStreamForward(ctx, streamID, dstID)
				if err != nil {
					return fmt.Errorf("unable to launch stream forward from '%s' to '%s': %w", streamID, dstID, err)
				}
			}
		}
	}

	return nil
}

func (s *StreamServer) ListServers(
	ctx context.Context,
) (_ret []types.PortServer) {
	logger.Tracef(ctx, "ListServers")
	defer func() { logger.Tracef(ctx, "/ListServers: %d servers", len(_ret)) }()
	s.Lock()
	defer s.Unlock()
	c := make([]types.PortServer, len(s.ServerHandlers))
	copy(c, s.ServerHandlers)
	return c
}

func (s *StreamServer) StartServer(
	ctx context.Context,
	serverType types.ServerType,
	listenAddr string,
) error {
	s.Lock()
	defer s.Unlock()
	err := s.startServer(ctx, serverType, listenAddr)
	if err != nil {
		return err
	}
	s.Config.Servers = append(s.Config.Servers, types.Server{
		Type:   serverType,
		Listen: listenAddr,
	})
	return nil
}

func (s *StreamServer) startServer(
	ctx context.Context,
	serverType types.ServerType,
	listenAddr string,
) (_ret error) {
	logger.Tracef(ctx, "startServer(%s, '%s')", serverType, listenAddr)
	defer func() { logger.Tracef(ctx, "/startServer(%s, '%s'): %v", serverType, listenAddr, _ret) }()
	var srv types.PortServer
	var err error
	switch serverType {
	case types.ServerTypeRTMP:
		var listener net.Listener
		listener, err = net.Listen("tcp", listenAddr)
		if err != nil {
			err = fmt.Errorf("unable to start listening '%s': %w", listenAddr, err)
			break
		}
		portSrv := &PortServer{
			Listener: listener,
		}
		portSrv.Server = rtmp.NewServer(&rtmp.ServerConfig{
			OnConnect: func(conn net.Conn) (io.ReadWriteCloser, *rtmp.ConnConfig) {
				ctx := belt.WithField(ctx, "client", conn.RemoteAddr().String())
				h := &Handler{
					relayService: s.RelayServer,
				}
				wrcc := types.NewReaderWriterCloseCounter(conn, &portSrv.ReadCount, &portSrv.WriteCount)
				return wrcc, &rtmp.ConnConfig{
					Handler: h,
					ControlState: rtmp.StreamControlStateConfig{
						DefaultBandwidthWindowSize: 20 * 1024 * 1024 / 8,
					},
					Logger: xlogger.LogrusFieldLoggerFromCtx(ctx),
				}
			},
		})
		go func() {
			err = portSrv.Serve(listener)
			if err != nil {
				err = fmt.Errorf("unable to start serving RTMP at '%s': %w", listener.Addr().String(), err)
				logger.Error(ctx, err)
			}
		}()
		srv = portSrv
	case types.ServerTypeRTSP:
		return fmt.Errorf("RTSP is not supported, yet")
	default:
		return fmt.Errorf("unexpected server type %v", serverType)
	}
	if err != nil {
		return err
	}

	s.ServerHandlers = append(s.ServerHandlers, srv)
	return nil
}

func (s *StreamServer) findServer(
	_ context.Context,
	server types.PortServer,
) (int, error) {
	for i := range s.ServerHandlers {
		if s.ServerHandlers[i] == server {
			return i, nil
		}
	}
	return -1, fmt.Errorf("server not found")
}

func (s *StreamServer) StopServer(
	ctx context.Context,
	server types.PortServer,
) error {
	s.Lock()
	defer s.Unlock()
	for idx, srv := range s.Config.Servers {
		if srv.Listen == server.ListenAddr() {
			s.Config.Servers = append(s.Config.Servers[:idx], s.Config.Servers[idx+1:]...)
			break
		}
	}
	return s.stopServer(ctx, server)
}

func (s *StreamServer) stopServer(
	ctx context.Context,
	server types.PortServer,
) error {
	idx, err := s.findServer(ctx, server)
	if err != nil {
		return err
	}

	s.ServerHandlers = append(s.ServerHandlers[:idx], s.ServerHandlers[idx+1:]...)
	return server.Close()
}

func (s *StreamServer) AddIncomingStream(
	ctx context.Context,
	streamID types.StreamID,
) error {
	s.Lock()
	defer s.Unlock()
	err := s.addIncomingStream(ctx, streamID)
	if err != nil {
		return err
	}
	s.Config.Streams[streamID] = &types.StreamConfig{}
	return nil
}

func (s *StreamServer) addIncomingStream(
	_ context.Context,
	streamID types.StreamID,
) error {
	return nil
	_, err := s.RelayServer.NewPubsub(string(streamID))
	if err != nil {
		return fmt.Errorf("unable to create the stream '%s': %w", streamID, err)
	}
	return nil
}

type IncomingStream struct {
	StreamID types.StreamID

	NumBytesWrote uint64
	NumBytesRead  uint64
}

func (s *StreamServer) ListIncomingStreams(
	ctx context.Context,
) []IncomingStream {
	s.Lock()
	defer s.Unlock()
	return s.listIncomingStreams(ctx)
}

func (s *StreamServer) listIncomingStreams(
	_ context.Context,
) []IncomingStream {
	var result []IncomingStream
	for name := range s.RelayServer.Pubsubs() {
		result = append(
			result,
			IncomingStream{
				StreamID: types.StreamID(name),
			},
		)
	}
	return result
}

func (s *StreamServer) RemoveIncomingStream(
	ctx context.Context,
	streamID types.StreamID,
) error {
	s.Lock()
	defer s.Unlock()
	delete(s.Config.Streams, streamID)
	return s.removeIncomingStream(ctx, streamID)
}

func (s *StreamServer) removeIncomingStream(
	_ context.Context,
	streamID types.StreamID,
) error {
	return nil
	if err := s.RelayServer.RemovePubsub(string(streamID)); err != nil {
		return fmt.Errorf("unable to remove stream '%s': %w", streamID, err)
	}
	return nil
}

type StreamForward struct {
	StreamID      types.StreamID
	DestinationID types.DestinationID
	Enabled       bool
	NumBytesWrote uint64
	NumBytesRead  uint64
}

func (s *StreamServer) AddStreamForward(
	ctx context.Context,
	streamID types.StreamID,
	destinationID types.DestinationID,
	enabled bool,
) error {
	s.Lock()
	defer s.Unlock()
	streamConfig := s.Config.Streams[streamID]
	if _, ok := streamConfig.Forwardings[destinationID]; ok {
		return fmt.Errorf("the forwarding %s->%s already exists", streamID, destinationID)
	}

	streamConfig.Forwardings[destinationID] = types.ForwardingConfig{
		Disabled: !enabled,
	}

	if enabled {
		err := s.addStreamForward(ctx, streamID, destinationID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *StreamServer) addStreamForward(
	ctx context.Context,
	streamID types.StreamID,
	destinationID types.DestinationID,
) error {
	ctx = belt.WithField(ctx, "stream_forward", fmt.Sprintf("%s->%s", streamID, destinationID))
	if _, ok := s.ActiveStreamForwardings[destinationID]; ok {
		return fmt.Errorf("there is already an active stream forwarding to '%s'", destinationID)
	}
	streamSrc := s.RelayServer.GetPubsub(string(streamID))
	if streamSrc == nil {
		return fmt.Errorf("unable to find stream ID '%s', available stream IDs: %s", streamID, strings.Join(s.RelayServer.PubsubNames(), ", "))
	}
	dst, err := s.findStreamDestinationByID(ctx, destinationID)
	if err != nil {
		return fmt.Errorf("unable to find stream destination '%s': %w", destinationID, err)
	}

	fwd, err := newActiveStreamForward(ctx, streamID, destinationID, streamSrc, dst.URL)
	if err != nil {
		return fmt.Errorf("unable to run the stream forwarding: %w", err)
	}
	s.ActiveStreamForwardings[destinationID] = fwd

	return nil
}

func (s *StreamServer) UpdateStreamForward(
	ctx context.Context,
	streamID types.StreamID,
	destinationID types.DestinationID,
	enabled bool,
) error {
	s.Lock()
	defer s.Unlock()
	streamConfig := s.Config.Streams[streamID]
	fwdCfg, ok := streamConfig.Forwardings[destinationID]
	if !ok {
		return fmt.Errorf("the forwarding %s->%s does not exist", streamID, destinationID)
	}

	if fwdCfg.Disabled && enabled {
		err := s.addStreamForward(ctx, streamID, destinationID)
		if err != nil {
			return err
		}
	}
	if !fwdCfg.Disabled && !enabled {
		err := s.removeStreamForward(ctx, streamID, destinationID)
		if err != nil {
			return err
		}
	}
	streamConfig.Forwardings[destinationID] = types.ForwardingConfig{
		Disabled: !enabled,
	}
	return nil
}

func (s *StreamServer) ListStreamForwards(
	ctx context.Context,
) ([]StreamForward, error) {
	s.Lock()
	defer s.Unlock()

	activeStreamForwards, err := s.listStreamForwards(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get the list of active stream forwardings: %w", err)
	}

	type fwdID struct {
		StreamID types.StreamID
		DestID   types.DestinationID
	}
	m := map[fwdID]*StreamForward{}
	for idx := range activeStreamForwards {
		fwd := &activeStreamForwards[idx]
		m[fwdID{
			StreamID: fwd.StreamID,
			DestID:   fwd.DestinationID,
		}] = fwd
	}

	var result []StreamForward
	for streamID, stream := range s.Config.Streams {
		for dstID, cfg := range stream.Forwardings {
			item := StreamForward{
				StreamID:      streamID,
				DestinationID: dstID,
				Enabled:       !cfg.Disabled,
			}
			if activeFwd, ok := m[fwdID{
				StreamID: streamID,
				DestID:   dstID,
			}]; ok {
				item.NumBytesWrote = activeFwd.NumBytesWrote
				item.NumBytesRead = activeFwd.NumBytesRead
			}
			logger.Tracef(ctx, "stream forwarding '%s->%s': %#+v", streamID, dstID, cfg)
			result = append(result, item)
		}
	}
	return result, nil
}

func (s *StreamServer) listStreamForwards(
	_ context.Context,
) ([]StreamForward, error) {
	var result []StreamForward
	for _, fwd := range s.ActiveStreamForwardings {
		result = append(result, StreamForward{
			StreamID:      fwd.StreamID,
			DestinationID: fwd.DestinationID,
			Enabled:       true,
			NumBytesWrote: fwd.WriteCount.Load(),
			NumBytesRead:  fwd.ReadCount.Load(),
		})
	}
	return result, nil
}

func (s *StreamServer) RemoveStreamForward(
	ctx context.Context,
	streamID types.StreamID,
	dstID types.DestinationID,
) error {
	s.Lock()
	defer s.Unlock()
	streamCfg := s.Config.Streams[streamID]
	if _, ok := streamCfg.Forwardings[dstID]; !ok {
		return fmt.Errorf("the forwarding %s->%s does not exist", streamID, dstID)
	}
	delete(streamCfg.Forwardings, dstID)
	return s.removeStreamForward(ctx, streamID, dstID)
}

func (s *StreamServer) removeStreamForward(
	_ context.Context,
	_ types.StreamID,
	dstID types.DestinationID,
) error {
	fwd := s.ActiveStreamForwardings[dstID]
	if fwd == nil {
		return nil
	}

	delete(s.ActiveStreamForwardings, dstID)
	err := fwd.Close()
	if err != nil {
		return fmt.Errorf("unable to close stream forwarding: %w", err)
	}

	return nil
}

func (s *StreamServer) ListStreamDestinations(
	ctx context.Context,
) ([]types.StreamDestination, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	return s.listStreamDestinations(ctx)
}

func (s *StreamServer) listStreamDestinations(
	_ context.Context,
) ([]types.StreamDestination, error) {
	c := make([]types.StreamDestination, len(s.StreamDestinations))
	copy(c, s.StreamDestinations)
	return c, nil
}

func (s *StreamServer) AddStreamDestination(
	ctx context.Context,
	destinationID types.DestinationID,
	url string,
) error {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	err := s.addStreamDestination(ctx, destinationID, url)
	if err != nil {
		return err
	}
	s.Config.Destinations[destinationID] = &types.DestinationConfig{URL: url}
	return nil
}

func (s *StreamServer) addStreamDestination(
	_ context.Context,
	destinationID types.DestinationID,
	url string,
) error {
	s.StreamDestinations = append(s.StreamDestinations, types.StreamDestination{
		ID:  destinationID,
		URL: url,
	})
	return nil
}

func (s *StreamServer) RemoveStreamDestination(
	ctx context.Context,
	destinationID types.DestinationID,
) error {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	for _, streamCfg := range s.Config.Streams {
		delete(streamCfg.Forwardings, destinationID)
	}
	delete(s.Config.Destinations, destinationID)
	return s.removeStreamDestination(ctx, destinationID)
}

func (s *StreamServer) removeStreamDestination(
	ctx context.Context,
	destinationID types.DestinationID,
) error {
	streamForwards, err := s.listStreamForwards(ctx)
	if err != nil {
		return fmt.Errorf("unable to list stream forwardings: %w", err)
	}
	for _, fwd := range streamForwards {
		if fwd.DestinationID == destinationID {
			s.removeStreamForward(ctx, fwd.StreamID, fwd.DestinationID)
		}
	}

	for i := range s.StreamDestinations {
		if s.StreamDestinations[i].ID == destinationID {
			s.StreamDestinations = append(s.StreamDestinations[:i], s.StreamDestinations[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("have not found stream destination with id %s", destinationID)
}

func (s *StreamServer) findStreamDestinationByID(
	_ context.Context,
	destinationID types.DestinationID,
) (types.StreamDestination, error) {
	for _, dst := range s.StreamDestinations {
		if dst.ID == destinationID {
			return dst, nil
		}
	}
	return types.StreamDestination{}, fmt.Errorf("unable to find a stream destination by StreamID '%s'", destinationID)
}
