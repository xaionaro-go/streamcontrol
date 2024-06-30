// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.12
// source: streamd.proto

package streamd_grpc

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// StreamDClient is the client API for StreamD service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type StreamDClient interface {
	GetConfig(ctx context.Context, in *GetConfigRequest, opts ...grpc.CallOption) (*GetConfigReply, error)
	SetConfig(ctx context.Context, in *SetConfigRequest, opts ...grpc.CallOption) (*SetConfigReply, error)
	SaveConfig(ctx context.Context, in *SaveConfigRequest, opts ...grpc.CallOption) (*SaveConfigReply, error)
	ResetCache(ctx context.Context, in *ResetCacheRequest, opts ...grpc.CallOption) (*ResetCacheReply, error)
	InitCache(ctx context.Context, in *InitCacheRequest, opts ...grpc.CallOption) (*InitCacheReply, error)
	StartStream(ctx context.Context, in *StartStreamRequest, opts ...grpc.CallOption) (*StartStreamReply, error)
	EndStream(ctx context.Context, in *EndStreamRequest, opts ...grpc.CallOption) (*EndStreamReply, error)
	GetStreamStatus(ctx context.Context, in *GetStreamStatusRequest, opts ...grpc.CallOption) (*GetStreamStatusReply, error)
	GetBackendInfo(ctx context.Context, in *GetBackendInfoRequest, opts ...grpc.CallOption) (*GetBackendInfoReply, error)
	Restart(ctx context.Context, in *RestartRequest, opts ...grpc.CallOption) (*RestartReply, error)
	SetTitle(ctx context.Context, in *SetTitleRequest, opts ...grpc.CallOption) (*SetTitleReply, error)
	SetDescription(ctx context.Context, in *SetDescriptionRequest, opts ...grpc.CallOption) (*SetDescriptionReply, error)
	SetApplyProfile(ctx context.Context, in *SetApplyProfileRequest, opts ...grpc.CallOption) (*SetApplyProfileReply, error)
	UpdateStream(ctx context.Context, in *UpdateStreamRequest, opts ...grpc.CallOption) (*UpdateStreamReply, error)
	EXPERIMENTAL_ReinitStreamControllers(ctx context.Context, in *EXPERIMENTAL_ReinitStreamControllersRequest, opts ...grpc.CallOption) (*EXPERIMENTAL_ReinitStreamControllersReply, error)
	OBSOLETE_FetchConfig(ctx context.Context, in *OBSOLETE_FetchConfigRequest, opts ...grpc.CallOption) (*OBSOLETE_FetchConfigReply, error)
	OBSOLETE_GitInfo(ctx context.Context, in *OBSOLETE_GetGitInfoRequest, opts ...grpc.CallOption) (*OBSOLETE_GetGitInfoReply, error)
	OBSOLETE_GitRelogin(ctx context.Context, in *OBSOLETE_GitReloginRequest, opts ...grpc.CallOption) (*OBSOLETE_GitReloginReply, error)
}

type streamDClient struct {
	cc grpc.ClientConnInterface
}

func NewStreamDClient(cc grpc.ClientConnInterface) StreamDClient {
	return &streamDClient{cc}
}

func (c *streamDClient) GetConfig(ctx context.Context, in *GetConfigRequest, opts ...grpc.CallOption) (*GetConfigReply, error) {
	out := new(GetConfigReply)
	err := c.cc.Invoke(ctx, "/StreamD/GetConfig", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *streamDClient) SetConfig(ctx context.Context, in *SetConfigRequest, opts ...grpc.CallOption) (*SetConfigReply, error) {
	out := new(SetConfigReply)
	err := c.cc.Invoke(ctx, "/StreamD/SetConfig", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *streamDClient) SaveConfig(ctx context.Context, in *SaveConfigRequest, opts ...grpc.CallOption) (*SaveConfigReply, error) {
	out := new(SaveConfigReply)
	err := c.cc.Invoke(ctx, "/StreamD/SaveConfig", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *streamDClient) ResetCache(ctx context.Context, in *ResetCacheRequest, opts ...grpc.CallOption) (*ResetCacheReply, error) {
	out := new(ResetCacheReply)
	err := c.cc.Invoke(ctx, "/StreamD/ResetCache", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *streamDClient) InitCache(ctx context.Context, in *InitCacheRequest, opts ...grpc.CallOption) (*InitCacheReply, error) {
	out := new(InitCacheReply)
	err := c.cc.Invoke(ctx, "/StreamD/InitCache", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *streamDClient) StartStream(ctx context.Context, in *StartStreamRequest, opts ...grpc.CallOption) (*StartStreamReply, error) {
	out := new(StartStreamReply)
	err := c.cc.Invoke(ctx, "/StreamD/StartStream", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *streamDClient) EndStream(ctx context.Context, in *EndStreamRequest, opts ...grpc.CallOption) (*EndStreamReply, error) {
	out := new(EndStreamReply)
	err := c.cc.Invoke(ctx, "/StreamD/EndStream", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *streamDClient) GetStreamStatus(ctx context.Context, in *GetStreamStatusRequest, opts ...grpc.CallOption) (*GetStreamStatusReply, error) {
	out := new(GetStreamStatusReply)
	err := c.cc.Invoke(ctx, "/StreamD/GetStreamStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *streamDClient) GetBackendInfo(ctx context.Context, in *GetBackendInfoRequest, opts ...grpc.CallOption) (*GetBackendInfoReply, error) {
	out := new(GetBackendInfoReply)
	err := c.cc.Invoke(ctx, "/StreamD/GetBackendInfo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *streamDClient) Restart(ctx context.Context, in *RestartRequest, opts ...grpc.CallOption) (*RestartReply, error) {
	out := new(RestartReply)
	err := c.cc.Invoke(ctx, "/StreamD/Restart", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *streamDClient) SetTitle(ctx context.Context, in *SetTitleRequest, opts ...grpc.CallOption) (*SetTitleReply, error) {
	out := new(SetTitleReply)
	err := c.cc.Invoke(ctx, "/StreamD/SetTitle", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *streamDClient) SetDescription(ctx context.Context, in *SetDescriptionRequest, opts ...grpc.CallOption) (*SetDescriptionReply, error) {
	out := new(SetDescriptionReply)
	err := c.cc.Invoke(ctx, "/StreamD/SetDescription", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *streamDClient) SetApplyProfile(ctx context.Context, in *SetApplyProfileRequest, opts ...grpc.CallOption) (*SetApplyProfileReply, error) {
	out := new(SetApplyProfileReply)
	err := c.cc.Invoke(ctx, "/StreamD/SetApplyProfile", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *streamDClient) UpdateStream(ctx context.Context, in *UpdateStreamRequest, opts ...grpc.CallOption) (*UpdateStreamReply, error) {
	out := new(UpdateStreamReply)
	err := c.cc.Invoke(ctx, "/StreamD/UpdateStream", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *streamDClient) EXPERIMENTAL_ReinitStreamControllers(ctx context.Context, in *EXPERIMENTAL_ReinitStreamControllersRequest, opts ...grpc.CallOption) (*EXPERIMENTAL_ReinitStreamControllersReply, error) {
	out := new(EXPERIMENTAL_ReinitStreamControllersReply)
	err := c.cc.Invoke(ctx, "/StreamD/EXPERIMENTAL_ReinitStreamControllers", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *streamDClient) OBSOLETE_FetchConfig(ctx context.Context, in *OBSOLETE_FetchConfigRequest, opts ...grpc.CallOption) (*OBSOLETE_FetchConfigReply, error) {
	out := new(OBSOLETE_FetchConfigReply)
	err := c.cc.Invoke(ctx, "/StreamD/OBSOLETE_FetchConfig", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *streamDClient) OBSOLETE_GitInfo(ctx context.Context, in *OBSOLETE_GetGitInfoRequest, opts ...grpc.CallOption) (*OBSOLETE_GetGitInfoReply, error) {
	out := new(OBSOLETE_GetGitInfoReply)
	err := c.cc.Invoke(ctx, "/StreamD/OBSOLETE_GitInfo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *streamDClient) OBSOLETE_GitRelogin(ctx context.Context, in *OBSOLETE_GitReloginRequest, opts ...grpc.CallOption) (*OBSOLETE_GitReloginReply, error) {
	out := new(OBSOLETE_GitReloginReply)
	err := c.cc.Invoke(ctx, "/StreamD/OBSOLETE_GitRelogin", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// StreamDServer is the server API for StreamD service.
// All implementations must embed UnimplementedStreamDServer
// for forward compatibility
type StreamDServer interface {
	GetConfig(context.Context, *GetConfigRequest) (*GetConfigReply, error)
	SetConfig(context.Context, *SetConfigRequest) (*SetConfigReply, error)
	SaveConfig(context.Context, *SaveConfigRequest) (*SaveConfigReply, error)
	ResetCache(context.Context, *ResetCacheRequest) (*ResetCacheReply, error)
	InitCache(context.Context, *InitCacheRequest) (*InitCacheReply, error)
	StartStream(context.Context, *StartStreamRequest) (*StartStreamReply, error)
	EndStream(context.Context, *EndStreamRequest) (*EndStreamReply, error)
	GetStreamStatus(context.Context, *GetStreamStatusRequest) (*GetStreamStatusReply, error)
	GetBackendInfo(context.Context, *GetBackendInfoRequest) (*GetBackendInfoReply, error)
	Restart(context.Context, *RestartRequest) (*RestartReply, error)
	SetTitle(context.Context, *SetTitleRequest) (*SetTitleReply, error)
	SetDescription(context.Context, *SetDescriptionRequest) (*SetDescriptionReply, error)
	SetApplyProfile(context.Context, *SetApplyProfileRequest) (*SetApplyProfileReply, error)
	UpdateStream(context.Context, *UpdateStreamRequest) (*UpdateStreamReply, error)
	EXPERIMENTAL_ReinitStreamControllers(context.Context, *EXPERIMENTAL_ReinitStreamControllersRequest) (*EXPERIMENTAL_ReinitStreamControllersReply, error)
	OBSOLETE_FetchConfig(context.Context, *OBSOLETE_FetchConfigRequest) (*OBSOLETE_FetchConfigReply, error)
	OBSOLETE_GitInfo(context.Context, *OBSOLETE_GetGitInfoRequest) (*OBSOLETE_GetGitInfoReply, error)
	OBSOLETE_GitRelogin(context.Context, *OBSOLETE_GitReloginRequest) (*OBSOLETE_GitReloginReply, error)
	mustEmbedUnimplementedStreamDServer()
}

// UnimplementedStreamDServer must be embedded to have forward compatible implementations.
type UnimplementedStreamDServer struct {
}

func (UnimplementedStreamDServer) GetConfig(context.Context, *GetConfigRequest) (*GetConfigReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetConfig not implemented")
}
func (UnimplementedStreamDServer) SetConfig(context.Context, *SetConfigRequest) (*SetConfigReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetConfig not implemented")
}
func (UnimplementedStreamDServer) SaveConfig(context.Context, *SaveConfigRequest) (*SaveConfigReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SaveConfig not implemented")
}
func (UnimplementedStreamDServer) ResetCache(context.Context, *ResetCacheRequest) (*ResetCacheReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ResetCache not implemented")
}
func (UnimplementedStreamDServer) InitCache(context.Context, *InitCacheRequest) (*InitCacheReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method InitCache not implemented")
}
func (UnimplementedStreamDServer) StartStream(context.Context, *StartStreamRequest) (*StartStreamReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StartStream not implemented")
}
func (UnimplementedStreamDServer) EndStream(context.Context, *EndStreamRequest) (*EndStreamReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method EndStream not implemented")
}
func (UnimplementedStreamDServer) GetStreamStatus(context.Context, *GetStreamStatusRequest) (*GetStreamStatusReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetStreamStatus not implemented")
}
func (UnimplementedStreamDServer) GetBackendInfo(context.Context, *GetBackendInfoRequest) (*GetBackendInfoReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBackendInfo not implemented")
}
func (UnimplementedStreamDServer) Restart(context.Context, *RestartRequest) (*RestartReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Restart not implemented")
}
func (UnimplementedStreamDServer) SetTitle(context.Context, *SetTitleRequest) (*SetTitleReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetTitle not implemented")
}
func (UnimplementedStreamDServer) SetDescription(context.Context, *SetDescriptionRequest) (*SetDescriptionReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetDescription not implemented")
}
func (UnimplementedStreamDServer) SetApplyProfile(context.Context, *SetApplyProfileRequest) (*SetApplyProfileReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetApplyProfile not implemented")
}
func (UnimplementedStreamDServer) UpdateStream(context.Context, *UpdateStreamRequest) (*UpdateStreamReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateStream not implemented")
}
func (UnimplementedStreamDServer) EXPERIMENTAL_ReinitStreamControllers(context.Context, *EXPERIMENTAL_ReinitStreamControllersRequest) (*EXPERIMENTAL_ReinitStreamControllersReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method EXPERIMENTAL_ReinitStreamControllers not implemented")
}
func (UnimplementedStreamDServer) OBSOLETE_FetchConfig(context.Context, *OBSOLETE_FetchConfigRequest) (*OBSOLETE_FetchConfigReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method OBSOLETE_FetchConfig not implemented")
}
func (UnimplementedStreamDServer) OBSOLETE_GitInfo(context.Context, *OBSOLETE_GetGitInfoRequest) (*OBSOLETE_GetGitInfoReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method OBSOLETE_GitInfo not implemented")
}
func (UnimplementedStreamDServer) OBSOLETE_GitRelogin(context.Context, *OBSOLETE_GitReloginRequest) (*OBSOLETE_GitReloginReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method OBSOLETE_GitRelogin not implemented")
}
func (UnimplementedStreamDServer) mustEmbedUnimplementedStreamDServer() {}

// UnsafeStreamDServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to StreamDServer will
// result in compilation errors.
type UnsafeStreamDServer interface {
	mustEmbedUnimplementedStreamDServer()
}

func RegisterStreamDServer(s grpc.ServiceRegistrar, srv StreamDServer) {
	s.RegisterService(&StreamD_ServiceDesc, srv)
}

func _StreamD_GetConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StreamDServer).GetConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/StreamD/GetConfig",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StreamDServer).GetConfig(ctx, req.(*GetConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _StreamD_SetConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StreamDServer).SetConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/StreamD/SetConfig",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StreamDServer).SetConfig(ctx, req.(*SetConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _StreamD_SaveConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SaveConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StreamDServer).SaveConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/StreamD/SaveConfig",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StreamDServer).SaveConfig(ctx, req.(*SaveConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _StreamD_ResetCache_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ResetCacheRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StreamDServer).ResetCache(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/StreamD/ResetCache",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StreamDServer).ResetCache(ctx, req.(*ResetCacheRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _StreamD_InitCache_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InitCacheRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StreamDServer).InitCache(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/StreamD/InitCache",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StreamDServer).InitCache(ctx, req.(*InitCacheRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _StreamD_StartStream_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StartStreamRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StreamDServer).StartStream(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/StreamD/StartStream",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StreamDServer).StartStream(ctx, req.(*StartStreamRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _StreamD_EndStream_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EndStreamRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StreamDServer).EndStream(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/StreamD/EndStream",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StreamDServer).EndStream(ctx, req.(*EndStreamRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _StreamD_GetStreamStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetStreamStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StreamDServer).GetStreamStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/StreamD/GetStreamStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StreamDServer).GetStreamStatus(ctx, req.(*GetStreamStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _StreamD_GetBackendInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetBackendInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StreamDServer).GetBackendInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/StreamD/GetBackendInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StreamDServer).GetBackendInfo(ctx, req.(*GetBackendInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _StreamD_Restart_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RestartRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StreamDServer).Restart(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/StreamD/Restart",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StreamDServer).Restart(ctx, req.(*RestartRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _StreamD_SetTitle_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetTitleRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StreamDServer).SetTitle(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/StreamD/SetTitle",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StreamDServer).SetTitle(ctx, req.(*SetTitleRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _StreamD_SetDescription_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetDescriptionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StreamDServer).SetDescription(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/StreamD/SetDescription",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StreamDServer).SetDescription(ctx, req.(*SetDescriptionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _StreamD_SetApplyProfile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetApplyProfileRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StreamDServer).SetApplyProfile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/StreamD/SetApplyProfile",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StreamDServer).SetApplyProfile(ctx, req.(*SetApplyProfileRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _StreamD_UpdateStream_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateStreamRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StreamDServer).UpdateStream(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/StreamD/UpdateStream",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StreamDServer).UpdateStream(ctx, req.(*UpdateStreamRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _StreamD_EXPERIMENTAL_ReinitStreamControllers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EXPERIMENTAL_ReinitStreamControllersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StreamDServer).EXPERIMENTAL_ReinitStreamControllers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/StreamD/EXPERIMENTAL_ReinitStreamControllers",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StreamDServer).EXPERIMENTAL_ReinitStreamControllers(ctx, req.(*EXPERIMENTAL_ReinitStreamControllersRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _StreamD_OBSOLETE_FetchConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OBSOLETE_FetchConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StreamDServer).OBSOLETE_FetchConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/StreamD/OBSOLETE_FetchConfig",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StreamDServer).OBSOLETE_FetchConfig(ctx, req.(*OBSOLETE_FetchConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _StreamD_OBSOLETE_GitInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OBSOLETE_GetGitInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StreamDServer).OBSOLETE_GitInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/StreamD/OBSOLETE_GitInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StreamDServer).OBSOLETE_GitInfo(ctx, req.(*OBSOLETE_GetGitInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _StreamD_OBSOLETE_GitRelogin_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OBSOLETE_GitReloginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StreamDServer).OBSOLETE_GitRelogin(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/StreamD/OBSOLETE_GitRelogin",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StreamDServer).OBSOLETE_GitRelogin(ctx, req.(*OBSOLETE_GitReloginRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// StreamD_ServiceDesc is the grpc.ServiceDesc for StreamD service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var StreamD_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "StreamD",
	HandlerType: (*StreamDServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetConfig",
			Handler:    _StreamD_GetConfig_Handler,
		},
		{
			MethodName: "SetConfig",
			Handler:    _StreamD_SetConfig_Handler,
		},
		{
			MethodName: "SaveConfig",
			Handler:    _StreamD_SaveConfig_Handler,
		},
		{
			MethodName: "ResetCache",
			Handler:    _StreamD_ResetCache_Handler,
		},
		{
			MethodName: "InitCache",
			Handler:    _StreamD_InitCache_Handler,
		},
		{
			MethodName: "StartStream",
			Handler:    _StreamD_StartStream_Handler,
		},
		{
			MethodName: "EndStream",
			Handler:    _StreamD_EndStream_Handler,
		},
		{
			MethodName: "GetStreamStatus",
			Handler:    _StreamD_GetStreamStatus_Handler,
		},
		{
			MethodName: "GetBackendInfo",
			Handler:    _StreamD_GetBackendInfo_Handler,
		},
		{
			MethodName: "Restart",
			Handler:    _StreamD_Restart_Handler,
		},
		{
			MethodName: "SetTitle",
			Handler:    _StreamD_SetTitle_Handler,
		},
		{
			MethodName: "SetDescription",
			Handler:    _StreamD_SetDescription_Handler,
		},
		{
			MethodName: "SetApplyProfile",
			Handler:    _StreamD_SetApplyProfile_Handler,
		},
		{
			MethodName: "UpdateStream",
			Handler:    _StreamD_UpdateStream_Handler,
		},
		{
			MethodName: "EXPERIMENTAL_ReinitStreamControllers",
			Handler:    _StreamD_EXPERIMENTAL_ReinitStreamControllers_Handler,
		},
		{
			MethodName: "OBSOLETE_FetchConfig",
			Handler:    _StreamD_OBSOLETE_FetchConfig_Handler,
		},
		{
			MethodName: "OBSOLETE_GitInfo",
			Handler:    _StreamD_OBSOLETE_GitInfo_Handler,
		},
		{
			MethodName: "OBSOLETE_GitRelogin",
			Handler:    _StreamD_OBSOLETE_GitRelogin_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "streamd.proto",
}
