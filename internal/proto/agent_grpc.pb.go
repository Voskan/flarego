// internal/proto/agent.proto
// RPC contract for control‐plane communication *from* the gateway *to* agents.
// In FlareGo v0.1 agents are largely autonomous (they push flamegraphs and do
// not await commands), but exposing this service early allows future features
// such as remote sampler tuning, version upgrade orchestration and health
// checks without breaking backward compatibility.
//
// Design principles:
//   • Keep messages minimal – only what is required for handshake and basic
//     ping/pong.  Fine‑grained control commands can be added with new RPCs or
//     oneof fields in ControlRequest.
//   • Streaming is unidirectional gateway→agent so that the gateway can push
//     config changes instantly while the agent needs only to ACK.
//   • Field numbers are frozen once released.  Reserve ranges for internal
//     use (100‑199) and experimental (200‑299).

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.3
// source: agent.proto

package agentpb

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
	AgentService_Handshake_FullMethodName = "/agentpb.AgentService/Handshake"
)

// AgentServiceClient is the client API for AgentService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// AgentService is implemented by *agents*; gateway acts as the client.
type AgentServiceClient interface {
	// Handshake opens a bidirectional stream: first message **from** agent must
	// be AgentInfo; afterwards gateway can push ControlRequest, agent responds
	// with ControlResponse.  Agent may also periodically send Heartbeat.
	Handshake(ctx context.Context, opts ...grpc.CallOption) (grpc.BidiStreamingClient[AgentEnvelope, ControlRequest], error)
}

type agentServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewAgentServiceClient(cc grpc.ClientConnInterface) AgentServiceClient {
	return &agentServiceClient{cc}
}

func (c *agentServiceClient) Handshake(ctx context.Context, opts ...grpc.CallOption) (grpc.BidiStreamingClient[AgentEnvelope, ControlRequest], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &AgentService_ServiceDesc.Streams[0], AgentService_Handshake_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[AgentEnvelope, ControlRequest]{ClientStream: stream}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type AgentService_HandshakeClient = grpc.BidiStreamingClient[AgentEnvelope, ControlRequest]

// AgentServiceServer is the server API for AgentService service.
// All implementations must embed UnimplementedAgentServiceServer
// for forward compatibility.
//
// AgentService is implemented by *agents*; gateway acts as the client.
type AgentServiceServer interface {
	// Handshake opens a bidirectional stream: first message **from** agent must
	// be AgentInfo; afterwards gateway can push ControlRequest, agent responds
	// with ControlResponse.  Agent may also periodically send Heartbeat.
	Handshake(grpc.BidiStreamingServer[AgentEnvelope, ControlRequest]) error
	mustEmbedUnimplementedAgentServiceServer()
}

// UnimplementedAgentServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedAgentServiceServer struct{}

func (UnimplementedAgentServiceServer) Handshake(grpc.BidiStreamingServer[AgentEnvelope, ControlRequest]) error {
	return status.Errorf(codes.Unimplemented, "method Handshake not implemented")
}
func (UnimplementedAgentServiceServer) mustEmbedUnimplementedAgentServiceServer() {}
func (UnimplementedAgentServiceServer) testEmbeddedByValue()                      {}

// UnsafeAgentServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AgentServiceServer will
// result in compilation errors.
type UnsafeAgentServiceServer interface {
	mustEmbedUnimplementedAgentServiceServer()
}

func RegisterAgentServiceServer(s grpc.ServiceRegistrar, srv AgentServiceServer) {
	// If the following call pancis, it indicates UnimplementedAgentServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&AgentService_ServiceDesc, srv)
}

func _AgentService_Handshake_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(AgentServiceServer).Handshake(&grpc.GenericServerStream[AgentEnvelope, ControlRequest]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type AgentService_HandshakeServer = grpc.BidiStreamingServer[AgentEnvelope, ControlRequest]

// AgentService_ServiceDesc is the grpc.ServiceDesc for AgentService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var AgentService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "agentpb.AgentService",
	HandlerType: (*AgentServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Handshake",
			Handler:       _AgentService_Handshake_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "agent.proto",
}
