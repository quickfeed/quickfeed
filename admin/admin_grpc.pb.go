// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.19.4
// source: admin/admin.proto

package admin

import (
	context "context"
	ag "github.com/autograde/quickfeed/ag"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// AdminServiceClient is the client API for AdminService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AdminServiceClient interface {
	GetUser(ctx context.Context, in *ag.Void, opts ...grpc.CallOption) (*ag.User, error)
	GetUsers(ctx context.Context, in *ag.Void, opts ...grpc.CallOption) (*ag.Users, error)
	UpdateUser(ctx context.Context, in *ag.User, opts ...grpc.CallOption) (*ag.Void, error)
	CreateCourse(ctx context.Context, in *ag.Course, opts ...grpc.CallOption) (*ag.Course, error)
	UpdateCourse(ctx context.Context, in *ag.Course, opts ...grpc.CallOption) (*ag.Void, error)
	GetOrganization(ctx context.Context, in *OrgRequest, opts ...grpc.CallOption) (*Organization, error)
}

type adminServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewAdminServiceClient(cc grpc.ClientConnInterface) AdminServiceClient {
	return &adminServiceClient{cc}
}

func (c *adminServiceClient) GetUser(ctx context.Context, in *ag.Void, opts ...grpc.CallOption) (*ag.User, error) {
	out := new(ag.User)
	err := c.cc.Invoke(ctx, "/admin.AdminService/GetUser", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminServiceClient) GetUsers(ctx context.Context, in *ag.Void, opts ...grpc.CallOption) (*ag.Users, error) {
	out := new(ag.Users)
	err := c.cc.Invoke(ctx, "/admin.AdminService/GetUsers", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminServiceClient) UpdateUser(ctx context.Context, in *ag.User, opts ...grpc.CallOption) (*ag.Void, error) {
	out := new(ag.Void)
	err := c.cc.Invoke(ctx, "/admin.AdminService/UpdateUser", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminServiceClient) CreateCourse(ctx context.Context, in *ag.Course, opts ...grpc.CallOption) (*ag.Course, error) {
	out := new(ag.Course)
	err := c.cc.Invoke(ctx, "/admin.AdminService/CreateCourse", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminServiceClient) UpdateCourse(ctx context.Context, in *ag.Course, opts ...grpc.CallOption) (*ag.Void, error) {
	out := new(ag.Void)
	err := c.cc.Invoke(ctx, "/admin.AdminService/UpdateCourse", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminServiceClient) GetOrganization(ctx context.Context, in *OrgRequest, opts ...grpc.CallOption) (*Organization, error) {
	out := new(Organization)
	err := c.cc.Invoke(ctx, "/admin.AdminService/GetOrganization", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AdminServiceServer is the server API for AdminService service.
// All implementations must embed UnimplementedAdminServiceServer
// for forward compatibility
type AdminServiceServer interface {
	GetUser(context.Context, *ag.Void) (*ag.User, error)
	GetUsers(context.Context, *ag.Void) (*ag.Users, error)
	UpdateUser(context.Context, *ag.User) (*ag.Void, error)
	CreateCourse(context.Context, *ag.Course) (*ag.Course, error)
	UpdateCourse(context.Context, *ag.Course) (*ag.Void, error)
	GetOrganization(context.Context, *OrgRequest) (*Organization, error)
	mustEmbedUnimplementedAdminServiceServer()
}

// UnimplementedAdminServiceServer must be embedded to have forward compatible implementations.
type UnimplementedAdminServiceServer struct {
}

func (UnimplementedAdminServiceServer) GetUser(context.Context, *ag.Void) (*ag.User, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUser not implemented")
}
func (UnimplementedAdminServiceServer) GetUsers(context.Context, *ag.Void) (*ag.Users, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUsers not implemented")
}
func (UnimplementedAdminServiceServer) UpdateUser(context.Context, *ag.User) (*ag.Void, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateUser not implemented")
}
func (UnimplementedAdminServiceServer) CreateCourse(context.Context, *ag.Course) (*ag.Course, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateCourse not implemented")
}
func (UnimplementedAdminServiceServer) UpdateCourse(context.Context, *ag.Course) (*ag.Void, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateCourse not implemented")
}
func (UnimplementedAdminServiceServer) GetOrganization(context.Context, *OrgRequest) (*Organization, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetOrganization not implemented")
}
func (UnimplementedAdminServiceServer) mustEmbedUnimplementedAdminServiceServer() {}

// UnsafeAdminServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AdminServiceServer will
// result in compilation errors.
type UnsafeAdminServiceServer interface {
	mustEmbedUnimplementedAdminServiceServer()
}

func RegisterAdminServiceServer(s grpc.ServiceRegistrar, srv AdminServiceServer) {
	s.RegisterService(&AdminService_ServiceDesc, srv)
}

func _AdminService_GetUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ag.Void)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServiceServer).GetUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/admin.AdminService/GetUser",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServiceServer).GetUser(ctx, req.(*ag.Void))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminService_GetUsers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ag.Void)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServiceServer).GetUsers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/admin.AdminService/GetUsers",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServiceServer).GetUsers(ctx, req.(*ag.Void))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminService_UpdateUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ag.User)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServiceServer).UpdateUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/admin.AdminService/UpdateUser",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServiceServer).UpdateUser(ctx, req.(*ag.User))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminService_CreateCourse_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ag.Course)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServiceServer).CreateCourse(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/admin.AdminService/CreateCourse",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServiceServer).CreateCourse(ctx, req.(*ag.Course))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminService_UpdateCourse_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ag.Course)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServiceServer).UpdateCourse(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/admin.AdminService/UpdateCourse",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServiceServer).UpdateCourse(ctx, req.(*ag.Course))
	}
	return interceptor(ctx, in, info, handler)
}

func _AdminService_GetOrganization_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OrgRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServiceServer).GetOrganization(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/admin.AdminService/GetOrganization",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServiceServer).GetOrganization(ctx, req.(*OrgRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// AdminService_ServiceDesc is the grpc.ServiceDesc for AdminService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var AdminService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "admin.AdminService",
	HandlerType: (*AdminServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetUser",
			Handler:    _AdminService_GetUser_Handler,
		},
		{
			MethodName: "GetUsers",
			Handler:    _AdminService_GetUsers_Handler,
		},
		{
			MethodName: "UpdateUser",
			Handler:    _AdminService_UpdateUser_Handler,
		},
		{
			MethodName: "CreateCourse",
			Handler:    _AdminService_CreateCourse_Handler,
		},
		{
			MethodName: "UpdateCourse",
			Handler:    _AdminService_UpdateCourse_Handler,
		},
		{
			MethodName: "GetOrganization",
			Handler:    _AdminService_GetOrganization_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "admin/admin.proto",
}
