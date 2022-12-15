// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1-devel
// 	protoc        v3.6.1
// source: mpc.proto

package mpcclient

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type CheckResultResponse_REQUEST_TYPE int32

const (
	CheckResultResponse_UNKNOWN_TYPE CheckResultResponse_REQUEST_TYPE = 0
	CheckResultResponse_KEYGEN       CheckResultResponse_REQUEST_TYPE = 1
	CheckResultResponse_SIGN         CheckResultResponse_REQUEST_TYPE = 2
)

// Enum value maps for CheckResultResponse_REQUEST_TYPE.
var (
	CheckResultResponse_REQUEST_TYPE_name = map[int32]string{
		0: "UNKNOWN_TYPE",
		1: "KEYGEN",
		2: "SIGN",
	}
	CheckResultResponse_REQUEST_TYPE_value = map[string]int32{
		"UNKNOWN_TYPE": 0,
		"KEYGEN":       1,
		"SIGN":         2,
	}
)

func (x CheckResultResponse_REQUEST_TYPE) Enum() *CheckResultResponse_REQUEST_TYPE {
	p := new(CheckResultResponse_REQUEST_TYPE)
	*p = x
	return p
}

func (x CheckResultResponse_REQUEST_TYPE) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (CheckResultResponse_REQUEST_TYPE) Descriptor() protoreflect.EnumDescriptor {
	return file_mpc_proto_enumTypes[0].Descriptor()
}

func (CheckResultResponse_REQUEST_TYPE) Type() protoreflect.EnumType {
	return &file_mpc_proto_enumTypes[0]
}

func (x CheckResultResponse_REQUEST_TYPE) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use CheckResultResponse_REQUEST_TYPE.Descriptor instead.
func (CheckResultResponse_REQUEST_TYPE) EnumDescriptor() ([]byte, []int) {
	return file_mpc_proto_rawDescGZIP(), []int{5, 0}
}

type CheckResultResponse_REQUEST_STATUS int32

const (
	CheckResultResponse_UNKNOWN_STATUS CheckResultResponse_REQUEST_STATUS = 0
	CheckResultResponse_RECEIVED       CheckResultResponse_REQUEST_STATUS = 1
	CheckResultResponse_PROCESSING     CheckResultResponse_REQUEST_STATUS = 2
	CheckResultResponse_DONE           CheckResultResponse_REQUEST_STATUS = 3
)

// Enum value maps for CheckResultResponse_REQUEST_STATUS.
var (
	CheckResultResponse_REQUEST_STATUS_name = map[int32]string{
		0: "UNKNOWN_STATUS",
		1: "RECEIVED",
		2: "PROCESSING",
		3: "DONE",
	}
	CheckResultResponse_REQUEST_STATUS_value = map[string]int32{
		"UNKNOWN_STATUS": 0,
		"RECEIVED":       1,
		"PROCESSING":     2,
		"DONE":           3,
	}
)

func (x CheckResultResponse_REQUEST_STATUS) Enum() *CheckResultResponse_REQUEST_STATUS {
	p := new(CheckResultResponse_REQUEST_STATUS)
	*p = x
	return p
}

func (x CheckResultResponse_REQUEST_STATUS) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (CheckResultResponse_REQUEST_STATUS) Descriptor() protoreflect.EnumDescriptor {
	return file_mpc_proto_enumTypes[1].Descriptor()
}

func (CheckResultResponse_REQUEST_STATUS) Type() protoreflect.EnumType {
	return &file_mpc_proto_enumTypes[1]
}

func (x CheckResultResponse_REQUEST_STATUS) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use CheckResultResponse_REQUEST_STATUS.Descriptor instead.
func (CheckResultResponse_REQUEST_STATUS) EnumDescriptor() ([]byte, []int) {
	return file_mpc_proto_rawDescGZIP(), []int{5, 1}
}

type KeygenRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RequestId             string   `protobuf:"bytes,1,opt,name=request_id,json=requestId,proto3" json:"request_id,omitempty"`
	ParticipantPublicKeys [][]byte `protobuf:"bytes,2,rep,name=participant_public_keys,json=participantPublicKeys,proto3" json:"participant_public_keys,omitempty"`
	Threshold             uint32   `protobuf:"varint,3,opt,name=threshold,proto3" json:"threshold,omitempty"`
}

func (x *KeygenRequest) Reset() {
	*x = KeygenRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mpc_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KeygenRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KeygenRequest) ProtoMessage() {}

func (x *KeygenRequest) ProtoReflect() protoreflect.Message {
	mi := &file_mpc_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KeygenRequest.ProtoReflect.Descriptor instead.
func (*KeygenRequest) Descriptor() ([]byte, []int) {
	return file_mpc_proto_rawDescGZIP(), []int{0}
}

func (x *KeygenRequest) GetRequestId() string {
	if x != nil {
		return x.RequestId
	}
	return ""
}

func (x *KeygenRequest) GetParticipantPublicKeys() [][]byte {
	if x != nil {
		return x.ParticipantPublicKeys
	}
	return nil
}

func (x *KeygenRequest) GetThreshold() uint32 {
	if x != nil {
		return x.Threshold
	}
	return 0
}

type KeygenResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RequestId string `protobuf:"bytes,1,opt,name=request_id,json=requestId,proto3" json:"request_id,omitempty"`
}

func (x *KeygenResponse) Reset() {
	*x = KeygenResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mpc_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KeygenResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KeygenResponse) ProtoMessage() {}

func (x *KeygenResponse) ProtoReflect() protoreflect.Message {
	mi := &file_mpc_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KeygenResponse.ProtoReflect.Descriptor instead.
func (*KeygenResponse) Descriptor() ([]byte, []int) {
	return file_mpc_proto_rawDescGZIP(), []int{1}
}

func (x *KeygenResponse) GetRequestId() string {
	if x != nil {
		return x.RequestId
	}
	return ""
}

type SignRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RequestId             string   `protobuf:"bytes,1,opt,name=request_id,json=requestId,proto3" json:"request_id,omitempty"`
	ParticipantPublicKeys [][]byte `protobuf:"bytes,2,rep,name=participant_public_keys,json=participantPublicKeys,proto3" json:"participant_public_keys,omitempty"`
	PublicKey             []byte   `protobuf:"bytes,3,opt,name=public_key,json=publicKey,proto3" json:"public_key,omitempty"`
	Hash                  []byte   `protobuf:"bytes,4,opt,name=hash,proto3" json:"hash,omitempty"`
}

func (x *SignRequest) Reset() {
	*x = SignRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mpc_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SignRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SignRequest) ProtoMessage() {}

func (x *SignRequest) ProtoReflect() protoreflect.Message {
	mi := &file_mpc_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SignRequest.ProtoReflect.Descriptor instead.
func (*SignRequest) Descriptor() ([]byte, []int) {
	return file_mpc_proto_rawDescGZIP(), []int{2}
}

func (x *SignRequest) GetRequestId() string {
	if x != nil {
		return x.RequestId
	}
	return ""
}

func (x *SignRequest) GetParticipantPublicKeys() [][]byte {
	if x != nil {
		return x.ParticipantPublicKeys
	}
	return nil
}

func (x *SignRequest) GetPublicKey() []byte {
	if x != nil {
		return x.PublicKey
	}
	return nil
}

func (x *SignRequest) GetHash() []byte {
	if x != nil {
		return x.Hash
	}
	return nil
}

type SignResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RequestId string `protobuf:"bytes,1,opt,name=request_id,json=requestId,proto3" json:"request_id,omitempty"`
}

func (x *SignResponse) Reset() {
	*x = SignResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mpc_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SignResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SignResponse) ProtoMessage() {}

func (x *SignResponse) ProtoReflect() protoreflect.Message {
	mi := &file_mpc_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SignResponse.ProtoReflect.Descriptor instead.
func (*SignResponse) Descriptor() ([]byte, []int) {
	return file_mpc_proto_rawDescGZIP(), []int{3}
}

func (x *SignResponse) GetRequestId() string {
	if x != nil {
		return x.RequestId
	}
	return ""
}

type CheckResultRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RequestId string `protobuf:"bytes,1,opt,name=request_id,json=requestId,proto3" json:"request_id,omitempty"`
}

func (x *CheckResultRequest) Reset() {
	*x = CheckResultRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mpc_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CheckResultRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CheckResultRequest) ProtoMessage() {}

func (x *CheckResultRequest) ProtoReflect() protoreflect.Message {
	mi := &file_mpc_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CheckResultRequest.ProtoReflect.Descriptor instead.
func (*CheckResultRequest) Descriptor() ([]byte, []int) {
	return file_mpc_proto_rawDescGZIP(), []int{4}
}

func (x *CheckResultRequest) GetRequestId() string {
	if x != nil {
		return x.RequestId
	}
	return ""
}

type CheckResultResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RequestId     string                             `protobuf:"bytes,1,opt,name=request_id,json=requestId,proto3" json:"request_id,omitempty"`
	Result        string                             `protobuf:"bytes,2,opt,name=result,proto3" json:"result,omitempty"`
	RequestType   CheckResultResponse_REQUEST_TYPE   `protobuf:"varint,3,opt,name=request_type,json=requestType,proto3,enum=mpc.CheckResultResponse_REQUEST_TYPE" json:"request_type,omitempty"`
	RequestStatus CheckResultResponse_REQUEST_STATUS `protobuf:"varint,4,opt,name=request_status,json=requestStatus,proto3,enum=mpc.CheckResultResponse_REQUEST_STATUS" json:"request_status,omitempty"`
}

func (x *CheckResultResponse) Reset() {
	*x = CheckResultResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mpc_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CheckResultResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CheckResultResponse) ProtoMessage() {}

func (x *CheckResultResponse) ProtoReflect() protoreflect.Message {
	mi := &file_mpc_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CheckResultResponse.ProtoReflect.Descriptor instead.
func (*CheckResultResponse) Descriptor() ([]byte, []int) {
	return file_mpc_proto_rawDescGZIP(), []int{5}
}

func (x *CheckResultResponse) GetRequestId() string {
	if x != nil {
		return x.RequestId
	}
	return ""
}

func (x *CheckResultResponse) GetResult() string {
	if x != nil {
		return x.Result
	}
	return ""
}

func (x *CheckResultResponse) GetRequestType() CheckResultResponse_REQUEST_TYPE {
	if x != nil {
		return x.RequestType
	}
	return CheckResultResponse_UNKNOWN_TYPE
}

func (x *CheckResultResponse) GetRequestStatus() CheckResultResponse_REQUEST_STATUS {
	if x != nil {
		return x.RequestStatus
	}
	return CheckResultResponse_UNKNOWN_STATUS
}

var File_mpc_proto protoreflect.FileDescriptor

var file_mpc_proto_rawDesc = []byte{
	0x0a, 0x09, 0x6d, 0x70, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x03, 0x6d, 0x70, 0x63,
	0x22, 0x84, 0x01, 0x0a, 0x0d, 0x4b, 0x65, 0x79, 0x67, 0x65, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x5f, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x49,
	0x64, 0x12, 0x36, 0x0a, 0x17, 0x70, 0x61, 0x72, 0x74, 0x69, 0x63, 0x69, 0x70, 0x61, 0x6e, 0x74,
	0x5f, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x5f, 0x6b, 0x65, 0x79, 0x73, 0x18, 0x02, 0x20, 0x03,
	0x28, 0x0c, 0x52, 0x15, 0x70, 0x61, 0x72, 0x74, 0x69, 0x63, 0x69, 0x70, 0x61, 0x6e, 0x74, 0x50,
	0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x73, 0x12, 0x1c, 0x0a, 0x09, 0x74, 0x68, 0x72,
	0x65, 0x73, 0x68, 0x6f, 0x6c, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x09, 0x74, 0x68,
	0x72, 0x65, 0x73, 0x68, 0x6f, 0x6c, 0x64, 0x22, 0x2f, 0x0a, 0x0e, 0x4b, 0x65, 0x79, 0x67, 0x65,
	0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1d, 0x0a, 0x0a, 0x72, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x72,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x49, 0x64, 0x22, 0x97, 0x01, 0x0a, 0x0b, 0x53, 0x69, 0x67,
	0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x72, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x72, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x49, 0x64, 0x12, 0x36, 0x0a, 0x17, 0x70, 0x61, 0x72, 0x74, 0x69,
	0x63, 0x69, 0x70, 0x61, 0x6e, 0x74, 0x5f, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x5f, 0x6b, 0x65,
	0x79, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x15, 0x70, 0x61, 0x72, 0x74, 0x69, 0x63,
	0x69, 0x70, 0x61, 0x6e, 0x74, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x73, 0x12,
	0x1d, 0x0a, 0x0a, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x09, 0x70, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x12, 0x12,
	0x0a, 0x04, 0x68, 0x61, 0x73, 0x68, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x68, 0x61,
	0x73, 0x68, 0x22, 0x2d, 0x0a, 0x0c, 0x53, 0x69, 0x67, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x1d, 0x0a, 0x0a, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x5f, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x49,
	0x64, 0x22, 0x33, 0x0a, 0x12, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x72, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x72, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x49, 0x64, 0x22, 0xec, 0x02, 0x0a, 0x13, 0x43, 0x68, 0x65, 0x63, 0x6b,
	0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1d,
	0x0a, 0x0a, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x09, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x49, 0x64, 0x12, 0x16, 0x0a,
	0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x72,
	0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x48, 0x0a, 0x0c, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x25, 0x2e, 0x6d, 0x70,
	0x63, 0x2e, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x2e, 0x52, 0x45, 0x51, 0x55, 0x45, 0x53, 0x54, 0x5f, 0x54, 0x59,
	0x50, 0x45, 0x52, 0x0b, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12,
	0x4e, 0x0a, 0x0e, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x27, 0x2e, 0x6d, 0x70, 0x63, 0x2e, 0x43, 0x68,
	0x65, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x2e, 0x52, 0x45, 0x51, 0x55, 0x45, 0x53, 0x54, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x55, 0x53,
	0x52, 0x0d, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22,
	0x36, 0x0a, 0x0c, 0x52, 0x45, 0x51, 0x55, 0x45, 0x53, 0x54, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x12,
	0x10, 0x0a, 0x0c, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x10,
	0x00, 0x12, 0x0a, 0x0a, 0x06, 0x4b, 0x45, 0x59, 0x47, 0x45, 0x4e, 0x10, 0x01, 0x12, 0x08, 0x0a,
	0x04, 0x53, 0x49, 0x47, 0x4e, 0x10, 0x02, 0x22, 0x4c, 0x0a, 0x0e, 0x52, 0x45, 0x51, 0x55, 0x45,
	0x53, 0x54, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x55, 0x53, 0x12, 0x12, 0x0a, 0x0e, 0x55, 0x4e, 0x4b,
	0x4e, 0x4f, 0x57, 0x4e, 0x5f, 0x53, 0x54, 0x41, 0x54, 0x55, 0x53, 0x10, 0x00, 0x12, 0x0c, 0x0a,
	0x08, 0x52, 0x45, 0x43, 0x45, 0x49, 0x56, 0x45, 0x44, 0x10, 0x01, 0x12, 0x0e, 0x0a, 0x0a, 0x50,
	0x52, 0x4f, 0x43, 0x45, 0x53, 0x53, 0x49, 0x4e, 0x47, 0x10, 0x02, 0x12, 0x08, 0x0a, 0x04, 0x44,
	0x4f, 0x4e, 0x45, 0x10, 0x03, 0x32, 0xa7, 0x01, 0x0a, 0x03, 0x4d, 0x70, 0x63, 0x12, 0x31, 0x0a,
	0x06, 0x4b, 0x65, 0x79, 0x67, 0x65, 0x6e, 0x12, 0x12, 0x2e, 0x6d, 0x70, 0x63, 0x2e, 0x4b, 0x65,
	0x79, 0x67, 0x65, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x13, 0x2e, 0x6d, 0x70,
	0x63, 0x2e, 0x4b, 0x65, 0x79, 0x67, 0x65, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x2b, 0x0a, 0x04, 0x53, 0x69, 0x67, 0x6e, 0x12, 0x10, 0x2e, 0x6d, 0x70, 0x63, 0x2e, 0x53,
	0x69, 0x67, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x11, 0x2e, 0x6d, 0x70, 0x63,
	0x2e, 0x53, 0x69, 0x67, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x40, 0x0a,
	0x0b, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x17, 0x2e, 0x6d,
	0x70, 0x63, 0x2e, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x18, 0x2e, 0x6d, 0x70, 0x63, 0x2e, 0x43, 0x68, 0x65, 0x63,
	0x6b, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42,
	0x2d, 0x5a, 0x2b, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x76,
	0x61, 0x6c, 0x69, 0x64, 0x6f, 0x2f, 0x6d, 0x70, 0x63, 0x2d, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f,
	0x6c, 0x6c, 0x65, 0x72, 0x2f, 0x6d, 0x70, 0x63, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_mpc_proto_rawDescOnce sync.Once
	file_mpc_proto_rawDescData = file_mpc_proto_rawDesc
)

func file_mpc_proto_rawDescGZIP() []byte {
	file_mpc_proto_rawDescOnce.Do(func() {
		file_mpc_proto_rawDescData = protoimpl.X.CompressGZIP(file_mpc_proto_rawDescData)
	})
	return file_mpc_proto_rawDescData
}

var file_mpc_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_mpc_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_mpc_proto_goTypes = []interface{}{
	(CheckResultResponse_REQUEST_TYPE)(0),   // 0: mpc.CheckResultResponse.REQUEST_TYPE
	(CheckResultResponse_REQUEST_STATUS)(0), // 1: mpc.CheckResultResponse.REQUEST_STATUS
	(*KeygenRequest)(nil),                   // 2: mpc.KeygenRequest
	(*KeygenResponse)(nil),                  // 3: mpc.KeygenResponse
	(*SignRequest)(nil),                     // 4: mpc.SignRequest
	(*SignResponse)(nil),                    // 5: mpc.SignResponse
	(*CheckResultRequest)(nil),              // 6: mpc.CheckResultRequest
	(*CheckResultResponse)(nil),             // 7: mpc.CheckResultResponse
}
var file_mpc_proto_depIdxs = []int32{
	0, // 0: mpc.CheckResultResponse.request_type:type_name -> mpc.CheckResultResponse.REQUEST_TYPE
	1, // 1: mpc.CheckResultResponse.request_status:type_name -> mpc.CheckResultResponse.REQUEST_STATUS
	2, // 2: mpc.Mpc.Keygen:input_type -> mpc.KeygenRequest
	4, // 3: mpc.Mpc.Sign:input_type -> mpc.SignRequest
	6, // 4: mpc.Mpc.CheckResult:input_type -> mpc.CheckResultRequest
	3, // 5: mpc.Mpc.Keygen:output_type -> mpc.KeygenResponse
	5, // 6: mpc.Mpc.Sign:output_type -> mpc.SignResponse
	7, // 7: mpc.Mpc.CheckResult:output_type -> mpc.CheckResultResponse
	5, // [5:8] is the sub-list for method output_type
	2, // [2:5] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_mpc_proto_init() }
func file_mpc_proto_init() {
	if File_mpc_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_mpc_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*KeygenRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_mpc_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*KeygenResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_mpc_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SignRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_mpc_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SignResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_mpc_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CheckResultRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_mpc_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CheckResultResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_mpc_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_mpc_proto_goTypes,
		DependencyIndexes: file_mpc_proto_depIdxs,
		EnumInfos:         file_mpc_proto_enumTypes,
		MessageInfos:      file_mpc_proto_msgTypes,
	}.Build()
	File_mpc_proto = out.File
	file_mpc_proto_rawDesc = nil
	file_mpc_proto_goTypes = nil
	file_mpc_proto_depIdxs = nil
}
