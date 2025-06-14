// internal/proto/ui.proto
// This schema defines the gRPC contract between the FlareGo gateway and the UI.
// The protocol is designed for real-time streaming of flamegraph data to the UI.

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.3
// source: ui.proto

package agentpb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	UIService_StreamFlamegraphs_FullMethodName = "/agentpb.UIService/StreamFlamegraphs"
)

// UIServiceClient is the client API for UIService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// UIService is implemented by the gateway; the UI connects to stream
// flamegraph data in real-time.
type UIServiceClient interface {
	// StreamFlamegraphs streams flamegraph data to the UI.
	StreamFlamegraphs(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (grpc.ServerStreamingClient[FlamegraphChunk], error)
}

type uIServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewUIServiceClient(cc grpc.ClientConnInterface) UIServiceClient {
	return &uIServiceClient{cc}
}

func (c *uIServiceClient) StreamFlamegraphs(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (grpc.ServerStreamingClient[FlamegraphChunk], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &UIService_ServiceDesc.Streams[0], UIService_StreamFlamegraphs_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[emptypb.Empty, FlamegraphChunk]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type UIService_StreamFlamegraphsClient = grpc.ServerStreamingClient[FlamegraphChunk]

// UIServiceServer is the server API for UIService service.
// All implementations must embed UnimplementedUIServiceServer
// for forward compatibility.
//
// UIService is implemented by the gateway; the UI connects to stream
// flamegraph data in real-time.
type UIServiceServer interface {
	// StreamFlamegraphs streams flamegraph data to the UI.
	StreamFlamegraphs(*emptypb.Empty, grpc.ServerStreamingServer[FlamegraphChunk]) error
	mustEmbedUnimplementedUIServiceServer()
}

// UnimplementedUIServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedUIServiceServer struct{}

func (UnimplementedUIServiceServer) StreamFlamegraphs(*emptypb.Empty, grpc.ServerStreamingServer[FlamegraphChunk]) error {
	return status.Errorf(codes.Unimplemented, "method StreamFlamegraphs not implemented")
}
func (UnimplementedUIServiceServer) mustEmbedUnimplementedUIServiceServer() {}
func (UnimplementedUIServiceServer) testEmbeddedByValue()                   {}

// UnsafeUIServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to UIServiceServer will
// result in compilation errors.
type UnsafeUIServiceServer interface {
	mustEmbedUnimplementedUIServiceServer()
}

func RegisterUIServiceServer(s grpc.ServiceRegistrar, srv UIServiceServer) {
	// If the following call pancis, it indicates UnimplementedUIServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&UIService_ServiceDesc, srv)
}

func _UIService_StreamFlamegraphs_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(emptypb.Empty)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(UIServiceServer).StreamFlamegraphs(m, &grpc.GenericServerStream[emptypb.Empty, FlamegraphChunk]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type UIService_StreamFlamegraphsServer = grpc.ServerStreamingServer[FlamegraphChunk]

// UIService_ServiceDesc is the grpc.ServiceDesc for UIService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var UIService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "agentpb.UIService",
	HandlerType: (*UIServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamFlamegraphs",
			Handler:       _UIService_StreamFlamegraphs_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "ui.proto",
}
