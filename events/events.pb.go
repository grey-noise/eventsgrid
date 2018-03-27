// Code generated by protoc-gen-go. DO NOT EDIT.
// source: events.proto

/*
Package events is a generated protocol buffer package.

It is generated from these files:
	events.proto

It has these top-level messages:
	Cursor
	Event
	Status
	Acknowledge
*/
package events

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Status_Code int32

const (
	Status_OK         Status_Code = 0
	Status_NOK_CONSUL Status_Code = 1
	Status_NOK_NATS   Status_Code = 2
	Status_NOK_VALID  Status_Code = 3
	Status_NOK_UNKOWN Status_Code = 100
)

var Status_Code_name = map[int32]string{
	0:   "OK",
	1:   "NOK_CONSUL",
	2:   "NOK_NATS",
	3:   "NOK_VALID",
	100: "NOK_UNKOWN",
}
var Status_Code_value = map[string]int32{
	"OK":         0,
	"NOK_CONSUL": 1,
	"NOK_NATS":   2,
	"NOK_VALID":  3,
	"NOK_UNKOWN": 100,
}

func (x Status_Code) String() string {
	return proto.EnumName(Status_Code_name, int32(x))
}
func (Status_Code) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{2, 0} }

type Cursor struct {
	Id uint64 `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	Ts int64  `protobuf:"varint,2,opt,name=ts" json:"ts,omitempty"`
}

func (m *Cursor) Reset()                    { *m = Cursor{} }
func (m *Cursor) String() string            { return proto.CompactTextString(m) }
func (*Cursor) ProtoMessage()               {}
func (*Cursor) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Cursor) GetId() uint64 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *Cursor) GetTs() int64 {
	if m != nil {
		return m.Ts
	}
	return 0
}

type Event struct {
	Cursor  *Cursor `protobuf:"bytes,1,opt,name=cursor" json:"cursor,omitempty"`
	Payload []byte  `protobuf:"bytes,3,opt,name=payload,proto3" json:"payload,omitempty"`
}

func (m *Event) Reset()                    { *m = Event{} }
func (m *Event) String() string            { return proto.CompactTextString(m) }
func (*Event) ProtoMessage()               {}
func (*Event) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Event) GetCursor() *Cursor {
	if m != nil {
		return m.Cursor
	}
	return nil
}

func (m *Event) GetPayload() []byte {
	if m != nil {
		return m.Payload
	}
	return nil
}

type Status struct {
	Code        Status_Code `protobuf:"varint,1,opt,name=code,enum=events.Status_Code" json:"code,omitempty"`
	Description string      `protobuf:"bytes,2,opt,name=description" json:"description,omitempty"`
}

func (m *Status) Reset()                    { *m = Status{} }
func (m *Status) String() string            { return proto.CompactTextString(m) }
func (*Status) ProtoMessage()               {}
func (*Status) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *Status) GetCode() Status_Code {
	if m != nil {
		return m.Code
	}
	return Status_OK
}

func (m *Status) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

type Acknowledge struct {
	Cursor *Cursor `protobuf:"bytes,1,opt,name=cursor" json:"cursor,omitempty"`
	Status *Status `protobuf:"bytes,3,opt,name=status" json:"status,omitempty"`
}

func (m *Acknowledge) Reset()                    { *m = Acknowledge{} }
func (m *Acknowledge) String() string            { return proto.CompactTextString(m) }
func (*Acknowledge) ProtoMessage()               {}
func (*Acknowledge) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *Acknowledge) GetCursor() *Cursor {
	if m != nil {
		return m.Cursor
	}
	return nil
}

func (m *Acknowledge) GetStatus() *Status {
	if m != nil {
		return m.Status
	}
	return nil
}

func init() {
	proto.RegisterType((*Cursor)(nil), "events.Cursor")
	proto.RegisterType((*Event)(nil), "events.Event")
	proto.RegisterType((*Status)(nil), "events.Status")
	proto.RegisterType((*Acknowledge)(nil), "events.Acknowledge")
	proto.RegisterEnum("events.Status_Code", Status_Code_name, Status_Code_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for EventsSender service

type EventsSenderClient interface {
	GetEvents(ctx context.Context, in *Acknowledge, opts ...grpc.CallOption) (EventsSender_GetEventsClient, error)
	SendEvents(ctx context.Context, opts ...grpc.CallOption) (EventsSender_SendEventsClient, error)
}

type eventsSenderClient struct {
	cc *grpc.ClientConn
}

func NewEventsSenderClient(cc *grpc.ClientConn) EventsSenderClient {
	return &eventsSenderClient{cc}
}

func (c *eventsSenderClient) GetEvents(ctx context.Context, in *Acknowledge, opts ...grpc.CallOption) (EventsSender_GetEventsClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_EventsSender_serviceDesc.Streams[0], c.cc, "/events.EventsSender/getEvents", opts...)
	if err != nil {
		return nil, err
	}
	x := &eventsSenderGetEventsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type EventsSender_GetEventsClient interface {
	Recv() (*Event, error)
	grpc.ClientStream
}

type eventsSenderGetEventsClient struct {
	grpc.ClientStream
}

func (x *eventsSenderGetEventsClient) Recv() (*Event, error) {
	m := new(Event)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *eventsSenderClient) SendEvents(ctx context.Context, opts ...grpc.CallOption) (EventsSender_SendEventsClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_EventsSender_serviceDesc.Streams[1], c.cc, "/events.EventsSender/sendEvents", opts...)
	if err != nil {
		return nil, err
	}
	x := &eventsSenderSendEventsClient{stream}
	return x, nil
}

type EventsSender_SendEventsClient interface {
	Send(*Event) error
	Recv() (*Acknowledge, error)
	grpc.ClientStream
}

type eventsSenderSendEventsClient struct {
	grpc.ClientStream
}

func (x *eventsSenderSendEventsClient) Send(m *Event) error {
	return x.ClientStream.SendMsg(m)
}

func (x *eventsSenderSendEventsClient) Recv() (*Acknowledge, error) {
	m := new(Acknowledge)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for EventsSender service

type EventsSenderServer interface {
	GetEvents(*Acknowledge, EventsSender_GetEventsServer) error
	SendEvents(EventsSender_SendEventsServer) error
}

func RegisterEventsSenderServer(s *grpc.Server, srv EventsSenderServer) {
	s.RegisterService(&_EventsSender_serviceDesc, srv)
}

func _EventsSender_GetEvents_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(Acknowledge)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(EventsSenderServer).GetEvents(m, &eventsSenderGetEventsServer{stream})
}

type EventsSender_GetEventsServer interface {
	Send(*Event) error
	grpc.ServerStream
}

type eventsSenderGetEventsServer struct {
	grpc.ServerStream
}

func (x *eventsSenderGetEventsServer) Send(m *Event) error {
	return x.ServerStream.SendMsg(m)
}

func _EventsSender_SendEvents_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(EventsSenderServer).SendEvents(&eventsSenderSendEventsServer{stream})
}

type EventsSender_SendEventsServer interface {
	Send(*Acknowledge) error
	Recv() (*Event, error)
	grpc.ServerStream
}

type eventsSenderSendEventsServer struct {
	grpc.ServerStream
}

func (x *eventsSenderSendEventsServer) Send(m *Acknowledge) error {
	return x.ServerStream.SendMsg(m)
}

func (x *eventsSenderSendEventsServer) Recv() (*Event, error) {
	m := new(Event)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _EventsSender_serviceDesc = grpc.ServiceDesc{
	ServiceName: "events.EventsSender",
	HandlerType: (*EventsSenderServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "getEvents",
			Handler:       _EventsSender_GetEvents_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "sendEvents",
			Handler:       _EventsSender_SendEvents_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "events.proto",
}

func init() { proto.RegisterFile("events.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 330 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x92, 0x41, 0x6b, 0xea, 0x40,
	0x14, 0x85, 0x9d, 0xe8, 0x9b, 0xf7, 0xbc, 0xc6, 0x10, 0xee, 0xdb, 0x04, 0x57, 0x61, 0x16, 0xbe,
	0xac, 0x44, 0x14, 0xde, 0x5e, 0x6c, 0x17, 0x12, 0x49, 0x60, 0x52, 0xdb, 0x55, 0x29, 0x36, 0x33,
	0x48, 0xa8, 0x64, 0x24, 0x33, 0xb6, 0x94, 0xfe, 0x99, 0xfe, 0xd4, 0x92, 0x49, 0x52, 0x2c, 0x74,
	0xd1, 0xe5, 0xb9, 0xf7, 0xbb, 0x87, 0x73, 0x86, 0x01, 0x57, 0x3e, 0xcb, 0xd2, 0xe8, 0xd9, 0xa9,
	0x52, 0x46, 0x21, 0x6d, 0x14, 0x8b, 0x80, 0xae, 0xcf, 0x95, 0x56, 0x15, 0x7a, 0xe0, 0x14, 0x22,
	0x20, 0x21, 0x89, 0x06, 0xdc, 0x29, 0x44, 0xad, 0x8d, 0x0e, 0x9c, 0x90, 0x44, 0x7d, 0xee, 0x18,
	0xcd, 0x36, 0xf0, 0xeb, 0xba, 0xbe, 0xc1, 0x29, 0xd0, 0xdc, 0x9e, 0x58, 0x78, 0xb4, 0xf0, 0x66,
	0xad, 0x73, 0x63, 0xc4, 0xdb, 0x2d, 0x06, 0xf0, 0xfb, 0xb4, 0x7f, 0x3d, 0xaa, 0xbd, 0x08, 0xfa,
	0x21, 0x89, 0x5c, 0xde, 0x49, 0xf6, 0x4e, 0x80, 0x66, 0x66, 0x6f, 0xce, 0x1a, 0xff, 0xc1, 0x20,
	0x57, 0x42, 0x5a, 0x2b, 0x6f, 0xf1, 0xb7, 0xb3, 0x6a, 0xb6, 0xb3, 0xb5, 0x12, 0x92, 0x5b, 0x00,
	0x43, 0x18, 0x09, 0xa9, 0xf3, 0xaa, 0x38, 0x99, 0x42, 0x95, 0x36, 0xd7, 0x90, 0x5f, 0x8e, 0x58,
	0x0c, 0x83, 0x9a, 0x47, 0x0a, 0x4e, 0x1a, 0xfb, 0x3d, 0xf4, 0x00, 0x92, 0x34, 0x7e, 0x58, 0xa7,
	0x49, 0xb6, 0xdb, 0xfa, 0x04, 0x5d, 0xf8, 0x53, 0xeb, 0x64, 0x75, 0x93, 0xf9, 0x0e, 0x8e, 0x61,
	0x58, 0xab, 0xdb, 0xd5, 0x76, 0x73, 0xe5, 0xf7, 0x3b, 0x78, 0x97, 0xc4, 0xe9, 0x5d, 0xe2, 0x0b,
	0x76, 0x0f, 0xa3, 0x55, 0xfe, 0x54, 0xaa, 0x97, 0xa3, 0x14, 0x07, 0xf9, 0xe3, 0xce, 0x53, 0xa0,
	0xda, 0x46, 0xb7, 0x95, 0x2f, 0xb8, 0xa6, 0x10, 0x6f, 0xb7, 0x8b, 0x37, 0x70, 0xed, 0x63, 0xea,
	0x4c, 0x96, 0x42, 0x56, 0xb8, 0x84, 0xe1, 0x41, 0x9a, 0x66, 0x84, 0x9f, 0xaf, 0x70, 0x91, 0x60,
	0x32, 0xee, 0x86, 0x16, 0x62, 0xbd, 0x39, 0xc1, 0xff, 0x00, 0x5a, 0x96, 0xa2, 0xbd, 0xfa, 0x0a,
	0x4c, 0xbe, 0x33, 0x61, 0xbd, 0x88, 0xcc, 0xc9, 0x23, 0xb5, 0x5f, 0x60, 0xf9, 0x11, 0x00, 0x00,
	0xff, 0xff, 0xb2, 0xd0, 0xc3, 0x6d, 0x12, 0x02, 0x00, 0x00,
}
