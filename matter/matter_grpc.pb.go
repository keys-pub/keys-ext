// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package matter

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// MatterClient is the client API for Matter service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MatterClient interface {
	Login(ctx context.Context, in *LoginRequest, opts ...grpc.CallOption) (*LoginResponse, error)
	TeamsForUser(ctx context.Context, in *TeamsForUserRequest, opts ...grpc.CallOption) (*TeamsForUserResponse, error)
	ChannelsForUser(ctx context.Context, in *ChannelsForUserRequest, opts ...grpc.CallOption) (*ChannelsForUserResponse, error)
	CreateChannel(ctx context.Context, in *CreateChannelRequest, opts ...grpc.CallOption) (*CreateChannelResponse, error)
	Listen(ctx context.Context, opts ...grpc.CallOption) (Matter_ListenClient, error)
}

type matterClient struct {
	cc grpc.ClientConnInterface
}

func NewMatterClient(cc grpc.ClientConnInterface) MatterClient {
	return &matterClient{cc}
}

func (c *matterClient) Login(ctx context.Context, in *LoginRequest, opts ...grpc.CallOption) (*LoginResponse, error) {
	out := new(LoginResponse)
	err := c.cc.Invoke(ctx, "/matter.Matter/Login", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *matterClient) TeamsForUser(ctx context.Context, in *TeamsForUserRequest, opts ...grpc.CallOption) (*TeamsForUserResponse, error) {
	out := new(TeamsForUserResponse)
	err := c.cc.Invoke(ctx, "/matter.Matter/TeamsForUser", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *matterClient) ChannelsForUser(ctx context.Context, in *ChannelsForUserRequest, opts ...grpc.CallOption) (*ChannelsForUserResponse, error) {
	out := new(ChannelsForUserResponse)
	err := c.cc.Invoke(ctx, "/matter.Matter/ChannelsForUser", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *matterClient) CreateChannel(ctx context.Context, in *CreateChannelRequest, opts ...grpc.CallOption) (*CreateChannelResponse, error) {
	out := new(CreateChannelResponse)
	err := c.cc.Invoke(ctx, "/matter.Matter/CreateChannel", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *matterClient) Listen(ctx context.Context, opts ...grpc.CallOption) (Matter_ListenClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Matter_serviceDesc.Streams[0], "/matter.Matter/Listen", opts...)
	if err != nil {
		return nil, err
	}
	x := &matterListenClient{stream}
	return x, nil
}

type Matter_ListenClient interface {
	Send(*ListenRequest) error
	Recv() (*ListenEvent, error)
	grpc.ClientStream
}

type matterListenClient struct {
	grpc.ClientStream
}

func (x *matterListenClient) Send(m *ListenRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *matterListenClient) Recv() (*ListenEvent, error) {
	m := new(ListenEvent)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// MatterServer is the server API for Matter service.
// All implementations must embed UnimplementedMatterServer
// for forward compatibility
type MatterServer interface {
	Login(context.Context, *LoginRequest) (*LoginResponse, error)
	TeamsForUser(context.Context, *TeamsForUserRequest) (*TeamsForUserResponse, error)
	ChannelsForUser(context.Context, *ChannelsForUserRequest) (*ChannelsForUserResponse, error)
	CreateChannel(context.Context, *CreateChannelRequest) (*CreateChannelResponse, error)
	Listen(Matter_ListenServer) error
	mustEmbedUnimplementedMatterServer()
}

// UnimplementedMatterServer must be embedded to have forward compatible implementations.
type UnimplementedMatterServer struct {
}

func (*UnimplementedMatterServer) Login(context.Context, *LoginRequest) (*LoginResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Login not implemented")
}
func (*UnimplementedMatterServer) TeamsForUser(context.Context, *TeamsForUserRequest) (*TeamsForUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TeamsForUser not implemented")
}
func (*UnimplementedMatterServer) ChannelsForUser(context.Context, *ChannelsForUserRequest) (*ChannelsForUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChannelsForUser not implemented")
}
func (*UnimplementedMatterServer) CreateChannel(context.Context, *CreateChannelRequest) (*CreateChannelResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateChannel not implemented")
}
func (*UnimplementedMatterServer) Listen(Matter_ListenServer) error {
	return status.Errorf(codes.Unimplemented, "method Listen not implemented")
}
func (*UnimplementedMatterServer) mustEmbedUnimplementedMatterServer() {}

func RegisterMatterServer(s *grpc.Server, srv MatterServer) {
	s.RegisterService(&_Matter_serviceDesc, srv)
}

func _Matter_Login_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LoginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MatterServer).Login(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/matter.Matter/Login",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MatterServer).Login(ctx, req.(*LoginRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Matter_TeamsForUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TeamsForUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MatterServer).TeamsForUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/matter.Matter/TeamsForUser",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MatterServer).TeamsForUser(ctx, req.(*TeamsForUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Matter_ChannelsForUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ChannelsForUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MatterServer).ChannelsForUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/matter.Matter/ChannelsForUser",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MatterServer).ChannelsForUser(ctx, req.(*ChannelsForUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Matter_CreateChannel_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateChannelRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MatterServer).CreateChannel(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/matter.Matter/CreateChannel",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MatterServer).CreateChannel(ctx, req.(*CreateChannelRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Matter_Listen_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(MatterServer).Listen(&matterListenServer{stream})
}

type Matter_ListenServer interface {
	Send(*ListenEvent) error
	Recv() (*ListenRequest, error)
	grpc.ServerStream
}

type matterListenServer struct {
	grpc.ServerStream
}

func (x *matterListenServer) Send(m *ListenEvent) error {
	return x.ServerStream.SendMsg(m)
}

func (x *matterListenServer) Recv() (*ListenRequest, error) {
	m := new(ListenRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _Matter_serviceDesc = grpc.ServiceDesc{
	ServiceName: "matter.Matter",
	HandlerType: (*MatterServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Login",
			Handler:    _Matter_Login_Handler,
		},
		{
			MethodName: "TeamsForUser",
			Handler:    _Matter_TeamsForUser_Handler,
		},
		{
			MethodName: "ChannelsForUser",
			Handler:    _Matter_ChannelsForUser_Handler,
		},
		{
			MethodName: "CreateChannel",
			Handler:    _Matter_CreateChannel_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Listen",
			Handler:       _Matter_Listen_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "matter.proto",
}
