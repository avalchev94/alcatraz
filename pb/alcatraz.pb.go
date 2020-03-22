// Code generated by protoc-gen-go. DO NOT EDIT.
// source: alcatraz.proto

package pb

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	empty "github.com/golang/protobuf/ptypes/empty"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type UploadRequest struct {
	// Types that are valid to be assigned to TestOneof:
	//	*UploadRequest_Chunk
	//	*UploadRequest_Name
	//	*UploadRequest_Hash
	TestOneof            isUploadRequest_TestOneof `protobuf_oneof:"test_oneof"`
	XXX_NoUnkeyedLiteral struct{}                  `json:"-"`
	XXX_unrecognized     []byte                    `json:"-"`
	XXX_sizecache        int32                     `json:"-"`
}

func (m *UploadRequest) Reset()         { *m = UploadRequest{} }
func (m *UploadRequest) String() string { return proto.CompactTextString(m) }
func (*UploadRequest) ProtoMessage()    {}
func (*UploadRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_73847c5369340d2a, []int{0}
}

func (m *UploadRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UploadRequest.Unmarshal(m, b)
}
func (m *UploadRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UploadRequest.Marshal(b, m, deterministic)
}
func (m *UploadRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UploadRequest.Merge(m, src)
}
func (m *UploadRequest) XXX_Size() int {
	return xxx_messageInfo_UploadRequest.Size(m)
}
func (m *UploadRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_UploadRequest.DiscardUnknown(m)
}

var xxx_messageInfo_UploadRequest proto.InternalMessageInfo

type isUploadRequest_TestOneof interface {
	isUploadRequest_TestOneof()
}

type UploadRequest_Chunk struct {
	Chunk []byte `protobuf:"bytes,1,opt,name=chunk,proto3,oneof"`
}

type UploadRequest_Name struct {
	Name string `protobuf:"bytes,2,opt,name=name,proto3,oneof"`
}

type UploadRequest_Hash struct {
	Hash string `protobuf:"bytes,3,opt,name=hash,proto3,oneof"`
}

func (*UploadRequest_Chunk) isUploadRequest_TestOneof() {}

func (*UploadRequest_Name) isUploadRequest_TestOneof() {}

func (*UploadRequest_Hash) isUploadRequest_TestOneof() {}

func (m *UploadRequest) GetTestOneof() isUploadRequest_TestOneof {
	if m != nil {
		return m.TestOneof
	}
	return nil
}

func (m *UploadRequest) GetChunk() []byte {
	if x, ok := m.GetTestOneof().(*UploadRequest_Chunk); ok {
		return x.Chunk
	}
	return nil
}

func (m *UploadRequest) GetName() string {
	if x, ok := m.GetTestOneof().(*UploadRequest_Name); ok {
		return x.Name
	}
	return ""
}

func (m *UploadRequest) GetHash() string {
	if x, ok := m.GetTestOneof().(*UploadRequest_Hash); ok {
		return x.Hash
	}
	return ""
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*UploadRequest) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*UploadRequest_Chunk)(nil),
		(*UploadRequest_Name)(nil),
		(*UploadRequest_Hash)(nil),
	}
}

func init() {
	proto.RegisterType((*UploadRequest)(nil), "pb.UploadRequest")
}

func init() {
	proto.RegisterFile("alcatraz.proto", fileDescriptor_73847c5369340d2a)
}

var fileDescriptor_73847c5369340d2a = []byte{
	// 194 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x4b, 0xcc, 0x49, 0x4e,
	0x2c, 0x29, 0x4a, 0xac, 0xd2, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x2a, 0x48, 0x92, 0x92,
	0x4e, 0xcf, 0xcf, 0x4f, 0xcf, 0x49, 0xd5, 0x07, 0x8b, 0x24, 0x95, 0xa6, 0xe9, 0xa7, 0xe6, 0x16,
	0x94, 0x54, 0x42, 0x14, 0x28, 0x25, 0x72, 0xf1, 0x86, 0x16, 0xe4, 0xe4, 0x27, 0xa6, 0x04, 0xa5,
	0x16, 0x96, 0xa6, 0x16, 0x97, 0x08, 0x89, 0x71, 0xb1, 0x26, 0x67, 0x94, 0xe6, 0x65, 0x4b, 0x30,
	0x2a, 0x30, 0x6a, 0xf0, 0x78, 0x30, 0x04, 0x41, 0xb8, 0x42, 0x22, 0x5c, 0x2c, 0x79, 0x89, 0xb9,
	0xa9, 0x12, 0x4c, 0x0a, 0x8c, 0x1a, 0x9c, 0x1e, 0x0c, 0x41, 0x60, 0x1e, 0x48, 0x34, 0x23, 0xb1,
	0x38, 0x43, 0x82, 0x19, 0x26, 0x0a, 0xe2, 0x39, 0xf1, 0x70, 0x71, 0x95, 0xa4, 0x16, 0x97, 0xc4,
	0xe7, 0xe7, 0xa5, 0xe6, 0xa7, 0x19, 0xb9, 0x73, 0x71, 0x38, 0x42, 0x5d, 0x25, 0x64, 0xcd, 0xc5,
	0x05, 0xb1, 0xce, 0x2d, 0x33, 0x27, 0x55, 0x48, 0x50, 0xaf, 0x20, 0x49, 0x0f, 0xc5, 0x7a, 0x29,
	0x31, 0x3d, 0x88, 0x6b, 0xf5, 0x60, 0xae, 0xd5, 0x73, 0x05, 0xb9, 0x56, 0x89, 0x41, 0x83, 0x31,
	0x89, 0x0d, 0x2c, 0x66, 0x0c, 0x08, 0x00, 0x00, 0xff, 0xff, 0x5e, 0x08, 0x67, 0xff, 0xe5, 0x00,
	0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// AlcatrazClient is the client API for Alcatraz service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type AlcatrazClient interface {
	UploadFile(ctx context.Context, opts ...grpc.CallOption) (Alcatraz_UploadFileClient, error)
}

type alcatrazClient struct {
	cc grpc.ClientConnInterface
}

func NewAlcatrazClient(cc grpc.ClientConnInterface) AlcatrazClient {
	return &alcatrazClient{cc}
}

func (c *alcatrazClient) UploadFile(ctx context.Context, opts ...grpc.CallOption) (Alcatraz_UploadFileClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Alcatraz_serviceDesc.Streams[0], "/pb.Alcatraz/UploadFile", opts...)
	if err != nil {
		return nil, err
	}
	x := &alcatrazUploadFileClient{stream}
	return x, nil
}

type Alcatraz_UploadFileClient interface {
	Send(*UploadRequest) error
	CloseAndRecv() (*empty.Empty, error)
	grpc.ClientStream
}

type alcatrazUploadFileClient struct {
	grpc.ClientStream
}

func (x *alcatrazUploadFileClient) Send(m *UploadRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *alcatrazUploadFileClient) CloseAndRecv() (*empty.Empty, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(empty.Empty)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// AlcatrazServer is the server API for Alcatraz service.
type AlcatrazServer interface {
	UploadFile(Alcatraz_UploadFileServer) error
}

// UnimplementedAlcatrazServer can be embedded to have forward compatible implementations.
type UnimplementedAlcatrazServer struct {
}

func (*UnimplementedAlcatrazServer) UploadFile(srv Alcatraz_UploadFileServer) error {
	return status.Errorf(codes.Unimplemented, "method UploadFile not implemented")
}

func RegisterAlcatrazServer(s *grpc.Server, srv AlcatrazServer) {
	s.RegisterService(&_Alcatraz_serviceDesc, srv)
}

func _Alcatraz_UploadFile_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(AlcatrazServer).UploadFile(&alcatrazUploadFileServer{stream})
}

type Alcatraz_UploadFileServer interface {
	SendAndClose(*empty.Empty) error
	Recv() (*UploadRequest, error)
	grpc.ServerStream
}

type alcatrazUploadFileServer struct {
	grpc.ServerStream
}

func (x *alcatrazUploadFileServer) SendAndClose(m *empty.Empty) error {
	return x.ServerStream.SendMsg(m)
}

func (x *alcatrazUploadFileServer) Recv() (*UploadRequest, error) {
	m := new(UploadRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _Alcatraz_serviceDesc = grpc.ServiceDesc{
	ServiceName: "pb.Alcatraz",
	HandlerType: (*AlcatrazServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "UploadFile",
			Handler:       _Alcatraz_UploadFile_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "alcatraz.proto",
}
