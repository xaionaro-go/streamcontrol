package streamserver

import (
	"context"
	"fmt"
	"net"

	"github.com/bluenviron/gortsplib/v4/pkg/auth"
	"github.com/xaionaro-go/mediamtx/pkg/conf"
	"github.com/xaionaro-go/mediamtx/pkg/externalcmd"
	"github.com/xaionaro-go/mediamtx/pkg/servers/hls"
	"github.com/xaionaro-go/mediamtx/pkg/servers/rtsp"
	"github.com/xaionaro-go/mediamtx/pkg/servers/srt"
	"github.com/xaionaro-go/streamctl/pkg/streamserver/types"
)

func newRTSPServer() *rtsp.Server {
	panic("not implemented")
	return &rtsp.Server{
		Address:             "",
		AuthMethods:         []auth.ValidateMethod{},
		ReadTimeout:         0,
		WriteTimeout:        0,
		WriteQueueSize:      0,
		UseUDP:              false,
		UseMulticast:        false,
		RTPAddress:          "",
		RTCPAddress:         "",
		MulticastIPRange:    "",
		MulticastRTPPort:    0,
		MulticastRTCPPort:   0,
		IsTLS:               false,
		ServerCert:          "",
		ServerKey:           "",
		RTSPAddress:         "",
		Protocols:           map[conf.Protocol]struct{}{},
		RunOnConnect:        "",
		RunOnConnectRestart: false,
		RunOnDisconnect:     "",
		ExternalCmdPool:     &externalcmd.Pool{},
		PathManager:         nil,
		Parent:              nil,
	}
}

func newSRTServer() *srt.Server {
	panic("not implemented")
	return &srt.Server{
		Address:             "",
		RTSPAddress:         "",
		ReadTimeout:         0,
		WriteTimeout:        0,
		UDPMaxPayloadSize:   0,
		RunOnConnect:        "",
		RunOnConnectRestart: false,
		RunOnDisconnect:     "",
		ExternalCmdPool:     &externalcmd.Pool{},
		PathManager:         nil,
		Parent:              nil,
	}
}

func newHLSServer() *hls.Server {
	panic("not implemented")
	return &hls.Server{
		Address:         "",
		Encryption:      false,
		ServerKey:       "",
		ServerCert:      "",
		AllowOrigin:     "",
		TrustedProxies:  []net.IPNet{},
		AlwaysRemux:     false,
		Variant:         0,
		SegmentCount:    0,
		SegmentDuration: 0,
		PartDuration:    0,
		SegmentMaxSize:  0,
		Directory:       "",
		ReadTimeout:     0,
		MuxerCloseAfter: 0,
		PathManager:     nil,
		Parent:          nil,
	}
}

func (s *StreamServer) newServerRTSP(
	ctx context.Context,
	listenAddr string,
	opts ...types.ServerOption,
) (_ types.PortServer, _ret error) {
	return nil, fmt.Errorf("support of RTSP is not implemented, yet")
}

func (s *StreamServer) newServerSRT(
	ctx context.Context,
	listenAddr string,
	opts ...types.ServerOption,
) (_ types.PortServer, _ret error) {
	return nil, fmt.Errorf("support of SRT is not implemented, yet")
}
func (s *StreamServer) newServerHLS(
	ctx context.Context,
	listenAddr string,
	opts ...types.ServerOption,
) (_ types.PortServer, _ret error) {
	return nil, fmt.Errorf("support of HLS is not implemented, yet")
}

func (s *StreamServer) newServerWebRTC(
	ctx context.Context,
	listenAddr string,
	opts ...types.ServerOption,
) (_ types.PortServer, _ret error) {
	return nil, fmt.Errorf("support of WebRTC is not implemented, yet")
}