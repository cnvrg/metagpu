// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        (unknown)
// source: device/v1/device.proto

package devicev1

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

type DeviceProcess struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Uuid            string `protobuf:"bytes,1,opt,name=uuid,proto3" json:"uuid,omitempty"`
	Pid             uint32 `protobuf:"varint,2,opt,name=pid,proto3" json:"pid,omitempty"`
	Memory          uint64 `protobuf:"varint,3,opt,name=memory,proto3" json:"memory,omitempty"`
	Cmdline         string `protobuf:"bytes,4,opt,name=cmdline,proto3" json:"cmdline,omitempty"`
	User            string `protobuf:"bytes,5,opt,name=user,proto3" json:"user,omitempty"`
	ContainerId     string `protobuf:"bytes,6,opt,name=container_id,json=containerId,proto3" json:"container_id,omitempty"`
	PodName         string `protobuf:"bytes,7,opt,name=pod_name,json=podName,proto3" json:"pod_name,omitempty"`
	PodNamespace    string `protobuf:"bytes,8,opt,name=pod_namespace,json=podNamespace,proto3" json:"pod_namespace,omitempty"`
	MetagpuRequests int64  `protobuf:"varint,9,opt,name=metagpu_requests,json=metagpuRequests,proto3" json:"metagpu_requests,omitempty"`
}

func (x *DeviceProcess) Reset() {
	*x = DeviceProcess{}
	if protoimpl.UnsafeEnabled {
		mi := &file_device_v1_device_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeviceProcess) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeviceProcess) ProtoMessage() {}

func (x *DeviceProcess) ProtoReflect() protoreflect.Message {
	mi := &file_device_v1_device_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeviceProcess.ProtoReflect.Descriptor instead.
func (*DeviceProcess) Descriptor() ([]byte, []int) {
	return file_device_v1_device_proto_rawDescGZIP(), []int{0}
}

func (x *DeviceProcess) GetUuid() string {
	if x != nil {
		return x.Uuid
	}
	return ""
}

func (x *DeviceProcess) GetPid() uint32 {
	if x != nil {
		return x.Pid
	}
	return 0
}

func (x *DeviceProcess) GetMemory() uint64 {
	if x != nil {
		return x.Memory
	}
	return 0
}

func (x *DeviceProcess) GetCmdline() string {
	if x != nil {
		return x.Cmdline
	}
	return ""
}

func (x *DeviceProcess) GetUser() string {
	if x != nil {
		return x.User
	}
	return ""
}

func (x *DeviceProcess) GetContainerId() string {
	if x != nil {
		return x.ContainerId
	}
	return ""
}

func (x *DeviceProcess) GetPodName() string {
	if x != nil {
		return x.PodName
	}
	return ""
}

func (x *DeviceProcess) GetPodNamespace() string {
	if x != nil {
		return x.PodNamespace
	}
	return ""
}

func (x *DeviceProcess) GetMetagpuRequests() int64 {
	if x != nil {
		return x.MetagpuRequests
	}
	return 0
}

type StreamDeviceProcessesRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PodId string `protobuf:"bytes,1,opt,name=pod_id,json=podId,proto3" json:"pod_id,omitempty"`
}

func (x *StreamDeviceProcessesRequest) Reset() {
	*x = StreamDeviceProcessesRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_device_v1_device_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StreamDeviceProcessesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StreamDeviceProcessesRequest) ProtoMessage() {}

func (x *StreamDeviceProcessesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_device_v1_device_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StreamDeviceProcessesRequest.ProtoReflect.Descriptor instead.
func (*StreamDeviceProcessesRequest) Descriptor() ([]byte, []int) {
	return file_device_v1_device_proto_rawDescGZIP(), []int{1}
}

func (x *StreamDeviceProcessesRequest) GetPodId() string {
	if x != nil {
		return x.PodId
	}
	return ""
}

type StreamDeviceProcessesResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	DevicesProcesses []*DeviceProcess `protobuf:"bytes,1,rep,name=devices_processes,json=devicesProcesses,proto3" json:"devices_processes,omitempty"`
}

func (x *StreamDeviceProcessesResponse) Reset() {
	*x = StreamDeviceProcessesResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_device_v1_device_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StreamDeviceProcessesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StreamDeviceProcessesResponse) ProtoMessage() {}

func (x *StreamDeviceProcessesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_device_v1_device_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StreamDeviceProcessesResponse.ProtoReflect.Descriptor instead.
func (*StreamDeviceProcessesResponse) Descriptor() ([]byte, []int) {
	return file_device_v1_device_proto_rawDescGZIP(), []int{2}
}

func (x *StreamDeviceProcessesResponse) GetDevicesProcesses() []*DeviceProcess {
	if x != nil {
		return x.DevicesProcesses
	}
	return nil
}

type ListDeviceProcessesRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PodId string `protobuf:"bytes,1,opt,name=pod_id,json=podId,proto3" json:"pod_id,omitempty"`
}

func (x *ListDeviceProcessesRequest) Reset() {
	*x = ListDeviceProcessesRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_device_v1_device_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListDeviceProcessesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListDeviceProcessesRequest) ProtoMessage() {}

func (x *ListDeviceProcessesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_device_v1_device_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListDeviceProcessesRequest.ProtoReflect.Descriptor instead.
func (*ListDeviceProcessesRequest) Descriptor() ([]byte, []int) {
	return file_device_v1_device_proto_rawDescGZIP(), []int{3}
}

func (x *ListDeviceProcessesRequest) GetPodId() string {
	if x != nil {
		return x.PodId
	}
	return ""
}

type ListDeviceProcessesResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	DevicesProcesses []*DeviceProcess `protobuf:"bytes,1,rep,name=devices_processes,json=devicesProcesses,proto3" json:"devices_processes,omitempty"`
}

func (x *ListDeviceProcessesResponse) Reset() {
	*x = ListDeviceProcessesResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_device_v1_device_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListDeviceProcessesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListDeviceProcessesResponse) ProtoMessage() {}

func (x *ListDeviceProcessesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_device_v1_device_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListDeviceProcessesResponse.ProtoReflect.Descriptor instead.
func (*ListDeviceProcessesResponse) Descriptor() ([]byte, []int) {
	return file_device_v1_device_proto_rawDescGZIP(), []int{4}
}

func (x *ListDeviceProcessesResponse) GetDevicesProcesses() []*DeviceProcess {
	if x != nil {
		return x.DevicesProcesses
	}
	return nil
}

type PingServerRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *PingServerRequest) Reset() {
	*x = PingServerRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_device_v1_device_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PingServerRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PingServerRequest) ProtoMessage() {}

func (x *PingServerRequest) ProtoReflect() protoreflect.Message {
	mi := &file_device_v1_device_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PingServerRequest.ProtoReflect.Descriptor instead.
func (*PingServerRequest) Descriptor() ([]byte, []int) {
	return file_device_v1_device_proto_rawDescGZIP(), []int{5}
}

type PingServerResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *PingServerResponse) Reset() {
	*x = PingServerResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_device_v1_device_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PingServerResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PingServerResponse) ProtoMessage() {}

func (x *PingServerResponse) ProtoReflect() protoreflect.Message {
	mi := &file_device_v1_device_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PingServerResponse.ProtoReflect.Descriptor instead.
func (*PingServerResponse) Descriptor() ([]byte, []int) {
	return file_device_v1_device_proto_rawDescGZIP(), []int{6}
}

var File_device_v1_device_proto protoreflect.FileDescriptor

var file_device_v1_device_proto_rawDesc = []byte{
	0x0a, 0x16, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x2f, 0x76, 0x31, 0x2f, 0x64, 0x65, 0x76, 0x69,
	0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65,
	0x2e, 0x76, 0x31, 0x22, 0x89, 0x02, 0x0a, 0x0d, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x50, 0x72,
	0x6f, 0x63, 0x65, 0x73, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x75, 0x75, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x04, 0x75, 0x75, 0x69, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x70, 0x69, 0x64,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x03, 0x70, 0x69, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x6d,
	0x65, 0x6d, 0x6f, 0x72, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x06, 0x6d, 0x65, 0x6d,
	0x6f, 0x72, 0x79, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x6d, 0x64, 0x6c, 0x69, 0x6e, 0x65, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x6d, 0x64, 0x6c, 0x69, 0x6e, 0x65, 0x12, 0x12, 0x0a,
	0x04, 0x75, 0x73, 0x65, 0x72, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x75, 0x73, 0x65,
	0x72, 0x12, 0x21, 0x0a, 0x0c, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x5f, 0x69,
	0x64, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e,
	0x65, 0x72, 0x49, 0x64, 0x12, 0x19, 0x0a, 0x08, 0x70, 0x6f, 0x64, 0x5f, 0x6e, 0x61, 0x6d, 0x65,
	0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x70, 0x6f, 0x64, 0x4e, 0x61, 0x6d, 0x65, 0x12,
	0x23, 0x0a, 0x0d, 0x70, 0x6f, 0x64, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65,
	0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x70, 0x6f, 0x64, 0x4e, 0x61, 0x6d, 0x65, 0x73,
	0x70, 0x61, 0x63, 0x65, 0x12, 0x29, 0x0a, 0x10, 0x6d, 0x65, 0x74, 0x61, 0x67, 0x70, 0x75, 0x5f,
	0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x73, 0x18, 0x09, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0f,
	0x6d, 0x65, 0x74, 0x61, 0x67, 0x70, 0x75, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x73, 0x22,
	0x35, 0x0a, 0x1c, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x50,
	0x72, 0x6f, 0x63, 0x65, 0x73, 0x73, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x15, 0x0a, 0x06, 0x70, 0x6f, 0x64, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x05, 0x70, 0x6f, 0x64, 0x49, 0x64, 0x22, 0x66, 0x0a, 0x1d, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d,
	0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x50, 0x72, 0x6f, 0x63, 0x65, 0x73, 0x73, 0x65, 0x73, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x45, 0x0a, 0x11, 0x64, 0x65, 0x76, 0x69, 0x63,
	0x65, 0x73, 0x5f, 0x70, 0x72, 0x6f, 0x63, 0x65, 0x73, 0x73, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x18, 0x2e, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x44,
	0x65, 0x76, 0x69, 0x63, 0x65, 0x50, 0x72, 0x6f, 0x63, 0x65, 0x73, 0x73, 0x52, 0x10, 0x64, 0x65,
	0x76, 0x69, 0x63, 0x65, 0x73, 0x50, 0x72, 0x6f, 0x63, 0x65, 0x73, 0x73, 0x65, 0x73, 0x22, 0x33,
	0x0a, 0x1a, 0x4c, 0x69, 0x73, 0x74, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x50, 0x72, 0x6f, 0x63,
	0x65, 0x73, 0x73, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x15, 0x0a, 0x06,
	0x70, 0x6f, 0x64, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x70, 0x6f,
	0x64, 0x49, 0x64, 0x22, 0x64, 0x0a, 0x1b, 0x4c, 0x69, 0x73, 0x74, 0x44, 0x65, 0x76, 0x69, 0x63,
	0x65, 0x50, 0x72, 0x6f, 0x63, 0x65, 0x73, 0x73, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x45, 0x0a, 0x11, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x73, 0x5f, 0x70, 0x72,
	0x6f, 0x63, 0x65, 0x73, 0x73, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x18, 0x2e,
	0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65,
	0x50, 0x72, 0x6f, 0x63, 0x65, 0x73, 0x73, 0x52, 0x10, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x73,
	0x50, 0x72, 0x6f, 0x63, 0x65, 0x73, 0x73, 0x65, 0x73, 0x22, 0x13, 0x0a, 0x11, 0x50, 0x69, 0x6e,
	0x67, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x14,
	0x0a, 0x12, 0x50, 0x69, 0x6e, 0x67, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x32, 0xb4, 0x02, 0x0a, 0x0d, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x53,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x66, 0x0a, 0x13, 0x4c, 0x69, 0x73, 0x74, 0x44, 0x65,
	0x76, 0x69, 0x63, 0x65, 0x50, 0x72, 0x6f, 0x63, 0x65, 0x73, 0x73, 0x65, 0x73, 0x12, 0x25, 0x2e,
	0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x44, 0x65,
	0x76, 0x69, 0x63, 0x65, 0x50, 0x72, 0x6f, 0x63, 0x65, 0x73, 0x73, 0x65, 0x73, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x26, 0x2e, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x76, 0x31,
	0x2e, 0x4c, 0x69, 0x73, 0x74, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x50, 0x72, 0x6f, 0x63, 0x65,
	0x73, 0x73, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x6e,
	0x0a, 0x15, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x50, 0x72,
	0x6f, 0x63, 0x65, 0x73, 0x73, 0x65, 0x73, 0x12, 0x27, 0x2e, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65,
	0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65,
	0x50, 0x72, 0x6f, 0x63, 0x65, 0x73, 0x73, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x28, 0x2e, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74, 0x72,
	0x65, 0x61, 0x6d, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x50, 0x72, 0x6f, 0x63, 0x65, 0x73, 0x73,
	0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x30, 0x01, 0x12, 0x4b,
	0x0a, 0x0a, 0x50, 0x69, 0x6e, 0x67, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x12, 0x1c, 0x2e, 0x64,
	0x65, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x69, 0x6e, 0x67, 0x53, 0x65, 0x72,
	0x76, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1d, 0x2e, 0x64, 0x65, 0x76,
	0x69, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x69, 0x6e, 0x67, 0x53, 0x65, 0x72, 0x76, 0x65,
	0x72, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0xb2, 0x01, 0x0a, 0x0d,
	0x63, 0x6f, 0x6d, 0x2e, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x76, 0x31, 0x42, 0x0b, 0x44,
	0x65, 0x76, 0x69, 0x63, 0x65, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x48, 0x02, 0x50, 0x01, 0x5a, 0x4d,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x41, 0x63, 0x63, 0x65, 0x73,
	0x73, 0x69, 0x62, 0x6c, 0x65, 0x41, 0x49, 0x2f, 0x6d, 0x65, 0x74, 0x61, 0x67, 0x70, 0x75, 0x2d,
	0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x2d, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x2f, 0x67, 0x65,
	0x6e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x6f, 0x2f, 0x64, 0x65, 0x76, 0x69, 0x63,
	0x65, 0x2f, 0x76, 0x31, 0x3b, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x76, 0x31, 0xa2, 0x02, 0x03,
	0x44, 0x58, 0x58, 0xaa, 0x02, 0x09, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x56, 0x31, 0xca,
	0x02, 0x09, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x5c, 0x56, 0x31, 0xe2, 0x02, 0x15, 0x44, 0x65,
	0x76, 0x69, 0x63, 0x65, 0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64,
	0x61, 0x74, 0x61, 0xea, 0x02, 0x0a, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x3a, 0x3a, 0x56, 0x31,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_device_v1_device_proto_rawDescOnce sync.Once
	file_device_v1_device_proto_rawDescData = file_device_v1_device_proto_rawDesc
)

func file_device_v1_device_proto_rawDescGZIP() []byte {
	file_device_v1_device_proto_rawDescOnce.Do(func() {
		file_device_v1_device_proto_rawDescData = protoimpl.X.CompressGZIP(file_device_v1_device_proto_rawDescData)
	})
	return file_device_v1_device_proto_rawDescData
}

var file_device_v1_device_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_device_v1_device_proto_goTypes = []interface{}{
	(*DeviceProcess)(nil),                 // 0: device.v1.DeviceProcess
	(*StreamDeviceProcessesRequest)(nil),  // 1: device.v1.StreamDeviceProcessesRequest
	(*StreamDeviceProcessesResponse)(nil), // 2: device.v1.StreamDeviceProcessesResponse
	(*ListDeviceProcessesRequest)(nil),    // 3: device.v1.ListDeviceProcessesRequest
	(*ListDeviceProcessesResponse)(nil),   // 4: device.v1.ListDeviceProcessesResponse
	(*PingServerRequest)(nil),             // 5: device.v1.PingServerRequest
	(*PingServerResponse)(nil),            // 6: device.v1.PingServerResponse
}
var file_device_v1_device_proto_depIdxs = []int32{
	0, // 0: device.v1.StreamDeviceProcessesResponse.devices_processes:type_name -> device.v1.DeviceProcess
	0, // 1: device.v1.ListDeviceProcessesResponse.devices_processes:type_name -> device.v1.DeviceProcess
	3, // 2: device.v1.DeviceService.ListDeviceProcesses:input_type -> device.v1.ListDeviceProcessesRequest
	1, // 3: device.v1.DeviceService.StreamDeviceProcesses:input_type -> device.v1.StreamDeviceProcessesRequest
	5, // 4: device.v1.DeviceService.PingServer:input_type -> device.v1.PingServerRequest
	4, // 5: device.v1.DeviceService.ListDeviceProcesses:output_type -> device.v1.ListDeviceProcessesResponse
	2, // 6: device.v1.DeviceService.StreamDeviceProcesses:output_type -> device.v1.StreamDeviceProcessesResponse
	6, // 7: device.v1.DeviceService.PingServer:output_type -> device.v1.PingServerResponse
	5, // [5:8] is the sub-list for method output_type
	2, // [2:5] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_device_v1_device_proto_init() }
func file_device_v1_device_proto_init() {
	if File_device_v1_device_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_device_v1_device_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeviceProcess); i {
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
		file_device_v1_device_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StreamDeviceProcessesRequest); i {
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
		file_device_v1_device_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StreamDeviceProcessesResponse); i {
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
		file_device_v1_device_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListDeviceProcessesRequest); i {
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
		file_device_v1_device_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListDeviceProcessesResponse); i {
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
		file_device_v1_device_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PingServerRequest); i {
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
		file_device_v1_device_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PingServerResponse); i {
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
			RawDescriptor: file_device_v1_device_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_device_v1_device_proto_goTypes,
		DependencyIndexes: file_device_v1_device_proto_depIdxs,
		MessageInfos:      file_device_v1_device_proto_msgTypes,
	}.Build()
	File_device_v1_device_proto = out.File
	file_device_v1_device_proto_rawDesc = nil
	file_device_v1_device_proto_goTypes = nil
	file_device_v1_device_proto_depIdxs = nil
}
