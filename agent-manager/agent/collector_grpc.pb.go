// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v3.21.12
// source: collector.proto

package agent

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	CollectorService_RegisterCollector_FullMethodName  = "/agent.CollectorService/RegisterCollector"
	CollectorService_DeleteCollector_FullMethodName    = "/agent.CollectorService/DeleteCollector"
	CollectorService_ListCollector_FullMethodName      = "/agent.CollectorService/ListCollector"
	CollectorService_CollectorStream_FullMethodName    = "/agent.CollectorService/CollectorStream"
	CollectorService_GetCollectorConfig_FullMethodName = "/agent.CollectorService/GetCollectorConfig"
)

// CollectorServiceClient is the client API for CollectorService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CollectorServiceClient interface {
	RegisterCollector(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*AuthResponse, error)
	DeleteCollector(ctx context.Context, in *DeleteRequest, opts ...grpc.CallOption) (*AuthResponse, error)
	ListCollector(ctx context.Context, in *ListRequest, opts ...grpc.CallOption) (*ListCollectorResponse, error)
	CollectorStream(ctx context.Context, opts ...grpc.CallOption) (grpc.BidiStreamingClient[CollectorMessages, CollectorMessages], error)
	GetCollectorConfig(ctx context.Context, in *ConfigRequest, opts ...grpc.CallOption) (*CollectorConfig, error)
}

type collectorServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewCollectorServiceClient(cc grpc.ClientConnInterface) CollectorServiceClient {
	return &collectorServiceClient{cc}
}

func (c *collectorServiceClient) RegisterCollector(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*AuthResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(AuthResponse)
	err := c.cc.Invoke(ctx, CollectorService_RegisterCollector_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *collectorServiceClient) DeleteCollector(ctx context.Context, in *DeleteRequest, opts ...grpc.CallOption) (*AuthResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(AuthResponse)
	err := c.cc.Invoke(ctx, CollectorService_DeleteCollector_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *collectorServiceClient) ListCollector(ctx context.Context, in *ListRequest, opts ...grpc.CallOption) (*ListCollectorResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ListCollectorResponse)
	err := c.cc.Invoke(ctx, CollectorService_ListCollector_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *collectorServiceClient) CollectorStream(ctx context.Context, opts ...grpc.CallOption) (grpc.BidiStreamingClient[CollectorMessages, CollectorMessages], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &CollectorService_ServiceDesc.Streams[0], CollectorService_CollectorStream_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[CollectorMessages, CollectorMessages]{ClientStream: stream}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type CollectorService_CollectorStreamClient = grpc.BidiStreamingClient[CollectorMessages, CollectorMessages]

func (c *collectorServiceClient) GetCollectorConfig(ctx context.Context, in *ConfigRequest, opts ...grpc.CallOption) (*CollectorConfig, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CollectorConfig)
	err := c.cc.Invoke(ctx, CollectorService_GetCollectorConfig_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CollectorServiceServer is the server API for CollectorService service.
// All implementations must embed UnimplementedCollectorServiceServer
// for forward compatibility.
type CollectorServiceServer interface {
	RegisterCollector(context.Context, *RegisterRequest) (*AuthResponse, error)
	DeleteCollector(context.Context, *DeleteRequest) (*AuthResponse, error)
	ListCollector(context.Context, *ListRequest) (*ListCollectorResponse, error)
	CollectorStream(grpc.BidiStreamingServer[CollectorMessages, CollectorMessages]) error
	GetCollectorConfig(context.Context, *ConfigRequest) (*CollectorConfig, error)
	mustEmbedUnimplementedCollectorServiceServer()
}

// UnimplementedCollectorServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedCollectorServiceServer struct{}

func (UnimplementedCollectorServiceServer) RegisterCollector(context.Context, *RegisterRequest) (*AuthResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegisterCollector not implemented")
}
func (UnimplementedCollectorServiceServer) DeleteCollector(context.Context, *DeleteRequest) (*AuthResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteCollector not implemented")
}
func (UnimplementedCollectorServiceServer) ListCollector(context.Context, *ListRequest) (*ListCollectorResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListCollector not implemented")
}
func (UnimplementedCollectorServiceServer) CollectorStream(grpc.BidiStreamingServer[CollectorMessages, CollectorMessages]) error {
	return status.Errorf(codes.Unimplemented, "method CollectorStream not implemented")
}
func (UnimplementedCollectorServiceServer) GetCollectorConfig(context.Context, *ConfigRequest) (*CollectorConfig, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCollectorConfig not implemented")
}
func (UnimplementedCollectorServiceServer) mustEmbedUnimplementedCollectorServiceServer() {}
func (UnimplementedCollectorServiceServer) testEmbeddedByValue()                          {}

// UnsafeCollectorServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CollectorServiceServer will
// result in compilation errors.
type UnsafeCollectorServiceServer interface {
	mustEmbedUnimplementedCollectorServiceServer()
}

func RegisterCollectorServiceServer(s grpc.ServiceRegistrar, srv CollectorServiceServer) {
	// If the following call pancis, it indicates UnimplementedCollectorServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&CollectorService_ServiceDesc, srv)
}

func _CollectorService_RegisterCollector_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CollectorServiceServer).RegisterCollector(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CollectorService_RegisterCollector_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CollectorServiceServer).RegisterCollector(ctx, req.(*RegisterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CollectorService_DeleteCollector_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CollectorServiceServer).DeleteCollector(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CollectorService_DeleteCollector_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CollectorServiceServer).DeleteCollector(ctx, req.(*DeleteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CollectorService_ListCollector_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CollectorServiceServer).ListCollector(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CollectorService_ListCollector_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CollectorServiceServer).ListCollector(ctx, req.(*ListRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CollectorService_CollectorStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(CollectorServiceServer).CollectorStream(&grpc.GenericServerStream[CollectorMessages, CollectorMessages]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type CollectorService_CollectorStreamServer = grpc.BidiStreamingServer[CollectorMessages, CollectorMessages]

func _CollectorService_GetCollectorConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CollectorServiceServer).GetCollectorConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CollectorService_GetCollectorConfig_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CollectorServiceServer).GetCollectorConfig(ctx, req.(*ConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// CollectorService_ServiceDesc is the grpc.ServiceDesc for CollectorService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var CollectorService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "agent.CollectorService",
	HandlerType: (*CollectorServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RegisterCollector",
			Handler:    _CollectorService_RegisterCollector_Handler,
		},
		{
			MethodName: "DeleteCollector",
			Handler:    _CollectorService_DeleteCollector_Handler,
		},
		{
			MethodName: "ListCollector",
			Handler:    _CollectorService_ListCollector_Handler,
		},
		{
			MethodName: "GetCollectorConfig",
			Handler:    _CollectorService_GetCollectorConfig_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "CollectorStream",
			Handler:       _CollectorService_CollectorStream_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "collector.proto",
}

const (
	PanelCollectorService_RegisterCollectorConfig_FullMethodName = "/agent.PanelCollectorService/RegisterCollectorConfig"
)

// PanelCollectorServiceClient is the client API for PanelCollectorService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type PanelCollectorServiceClient interface {
	RegisterCollectorConfig(ctx context.Context, in *CollectorConfig, opts ...grpc.CallOption) (*ConfigKnowledge, error)
}

type panelCollectorServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewPanelCollectorServiceClient(cc grpc.ClientConnInterface) PanelCollectorServiceClient {
	return &panelCollectorServiceClient{cc}
}

func (c *panelCollectorServiceClient) RegisterCollectorConfig(ctx context.Context, in *CollectorConfig, opts ...grpc.CallOption) (*ConfigKnowledge, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ConfigKnowledge)
	err := c.cc.Invoke(ctx, PanelCollectorService_RegisterCollectorConfig_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PanelCollectorServiceServer is the server API for PanelCollectorService service.
// All implementations must embed UnimplementedPanelCollectorServiceServer
// for forward compatibility.
type PanelCollectorServiceServer interface {
	RegisterCollectorConfig(context.Context, *CollectorConfig) (*ConfigKnowledge, error)
	mustEmbedUnimplementedPanelCollectorServiceServer()
}

// UnimplementedPanelCollectorServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedPanelCollectorServiceServer struct{}

func (UnimplementedPanelCollectorServiceServer) RegisterCollectorConfig(context.Context, *CollectorConfig) (*ConfigKnowledge, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegisterCollectorConfig not implemented")
}
func (UnimplementedPanelCollectorServiceServer) mustEmbedUnimplementedPanelCollectorServiceServer() {}
func (UnimplementedPanelCollectorServiceServer) testEmbeddedByValue()                               {}

// UnsafePanelCollectorServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to PanelCollectorServiceServer will
// result in compilation errors.
type UnsafePanelCollectorServiceServer interface {
	mustEmbedUnimplementedPanelCollectorServiceServer()
}

func RegisterPanelCollectorServiceServer(s grpc.ServiceRegistrar, srv PanelCollectorServiceServer) {
	// If the following call pancis, it indicates UnimplementedPanelCollectorServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&PanelCollectorService_ServiceDesc, srv)
}

func _PanelCollectorService_RegisterCollectorConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CollectorConfig)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PanelCollectorServiceServer).RegisterCollectorConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: PanelCollectorService_RegisterCollectorConfig_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PanelCollectorServiceServer).RegisterCollectorConfig(ctx, req.(*CollectorConfig))
	}
	return interceptor(ctx, in, info, handler)
}

// PanelCollectorService_ServiceDesc is the grpc.ServiceDesc for PanelCollectorService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var PanelCollectorService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "agent.PanelCollectorService",
	HandlerType: (*PanelCollectorServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RegisterCollectorConfig",
			Handler:    _PanelCollectorService_RegisterCollectorConfig_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "collector.proto",
}
