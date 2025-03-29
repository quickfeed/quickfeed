// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.5
// 	protoc        (unknown)
// source: qf/requests.proto

package qf

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type SubmissionRequest_SubmissionType int32

const (
	SubmissionRequest_ALL   SubmissionRequest_SubmissionType = 0 // fetch all submissions
	SubmissionRequest_USER  SubmissionRequest_SubmissionType = 1 // fetch all user submissions
	SubmissionRequest_GROUP SubmissionRequest_SubmissionType = 2 // fetch all group submissions
)

// Enum value maps for SubmissionRequest_SubmissionType.
var (
	SubmissionRequest_SubmissionType_name = map[int32]string{
		0: "ALL",
		1: "USER",
		2: "GROUP",
	}
	SubmissionRequest_SubmissionType_value = map[string]int32{
		"ALL":   0,
		"USER":  1,
		"GROUP": 2,
	}
)

func (x SubmissionRequest_SubmissionType) Enum() *SubmissionRequest_SubmissionType {
	p := new(SubmissionRequest_SubmissionType)
	*p = x
	return p
}

func (x SubmissionRequest_SubmissionType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (SubmissionRequest_SubmissionType) Descriptor() protoreflect.EnumDescriptor {
	return file_qf_requests_proto_enumTypes[0].Descriptor()
}

func (SubmissionRequest_SubmissionType) Type() protoreflect.EnumType {
	return &file_qf_requests_proto_enumTypes[0]
}

func (x SubmissionRequest_SubmissionType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use SubmissionRequest_SubmissionType.Descriptor instead.
func (SubmissionRequest_SubmissionType) EnumDescriptor() ([]byte, []int) {
	return file_qf_requests_proto_rawDescGZIP(), []int{6, 0}
}

type CourseSubmissions struct {
	state         protoimpl.MessageState  `protogen:"open.v1"`
	Submissions   map[uint64]*Submissions `protobuf:"bytes,1,rep,name=submissions,proto3" json:"submissions,omitempty" protobuf_key:"varint,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CourseSubmissions) Reset() {
	*x = CourseSubmissions{}
	mi := &file_qf_requests_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CourseSubmissions) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CourseSubmissions) ProtoMessage() {}

func (x *CourseSubmissions) ProtoReflect() protoreflect.Message {
	mi := &file_qf_requests_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CourseSubmissions.ProtoReflect.Descriptor instead.
func (*CourseSubmissions) Descriptor() ([]byte, []int) {
	return file_qf_requests_proto_rawDescGZIP(), []int{0}
}

func (x *CourseSubmissions) GetSubmissions() map[uint64]*Submissions {
	if x != nil {
		return x.Submissions
	}
	return nil
}

type ReviewRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	CourseID      uint64                 `protobuf:"varint,1,opt,name=courseID,proto3" json:"courseID,omitempty"`
	Review        *Review                `protobuf:"bytes,2,opt,name=review,proto3" json:"review,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ReviewRequest) Reset() {
	*x = ReviewRequest{}
	mi := &file_qf_requests_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ReviewRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReviewRequest) ProtoMessage() {}

func (x *ReviewRequest) ProtoReflect() protoreflect.Message {
	mi := &file_qf_requests_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReviewRequest.ProtoReflect.Descriptor instead.
func (*ReviewRequest) Descriptor() ([]byte, []int) {
	return file_qf_requests_proto_rawDescGZIP(), []int{1}
}

func (x *ReviewRequest) GetCourseID() uint64 {
	if x != nil {
		return x.CourseID
	}
	return 0
}

func (x *ReviewRequest) GetReview() *Review {
	if x != nil {
		return x.Review
	}
	return nil
}

type CourseRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	CourseID      uint64                 `protobuf:"varint,1,opt,name=courseID,proto3" json:"courseID,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CourseRequest) Reset() {
	*x = CourseRequest{}
	mi := &file_qf_requests_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CourseRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CourseRequest) ProtoMessage() {}

func (x *CourseRequest) ProtoReflect() protoreflect.Message {
	mi := &file_qf_requests_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CourseRequest.ProtoReflect.Descriptor instead.
func (*CourseRequest) Descriptor() ([]byte, []int) {
	return file_qf_requests_proto_rawDescGZIP(), []int{2}
}

func (x *CourseRequest) GetCourseID() uint64 {
	if x != nil {
		return x.CourseID
	}
	return 0
}

type GroupRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	CourseID      uint64                 `protobuf:"varint,1,opt,name=courseID,proto3" json:"courseID,omitempty"`
	UserID        uint64                 `protobuf:"varint,2,opt,name=userID,proto3" json:"userID,omitempty"`
	GroupID       uint64                 `protobuf:"varint,3,opt,name=groupID,proto3" json:"groupID,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GroupRequest) Reset() {
	*x = GroupRequest{}
	mi := &file_qf_requests_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GroupRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GroupRequest) ProtoMessage() {}

func (x *GroupRequest) ProtoReflect() protoreflect.Message {
	mi := &file_qf_requests_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GroupRequest.ProtoReflect.Descriptor instead.
func (*GroupRequest) Descriptor() ([]byte, []int) {
	return file_qf_requests_proto_rawDescGZIP(), []int{3}
}

func (x *GroupRequest) GetCourseID() uint64 {
	if x != nil {
		return x.CourseID
	}
	return 0
}

func (x *GroupRequest) GetUserID() uint64 {
	if x != nil {
		return x.UserID
	}
	return 0
}

func (x *GroupRequest) GetGroupID() uint64 {
	if x != nil {
		return x.GroupID
	}
	return 0
}

type Organization struct {
	state               protoimpl.MessageState `protogen:"open.v1"`
	ScmOrganizationID   uint64                 `protobuf:"varint,1,opt,name=ScmOrganizationID,proto3" json:"ScmOrganizationID,omitempty"`
	ScmOrganizationName string                 `protobuf:"bytes,2,opt,name=ScmOrganizationName,proto3" json:"ScmOrganizationName,omitempty"`
	unknownFields       protoimpl.UnknownFields
	sizeCache           protoimpl.SizeCache
}

func (x *Organization) Reset() {
	*x = Organization{}
	mi := &file_qf_requests_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Organization) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Organization) ProtoMessage() {}

func (x *Organization) ProtoReflect() protoreflect.Message {
	mi := &file_qf_requests_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Organization.ProtoReflect.Descriptor instead.
func (*Organization) Descriptor() ([]byte, []int) {
	return file_qf_requests_proto_rawDescGZIP(), []int{4}
}

func (x *Organization) GetScmOrganizationID() uint64 {
	if x != nil {
		return x.ScmOrganizationID
	}
	return 0
}

func (x *Organization) GetScmOrganizationName() string {
	if x != nil {
		return x.ScmOrganizationName
	}
	return ""
}

type EnrollmentRequest struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Types that are valid to be assigned to FetchMode:
	//
	//	*EnrollmentRequest_CourseID
	//	*EnrollmentRequest_UserID
	FetchMode     isEnrollmentRequest_FetchMode `protobuf_oneof:"FetchMode"`
	Statuses      []Enrollment_UserStatus       `protobuf:"varint,3,rep,packed,name=statuses,proto3,enum=qf.Enrollment_UserStatus" json:"statuses,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *EnrollmentRequest) Reset() {
	*x = EnrollmentRequest{}
	mi := &file_qf_requests_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *EnrollmentRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EnrollmentRequest) ProtoMessage() {}

func (x *EnrollmentRequest) ProtoReflect() protoreflect.Message {
	mi := &file_qf_requests_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EnrollmentRequest.ProtoReflect.Descriptor instead.
func (*EnrollmentRequest) Descriptor() ([]byte, []int) {
	return file_qf_requests_proto_rawDescGZIP(), []int{5}
}

func (x *EnrollmentRequest) GetFetchMode() isEnrollmentRequest_FetchMode {
	if x != nil {
		return x.FetchMode
	}
	return nil
}

func (x *EnrollmentRequest) GetCourseID() uint64 {
	if x != nil {
		if x, ok := x.FetchMode.(*EnrollmentRequest_CourseID); ok {
			return x.CourseID
		}
	}
	return 0
}

func (x *EnrollmentRequest) GetUserID() uint64 {
	if x != nil {
		if x, ok := x.FetchMode.(*EnrollmentRequest_UserID); ok {
			return x.UserID
		}
	}
	return 0
}

func (x *EnrollmentRequest) GetStatuses() []Enrollment_UserStatus {
	if x != nil {
		return x.Statuses
	}
	return nil
}

type isEnrollmentRequest_FetchMode interface {
	isEnrollmentRequest_FetchMode()
}

type EnrollmentRequest_CourseID struct {
	CourseID uint64 `protobuf:"varint,1,opt,name=courseID,proto3,oneof"`
}

type EnrollmentRequest_UserID struct {
	UserID uint64 `protobuf:"varint,2,opt,name=userID,proto3,oneof"`
}

func (*EnrollmentRequest_CourseID) isEnrollmentRequest_FetchMode() {}

func (*EnrollmentRequest_UserID) isEnrollmentRequest_FetchMode() {}

type SubmissionRequest struct {
	state        protoimpl.MessageState `protogen:"open.v1"`
	CourseID     uint64                 `protobuf:"varint,1,opt,name=CourseID,proto3" json:"CourseID,omitempty"`
	AssignmentID uint64                 `protobuf:"varint,2,opt,name=AssignmentID,proto3" json:"AssignmentID,omitempty"` // only used for user and group submissions
	// Types that are valid to be assigned to FetchMode:
	//
	//	*SubmissionRequest_UserID
	//	*SubmissionRequest_GroupID
	//	*SubmissionRequest_SubmissionID
	//	*SubmissionRequest_Type
	FetchMode     isSubmissionRequest_FetchMode `protobuf_oneof:"FetchMode"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SubmissionRequest) Reset() {
	*x = SubmissionRequest{}
	mi := &file_qf_requests_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SubmissionRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SubmissionRequest) ProtoMessage() {}

func (x *SubmissionRequest) ProtoReflect() protoreflect.Message {
	mi := &file_qf_requests_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SubmissionRequest.ProtoReflect.Descriptor instead.
func (*SubmissionRequest) Descriptor() ([]byte, []int) {
	return file_qf_requests_proto_rawDescGZIP(), []int{6}
}

func (x *SubmissionRequest) GetCourseID() uint64 {
	if x != nil {
		return x.CourseID
	}
	return 0
}

func (x *SubmissionRequest) GetAssignmentID() uint64 {
	if x != nil {
		return x.AssignmentID
	}
	return 0
}

func (x *SubmissionRequest) GetFetchMode() isSubmissionRequest_FetchMode {
	if x != nil {
		return x.FetchMode
	}
	return nil
}

func (x *SubmissionRequest) GetUserID() uint64 {
	if x != nil {
		if x, ok := x.FetchMode.(*SubmissionRequest_UserID); ok {
			return x.UserID
		}
	}
	return 0
}

func (x *SubmissionRequest) GetGroupID() uint64 {
	if x != nil {
		if x, ok := x.FetchMode.(*SubmissionRequest_GroupID); ok {
			return x.GroupID
		}
	}
	return 0
}

func (x *SubmissionRequest) GetSubmissionID() uint64 {
	if x != nil {
		if x, ok := x.FetchMode.(*SubmissionRequest_SubmissionID); ok {
			return x.SubmissionID
		}
	}
	return 0
}

func (x *SubmissionRequest) GetType() SubmissionRequest_SubmissionType {
	if x != nil {
		if x, ok := x.FetchMode.(*SubmissionRequest_Type); ok {
			return x.Type
		}
	}
	return SubmissionRequest_ALL
}

type isSubmissionRequest_FetchMode interface {
	isSubmissionRequest_FetchMode()
}

type SubmissionRequest_UserID struct {
	UserID uint64 `protobuf:"varint,3,opt,name=UserID,proto3,oneof"` // fetch single user's submissions with build info
}

type SubmissionRequest_GroupID struct {
	GroupID uint64 `protobuf:"varint,4,opt,name=GroupID,proto3,oneof"` // fetch single group's submissions with build info
}

type SubmissionRequest_SubmissionID struct {
	SubmissionID uint64 `protobuf:"varint,5,opt,name=SubmissionID,proto3,oneof"` // fetch single specific submission with build info
}

type SubmissionRequest_Type struct {
	Type SubmissionRequest_SubmissionType `protobuf:"varint,6,opt,name=Type,proto3,enum=qf.SubmissionRequest_SubmissionType,oneof"` // fetch all submissions of given type without build info
}

func (*SubmissionRequest_UserID) isSubmissionRequest_FetchMode() {}

func (*SubmissionRequest_GroupID) isSubmissionRequest_FetchMode() {}

func (*SubmissionRequest_SubmissionID) isSubmissionRequest_FetchMode() {}

func (*SubmissionRequest_Type) isSubmissionRequest_FetchMode() {}

// UpdateSubmissionRequest is used to update manually reviewed submissions.
type UpdateSubmissionRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	CourseID      uint64                 `protobuf:"varint,1,opt,name=courseID,proto3" json:"courseID,omitempty"`
	AssignmentID  uint64                 `protobuf:"varint,2,opt,name=assignmentID,proto3" json:"assignmentID,omitempty"` // if non-zero, update all submissions
	SubmissionID  uint64                 `protobuf:"varint,3,opt,name=submissionID,proto3" json:"submissionID,omitempty"` // if non-zero, update single specific submission
	Score         uint32                 `protobuf:"varint,4,opt,name=score,proto3" json:"score,omitempty"`               // only used for single submission
	Release       bool                   `protobuf:"varint,5,opt,name=release,proto3" json:"release,omitempty"`           // indicate whether or not to release submission(s) to students
	Status        Submission_Status      `protobuf:"varint,6,opt,name=status,proto3,enum=qf.Submission_Status" json:"status,omitempty"`
	Grades        []*Grade               `protobuf:"bytes,7,rep,name=grades,proto3" json:"grades,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UpdateSubmissionRequest) Reset() {
	*x = UpdateSubmissionRequest{}
	mi := &file_qf_requests_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UpdateSubmissionRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateSubmissionRequest) ProtoMessage() {}

func (x *UpdateSubmissionRequest) ProtoReflect() protoreflect.Message {
	mi := &file_qf_requests_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateSubmissionRequest.ProtoReflect.Descriptor instead.
func (*UpdateSubmissionRequest) Descriptor() ([]byte, []int) {
	return file_qf_requests_proto_rawDescGZIP(), []int{7}
}

func (x *UpdateSubmissionRequest) GetCourseID() uint64 {
	if x != nil {
		return x.CourseID
	}
	return 0
}

func (x *UpdateSubmissionRequest) GetAssignmentID() uint64 {
	if x != nil {
		return x.AssignmentID
	}
	return 0
}

func (x *UpdateSubmissionRequest) GetSubmissionID() uint64 {
	if x != nil {
		return x.SubmissionID
	}
	return 0
}

func (x *UpdateSubmissionRequest) GetScore() uint32 {
	if x != nil {
		return x.Score
	}
	return 0
}

func (x *UpdateSubmissionRequest) GetRelease() bool {
	if x != nil {
		return x.Release
	}
	return false
}

func (x *UpdateSubmissionRequest) GetStatus() Submission_Status {
	if x != nil {
		return x.Status
	}
	return Submission_NONE
}

func (x *UpdateSubmissionRequest) GetGrades() []*Grade {
	if x != nil {
		return x.Grades
	}
	return nil
}

// used to check whether student/group submission repo is empty
type RepositoryRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	UserID        uint64                 `protobuf:"varint,1,opt,name=userID,proto3" json:"userID,omitempty"`
	GroupID       uint64                 `protobuf:"varint,2,opt,name=groupID,proto3" json:"groupID,omitempty"`
	CourseID      uint64                 `protobuf:"varint,3,opt,name=courseID,proto3" json:"courseID,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RepositoryRequest) Reset() {
	*x = RepositoryRequest{}
	mi := &file_qf_requests_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RepositoryRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RepositoryRequest) ProtoMessage() {}

func (x *RepositoryRequest) ProtoReflect() protoreflect.Message {
	mi := &file_qf_requests_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RepositoryRequest.ProtoReflect.Descriptor instead.
func (*RepositoryRequest) Descriptor() ([]byte, []int) {
	return file_qf_requests_proto_rawDescGZIP(), []int{8}
}

func (x *RepositoryRequest) GetUserID() uint64 {
	if x != nil {
		return x.UserID
	}
	return 0
}

func (x *RepositoryRequest) GetGroupID() uint64 {
	if x != nil {
		return x.GroupID
	}
	return 0
}

func (x *RepositoryRequest) GetCourseID() uint64 {
	if x != nil {
		return x.CourseID
	}
	return 0
}

type Repositories struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	URLs          map[uint32]string      `protobuf:"bytes,1,rep,name=URLs,proto3" json:"URLs,omitempty" protobuf_key:"varint,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"` // Map key is the Repository.Type
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Repositories) Reset() {
	*x = Repositories{}
	mi := &file_qf_requests_proto_msgTypes[9]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Repositories) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Repositories) ProtoMessage() {}

func (x *Repositories) ProtoReflect() protoreflect.Message {
	mi := &file_qf_requests_proto_msgTypes[9]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Repositories.ProtoReflect.Descriptor instead.
func (*Repositories) Descriptor() ([]byte, []int) {
	return file_qf_requests_proto_rawDescGZIP(), []int{9}
}

func (x *Repositories) GetURLs() map[uint32]string {
	if x != nil {
		return x.URLs
	}
	return nil
}

type RebuildRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	CourseID      uint64                 `protobuf:"varint,1,opt,name=courseID,proto3" json:"courseID,omitempty"`
	AssignmentID  uint64                 `protobuf:"varint,2,opt,name=assignmentID,proto3" json:"assignmentID,omitempty"`
	SubmissionID  uint64                 `protobuf:"varint,3,opt,name=submissionID,proto3" json:"submissionID,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RebuildRequest) Reset() {
	*x = RebuildRequest{}
	mi := &file_qf_requests_proto_msgTypes[10]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RebuildRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RebuildRequest) ProtoMessage() {}

func (x *RebuildRequest) ProtoReflect() protoreflect.Message {
	mi := &file_qf_requests_proto_msgTypes[10]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RebuildRequest.ProtoReflect.Descriptor instead.
func (*RebuildRequest) Descriptor() ([]byte, []int) {
	return file_qf_requests_proto_rawDescGZIP(), []int{10}
}

func (x *RebuildRequest) GetCourseID() uint64 {
	if x != nil {
		return x.CourseID
	}
	return 0
}

func (x *RebuildRequest) GetAssignmentID() uint64 {
	if x != nil {
		return x.AssignmentID
	}
	return 0
}

func (x *RebuildRequest) GetSubmissionID() uint64 {
	if x != nil {
		return x.SubmissionID
	}
	return 0
}

type Void struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Void) Reset() {
	*x = Void{}
	mi := &file_qf_requests_proto_msgTypes[11]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Void) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Void) ProtoMessage() {}

func (x *Void) ProtoReflect() protoreflect.Message {
	mi := &file_qf_requests_proto_msgTypes[11]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Void.ProtoReflect.Descriptor instead.
func (*Void) Descriptor() ([]byte, []int) {
	return file_qf_requests_proto_rawDescGZIP(), []int{11}
}

var File_qf_requests_proto protoreflect.FileDescriptor

var file_qf_requests_proto_rawDesc = string([]byte{
	0x0a, 0x11, 0x71, 0x66, 0x2f, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x73, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x02, 0x71, 0x66, 0x1a, 0x0e, 0x71, 0x66, 0x2f, 0x74, 0x79, 0x70, 0x65,
	0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xae, 0x01, 0x0a, 0x11, 0x43, 0x6f, 0x75, 0x72,
	0x73, 0x65, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x48, 0x0a,
	0x0b, 0x73, 0x75, 0x62, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x01, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x26, 0x2e, 0x71, 0x66, 0x2e, 0x43, 0x6f, 0x75, 0x72, 0x73, 0x65, 0x53, 0x75,
	0x62, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x73,
	0x73, 0x69, 0x6f, 0x6e, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x0b, 0x73, 0x75, 0x62, 0x6d,
	0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x1a, 0x4f, 0x0a, 0x10, 0x53, 0x75, 0x62, 0x6d, 0x69,
	0x73, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b,
	0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x25, 0x0a,
	0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x71,
	0x66, 0x2e, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x4f, 0x0a, 0x0d, 0x52, 0x65, 0x76, 0x69,
	0x65, 0x77, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x63, 0x6f, 0x75,
	0x72, 0x73, 0x65, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x08, 0x63, 0x6f, 0x75,
	0x72, 0x73, 0x65, 0x49, 0x44, 0x12, 0x22, 0x0a, 0x06, 0x72, 0x65, 0x76, 0x69, 0x65, 0x77, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0a, 0x2e, 0x71, 0x66, 0x2e, 0x52, 0x65, 0x76, 0x69, 0x65,
	0x77, 0x52, 0x06, 0x72, 0x65, 0x76, 0x69, 0x65, 0x77, 0x22, 0x2b, 0x0a, 0x0d, 0x43, 0x6f, 0x75,
	0x72, 0x73, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x63, 0x6f,
	0x75, 0x72, 0x73, 0x65, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x08, 0x63, 0x6f,
	0x75, 0x72, 0x73, 0x65, 0x49, 0x44, 0x22, 0x5c, 0x0a, 0x0c, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x63, 0x6f, 0x75, 0x72, 0x73, 0x65,
	0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x08, 0x63, 0x6f, 0x75, 0x72, 0x73, 0x65,
	0x49, 0x44, 0x12, 0x16, 0x0a, 0x06, 0x75, 0x73, 0x65, 0x72, 0x49, 0x44, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x04, 0x52, 0x06, 0x75, 0x73, 0x65, 0x72, 0x49, 0x44, 0x12, 0x18, 0x0a, 0x07, 0x67, 0x72,
	0x6f, 0x75, 0x70, 0x49, 0x44, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x07, 0x67, 0x72, 0x6f,
	0x75, 0x70, 0x49, 0x44, 0x22, 0x6e, 0x0a, 0x0c, 0x4f, 0x72, 0x67, 0x61, 0x6e, 0x69, 0x7a, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x12, 0x2c, 0x0a, 0x11, 0x53, 0x63, 0x6d, 0x4f, 0x72, 0x67, 0x61, 0x6e,
	0x69, 0x7a, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52,
	0x11, 0x53, 0x63, 0x6d, 0x4f, 0x72, 0x67, 0x61, 0x6e, 0x69, 0x7a, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x49, 0x44, 0x12, 0x30, 0x0a, 0x13, 0x53, 0x63, 0x6d, 0x4f, 0x72, 0x67, 0x61, 0x6e, 0x69, 0x7a,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x13, 0x53, 0x63, 0x6d, 0x4f, 0x72, 0x67, 0x61, 0x6e, 0x69, 0x7a, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x4e, 0x61, 0x6d, 0x65, 0x22, 0x8f, 0x01, 0x0a, 0x11, 0x45, 0x6e, 0x72, 0x6f, 0x6c, 0x6c, 0x6d,
	0x65, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1c, 0x0a, 0x08, 0x63, 0x6f,
	0x75, 0x72, 0x73, 0x65, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x48, 0x00, 0x52, 0x08,
	0x63, 0x6f, 0x75, 0x72, 0x73, 0x65, 0x49, 0x44, 0x12, 0x18, 0x0a, 0x06, 0x75, 0x73, 0x65, 0x72,
	0x49, 0x44, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x48, 0x00, 0x52, 0x06, 0x75, 0x73, 0x65, 0x72,
	0x49, 0x44, 0x12, 0x35, 0x0a, 0x08, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x65, 0x73, 0x18, 0x03,
	0x20, 0x03, 0x28, 0x0e, 0x32, 0x19, 0x2e, 0x71, 0x66, 0x2e, 0x45, 0x6e, 0x72, 0x6f, 0x6c, 0x6c,
	0x6d, 0x65, 0x6e, 0x74, 0x2e, 0x55, 0x73, 0x65, 0x72, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52,
	0x08, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x65, 0x73, 0x42, 0x0b, 0x0a, 0x09, 0x46, 0x65, 0x74,
	0x63, 0x68, 0x4d, 0x6f, 0x64, 0x65, 0x22, 0xa8, 0x02, 0x0a, 0x11, 0x53, 0x75, 0x62, 0x6d, 0x69,
	0x73, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1a, 0x0a, 0x08,
	0x43, 0x6f, 0x75, 0x72, 0x73, 0x65, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x08,
	0x43, 0x6f, 0x75, 0x72, 0x73, 0x65, 0x49, 0x44, 0x12, 0x22, 0x0a, 0x0c, 0x41, 0x73, 0x73, 0x69,
	0x67, 0x6e, 0x6d, 0x65, 0x6e, 0x74, 0x49, 0x44, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0c,
	0x41, 0x73, 0x73, 0x69, 0x67, 0x6e, 0x6d, 0x65, 0x6e, 0x74, 0x49, 0x44, 0x12, 0x18, 0x0a, 0x06,
	0x55, 0x73, 0x65, 0x72, 0x49, 0x44, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04, 0x48, 0x00, 0x52, 0x06,
	0x55, 0x73, 0x65, 0x72, 0x49, 0x44, 0x12, 0x1a, 0x0a, 0x07, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x49,
	0x44, 0x18, 0x04, 0x20, 0x01, 0x28, 0x04, 0x48, 0x00, 0x52, 0x07, 0x47, 0x72, 0x6f, 0x75, 0x70,
	0x49, 0x44, 0x12, 0x24, 0x0a, 0x0c, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e,
	0x49, 0x44, 0x18, 0x05, 0x20, 0x01, 0x28, 0x04, 0x48, 0x00, 0x52, 0x0c, 0x53, 0x75, 0x62, 0x6d,
	0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x49, 0x44, 0x12, 0x3a, 0x0a, 0x04, 0x54, 0x79, 0x70, 0x65,
	0x18, 0x06, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x24, 0x2e, 0x71, 0x66, 0x2e, 0x53, 0x75, 0x62, 0x6d,
	0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x53, 0x75,
	0x62, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x54, 0x79, 0x70, 0x65, 0x48, 0x00, 0x52, 0x04,
	0x54, 0x79, 0x70, 0x65, 0x22, 0x2e, 0x0a, 0x0e, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x73, 0x73, 0x69,
	0x6f, 0x6e, 0x54, 0x79, 0x70, 0x65, 0x12, 0x07, 0x0a, 0x03, 0x41, 0x4c, 0x4c, 0x10, 0x00, 0x12,
	0x08, 0x0a, 0x04, 0x55, 0x53, 0x45, 0x52, 0x10, 0x01, 0x12, 0x09, 0x0a, 0x05, 0x47, 0x52, 0x4f,
	0x55, 0x50, 0x10, 0x02, 0x42, 0x0b, 0x0a, 0x09, 0x46, 0x65, 0x74, 0x63, 0x68, 0x4d, 0x6f, 0x64,
	0x65, 0x22, 0xff, 0x01, 0x0a, 0x17, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x53, 0x75, 0x62, 0x6d,
	0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1a, 0x0a,
	0x08, 0x63, 0x6f, 0x75, 0x72, 0x73, 0x65, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52,
	0x08, 0x63, 0x6f, 0x75, 0x72, 0x73, 0x65, 0x49, 0x44, 0x12, 0x22, 0x0a, 0x0c, 0x61, 0x73, 0x73,
	0x69, 0x67, 0x6e, 0x6d, 0x65, 0x6e, 0x74, 0x49, 0x44, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52,
	0x0c, 0x61, 0x73, 0x73, 0x69, 0x67, 0x6e, 0x6d, 0x65, 0x6e, 0x74, 0x49, 0x44, 0x12, 0x22, 0x0a,
	0x0c, 0x73, 0x75, 0x62, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x49, 0x44, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x04, 0x52, 0x0c, 0x73, 0x75, 0x62, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x49,
	0x44, 0x12, 0x14, 0x0a, 0x05, 0x73, 0x63, 0x6f, 0x72, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0d,
	0x52, 0x05, 0x73, 0x63, 0x6f, 0x72, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x72, 0x65, 0x6c, 0x65, 0x61,
	0x73, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x72, 0x65, 0x6c, 0x65, 0x61, 0x73,
	0x65, 0x12, 0x2d, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x06, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x15, 0x2e, 0x71, 0x66, 0x2e, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f,
	0x6e, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x12, 0x21, 0x0a, 0x06, 0x67, 0x72, 0x61, 0x64, 0x65, 0x73, 0x18, 0x07, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x09, 0x2e, 0x71, 0x66, 0x2e, 0x47, 0x72, 0x61, 0x64, 0x65, 0x52, 0x06, 0x67, 0x72, 0x61,
	0x64, 0x65, 0x73, 0x22, 0x61, 0x0a, 0x11, 0x52, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x6f, 0x72,
	0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x75, 0x73, 0x65, 0x72,
	0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x06, 0x75, 0x73, 0x65, 0x72, 0x49, 0x44,
	0x12, 0x18, 0x0a, 0x07, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x49, 0x44, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x04, 0x52, 0x07, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x49, 0x44, 0x12, 0x1a, 0x0a, 0x08, 0x63, 0x6f,
	0x75, 0x72, 0x73, 0x65, 0x49, 0x44, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x08, 0x63, 0x6f,
	0x75, 0x72, 0x73, 0x65, 0x49, 0x44, 0x22, 0x77, 0x0a, 0x0c, 0x52, 0x65, 0x70, 0x6f, 0x73, 0x69,
	0x74, 0x6f, 0x72, 0x69, 0x65, 0x73, 0x12, 0x2e, 0x0a, 0x04, 0x55, 0x52, 0x4c, 0x73, 0x18, 0x01,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x71, 0x66, 0x2e, 0x52, 0x65, 0x70, 0x6f, 0x73, 0x69,
	0x74, 0x6f, 0x72, 0x69, 0x65, 0x73, 0x2e, 0x55, 0x52, 0x4c, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79,
	0x52, 0x04, 0x55, 0x52, 0x4c, 0x73, 0x1a, 0x37, 0x0a, 0x09, 0x55, 0x52, 0x4c, 0x73, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d,
	0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22,
	0x74, 0x0a, 0x0e, 0x52, 0x65, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x1a, 0x0a, 0x08, 0x63, 0x6f, 0x75, 0x72, 0x73, 0x65, 0x49, 0x44, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x04, 0x52, 0x08, 0x63, 0x6f, 0x75, 0x72, 0x73, 0x65, 0x49, 0x44, 0x12, 0x22, 0x0a,
	0x0c, 0x61, 0x73, 0x73, 0x69, 0x67, 0x6e, 0x6d, 0x65, 0x6e, 0x74, 0x49, 0x44, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x04, 0x52, 0x0c, 0x61, 0x73, 0x73, 0x69, 0x67, 0x6e, 0x6d, 0x65, 0x6e, 0x74, 0x49,
	0x44, 0x12, 0x22, 0x0a, 0x0c, 0x73, 0x75, 0x62, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x49,
	0x44, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0c, 0x73, 0x75, 0x62, 0x6d, 0x69, 0x73, 0x73,
	0x69, 0x6f, 0x6e, 0x49, 0x44, 0x22, 0x06, 0x0a, 0x04, 0x56, 0x6f, 0x69, 0x64, 0x42, 0x26, 0x5a,
	0x21, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x71, 0x75, 0x69, 0x63,
	0x6b, 0x66, 0x65, 0x65, 0x64, 0x2f, 0x71, 0x75, 0x69, 0x63, 0x6b, 0x66, 0x65, 0x65, 0x64, 0x2f,
	0x71, 0x66, 0xba, 0x02, 0x00, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
})

var (
	file_qf_requests_proto_rawDescOnce sync.Once
	file_qf_requests_proto_rawDescData []byte
)

func file_qf_requests_proto_rawDescGZIP() []byte {
	file_qf_requests_proto_rawDescOnce.Do(func() {
		file_qf_requests_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_qf_requests_proto_rawDesc), len(file_qf_requests_proto_rawDesc)))
	})
	return file_qf_requests_proto_rawDescData
}

var file_qf_requests_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_qf_requests_proto_msgTypes = make([]protoimpl.MessageInfo, 14)
var file_qf_requests_proto_goTypes = []any{
	(SubmissionRequest_SubmissionType)(0), // 0: qf.SubmissionRequest.SubmissionType
	(*CourseSubmissions)(nil),             // 1: qf.CourseSubmissions
	(*ReviewRequest)(nil),                 // 2: qf.ReviewRequest
	(*CourseRequest)(nil),                 // 3: qf.CourseRequest
	(*GroupRequest)(nil),                  // 4: qf.GroupRequest
	(*Organization)(nil),                  // 5: qf.Organization
	(*EnrollmentRequest)(nil),             // 6: qf.EnrollmentRequest
	(*SubmissionRequest)(nil),             // 7: qf.SubmissionRequest
	(*UpdateSubmissionRequest)(nil),       // 8: qf.UpdateSubmissionRequest
	(*RepositoryRequest)(nil),             // 9: qf.RepositoryRequest
	(*Repositories)(nil),                  // 10: qf.Repositories
	(*RebuildRequest)(nil),                // 11: qf.RebuildRequest
	(*Void)(nil),                          // 12: qf.Void
	nil,                                   // 13: qf.CourseSubmissions.SubmissionsEntry
	nil,                                   // 14: qf.Repositories.URLsEntry
	(*Review)(nil),                        // 15: qf.Review
	(Enrollment_UserStatus)(0),            // 16: qf.Enrollment.UserStatus
	(Submission_Status)(0),                // 17: qf.Submission.Status
	(*Grade)(nil),                         // 18: qf.Grade
	(*Submissions)(nil),                   // 19: qf.Submissions
}
var file_qf_requests_proto_depIdxs = []int32{
	13, // 0: qf.CourseSubmissions.submissions:type_name -> qf.CourseSubmissions.SubmissionsEntry
	15, // 1: qf.ReviewRequest.review:type_name -> qf.Review
	16, // 2: qf.EnrollmentRequest.statuses:type_name -> qf.Enrollment.UserStatus
	0,  // 3: qf.SubmissionRequest.Type:type_name -> qf.SubmissionRequest.SubmissionType
	17, // 4: qf.UpdateSubmissionRequest.status:type_name -> qf.Submission.Status
	18, // 5: qf.UpdateSubmissionRequest.grades:type_name -> qf.Grade
	14, // 6: qf.Repositories.URLs:type_name -> qf.Repositories.URLsEntry
	19, // 7: qf.CourseSubmissions.SubmissionsEntry.value:type_name -> qf.Submissions
	8,  // [8:8] is the sub-list for method output_type
	8,  // [8:8] is the sub-list for method input_type
	8,  // [8:8] is the sub-list for extension type_name
	8,  // [8:8] is the sub-list for extension extendee
	0,  // [0:8] is the sub-list for field type_name
}

func init() { file_qf_requests_proto_init() }
func file_qf_requests_proto_init() {
	if File_qf_requests_proto != nil {
		return
	}
	file_qf_types_proto_init()
	file_qf_requests_proto_msgTypes[5].OneofWrappers = []any{
		(*EnrollmentRequest_CourseID)(nil),
		(*EnrollmentRequest_UserID)(nil),
	}
	file_qf_requests_proto_msgTypes[6].OneofWrappers = []any{
		(*SubmissionRequest_UserID)(nil),
		(*SubmissionRequest_GroupID)(nil),
		(*SubmissionRequest_SubmissionID)(nil),
		(*SubmissionRequest_Type)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_qf_requests_proto_rawDesc), len(file_qf_requests_proto_rawDesc)),
			NumEnums:      1,
			NumMessages:   14,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_qf_requests_proto_goTypes,
		DependencyIndexes: file_qf_requests_proto_depIdxs,
		EnumInfos:         file_qf_requests_proto_enumTypes,
		MessageInfos:      file_qf_requests_proto_msgTypes,
	}.Build()
	File_qf_requests_proto = out.File
	file_qf_requests_proto_goTypes = nil
	file_qf_requests_proto_depIdxs = nil
}
