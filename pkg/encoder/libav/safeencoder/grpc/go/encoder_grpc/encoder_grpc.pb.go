// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package encoder_grpc

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// EncoderClient is the client API for Encoder service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type EncoderClient interface {
	SetLoggingLevel(ctx context.Context, in *SetLoggingLevelRequest, opts ...grpc.CallOption) (*SetLoggingLevelReply, error)
	NewInput(ctx context.Context, in *NewInputRequest, opts ...grpc.CallOption) (*NewInputReply, error)
	NewOutput(ctx context.Context, in *NewOutputRequest, opts ...grpc.CallOption) (*NewOutputReply, error)
	NewEncoder(ctx context.Context, in *NewEncoderRequest, opts ...grpc.CallOption) (*NewEncoderReply, error)
	SetEncoderConfig(ctx context.Context, in *SetEncoderConfigRequest, opts ...grpc.CallOption) (*SetEncoderConfigReply, error)
	CloseInput(ctx context.Context, in *CloseInputRequest, opts ...grpc.CallOption) (*CloseInputReply, error)
	CloseOutput(ctx context.Context, in *CloseOutputRequest, opts ...grpc.CallOption) (*CloseOutputReply, error)
	GetEncoderStats(ctx context.Context, in *GetEncoderStatsRequest, opts ...grpc.CallOption) (*GetEncoderStatsReply, error)
	StartEncoding(ctx context.Context, in *StartEncodingRequest, opts ...grpc.CallOption) (*StartEncodingReply, error)
	EncodingEndedChan(ctx context.Context, in *EncodingEndedChanRequest, opts ...grpc.CallOption) (Encoder_EncodingEndedChanClient, error)
}

type encoderClient struct {
	cc grpc.ClientConnInterface
}

func NewEncoderClient(cc grpc.ClientConnInterface) EncoderClient {
	return &encoderClient{cc}
}

func (c *encoderClient) SetLoggingLevel(ctx context.Context, in *SetLoggingLevelRequest, opts ...grpc.CallOption) (*SetLoggingLevelReply, error) {
	out := new(SetLoggingLevelReply)
	err := c.cc.Invoke(ctx, "/encoder_grpc.Encoder/SetLoggingLevel", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *encoderClient) NewInput(ctx context.Context, in *NewInputRequest, opts ...grpc.CallOption) (*NewInputReply, error) {
	out := new(NewInputReply)
	err := c.cc.Invoke(ctx, "/encoder_grpc.Encoder/NewInput", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *encoderClient) NewOutput(ctx context.Context, in *NewOutputRequest, opts ...grpc.CallOption) (*NewOutputReply, error) {
	out := new(NewOutputReply)
	err := c.cc.Invoke(ctx, "/encoder_grpc.Encoder/NewOutput", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *encoderClient) NewEncoder(ctx context.Context, in *NewEncoderRequest, opts ...grpc.CallOption) (*NewEncoderReply, error) {
	out := new(NewEncoderReply)
	err := c.cc.Invoke(ctx, "/encoder_grpc.Encoder/NewEncoder", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *encoderClient) SetEncoderConfig(ctx context.Context, in *SetEncoderConfigRequest, opts ...grpc.CallOption) (*SetEncoderConfigReply, error) {
	out := new(SetEncoderConfigReply)
	err := c.cc.Invoke(ctx, "/encoder_grpc.Encoder/SetEncoderConfig", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *encoderClient) CloseInput(ctx context.Context, in *CloseInputRequest, opts ...grpc.CallOption) (*CloseInputReply, error) {
	out := new(CloseInputReply)
	err := c.cc.Invoke(ctx, "/encoder_grpc.Encoder/CloseInput", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *encoderClient) CloseOutput(ctx context.Context, in *CloseOutputRequest, opts ...grpc.CallOption) (*CloseOutputReply, error) {
	out := new(CloseOutputReply)
	err := c.cc.Invoke(ctx, "/encoder_grpc.Encoder/CloseOutput", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *encoderClient) GetEncoderStats(ctx context.Context, in *GetEncoderStatsRequest, opts ...grpc.CallOption) (*GetEncoderStatsReply, error) {
	out := new(GetEncoderStatsReply)
	err := c.cc.Invoke(ctx, "/encoder_grpc.Encoder/GetEncoderStats", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *encoderClient) StartEncoding(ctx context.Context, in *StartEncodingRequest, opts ...grpc.CallOption) (*StartEncodingReply, error) {
	out := new(StartEncodingReply)
	err := c.cc.Invoke(ctx, "/encoder_grpc.Encoder/StartEncoding", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *encoderClient) EncodingEndedChan(ctx context.Context, in *EncodingEndedChanRequest, opts ...grpc.CallOption) (Encoder_EncodingEndedChanClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Encoder_serviceDesc.Streams[0], "/encoder_grpc.Encoder/EncodingEndedChan", opts...)
	if err != nil {
		return nil, err
	}
	x := &encoderEncodingEndedChanClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Encoder_EncodingEndedChanClient interface {
	Recv() (*EncodingEndedChanReply, error)
	grpc.ClientStream
}

type encoderEncodingEndedChanClient struct {
	grpc.ClientStream
}

func (x *encoderEncodingEndedChanClient) Recv() (*EncodingEndedChanReply, error) {
	m := new(EncodingEndedChanReply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// EncoderServer is the server API for Encoder service.
// All implementations must embed UnimplementedEncoderServer
// for forward compatibility
type EncoderServer interface {
	SetLoggingLevel(context.Context, *SetLoggingLevelRequest) (*SetLoggingLevelReply, error)
	NewInput(context.Context, *NewInputRequest) (*NewInputReply, error)
	NewOutput(context.Context, *NewOutputRequest) (*NewOutputReply, error)
	NewEncoder(context.Context, *NewEncoderRequest) (*NewEncoderReply, error)
	SetEncoderConfig(context.Context, *SetEncoderConfigRequest) (*SetEncoderConfigReply, error)
	CloseInput(context.Context, *CloseInputRequest) (*CloseInputReply, error)
	CloseOutput(context.Context, *CloseOutputRequest) (*CloseOutputReply, error)
	GetEncoderStats(context.Context, *GetEncoderStatsRequest) (*GetEncoderStatsReply, error)
	StartEncoding(context.Context, *StartEncodingRequest) (*StartEncodingReply, error)
	EncodingEndedChan(*EncodingEndedChanRequest, Encoder_EncodingEndedChanServer) error
	mustEmbedUnimplementedEncoderServer()
}

// UnimplementedEncoderServer must be embedded to have forward compatible implementations.
type UnimplementedEncoderServer struct {
}

func (UnimplementedEncoderServer) SetLoggingLevel(context.Context, *SetLoggingLevelRequest) (*SetLoggingLevelReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetLoggingLevel not implemented")
}
func (UnimplementedEncoderServer) NewInput(context.Context, *NewInputRequest) (*NewInputReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NewInput not implemented")
}
func (UnimplementedEncoderServer) NewOutput(context.Context, *NewOutputRequest) (*NewOutputReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NewOutput not implemented")
}
func (UnimplementedEncoderServer) NewEncoder(context.Context, *NewEncoderRequest) (*NewEncoderReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NewEncoder not implemented")
}
func (UnimplementedEncoderServer) SetEncoderConfig(context.Context, *SetEncoderConfigRequest) (*SetEncoderConfigReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetEncoderConfig not implemented")
}
func (UnimplementedEncoderServer) CloseInput(context.Context, *CloseInputRequest) (*CloseInputReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CloseInput not implemented")
}
func (UnimplementedEncoderServer) CloseOutput(context.Context, *CloseOutputRequest) (*CloseOutputReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CloseOutput not implemented")
}
func (UnimplementedEncoderServer) GetEncoderStats(context.Context, *GetEncoderStatsRequest) (*GetEncoderStatsReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetEncoderStats not implemented")
}
func (UnimplementedEncoderServer) StartEncoding(context.Context, *StartEncodingRequest) (*StartEncodingReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StartEncoding not implemented")
}
func (UnimplementedEncoderServer) EncodingEndedChan(*EncodingEndedChanRequest, Encoder_EncodingEndedChanServer) error {
	return status.Errorf(codes.Unimplemented, "method EncodingEndedChan not implemented")
}
func (UnimplementedEncoderServer) mustEmbedUnimplementedEncoderServer() {}

// UnsafeEncoderServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to EncoderServer will
// result in compilation errors.
type UnsafeEncoderServer interface {
	mustEmbedUnimplementedEncoderServer()
}

func RegisterEncoderServer(s *grpc.Server, srv EncoderServer) {
	s.RegisterService(&_Encoder_serviceDesc, srv)
}

func _Encoder_SetLoggingLevel_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetLoggingLevelRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EncoderServer).SetLoggingLevel(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/encoder_grpc.Encoder/SetLoggingLevel",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EncoderServer).SetLoggingLevel(ctx, req.(*SetLoggingLevelRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Encoder_NewInput_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NewInputRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EncoderServer).NewInput(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/encoder_grpc.Encoder/NewInput",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EncoderServer).NewInput(ctx, req.(*NewInputRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Encoder_NewOutput_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NewOutputRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EncoderServer).NewOutput(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/encoder_grpc.Encoder/NewOutput",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EncoderServer).NewOutput(ctx, req.(*NewOutputRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Encoder_NewEncoder_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NewEncoderRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EncoderServer).NewEncoder(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/encoder_grpc.Encoder/NewEncoder",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EncoderServer).NewEncoder(ctx, req.(*NewEncoderRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Encoder_SetEncoderConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetEncoderConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EncoderServer).SetEncoderConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/encoder_grpc.Encoder/SetEncoderConfig",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EncoderServer).SetEncoderConfig(ctx, req.(*SetEncoderConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Encoder_CloseInput_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CloseInputRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EncoderServer).CloseInput(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/encoder_grpc.Encoder/CloseInput",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EncoderServer).CloseInput(ctx, req.(*CloseInputRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Encoder_CloseOutput_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CloseOutputRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EncoderServer).CloseOutput(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/encoder_grpc.Encoder/CloseOutput",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EncoderServer).CloseOutput(ctx, req.(*CloseOutputRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Encoder_GetEncoderStats_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetEncoderStatsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EncoderServer).GetEncoderStats(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/encoder_grpc.Encoder/GetEncoderStats",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EncoderServer).GetEncoderStats(ctx, req.(*GetEncoderStatsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Encoder_StartEncoding_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StartEncodingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EncoderServer).StartEncoding(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/encoder_grpc.Encoder/StartEncoding",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EncoderServer).StartEncoding(ctx, req.(*StartEncodingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Encoder_EncodingEndedChan_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(EncodingEndedChanRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(EncoderServer).EncodingEndedChan(m, &encoderEncodingEndedChanServer{stream})
}

type Encoder_EncodingEndedChanServer interface {
	Send(*EncodingEndedChanReply) error
	grpc.ServerStream
}

type encoderEncodingEndedChanServer struct {
	grpc.ServerStream
}

func (x *encoderEncodingEndedChanServer) Send(m *EncodingEndedChanReply) error {
	return x.ServerStream.SendMsg(m)
}

var _Encoder_serviceDesc = grpc.ServiceDesc{
	ServiceName: "encoder_grpc.Encoder",
	HandlerType: (*EncoderServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SetLoggingLevel",
			Handler:    _Encoder_SetLoggingLevel_Handler,
		},
		{
			MethodName: "NewInput",
			Handler:    _Encoder_NewInput_Handler,
		},
		{
			MethodName: "NewOutput",
			Handler:    _Encoder_NewOutput_Handler,
		},
		{
			MethodName: "NewEncoder",
			Handler:    _Encoder_NewEncoder_Handler,
		},
		{
			MethodName: "SetEncoderConfig",
			Handler:    _Encoder_SetEncoderConfig_Handler,
		},
		{
			MethodName: "CloseInput",
			Handler:    _Encoder_CloseInput_Handler,
		},
		{
			MethodName: "CloseOutput",
			Handler:    _Encoder_CloseOutput_Handler,
		},
		{
			MethodName: "GetEncoderStats",
			Handler:    _Encoder_GetEncoderStats_Handler,
		},
		{
			MethodName: "StartEncoding",
			Handler:    _Encoder_StartEncoding_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "EncodingEndedChan",
			Handler:       _Encoder_EncodingEndedChan_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "encoder.proto",
}
