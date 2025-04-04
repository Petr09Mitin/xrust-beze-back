// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.3
// source: proto/file/file.proto

package file

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
	FileService_MoveTempFileToAvatars_FullMethodName = "/file.FileService/MoveTempFileToAvatars"
	FileService_DeleteAvatar_FullMethodName          = "/file.FileService/DeleteAvatar"
)

// FileServiceClient is the client API for FileService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type FileServiceClient interface {
	MoveTempFileToAvatars(ctx context.Context, in *MoveTempFileToAvatarsRequest, opts ...grpc.CallOption) (*MoveTempFileToAvatarsResponse, error)
	DeleteAvatar(ctx context.Context, in *DeleteAvatarRequest, opts ...grpc.CallOption) (*DeleteAvatarResponse, error)
}

type fileServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewFileServiceClient(cc grpc.ClientConnInterface) FileServiceClient {
	return &fileServiceClient{cc}
}

func (c *fileServiceClient) MoveTempFileToAvatars(ctx context.Context, in *MoveTempFileToAvatarsRequest, opts ...grpc.CallOption) (*MoveTempFileToAvatarsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(MoveTempFileToAvatarsResponse)
	err := c.cc.Invoke(ctx, FileService_MoveTempFileToAvatars_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fileServiceClient) DeleteAvatar(ctx context.Context, in *DeleteAvatarRequest, opts ...grpc.CallOption) (*DeleteAvatarResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(DeleteAvatarResponse)
	err := c.cc.Invoke(ctx, FileService_DeleteAvatar_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FileServiceServer is the server API for FileService service.
// All implementations must embed UnimplementedFileServiceServer
// for forward compatibility.
type FileServiceServer interface {
	MoveTempFileToAvatars(context.Context, *MoveTempFileToAvatarsRequest) (*MoveTempFileToAvatarsResponse, error)
	DeleteAvatar(context.Context, *DeleteAvatarRequest) (*DeleteAvatarResponse, error)
	mustEmbedUnimplementedFileServiceServer()
}

// UnimplementedFileServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedFileServiceServer struct{}

func (UnimplementedFileServiceServer) MoveTempFileToAvatars(context.Context, *MoveTempFileToAvatarsRequest) (*MoveTempFileToAvatarsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MoveTempFileToAvatars not implemented")
}
func (UnimplementedFileServiceServer) DeleteAvatar(context.Context, *DeleteAvatarRequest) (*DeleteAvatarResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteAvatar not implemented")
}
func (UnimplementedFileServiceServer) mustEmbedUnimplementedFileServiceServer() {}
func (UnimplementedFileServiceServer) testEmbeddedByValue()                     {}

// UnsafeFileServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to FileServiceServer will
// result in compilation errors.
type UnsafeFileServiceServer interface {
	mustEmbedUnimplementedFileServiceServer()
}

func RegisterFileServiceServer(s grpc.ServiceRegistrar, srv FileServiceServer) {
	// If the following call pancis, it indicates UnimplementedFileServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&FileService_ServiceDesc, srv)
}

func _FileService_MoveTempFileToAvatars_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MoveTempFileToAvatarsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FileServiceServer).MoveTempFileToAvatars(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FileService_MoveTempFileToAvatars_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FileServiceServer).MoveTempFileToAvatars(ctx, req.(*MoveTempFileToAvatarsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FileService_DeleteAvatar_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteAvatarRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FileServiceServer).DeleteAvatar(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: FileService_DeleteAvatar_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FileServiceServer).DeleteAvatar(ctx, req.(*DeleteAvatarRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// FileService_ServiceDesc is the grpc.ServiceDesc for FileService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var FileService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "file.FileService",
	HandlerType: (*FileServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "MoveTempFileToAvatars",
			Handler:    _FileService_MoveTempFileToAvatars_Handler,
		},
		{
			MethodName: "DeleteAvatar",
			Handler:    _FileService_DeleteAvatar_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/file/file.proto",
}
