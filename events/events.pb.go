// Code generated by protoc-gen-go. DO NOT EDIT.
// source: events.proto

/*
Package events is a generated protocol buffer package.

It is generated from these files:
	events.proto

It has these top-level messages:
	Cursor
	Header
	Event
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

type Acknowledge struct {
	Cursor *Cursor `protobuf:"bytes,1,opt,name=cursor" json:"cursor,omitempty"`
	Header *Header `protobuf:"bytes,2,opt,name=header" json:"header,omitempty"`
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

func (m *Acknowledge) GetHeader() *Header {
	if m != nil {
		return m.Header
	}
	return nil
}

func init() {
	proto.RegisterType((*Cursor)(nil), "events.Cursor")
	proto.RegisterType((*Header)(nil), "events.Header")
	proto.RegisterType((*Event)(nil), "events.Event")
	proto.RegisterType((*Acknowledge)(nil), "events.Acknowledge")
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

// Server API for EventsSender service

type EventsSenderServer interface {
	GetEvents(*Acknowledge, EventsSender_GetEventsServer) error
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
	},
	Metadata: "events.proto",
}

func init() { proto.RegisterFile("events.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 311 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x92, 0x41, 0x4e, 0xc3, 0x30,
	0x10, 0x45, 0x49, 0xda, 0x06, 0x3a, 0x2d, 0x45, 0x32, 0x12, 0x8a, 0xca, 0xa6, 0xca, 0x02, 0x75,
	0x15, 0x55, 0xed, 0x09, 0x68, 0x85, 0xa0, 0xbb, 0xca, 0xb0, 0x65, 0x11, 0xe2, 0x51, 0x88, 0x08,
	0xb1, 0xb1, 0x1d, 0xa0, 0xd7, 0xe0, 0xc4, 0xc8, 0xe3, 0x04, 0xe5, 0x00, 0xec, 0xfc, 0x66, 0xfe,
	0x97, 0xff, 0x97, 0x06, 0xa6, 0xf8, 0x89, 0xb5, 0x35, 0xa9, 0xd2, 0xd2, 0x4a, 0x16, 0x79, 0x4a,
	0x96, 0x10, 0xed, 0x1a, 0x6d, 0xa4, 0x66, 0x33, 0x08, 0x4b, 0x11, 0x07, 0x8b, 0x60, 0x39, 0xe4,
	0x61, 0x29, 0x1c, 0x5b, 0x13, 0x87, 0x8b, 0x60, 0x39, 0xe0, 0xa1, 0x35, 0xc9, 0x4f, 0x00, 0xd1,
	0x03, 0x66, 0x02, 0x35, 0x9b, 0xc3, 0x59, 0x5e, 0x95, 0x58, 0xdb, 0xbd, 0x37, 0x8c, 0xf9, 0x1f,
	0xb3, 0x18, 0x4e, 0x0b, 0x2d, 0x1b, 0xb5, 0x17, 0xe4, 0x1d, 0xf3, 0x0e, 0xd9, 0x02, 0x26, 0xf4,
	0xe9, 0xd3, 0x51, 0xe1, 0x5e, 0xc4, 0x03, 0xda, 0xf6, 0x47, 0xce, 0x4b, 0x58, 0x8a, 0x78, 0xe8,
	0xbd, 0x2d, 0xb2, 0x2b, 0x88, 0x8c, 0x6c, 0x74, 0x8e, 0xf1, 0x88, 0x16, 0x2d, 0x25, 0x1f, 0x30,
	0xba, 0x73, 0x12, 0x76, 0x03, 0x51, 0x4e, 0x3d, 0x28, 0xd0, 0x64, 0x3d, 0x4b, 0xdb, 0xba, 0xbe,
	0x1d, 0x6f, 0xb7, 0x4e, 0xf7, 0x4a, 0x25, 0x28, 0x5d, 0x4f, 0xe7, 0xab, 0xf1, 0x76, 0xeb, 0xa2,
	0xa8, 0xec, 0x58, 0xc9, 0xcc, 0x07, 0x9d, 0xf2, 0x0e, 0x93, 0x67, 0x98, 0xdc, 0xe6, 0x6f, 0xb5,
	0xfc, 0xaa, 0x50, 0x14, 0xf8, 0xdf, 0x1f, 0xaf, 0x77, 0x30, 0xa5, 0x46, 0xe6, 0x11, 0x6b, 0x17,
	0x64, 0x03, 0xe3, 0x02, 0xad, 0x1f, 0xb1, 0xcb, 0xce, 0xd4, 0x4b, 0x30, 0x3f, 0xef, 0x86, 0x24,
	0x4a, 0x4e, 0x56, 0xc1, 0x76, 0x05, 0xd7, 0xa5, 0x4c, 0x0b, 0xad, 0xf2, 0x14, 0xbf, 0xb3, 0x77,
	0x55, 0xa1, 0x49, 0xb5, 0x6c, 0x2c, 0x16, 0x4d, 0x29, 0x70, 0x7b, 0xc1, 0xdd, 0xfb, 0xde, 0xbd,
	0x0f, 0xee, 0x1a, 0x0e, 0xc1, 0x4b, 0x44, 0x67, 0xb1, 0xf9, 0x0d, 0x00, 0x00, 0xff, 0xff, 0x5b,
	0xf0, 0x39, 0xdf, 0x26, 0x02, 0x00, 0x00,
}
