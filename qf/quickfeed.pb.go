// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        (unknown)
// source: qf/quickfeed.proto

package qf

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

var File_qf_quickfeed_proto protoreflect.FileDescriptor

var file_qf_quickfeed_proto_rawDesc = []byte{
	0x0a, 0x12, 0x71, 0x66, 0x2f, 0x71, 0x75, 0x69, 0x63, 0x6b, 0x66, 0x65, 0x65, 0x64, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x02, 0x71, 0x66, 0x1a, 0x0e, 0x71, 0x66, 0x2f, 0x74, 0x79, 0x70,
	0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x11, 0x71, 0x66, 0x2f, 0x72, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x32, 0xd9, 0x10, 0x0a, 0x10,
	0x51, 0x75, 0x69, 0x63, 0x6b, 0x46, 0x65, 0x65, 0x64, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x12, 0x1f, 0x0a, 0x07, 0x47, 0x65, 0x74, 0x55, 0x73, 0x65, 0x72, 0x12, 0x08, 0x2e, 0x71, 0x66,
	0x2e, 0x56, 0x6f, 0x69, 0x64, 0x1a, 0x08, 0x2e, 0x71, 0x66, 0x2e, 0x55, 0x73, 0x65, 0x72, 0x22,
	0x00, 0x12, 0x21, 0x0a, 0x08, 0x47, 0x65, 0x74, 0x55, 0x73, 0x65, 0x72, 0x73, 0x12, 0x08, 0x2e,
	0x71, 0x66, 0x2e, 0x56, 0x6f, 0x69, 0x64, 0x1a, 0x09, 0x2e, 0x71, 0x66, 0x2e, 0x55, 0x73, 0x65,
	0x72, 0x73, 0x22, 0x00, 0x12, 0x34, 0x0a, 0x0f, 0x47, 0x65, 0x74, 0x55, 0x73, 0x65, 0x72, 0x42,
	0x79, 0x43, 0x6f, 0x75, 0x72, 0x73, 0x65, 0x12, 0x15, 0x2e, 0x71, 0x66, 0x2e, 0x43, 0x6f, 0x75,
	0x72, 0x73, 0x65, 0x55, 0x73, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x08,
	0x2e, 0x71, 0x66, 0x2e, 0x55, 0x73, 0x65, 0x72, 0x22, 0x00, 0x12, 0x22, 0x0a, 0x0a, 0x55, 0x70,
	0x64, 0x61, 0x74, 0x65, 0x55, 0x73, 0x65, 0x72, 0x12, 0x08, 0x2e, 0x71, 0x66, 0x2e, 0x55, 0x73,
	0x65, 0x72, 0x1a, 0x08, 0x2e, 0x71, 0x66, 0x2e, 0x56, 0x6f, 0x69, 0x64, 0x22, 0x00, 0x12, 0x3c,
	0x0a, 0x13, 0x49, 0x73, 0x41, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x69, 0x7a, 0x65, 0x64, 0x54, 0x65,
	0x61, 0x63, 0x68, 0x65, 0x72, 0x12, 0x08, 0x2e, 0x71, 0x66, 0x2e, 0x56, 0x6f, 0x69, 0x64, 0x1a,
	0x19, 0x2e, 0x71, 0x66, 0x2e, 0x41, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x69, 0x7a, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x2c, 0x0a, 0x08,
	0x47, 0x65, 0x74, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x12, 0x13, 0x2e, 0x71, 0x66, 0x2e, 0x47, 0x65,
	0x74, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x09, 0x2e,
	0x71, 0x66, 0x2e, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x22, 0x00, 0x12, 0x38, 0x0a, 0x17, 0x47, 0x65,
	0x74, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x42, 0x79, 0x55, 0x73, 0x65, 0x72, 0x41, 0x6e, 0x64, 0x43,
	0x6f, 0x75, 0x72, 0x73, 0x65, 0x12, 0x10, 0x2e, 0x71, 0x66, 0x2e, 0x47, 0x72, 0x6f, 0x75, 0x70,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x09, 0x2e, 0x71, 0x66, 0x2e, 0x47, 0x72, 0x6f,
	0x75, 0x70, 0x22, 0x00, 0x12, 0x34, 0x0a, 0x11, 0x47, 0x65, 0x74, 0x47, 0x72, 0x6f, 0x75, 0x70,
	0x73, 0x42, 0x79, 0x43, 0x6f, 0x75, 0x72, 0x73, 0x65, 0x12, 0x11, 0x2e, 0x71, 0x66, 0x2e, 0x43,
	0x6f, 0x75, 0x72, 0x73, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0a, 0x2e, 0x71,
	0x66, 0x2e, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x73, 0x22, 0x00, 0x12, 0x25, 0x0a, 0x0b, 0x43, 0x72,
	0x65, 0x61, 0x74, 0x65, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x12, 0x09, 0x2e, 0x71, 0x66, 0x2e, 0x47,
	0x72, 0x6f, 0x75, 0x70, 0x1a, 0x09, 0x2e, 0x71, 0x66, 0x2e, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x22,
	0x00, 0x12, 0x25, 0x0a, 0x0b, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x47, 0x72, 0x6f, 0x75, 0x70,
	0x12, 0x09, 0x2e, 0x71, 0x66, 0x2e, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x1a, 0x09, 0x2e, 0x71, 0x66,
	0x2e, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x22, 0x00, 0x12, 0x2b, 0x0a, 0x0b, 0x44, 0x65, 0x6c, 0x65,
	0x74, 0x65, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x12, 0x10, 0x2e, 0x71, 0x66, 0x2e, 0x47, 0x72, 0x6f,
	0x75, 0x70, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x08, 0x2e, 0x71, 0x66, 0x2e, 0x56,
	0x6f, 0x69, 0x64, 0x22, 0x00, 0x12, 0x2c, 0x0a, 0x09, 0x47, 0x65, 0x74, 0x43, 0x6f, 0x75, 0x72,
	0x73, 0x65, 0x12, 0x11, 0x2e, 0x71, 0x66, 0x2e, 0x43, 0x6f, 0x75, 0x72, 0x73, 0x65, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0a, 0x2e, 0x71, 0x66, 0x2e, 0x43, 0x6f, 0x75, 0x72, 0x73,
	0x65, 0x22, 0x00, 0x12, 0x25, 0x0a, 0x0a, 0x47, 0x65, 0x74, 0x43, 0x6f, 0x75, 0x72, 0x73, 0x65,
	0x73, 0x12, 0x08, 0x2e, 0x71, 0x66, 0x2e, 0x56, 0x6f, 0x69, 0x64, 0x1a, 0x0b, 0x2e, 0x71, 0x66,
	0x2e, 0x43, 0x6f, 0x75, 0x72, 0x73, 0x65, 0x73, 0x22, 0x00, 0x12, 0x3e, 0x0a, 0x10, 0x47, 0x65,
	0x74, 0x43, 0x6f, 0x75, 0x72, 0x73, 0x65, 0x73, 0x42, 0x79, 0x55, 0x73, 0x65, 0x72, 0x12, 0x1b,
	0x2e, 0x71, 0x66, 0x2e, 0x45, 0x6e, 0x72, 0x6f, 0x6c, 0x6c, 0x6d, 0x65, 0x6e, 0x74, 0x53, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0b, 0x2e, 0x71, 0x66,
	0x2e, 0x43, 0x6f, 0x75, 0x72, 0x73, 0x65, 0x73, 0x22, 0x00, 0x12, 0x28, 0x0a, 0x0c, 0x43, 0x72,
	0x65, 0x61, 0x74, 0x65, 0x43, 0x6f, 0x75, 0x72, 0x73, 0x65, 0x12, 0x0a, 0x2e, 0x71, 0x66, 0x2e,
	0x43, 0x6f, 0x75, 0x72, 0x73, 0x65, 0x1a, 0x0a, 0x2e, 0x71, 0x66, 0x2e, 0x43, 0x6f, 0x75, 0x72,
	0x73, 0x65, 0x22, 0x00, 0x12, 0x26, 0x0a, 0x0c, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x43, 0x6f,
	0x75, 0x72, 0x73, 0x65, 0x12, 0x0a, 0x2e, 0x71, 0x66, 0x2e, 0x43, 0x6f, 0x75, 0x72, 0x73, 0x65,
	0x1a, 0x08, 0x2e, 0x71, 0x66, 0x2e, 0x56, 0x6f, 0x69, 0x64, 0x22, 0x00, 0x12, 0x34, 0x0a, 0x16,
	0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x43, 0x6f, 0x75, 0x72, 0x73, 0x65, 0x56, 0x69, 0x73, 0x69,
	0x62, 0x69, 0x6c, 0x69, 0x74, 0x79, 0x12, 0x0e, 0x2e, 0x71, 0x66, 0x2e, 0x45, 0x6e, 0x72, 0x6f,
	0x6c, 0x6c, 0x6d, 0x65, 0x6e, 0x74, 0x1a, 0x08, 0x2e, 0x71, 0x66, 0x2e, 0x56, 0x6f, 0x69, 0x64,
	0x22, 0x00, 0x12, 0x36, 0x0a, 0x0e, 0x47, 0x65, 0x74, 0x41, 0x73, 0x73, 0x69, 0x67, 0x6e, 0x6d,
	0x65, 0x6e, 0x74, 0x73, 0x12, 0x11, 0x2e, 0x71, 0x66, 0x2e, 0x43, 0x6f, 0x75, 0x72, 0x73, 0x65,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0f, 0x2e, 0x71, 0x66, 0x2e, 0x41, 0x73, 0x73,
	0x69, 0x67, 0x6e, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x22, 0x00, 0x12, 0x32, 0x0a, 0x11, 0x55, 0x70,
	0x64, 0x61, 0x74, 0x65, 0x41, 0x73, 0x73, 0x69, 0x67, 0x6e, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x12,
	0x11, 0x2e, 0x71, 0x66, 0x2e, 0x43, 0x6f, 0x75, 0x72, 0x73, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x08, 0x2e, 0x71, 0x66, 0x2e, 0x56, 0x6f, 0x69, 0x64, 0x22, 0x00, 0x12, 0x46,
	0x0a, 0x14, 0x47, 0x65, 0x74, 0x45, 0x6e, 0x72, 0x6f, 0x6c, 0x6c, 0x6d, 0x65, 0x6e, 0x74, 0x73,
	0x42, 0x79, 0x55, 0x73, 0x65, 0x72, 0x12, 0x1b, 0x2e, 0x71, 0x66, 0x2e, 0x45, 0x6e, 0x72, 0x6f,
	0x6c, 0x6c, 0x6d, 0x65, 0x6e, 0x74, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x0f, 0x2e, 0x71, 0x66, 0x2e, 0x45, 0x6e, 0x72, 0x6f, 0x6c, 0x6c, 0x6d,
	0x65, 0x6e, 0x74, 0x73, 0x22, 0x00, 0x12, 0x42, 0x0a, 0x16, 0x47, 0x65, 0x74, 0x45, 0x6e, 0x72,
	0x6f, 0x6c, 0x6c, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x42, 0x79, 0x43, 0x6f, 0x75, 0x72, 0x73, 0x65,
	0x12, 0x15, 0x2e, 0x71, 0x66, 0x2e, 0x45, 0x6e, 0x72, 0x6f, 0x6c, 0x6c, 0x6d, 0x65, 0x6e, 0x74,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0f, 0x2e, 0x71, 0x66, 0x2e, 0x45, 0x6e, 0x72,
	0x6f, 0x6c, 0x6c, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x22, 0x00, 0x12, 0x2e, 0x0a, 0x10, 0x43, 0x72,
	0x65, 0x61, 0x74, 0x65, 0x45, 0x6e, 0x72, 0x6f, 0x6c, 0x6c, 0x6d, 0x65, 0x6e, 0x74, 0x12, 0x0e,
	0x2e, 0x71, 0x66, 0x2e, 0x45, 0x6e, 0x72, 0x6f, 0x6c, 0x6c, 0x6d, 0x65, 0x6e, 0x74, 0x1a, 0x08,
	0x2e, 0x71, 0x66, 0x2e, 0x56, 0x6f, 0x69, 0x64, 0x22, 0x00, 0x12, 0x30, 0x0a, 0x11, 0x55, 0x70,
	0x64, 0x61, 0x74, 0x65, 0x45, 0x6e, 0x72, 0x6f, 0x6c, 0x6c, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x12,
	0x0f, 0x2e, 0x71, 0x66, 0x2e, 0x45, 0x6e, 0x72, 0x6f, 0x6c, 0x6c, 0x6d, 0x65, 0x6e, 0x74, 0x73,
	0x1a, 0x08, 0x2e, 0x71, 0x66, 0x2e, 0x56, 0x6f, 0x69, 0x64, 0x22, 0x00, 0x12, 0x3a, 0x0a, 0x0e,
	0x47, 0x65, 0x74, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x15,
	0x2e, 0x71, 0x66, 0x2e, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0f, 0x2e, 0x71, 0x66, 0x2e, 0x53, 0x75, 0x62, 0x6d, 0x69,
	0x73, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x22, 0x00, 0x12, 0x52, 0x0a, 0x16, 0x47, 0x65, 0x74, 0x53,
	0x75, 0x62, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x42, 0x79, 0x43, 0x6f, 0x75, 0x72,
	0x73, 0x65, 0x12, 0x1f, 0x2e, 0x71, 0x66, 0x2e, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x73, 0x73, 0x69,
	0x6f, 0x6e, 0x73, 0x46, 0x6f, 0x72, 0x43, 0x6f, 0x75, 0x72, 0x73, 0x65, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x15, 0x2e, 0x71, 0x66, 0x2e, 0x43, 0x6f, 0x75, 0x72, 0x73, 0x65, 0x53,
	0x75, 0x62, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x22, 0x00, 0x12, 0x3b, 0x0a, 0x10,
	0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e,
	0x12, 0x1b, 0x2e, 0x71, 0x66, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x53, 0x75, 0x62, 0x6d,
	0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x08, 0x2e,
	0x71, 0x66, 0x2e, 0x56, 0x6f, 0x69, 0x64, 0x22, 0x00, 0x12, 0x3d, 0x0a, 0x11, 0x55, 0x70, 0x64,
	0x61, 0x74, 0x65, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x1c,
	0x2e, 0x71, 0x66, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x73,
	0x73, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x08, 0x2e, 0x71,
	0x66, 0x2e, 0x56, 0x6f, 0x69, 0x64, 0x22, 0x00, 0x12, 0x34, 0x0a, 0x12, 0x52, 0x65, 0x62, 0x75,
	0x69, 0x6c, 0x64, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x12,
	0x2e, 0x71, 0x66, 0x2e, 0x52, 0x65, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x08, 0x2e, 0x71, 0x66, 0x2e, 0x56, 0x6f, 0x69, 0x64, 0x22, 0x00, 0x12, 0x3f,
	0x0a, 0x0f, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x42, 0x65, 0x6e, 0x63, 0x68, 0x6d, 0x61, 0x72,
	0x6b, 0x12, 0x14, 0x2e, 0x71, 0x66, 0x2e, 0x47, 0x72, 0x61, 0x64, 0x69, 0x6e, 0x67, 0x42, 0x65,
	0x6e, 0x63, 0x68, 0x6d, 0x61, 0x72, 0x6b, 0x1a, 0x14, 0x2e, 0x71, 0x66, 0x2e, 0x47, 0x72, 0x61,
	0x64, 0x69, 0x6e, 0x67, 0x42, 0x65, 0x6e, 0x63, 0x68, 0x6d, 0x61, 0x72, 0x6b, 0x22, 0x00, 0x12,
	0x33, 0x0a, 0x0f, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x42, 0x65, 0x6e, 0x63, 0x68, 0x6d, 0x61,
	0x72, 0x6b, 0x12, 0x14, 0x2e, 0x71, 0x66, 0x2e, 0x47, 0x72, 0x61, 0x64, 0x69, 0x6e, 0x67, 0x42,
	0x65, 0x6e, 0x63, 0x68, 0x6d, 0x61, 0x72, 0x6b, 0x1a, 0x08, 0x2e, 0x71, 0x66, 0x2e, 0x56, 0x6f,
	0x69, 0x64, 0x22, 0x00, 0x12, 0x33, 0x0a, 0x0f, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x42, 0x65,
	0x6e, 0x63, 0x68, 0x6d, 0x61, 0x72, 0x6b, 0x12, 0x14, 0x2e, 0x71, 0x66, 0x2e, 0x47, 0x72, 0x61,
	0x64, 0x69, 0x6e, 0x67, 0x42, 0x65, 0x6e, 0x63, 0x68, 0x6d, 0x61, 0x72, 0x6b, 0x1a, 0x08, 0x2e,
	0x71, 0x66, 0x2e, 0x56, 0x6f, 0x69, 0x64, 0x22, 0x00, 0x12, 0x3f, 0x0a, 0x0f, 0x43, 0x72, 0x65,
	0x61, 0x74, 0x65, 0x43, 0x72, 0x69, 0x74, 0x65, 0x72, 0x69, 0x6f, 0x6e, 0x12, 0x14, 0x2e, 0x71,
	0x66, 0x2e, 0x47, 0x72, 0x61, 0x64, 0x69, 0x6e, 0x67, 0x43, 0x72, 0x69, 0x74, 0x65, 0x72, 0x69,
	0x6f, 0x6e, 0x1a, 0x14, 0x2e, 0x71, 0x66, 0x2e, 0x47, 0x72, 0x61, 0x64, 0x69, 0x6e, 0x67, 0x43,
	0x72, 0x69, 0x74, 0x65, 0x72, 0x69, 0x6f, 0x6e, 0x22, 0x00, 0x12, 0x33, 0x0a, 0x0f, 0x55, 0x70,
	0x64, 0x61, 0x74, 0x65, 0x43, 0x72, 0x69, 0x74, 0x65, 0x72, 0x69, 0x6f, 0x6e, 0x12, 0x14, 0x2e,
	0x71, 0x66, 0x2e, 0x47, 0x72, 0x61, 0x64, 0x69, 0x6e, 0x67, 0x43, 0x72, 0x69, 0x74, 0x65, 0x72,
	0x69, 0x6f, 0x6e, 0x1a, 0x08, 0x2e, 0x71, 0x66, 0x2e, 0x56, 0x6f, 0x69, 0x64, 0x22, 0x00, 0x12,
	0x33, 0x0a, 0x0f, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x43, 0x72, 0x69, 0x74, 0x65, 0x72, 0x69,
	0x6f, 0x6e, 0x12, 0x14, 0x2e, 0x71, 0x66, 0x2e, 0x47, 0x72, 0x61, 0x64, 0x69, 0x6e, 0x67, 0x43,
	0x72, 0x69, 0x74, 0x65, 0x72, 0x69, 0x6f, 0x6e, 0x1a, 0x08, 0x2e, 0x71, 0x66, 0x2e, 0x56, 0x6f,
	0x69, 0x64, 0x22, 0x00, 0x12, 0x2f, 0x0a, 0x0c, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x52, 0x65,
	0x76, 0x69, 0x65, 0x77, 0x12, 0x11, 0x2e, 0x71, 0x66, 0x2e, 0x52, 0x65, 0x76, 0x69, 0x65, 0x77,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0a, 0x2e, 0x71, 0x66, 0x2e, 0x52, 0x65, 0x76,
	0x69, 0x65, 0x77, 0x22, 0x00, 0x12, 0x2f, 0x0a, 0x0c, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52,
	0x65, 0x76, 0x69, 0x65, 0x77, 0x12, 0x11, 0x2e, 0x71, 0x66, 0x2e, 0x52, 0x65, 0x76, 0x69, 0x65,
	0x77, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0a, 0x2e, 0x71, 0x66, 0x2e, 0x52, 0x65,
	0x76, 0x69, 0x65, 0x77, 0x22, 0x00, 0x12, 0x3f, 0x0a, 0x0c, 0x47, 0x65, 0x74, 0x52, 0x65, 0x76,
	0x69, 0x65, 0x77, 0x65, 0x72, 0x73, 0x12, 0x1e, 0x2e, 0x71, 0x66, 0x2e, 0x53, 0x75, 0x62, 0x6d,
	0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x76, 0x69, 0x65, 0x77, 0x65, 0x72, 0x73, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0d, 0x2e, 0x71, 0x66, 0x2e, 0x52, 0x65, 0x76, 0x69,
	0x65, 0x77, 0x65, 0x72, 0x73, 0x22, 0x00, 0x12, 0x35, 0x0a, 0x0f, 0x47, 0x65, 0x74, 0x4f, 0x72,
	0x67, 0x61, 0x6e, 0x69, 0x7a, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x0e, 0x2e, 0x71, 0x66, 0x2e,
	0x4f, 0x72, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x10, 0x2e, 0x71, 0x66, 0x2e,
	0x4f, 0x72, 0x67, 0x61, 0x6e, 0x69, 0x7a, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x00, 0x12, 0x35,
	0x0a, 0x0f, 0x47, 0x65, 0x74, 0x52, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x65,
	0x73, 0x12, 0x0e, 0x2e, 0x71, 0x66, 0x2e, 0x55, 0x52, 0x4c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x10, 0x2e, 0x71, 0x66, 0x2e, 0x52, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x6f, 0x72,
	0x69, 0x65, 0x73, 0x22, 0x00, 0x12, 0x30, 0x0a, 0x0b, 0x49, 0x73, 0x45, 0x6d, 0x70, 0x74, 0x79,
	0x52, 0x65, 0x70, 0x6f, 0x12, 0x15, 0x2e, 0x71, 0x66, 0x2e, 0x52, 0x65, 0x70, 0x6f, 0x73, 0x69,
	0x74, 0x6f, 0x72, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x08, 0x2e, 0x71, 0x66,
	0x2e, 0x56, 0x6f, 0x69, 0x64, 0x22, 0x00, 0x42, 0x26, 0x5a, 0x21, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x71, 0x75, 0x69, 0x63, 0x6b, 0x66, 0x65, 0x65, 0x64, 0x2f,
	0x71, 0x75, 0x69, 0x63, 0x6b, 0x66, 0x65, 0x65, 0x64, 0x2f, 0x71, 0x66, 0xba, 0x02, 0x00, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var file_qf_quickfeed_proto_goTypes = []interface{}{
	(*Void)(nil),                        // 0: qf.Void
	(*CourseUserRequest)(nil),           // 1: qf.CourseUserRequest
	(*User)(nil),                        // 2: qf.User
	(*GetGroupRequest)(nil),             // 3: qf.GetGroupRequest
	(*GroupRequest)(nil),                // 4: qf.GroupRequest
	(*CourseRequest)(nil),               // 5: qf.CourseRequest
	(*Group)(nil),                       // 6: qf.Group
	(*EnrollmentStatusRequest)(nil),     // 7: qf.EnrollmentStatusRequest
	(*Course)(nil),                      // 8: qf.Course
	(*Enrollment)(nil),                  // 9: qf.Enrollment
	(*EnrollmentRequest)(nil),           // 10: qf.EnrollmentRequest
	(*Enrollments)(nil),                 // 11: qf.Enrollments
	(*SubmissionRequest)(nil),           // 12: qf.SubmissionRequest
	(*SubmissionsForCourseRequest)(nil), // 13: qf.SubmissionsForCourseRequest
	(*UpdateSubmissionRequest)(nil),     // 14: qf.UpdateSubmissionRequest
	(*UpdateSubmissionsRequest)(nil),    // 15: qf.UpdateSubmissionsRequest
	(*RebuildRequest)(nil),              // 16: qf.RebuildRequest
	(*GradingBenchmark)(nil),            // 17: qf.GradingBenchmark
	(*GradingCriterion)(nil),            // 18: qf.GradingCriterion
	(*ReviewRequest)(nil),               // 19: qf.ReviewRequest
	(*SubmissionReviewersRequest)(nil),  // 20: qf.SubmissionReviewersRequest
	(*OrgRequest)(nil),                  // 21: qf.OrgRequest
	(*URLRequest)(nil),                  // 22: qf.URLRequest
	(*RepositoryRequest)(nil),           // 23: qf.RepositoryRequest
	(*Users)(nil),                       // 24: qf.Users
	(*AuthorizationResponse)(nil),       // 25: qf.AuthorizationResponse
	(*Groups)(nil),                      // 26: qf.Groups
	(*Courses)(nil),                     // 27: qf.Courses
	(*Assignments)(nil),                 // 28: qf.Assignments
	(*Submissions)(nil),                 // 29: qf.Submissions
	(*CourseSubmissions)(nil),           // 30: qf.CourseSubmissions
	(*Review)(nil),                      // 31: qf.Review
	(*Reviewers)(nil),                   // 32: qf.Reviewers
	(*Organization)(nil),                // 33: qf.Organization
	(*Repositories)(nil),                // 34: qf.Repositories
}
var file_qf_quickfeed_proto_depIdxs = []int32{
	0,  // 0: qf.QuickFeedService.GetUser:input_type -> qf.Void
	0,  // 1: qf.QuickFeedService.GetUsers:input_type -> qf.Void
	1,  // 2: qf.QuickFeedService.GetUserByCourse:input_type -> qf.CourseUserRequest
	2,  // 3: qf.QuickFeedService.UpdateUser:input_type -> qf.User
	0,  // 4: qf.QuickFeedService.IsAuthorizedTeacher:input_type -> qf.Void
	3,  // 5: qf.QuickFeedService.GetGroup:input_type -> qf.GetGroupRequest
	4,  // 6: qf.QuickFeedService.GetGroupByUserAndCourse:input_type -> qf.GroupRequest
	5,  // 7: qf.QuickFeedService.GetGroupsByCourse:input_type -> qf.CourseRequest
	6,  // 8: qf.QuickFeedService.CreateGroup:input_type -> qf.Group
	6,  // 9: qf.QuickFeedService.UpdateGroup:input_type -> qf.Group
	4,  // 10: qf.QuickFeedService.DeleteGroup:input_type -> qf.GroupRequest
	5,  // 11: qf.QuickFeedService.GetCourse:input_type -> qf.CourseRequest
	0,  // 12: qf.QuickFeedService.GetCourses:input_type -> qf.Void
	7,  // 13: qf.QuickFeedService.GetCoursesByUser:input_type -> qf.EnrollmentStatusRequest
	8,  // 14: qf.QuickFeedService.CreateCourse:input_type -> qf.Course
	8,  // 15: qf.QuickFeedService.UpdateCourse:input_type -> qf.Course
	9,  // 16: qf.QuickFeedService.UpdateCourseVisibility:input_type -> qf.Enrollment
	5,  // 17: qf.QuickFeedService.GetAssignments:input_type -> qf.CourseRequest
	5,  // 18: qf.QuickFeedService.UpdateAssignments:input_type -> qf.CourseRequest
	7,  // 19: qf.QuickFeedService.GetEnrollmentsByUser:input_type -> qf.EnrollmentStatusRequest
	10, // 20: qf.QuickFeedService.GetEnrollmentsByCourse:input_type -> qf.EnrollmentRequest
	9,  // 21: qf.QuickFeedService.CreateEnrollment:input_type -> qf.Enrollment
	11, // 22: qf.QuickFeedService.UpdateEnrollments:input_type -> qf.Enrollments
	12, // 23: qf.QuickFeedService.GetSubmissions:input_type -> qf.SubmissionRequest
	13, // 24: qf.QuickFeedService.GetSubmissionsByCourse:input_type -> qf.SubmissionsForCourseRequest
	14, // 25: qf.QuickFeedService.UpdateSubmission:input_type -> qf.UpdateSubmissionRequest
	15, // 26: qf.QuickFeedService.UpdateSubmissions:input_type -> qf.UpdateSubmissionsRequest
	16, // 27: qf.QuickFeedService.RebuildSubmissions:input_type -> qf.RebuildRequest
	17, // 28: qf.QuickFeedService.CreateBenchmark:input_type -> qf.GradingBenchmark
	17, // 29: qf.QuickFeedService.UpdateBenchmark:input_type -> qf.GradingBenchmark
	17, // 30: qf.QuickFeedService.DeleteBenchmark:input_type -> qf.GradingBenchmark
	18, // 31: qf.QuickFeedService.CreateCriterion:input_type -> qf.GradingCriterion
	18, // 32: qf.QuickFeedService.UpdateCriterion:input_type -> qf.GradingCriterion
	18, // 33: qf.QuickFeedService.DeleteCriterion:input_type -> qf.GradingCriterion
	19, // 34: qf.QuickFeedService.CreateReview:input_type -> qf.ReviewRequest
	19, // 35: qf.QuickFeedService.UpdateReview:input_type -> qf.ReviewRequest
	20, // 36: qf.QuickFeedService.GetReviewers:input_type -> qf.SubmissionReviewersRequest
	21, // 37: qf.QuickFeedService.GetOrganization:input_type -> qf.OrgRequest
	22, // 38: qf.QuickFeedService.GetRepositories:input_type -> qf.URLRequest
	23, // 39: qf.QuickFeedService.IsEmptyRepo:input_type -> qf.RepositoryRequest
	2,  // 40: qf.QuickFeedService.GetUser:output_type -> qf.User
	24, // 41: qf.QuickFeedService.GetUsers:output_type -> qf.Users
	2,  // 42: qf.QuickFeedService.GetUserByCourse:output_type -> qf.User
	0,  // 43: qf.QuickFeedService.UpdateUser:output_type -> qf.Void
	25, // 44: qf.QuickFeedService.IsAuthorizedTeacher:output_type -> qf.AuthorizationResponse
	6,  // 45: qf.QuickFeedService.GetGroup:output_type -> qf.Group
	6,  // 46: qf.QuickFeedService.GetGroupByUserAndCourse:output_type -> qf.Group
	26, // 47: qf.QuickFeedService.GetGroupsByCourse:output_type -> qf.Groups
	6,  // 48: qf.QuickFeedService.CreateGroup:output_type -> qf.Group
	6,  // 49: qf.QuickFeedService.UpdateGroup:output_type -> qf.Group
	0,  // 50: qf.QuickFeedService.DeleteGroup:output_type -> qf.Void
	8,  // 51: qf.QuickFeedService.GetCourse:output_type -> qf.Course
	27, // 52: qf.QuickFeedService.GetCourses:output_type -> qf.Courses
	27, // 53: qf.QuickFeedService.GetCoursesByUser:output_type -> qf.Courses
	8,  // 54: qf.QuickFeedService.CreateCourse:output_type -> qf.Course
	0,  // 55: qf.QuickFeedService.UpdateCourse:output_type -> qf.Void
	0,  // 56: qf.QuickFeedService.UpdateCourseVisibility:output_type -> qf.Void
	28, // 57: qf.QuickFeedService.GetAssignments:output_type -> qf.Assignments
	0,  // 58: qf.QuickFeedService.UpdateAssignments:output_type -> qf.Void
	11, // 59: qf.QuickFeedService.GetEnrollmentsByUser:output_type -> qf.Enrollments
	11, // 60: qf.QuickFeedService.GetEnrollmentsByCourse:output_type -> qf.Enrollments
	0,  // 61: qf.QuickFeedService.CreateEnrollment:output_type -> qf.Void
	0,  // 62: qf.QuickFeedService.UpdateEnrollments:output_type -> qf.Void
	29, // 63: qf.QuickFeedService.GetSubmissions:output_type -> qf.Submissions
	30, // 64: qf.QuickFeedService.GetSubmissionsByCourse:output_type -> qf.CourseSubmissions
	0,  // 65: qf.QuickFeedService.UpdateSubmission:output_type -> qf.Void
	0,  // 66: qf.QuickFeedService.UpdateSubmissions:output_type -> qf.Void
	0,  // 67: qf.QuickFeedService.RebuildSubmissions:output_type -> qf.Void
	17, // 68: qf.QuickFeedService.CreateBenchmark:output_type -> qf.GradingBenchmark
	0,  // 69: qf.QuickFeedService.UpdateBenchmark:output_type -> qf.Void
	0,  // 70: qf.QuickFeedService.DeleteBenchmark:output_type -> qf.Void
	18, // 71: qf.QuickFeedService.CreateCriterion:output_type -> qf.GradingCriterion
	0,  // 72: qf.QuickFeedService.UpdateCriterion:output_type -> qf.Void
	0,  // 73: qf.QuickFeedService.DeleteCriterion:output_type -> qf.Void
	31, // 74: qf.QuickFeedService.CreateReview:output_type -> qf.Review
	31, // 75: qf.QuickFeedService.UpdateReview:output_type -> qf.Review
	32, // 76: qf.QuickFeedService.GetReviewers:output_type -> qf.Reviewers
	33, // 77: qf.QuickFeedService.GetOrganization:output_type -> qf.Organization
	34, // 78: qf.QuickFeedService.GetRepositories:output_type -> qf.Repositories
	0,  // 79: qf.QuickFeedService.IsEmptyRepo:output_type -> qf.Void
	40, // [40:80] is the sub-list for method output_type
	0,  // [0:40] is the sub-list for method input_type
	0,  // [0:0] is the sub-list for extension type_name
	0,  // [0:0] is the sub-list for extension extendee
	0,  // [0:0] is the sub-list for field type_name
}

func init() { file_qf_quickfeed_proto_init() }
func file_qf_quickfeed_proto_init() {
	if File_qf_quickfeed_proto != nil {
		return
	}
	file_qf_types_proto_init()
	file_qf_requests_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_qf_quickfeed_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_qf_quickfeed_proto_goTypes,
		DependencyIndexes: file_qf_quickfeed_proto_depIdxs,
	}.Build()
	File_qf_quickfeed_proto = out.File
	file_qf_quickfeed_proto_rawDesc = nil
	file_qf_quickfeed_proto_goTypes = nil
	file_qf_quickfeed_proto_depIdxs = nil
}
