// Code generated by protoc-gen-go. DO NOT EDIT.
// source: events.proto

/*
Package events is a generated protocol buffer package.

It is generated from these files
	events.proto

It has these top-level messages:
	Cursor
	Header
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
func (Status_Code) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{3, 0} }

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

type Header struct {
	ClientId    string `protobuf:"bytes,1,opt,name=clientId" json:"clientId,omitempty"`
	GroupId     string `protobuf:"bytes,2,opt,name=groupId" json:"groupId,omitempty"`
	EventTypeId string `protobuf:"bytes,3,opt,name=eventTypeId" json:"eventTypeId,omitempty"`
	Eventid     string `protobuf:"bytes,4,opt,name=eventid" json:"eventid,omitempty"`
	Source      string `protobuf:"bytes,5,opt,name=source" json:"source,omitempty"`
}

func (m *Header) Reset()                    { *m = Header{} }
func (m *Header) String() string            { return proto.CompactTextString(m) }
func (*Header) ProtoMessage()               {}
func (*Header) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Header) GetClientId() string {
	if m != nil {
		return m.ClientId
	}
	return ""
}

func (m *Header) GetGroupId() string {
	if m != nil {
		return m.GroupId
	}
	return ""
}

func (m *Header) GetEventTypeId() string {
	if m != nil {
		return m.EventTypeId
	}
	return ""
}

func (m *Header) GetEventid() string {
	if m != nil {
		return m.Eventid
	}
	return ""
}

func (m *Header) GetSource() string {
	if m != nil {
		return m.Source
	}
	return ""
}

type Event struct {
	Cursor  *Cursor `protobuf:"bytes,1,opt,name=cursor" json:"cursor,omitempty"`
	Header  *Header `protobuf:"bytes,2,opt,name=header" json:"header,omitempty"`
	Payload []byte  `protobuf:"bytes,3,opt,name=payload,proto3" json:"payload,omitempty"`
}

func (m *Event) Reset()                    { *m = Event{} }
func (m *Event) String() string            { return proto.CompactTextString(m) }
func (*Event) ProtoMessage()               {}
func (*Event) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *Event) GetCursor() *Cursor {
	if m != nil {
		return m.Cursor
	}
	return nil
}

func (m *Event) GetHeader() *Header {
	if m != nil {
		return m.Header
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
func (*Status) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

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
	Header *Header `protobuf:"bytes,2,opt,name=header" json:"header,omitempty"`
	Status *Status `protobuf:"bytes,3,opt,name=status" json:"status,omitempty"`
}

func (m *Acknowledge) Reset()                    { *m = Acknowledge{} }
func (m *Acknowledge) String() string            { return proto.CompactTextString(m) }
func (*Acknowledge) ProtoMessage()               {}
func (*Acknowledge) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *Acknowledge) GetCursor() *Cursor {
	if m != nil {
		return m.Cursor
	}
	return nil
}

func (m *Acknowledge) GetHeader() *Header {
	if m != nil {
		return m.Header
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
	proto.RegisterType((*Header)(nil), "events.Header")
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
	// 448 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x53, 0x41, 0x6f, 0xd3, 0x30,
	0x18, 0x9d, 0xd3, 0xce, 0xac, 0x5f, 0xbb, 0x12, 0x19, 0x09, 0x45, 0xe5, 0x52, 0xe5, 0x30, 0x7a,
	0x8a, 0xaa, 0x4e, 0xe2, 0xde, 0x15, 0x04, 0x55, 0xa7, 0x74, 0x72, 0x37, 0x38, 0xa2, 0x10, 0x7f,
	0x0a, 0x11, 0x25, 0x0e, 0xb6, 0x03, 0x4c, 0x9c, 0xf9, 0x03, 0xfc, 0x02, 0x7e, 0x2a, 0xb2, 0x9d,
	0x4c, 0x41, 0xe2, 0xc8, 0x2d, 0xef, 0x7b, 0xef, 0xd9, 0xef, 0x7d, 0x91, 0x61, 0x82, 0x5f, 0xb1,
	0x32, 0x3a, 0xa9, 0x95, 0x34, 0x92, 0x51, 0x8f, 0xe2, 0x05, 0xd0, 0x4d, 0xa3, 0xb4, 0x54, 0x6c,
	0x0a, 0x41, 0x29, 0x22, 0x32, 0x27, 0x8b, 0x21, 0x0f, 0x4a, 0x61, 0xb1, 0xd1, 0x51, 0x30, 0x27,
	0x8b, 0x01, 0x0f, 0x8c, 0x8e, 0x7f, 0x11, 0xa0, 0x6f, 0x30, 0x13, 0xa8, 0xd8, 0x0c, 0xce, 0xf2,
	0x63, 0x89, 0x95, 0xd9, 0x7a, 0xc3, 0x88, 0x3f, 0x60, 0x16, 0xc1, 0xa3, 0x42, 0xc9, 0xa6, 0xde,
	0x0a, 0xe7, 0x1d, 0xf1, 0x0e, 0xb2, 0x39, 0x8c, 0xdd, 0xa5, 0xb7, 0xf7, 0x35, 0x6e, 0x45, 0x34,
	0x70, 0x6c, 0x7f, 0x64, 0xbd, 0x0e, 0x96, 0x22, 0x1a, 0x7a, 0x6f, 0x0b, 0xd9, 0x53, 0xa0, 0x5a,
	0x36, 0x2a, 0xc7, 0xe8, 0xd4, 0x11, 0x2d, 0x8a, 0xbf, 0xc0, 0xe9, 0x2b, 0x2b, 0x61, 0x17, 0x40,
	0x73, 0xd7, 0xc3, 0x05, 0x1a, 0xaf, 0xa6, 0x49, 0x5b, 0xd7, 0xb7, 0xe3, 0x2d, 0x6b, 0x75, 0x1f,
	0x5d, 0x09, 0x97, 0xae, 0xa7, 0xf3, 0xd5, 0x78, 0xcb, 0xda, 0x28, 0x75, 0x76, 0x7f, 0x94, 0x99,
	0x0f, 0x3a, 0xe1, 0x1d, 0x8c, 0x7f, 0x13, 0xa0, 0x07, 0x93, 0x99, 0x46, 0xb3, 0xe7, 0x30, 0xcc,
	0xa5, 0x40, 0x77, 0xe5, 0x74, 0xf5, 0xa4, 0x3b, 0xca, 0xb3, 0xc9, 0x46, 0x0a, 0xe4, 0x4e, 0x60,
	0xab, 0x0b, 0xd4, 0xb9, 0x2a, 0x6b, 0x53, 0xca, 0xaa, 0x5d, 0x4c, 0x7f, 0x14, 0xef, 0x60, 0x68,
	0xf5, 0x8c, 0x42, 0xb0, 0xdf, 0x85, 0x27, 0x6c, 0x0a, 0x90, 0xee, 0x77, 0xef, 0x37, 0xfb, 0xf4,
	0x70, 0x77, 0x1d, 0x12, 0x36, 0x81, 0x33, 0x8b, 0xd3, 0xf5, 0xed, 0x21, 0x0c, 0xd8, 0x39, 0x8c,
	0x2c, 0x7a, 0xbb, 0xbe, 0xde, 0xbe, 0x0c, 0x07, 0x9d, 0xf8, 0x2e, 0xdd, 0xed, 0xdf, 0xa5, 0xa1,
	0x88, 0x7f, 0x12, 0x18, 0xaf, 0xf3, 0x4f, 0x95, 0xfc, 0x76, 0x44, 0x51, 0xe0, 0x7f, 0x5f, 0xce,
	0x05, 0x50, 0xed, 0x3a, 0xba, 0xdd, 0xf4, 0x74, 0xbe, 0x39, 0x6f, 0xd9, 0xd5, 0x0f, 0x98, 0xb8,
	0xbf, 0xa3, 0x0f, 0x58, 0x59, 0xdf, 0x25, 0x8c, 0x0a, 0x34, 0x7e, 0xc4, 0x1e, 0xd6, 0xd5, 0x4b,
	0x3a, 0x3b, 0xef, 0x86, 0x4e, 0x14, 0x9f, 0x2c, 0x09, 0x7b, 0x01, 0xa0, 0xb1, 0x12, 0xad, 0xeb,
	0x6f, 0xc1, 0xec, 0x5f, 0x87, 0xc4, 0x27, 0x0b, 0xb2, 0x24, 0x57, 0x4b, 0x78, 0x56, 0xca, 0xa4,
	0x50, 0x75, 0x9e, 0xe0, 0xf7, 0xec, 0x73, 0x7d, 0x44, 0x9d, 0x28, 0xd9, 0x18, 0x2c, 0x9a, 0x52,
	0xe0, 0xd5, 0x63, 0x6e, 0xbf, 0x5f, 0xdb, 0xef, 0x1b, 0xfb, 0x22, 0x6e, 0xc8, 0x07, 0xea, 0x9e,
	0xc6, 0xe5, 0x9f, 0x00, 0x00, 0x00, 0xff, 0xff, 0xb4, 0xd4, 0x02, 0x2c, 0x2a, 0x03, 0x00, 0x00,
}
