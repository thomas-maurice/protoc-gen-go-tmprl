// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.1
// 	protoc        (unknown)
// source: temporal/v1/temporal.proto

package temporalv1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	descriptorpb "google.golang.org/protobuf/types/descriptorpb"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ActivityOptions struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Timeout from schedule to close - in seconds
	ScheduleToCloseTimeout *wrapperspb.Int32Value `protobuf:"bytes,2,opt,name=schedule_to_close_timeout,json=scheduleToCloseTimeout,proto3" json:"schedule_to_close_timeout,omitempty"`
	// Timeout from start to close - in seconds
	StartToCloseTimeout *wrapperspb.Int32Value `protobuf:"bytes,3,opt,name=start_to_close_timeout,json=startToCloseTimeout,proto3" json:"start_to_close_timeout,omitempty"`
	// Timeout from schedule to  - in seconds
	ScheduleToStartTimeout *wrapperspb.Int32Value `protobuf:"bytes,4,opt,name=schedule_to_start_timeout,json=scheduleToStartTimeout,proto3" json:"schedule_to_start_timeout,omitempty"`
	// Default retry policy
	RetryPolicy *RetryPolicy `protobuf:"bytes,5,opt,name=retry_policy,json=retryPolicy,proto3" json:"retry_policy,omitempty"`
}

func (x *ActivityOptions) Reset() {
	*x = ActivityOptions{}
	mi := &file_temporal_v1_temporal_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ActivityOptions) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ActivityOptions) ProtoMessage() {}

func (x *ActivityOptions) ProtoReflect() protoreflect.Message {
	mi := &file_temporal_v1_temporal_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ActivityOptions.ProtoReflect.Descriptor instead.
func (*ActivityOptions) Descriptor() ([]byte, []int) {
	return file_temporal_v1_temporal_proto_rawDescGZIP(), []int{0}
}

func (x *ActivityOptions) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *ActivityOptions) GetScheduleToCloseTimeout() *wrapperspb.Int32Value {
	if x != nil {
		return x.ScheduleToCloseTimeout
	}
	return nil
}

func (x *ActivityOptions) GetStartToCloseTimeout() *wrapperspb.Int32Value {
	if x != nil {
		return x.StartToCloseTimeout
	}
	return nil
}

func (x *ActivityOptions) GetScheduleToStartTimeout() *wrapperspb.Int32Value {
	if x != nil {
		return x.ScheduleToStartTimeout
	}
	return nil
}

func (x *ActivityOptions) GetRetryPolicy() *RetryPolicy {
	if x != nil {
		return x.RetryPolicy
	}
	return nil
}

type WorkflowOptions struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Execution timeout for the workflow - in seconds
	WorkflowExecutionTimeout *wrapperspb.Int32Value `protobuf:"bytes,2,opt,name=workflow_execution_timeout,json=workflowExecutionTimeout,proto3" json:"workflow_execution_timeout,omitempty"`
	// The timeout for duration of a single workflow run - in seconds
	WorkflowRunTimeout *wrapperspb.Int32Value `protobuf:"bytes,3,opt,name=workflow_run_timeout,json=workflowRunTimeout,proto3" json:"workflow_run_timeout,omitempty"`
	// The timeout for processing workflow task from the time the worker
	// pulled this task
	WorkflowTaskTimeout *wrapperspb.Int32Value `protobuf:"bytes,4,opt,name=workflow_task_timeout,json=workflowTaskTimeout,proto3" json:"workflow_task_timeout,omitempty"`
	// Default retry policy
	RetryPolicy *RetryPolicy `protobuf:"bytes,5,opt,name=retry_policy,json=retryPolicy,proto3" json:"retry_policy,omitempty"`
}

func (x *WorkflowOptions) Reset() {
	*x = WorkflowOptions{}
	mi := &file_temporal_v1_temporal_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *WorkflowOptions) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WorkflowOptions) ProtoMessage() {}

func (x *WorkflowOptions) ProtoReflect() protoreflect.Message {
	mi := &file_temporal_v1_temporal_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WorkflowOptions.ProtoReflect.Descriptor instead.
func (*WorkflowOptions) Descriptor() ([]byte, []int) {
	return file_temporal_v1_temporal_proto_rawDescGZIP(), []int{1}
}

func (x *WorkflowOptions) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *WorkflowOptions) GetWorkflowExecutionTimeout() *wrapperspb.Int32Value {
	if x != nil {
		return x.WorkflowExecutionTimeout
	}
	return nil
}

func (x *WorkflowOptions) GetWorkflowRunTimeout() *wrapperspb.Int32Value {
	if x != nil {
		return x.WorkflowRunTimeout
	}
	return nil
}

func (x *WorkflowOptions) GetWorkflowTaskTimeout() *wrapperspb.Int32Value {
	if x != nil {
		return x.WorkflowTaskTimeout
	}
	return nil
}

func (x *WorkflowOptions) GetRetryPolicy() *RetryPolicy {
	if x != nil {
		return x.RetryPolicy
	}
	return nil
}

type ServiceOptions struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TaskQueue string `protobuf:"bytes,1,opt,name=task_queue,json=taskQueue,proto3" json:"task_queue,omitempty"`
	// These will apply to all workflows unless defined otherwise
	// appart from the `name` one that is ignored here
	DefaultWorkflowOptions *WorkflowOptions `protobuf:"bytes,2,opt,name=default_workflow_options,json=defaultWorkflowOptions,proto3" json:"default_workflow_options,omitempty"`
	// These settings will apply to all activities unless defined otherwise
	// appart from the `name` one that is ignored here
	DefaultActivityOptions *ActivityOptions `protobuf:"bytes,3,opt,name=default_activity_options,json=defaultActivityOptions,proto3" json:"default_activity_options,omitempty"`
}

func (x *ServiceOptions) Reset() {
	*x = ServiceOptions{}
	mi := &file_temporal_v1_temporal_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ServiceOptions) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServiceOptions) ProtoMessage() {}

func (x *ServiceOptions) ProtoReflect() protoreflect.Message {
	mi := &file_temporal_v1_temporal_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServiceOptions.ProtoReflect.Descriptor instead.
func (*ServiceOptions) Descriptor() ([]byte, []int) {
	return file_temporal_v1_temporal_proto_rawDescGZIP(), []int{2}
}

func (x *ServiceOptions) GetTaskQueue() string {
	if x != nil {
		return x.TaskQueue
	}
	return ""
}

func (x *ServiceOptions) GetDefaultWorkflowOptions() *WorkflowOptions {
	if x != nil {
		return x.DefaultWorkflowOptions
	}
	return nil
}

func (x *ServiceOptions) GetDefaultActivityOptions() *ActivityOptions {
	if x != nil {
		return x.DefaultActivityOptions
	}
	return nil
}

type RetryPolicy struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Initial interval in seconds for the first retry
	InitialInterval *wrapperspb.Int32Value `protobuf:"bytes,1,opt,name=initial_interval,json=initialInterval,proto3" json:"initial_interval,omitempty"`
	// Backoff coefficient for exponential backoff
	BackoffCoefficient *wrapperspb.FloatValue `protobuf:"bytes,2,opt,name=backoff_coefficient,json=backoffCoefficient,proto3" json:"backoff_coefficient,omitempty"`
	// Max inteval between two retries
	MaximumInterval *wrapperspb.Int32Value `protobuf:"bytes,3,opt,name=maximum_interval,json=maximumInterval,proto3" json:"maximum_interval,omitempty"`
	// Maximum of attempts
	MaximumAttempts *wrapperspb.Int32Value `protobuf:"bytes,4,opt,name=maximum_attempts,json=maximumAttempts,proto3" json:"maximum_attempts,omitempty"`
	// Non retryable error types
	NonRetryableErrorTypes []*wrapperspb.StringValue `protobuf:"bytes,5,rep,name=non_retryable_error_types,json=nonRetryableErrorTypes,proto3" json:"non_retryable_error_types,omitempty"`
}

func (x *RetryPolicy) Reset() {
	*x = RetryPolicy{}
	mi := &file_temporal_v1_temporal_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RetryPolicy) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RetryPolicy) ProtoMessage() {}

func (x *RetryPolicy) ProtoReflect() protoreflect.Message {
	mi := &file_temporal_v1_temporal_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RetryPolicy.ProtoReflect.Descriptor instead.
func (*RetryPolicy) Descriptor() ([]byte, []int) {
	return file_temporal_v1_temporal_proto_rawDescGZIP(), []int{3}
}

func (x *RetryPolicy) GetInitialInterval() *wrapperspb.Int32Value {
	if x != nil {
		return x.InitialInterval
	}
	return nil
}

func (x *RetryPolicy) GetBackoffCoefficient() *wrapperspb.FloatValue {
	if x != nil {
		return x.BackoffCoefficient
	}
	return nil
}

func (x *RetryPolicy) GetMaximumInterval() *wrapperspb.Int32Value {
	if x != nil {
		return x.MaximumInterval
	}
	return nil
}

func (x *RetryPolicy) GetMaximumAttempts() *wrapperspb.Int32Value {
	if x != nil {
		return x.MaximumAttempts
	}
	return nil
}

func (x *RetryPolicy) GetNonRetryableErrorTypes() []*wrapperspb.StringValue {
	if x != nil {
		return x.NonRetryableErrorTypes
	}
	return nil
}

type SignalOptions struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Name is the name of the signal, better left auto generated
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *SignalOptions) Reset() {
	*x = SignalOptions{}
	mi := &file_temporal_v1_temporal_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SignalOptions) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SignalOptions) ProtoMessage() {}

func (x *SignalOptions) ProtoReflect() protoreflect.Message {
	mi := &file_temporal_v1_temporal_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SignalOptions.ProtoReflect.Descriptor instead.
func (*SignalOptions) Descriptor() ([]byte, []int) {
	return file_temporal_v1_temporal_proto_rawDescGZIP(), []int{4}
}

func (x *SignalOptions) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type QueryOptions struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Name is the name of the query, better left auto generated
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *QueryOptions) Reset() {
	*x = QueryOptions{}
	mi := &file_temporal_v1_temporal_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *QueryOptions) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QueryOptions) ProtoMessage() {}

func (x *QueryOptions) ProtoReflect() protoreflect.Message {
	mi := &file_temporal_v1_temporal_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QueryOptions.ProtoReflect.Descriptor instead.
func (*QueryOptions) Descriptor() ([]byte, []int) {
	return file_temporal_v1_temporal_proto_rawDescGZIP(), []int{5}
}

func (x *QueryOptions) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

var file_temporal_v1_temporal_proto_extTypes = []protoimpl.ExtensionInfo{
	{
		ExtendedType:  (*descriptorpb.MethodOptions)(nil),
		ExtensionType: (*ActivityOptions)(nil),
		Field:         50000,
		Name:          "temporal.v1.activity",
		Tag:           "bytes,50000,opt,name=activity",
		Filename:      "temporal/v1/temporal.proto",
	},
	{
		ExtendedType:  (*descriptorpb.MethodOptions)(nil),
		ExtensionType: (*WorkflowOptions)(nil),
		Field:         50001,
		Name:          "temporal.v1.workflow",
		Tag:           "bytes,50001,opt,name=workflow",
		Filename:      "temporal/v1/temporal.proto",
	},
	{
		ExtendedType:  (*descriptorpb.MethodOptions)(nil),
		ExtensionType: (*SignalOptions)(nil),
		Field:         50002,
		Name:          "temporal.v1.signal",
		Tag:           "bytes,50002,opt,name=signal",
		Filename:      "temporal/v1/temporal.proto",
	},
	{
		ExtendedType:  (*descriptorpb.MethodOptions)(nil),
		ExtensionType: (*QueryOptions)(nil),
		Field:         50003,
		Name:          "temporal.v1.query",
		Tag:           "bytes,50003,opt,name=query",
		Filename:      "temporal/v1/temporal.proto",
	},
	{
		ExtendedType:  (*descriptorpb.ServiceOptions)(nil),
		ExtensionType: (*ServiceOptions)(nil),
		Field:         50002,
		Name:          "temporal.v1.service",
		Tag:           "bytes,50002,opt,name=service",
		Filename:      "temporal/v1/temporal.proto",
	},
}

// Extension fields to descriptorpb.MethodOptions.
var (
	// optional temporal.v1.ActivityOptions activity = 50000;
	E_Activity = &file_temporal_v1_temporal_proto_extTypes[0]
	// optional temporal.v1.WorkflowOptions workflow = 50001;
	E_Workflow = &file_temporal_v1_temporal_proto_extTypes[1]
	// optional temporal.v1.SignalOptions signal = 50002;
	E_Signal = &file_temporal_v1_temporal_proto_extTypes[2]
	// optional temporal.v1.QueryOptions query = 50003;
	E_Query = &file_temporal_v1_temporal_proto_extTypes[3]
)

// Extension fields to descriptorpb.ServiceOptions.
var (
	// optional temporal.v1.ServiceOptions service = 50002;
	E_Service = &file_temporal_v1_temporal_proto_extTypes[4]
)

var File_temporal_v1_temporal_proto protoreflect.FileDescriptor

var file_temporal_v1_temporal_proto_rawDesc = []byte{
	0x0a, 0x1a, 0x74, 0x65, 0x6d, 0x70, 0x6f, 0x72, 0x61, 0x6c, 0x2f, 0x76, 0x31, 0x2f, 0x74, 0x65,
	0x6d, 0x70, 0x6f, 0x72, 0x61, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0b, 0x74, 0x65,
	0x6d, 0x70, 0x6f, 0x72, 0x61, 0x6c, 0x2e, 0x76, 0x31, 0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x65, 0x73, 0x63, 0x72,
	0x69, 0x70, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x77, 0x72, 0x61,
	0x70, 0x70, 0x65, 0x72, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xe4, 0x02, 0x0a, 0x0f,
	0x41, 0x63, 0x74, 0x69, 0x76, 0x69, 0x74, 0x79, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12,
	0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x12, 0x56, 0x0a, 0x19, 0x73, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x65, 0x5f,
	0x74, 0x6f, 0x5f, 0x63, 0x6c, 0x6f, 0x73, 0x65, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x49, 0x6e, 0x74, 0x33, 0x32, 0x56, 0x61,
	0x6c, 0x75, 0x65, 0x52, 0x16, 0x73, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x65, 0x54, 0x6f, 0x43,
	0x6c, 0x6f, 0x73, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x12, 0x50, 0x0a, 0x16, 0x73,
	0x74, 0x61, 0x72, 0x74, 0x5f, 0x74, 0x6f, 0x5f, 0x63, 0x6c, 0x6f, 0x73, 0x65, 0x5f, 0x74, 0x69,
	0x6d, 0x65, 0x6f, 0x75, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x49, 0x6e,
	0x74, 0x33, 0x32, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x13, 0x73, 0x74, 0x61, 0x72, 0x74, 0x54,
	0x6f, 0x43, 0x6c, 0x6f, 0x73, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x12, 0x56, 0x0a,
	0x19, 0x73, 0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x65, 0x5f, 0x74, 0x6f, 0x5f, 0x73, 0x74, 0x61,
	0x72, 0x74, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x1b, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2e, 0x49, 0x6e, 0x74, 0x33, 0x32, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x16, 0x73,
	0x63, 0x68, 0x65, 0x64, 0x75, 0x6c, 0x65, 0x54, 0x6f, 0x53, 0x74, 0x61, 0x72, 0x74, 0x54, 0x69,
	0x6d, 0x65, 0x6f, 0x75, 0x74, 0x12, 0x3b, 0x0a, 0x0c, 0x72, 0x65, 0x74, 0x72, 0x79, 0x5f, 0x70,
	0x6f, 0x6c, 0x69, 0x63, 0x79, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x74, 0x65,
	0x6d, 0x70, 0x6f, 0x72, 0x61, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x74, 0x72, 0x79, 0x50,
	0x6f, 0x6c, 0x69, 0x63, 0x79, 0x52, 0x0b, 0x72, 0x65, 0x74, 0x72, 0x79, 0x50, 0x6f, 0x6c, 0x69,
	0x63, 0x79, 0x22, 0xdd, 0x02, 0x0a, 0x0f, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x4f,
	0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x59, 0x0a, 0x1a, 0x77, 0x6f,
	0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x5f, 0x65, 0x78, 0x65, 0x63, 0x75, 0x74, 0x69, 0x6f, 0x6e,
	0x5f, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x49, 0x6e, 0x74, 0x33, 0x32, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x18, 0x77, 0x6f, 0x72,
	0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x54, 0x69,
	0x6d, 0x65, 0x6f, 0x75, 0x74, 0x12, 0x4d, 0x0a, 0x14, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f,
	0x77, 0x5f, 0x72, 0x75, 0x6e, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x49, 0x6e, 0x74, 0x33, 0x32, 0x56, 0x61, 0x6c, 0x75, 0x65,
	0x52, 0x12, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x52, 0x75, 0x6e, 0x54, 0x69, 0x6d,
	0x65, 0x6f, 0x75, 0x74, 0x12, 0x4f, 0x0a, 0x15, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77,
	0x5f, 0x74, 0x61, 0x73, 0x6b, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x49, 0x6e, 0x74, 0x33, 0x32, 0x56, 0x61, 0x6c, 0x75, 0x65,
	0x52, 0x13, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x54, 0x61, 0x73, 0x6b, 0x54, 0x69,
	0x6d, 0x65, 0x6f, 0x75, 0x74, 0x12, 0x3b, 0x0a, 0x0c, 0x72, 0x65, 0x74, 0x72, 0x79, 0x5f, 0x70,
	0x6f, 0x6c, 0x69, 0x63, 0x79, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x74, 0x65,
	0x6d, 0x70, 0x6f, 0x72, 0x61, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x74, 0x72, 0x79, 0x50,
	0x6f, 0x6c, 0x69, 0x63, 0x79, 0x52, 0x0b, 0x72, 0x65, 0x74, 0x72, 0x79, 0x50, 0x6f, 0x6c, 0x69,
	0x63, 0x79, 0x22, 0xdf, 0x01, 0x0a, 0x0e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x4f, 0x70,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x1d, 0x0a, 0x0a, 0x74, 0x61, 0x73, 0x6b, 0x5f, 0x71, 0x75,
	0x65, 0x75, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x74, 0x61, 0x73, 0x6b, 0x51,
	0x75, 0x65, 0x75, 0x65, 0x12, 0x56, 0x0a, 0x18, 0x64, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x5f,
	0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x5f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x74, 0x65, 0x6d, 0x70, 0x6f, 0x72, 0x61,
	0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x57, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x4f, 0x70, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x52, 0x16, 0x64, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x57, 0x6f, 0x72,
	0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x56, 0x0a, 0x18,
	0x64, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x5f, 0x61, 0x63, 0x74, 0x69, 0x76, 0x69, 0x74, 0x79,
	0x5f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c,
	0x2e, 0x74, 0x65, 0x6d, 0x70, 0x6f, 0x72, 0x61, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x41, 0x63, 0x74,
	0x69, 0x76, 0x69, 0x74, 0x79, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x16, 0x64, 0x65,
	0x66, 0x61, 0x75, 0x6c, 0x74, 0x41, 0x63, 0x74, 0x69, 0x76, 0x69, 0x74, 0x79, 0x4f, 0x70, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x22, 0x8c, 0x03, 0x0a, 0x0b, 0x52, 0x65, 0x74, 0x72, 0x79, 0x50, 0x6f,
	0x6c, 0x69, 0x63, 0x79, 0x12, 0x46, 0x0a, 0x10, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x61, 0x6c, 0x5f,
	0x69, 0x6e, 0x74, 0x65, 0x72, 0x76, 0x61, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x49, 0x6e, 0x74, 0x33, 0x32, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x0f, 0x69, 0x6e, 0x69,
	0x74, 0x69, 0x61, 0x6c, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x76, 0x61, 0x6c, 0x12, 0x4c, 0x0a, 0x13,
	0x62, 0x61, 0x63, 0x6b, 0x6f, 0x66, 0x66, 0x5f, 0x63, 0x6f, 0x65, 0x66, 0x66, 0x69, 0x63, 0x69,
	0x65, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x46, 0x6c, 0x6f, 0x61,
	0x74, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x12, 0x62, 0x61, 0x63, 0x6b, 0x6f, 0x66, 0x66, 0x43,
	0x6f, 0x65, 0x66, 0x66, 0x69, 0x63, 0x69, 0x65, 0x6e, 0x74, 0x12, 0x46, 0x0a, 0x10, 0x6d, 0x61,
	0x78, 0x69, 0x6d, 0x75, 0x6d, 0x5f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x76, 0x61, 0x6c, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x49, 0x6e, 0x74, 0x33, 0x32, 0x56, 0x61, 0x6c, 0x75,
	0x65, 0x52, 0x0f, 0x6d, 0x61, 0x78, 0x69, 0x6d, 0x75, 0x6d, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x76,
	0x61, 0x6c, 0x12, 0x46, 0x0a, 0x10, 0x6d, 0x61, 0x78, 0x69, 0x6d, 0x75, 0x6d, 0x5f, 0x61, 0x74,
	0x74, 0x65, 0x6d, 0x70, 0x74, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x49,
	0x6e, 0x74, 0x33, 0x32, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x0f, 0x6d, 0x61, 0x78, 0x69, 0x6d,
	0x75, 0x6d, 0x41, 0x74, 0x74, 0x65, 0x6d, 0x70, 0x74, 0x73, 0x12, 0x57, 0x0a, 0x19, 0x6e, 0x6f,
	0x6e, 0x5f, 0x72, 0x65, 0x74, 0x72, 0x79, 0x61, 0x62, 0x6c, 0x65, 0x5f, 0x65, 0x72, 0x72, 0x6f,
	0x72, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1c, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e,
	0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x16, 0x6e, 0x6f, 0x6e,
	0x52, 0x65, 0x74, 0x72, 0x79, 0x61, 0x62, 0x6c, 0x65, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x54, 0x79,
	0x70, 0x65, 0x73, 0x22, 0x23, 0x0a, 0x0d, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x6c, 0x4f, 0x70, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0x22, 0x0a, 0x0c, 0x51, 0x75, 0x65, 0x72,
	0x79, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x3a, 0x5d, 0x0a, 0x08,
	0x61, 0x63, 0x74, 0x69, 0x76, 0x69, 0x74, 0x79, 0x12, 0x1e, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x4d, 0x65, 0x74, 0x68, 0x6f,
	0x64, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xd0, 0x86, 0x03, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x1c, 0x2e, 0x74, 0x65, 0x6d, 0x70, 0x6f, 0x72, 0x61, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x41,
	0x63, 0x74, 0x69, 0x76, 0x69, 0x74, 0x79, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x08,
	0x61, 0x63, 0x74, 0x69, 0x76, 0x69, 0x74, 0x79, 0x88, 0x01, 0x01, 0x3a, 0x5d, 0x0a, 0x08, 0x77,
	0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x12, 0x1e, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64,
	0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xd1, 0x86, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x1c, 0x2e, 0x74, 0x65, 0x6d, 0x70, 0x6f, 0x72, 0x61, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x57, 0x6f,
	0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x08, 0x77,
	0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x88, 0x01, 0x01, 0x3a, 0x57, 0x0a, 0x06, 0x73, 0x69,
	0x67, 0x6e, 0x61, 0x6c, 0x12, 0x1e, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64, 0x4f, 0x70, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x18, 0xd2, 0x86, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x74,
	0x65, 0x6d, 0x70, 0x6f, 0x72, 0x61, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x69, 0x67, 0x6e, 0x61,
	0x6c, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x06, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x6c,
	0x88, 0x01, 0x01, 0x3a, 0x54, 0x0a, 0x05, 0x71, 0x75, 0x65, 0x72, 0x79, 0x12, 0x1e, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x4d,
	0x65, 0x74, 0x68, 0x6f, 0x64, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xd3, 0x86, 0x03,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x74, 0x65, 0x6d, 0x70, 0x6f, 0x72, 0x61, 0x6c, 0x2e,
	0x76, 0x31, 0x2e, 0x51, 0x75, 0x65, 0x72, 0x79, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52,
	0x05, 0x71, 0x75, 0x65, 0x72, 0x79, 0x88, 0x01, 0x01, 0x3a, 0x5b, 0x0a, 0x07, 0x73, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x12, 0x1f, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x4f, 0x70,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xd2, 0x86, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e,
	0x74, 0x65, 0x6d, 0x70, 0x6f, 0x72, 0x61, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x07, 0x73, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x88, 0x01, 0x01, 0x42, 0xb7, 0x01, 0x0a, 0x0f, 0x63, 0x6f, 0x6d, 0x2e, 0x74,
	0x65, 0x6d, 0x70, 0x6f, 0x72, 0x61, 0x6c, 0x2e, 0x76, 0x31, 0x42, 0x0d, 0x54, 0x65, 0x6d, 0x70,
	0x6f, 0x72, 0x61, 0x6c, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x48, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x74, 0x68, 0x6f, 0x6d, 0x61, 0x73, 0x2d, 0x6d,
	0x61, 0x75, 0x72, 0x69, 0x63, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x2d, 0x67, 0x65,
	0x6e, 0x2d, 0x67, 0x6f, 0x2d, 0x74, 0x6d, 0x70, 0x72, 0x6c, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x74,
	0x65, 0x6d, 0x70, 0x6f, 0x72, 0x61, 0x6c, 0x2f, 0x76, 0x31, 0x3b, 0x74, 0x65, 0x6d, 0x70, 0x6f,
	0x72, 0x61, 0x6c, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x54, 0x58, 0x58, 0xaa, 0x02, 0x0b, 0x54, 0x65,
	0x6d, 0x70, 0x6f, 0x72, 0x61, 0x6c, 0x2e, 0x56, 0x31, 0xca, 0x02, 0x0b, 0x54, 0x65, 0x6d, 0x70,
	0x6f, 0x72, 0x61, 0x6c, 0x5c, 0x56, 0x31, 0xe2, 0x02, 0x17, 0x54, 0x65, 0x6d, 0x70, 0x6f, 0x72,
	0x61, 0x6c, 0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74,
	0x61, 0xea, 0x02, 0x0c, 0x54, 0x65, 0x6d, 0x70, 0x6f, 0x72, 0x61, 0x6c, 0x3a, 0x3a, 0x56, 0x31,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_temporal_v1_temporal_proto_rawDescOnce sync.Once
	file_temporal_v1_temporal_proto_rawDescData = file_temporal_v1_temporal_proto_rawDesc
)

func file_temporal_v1_temporal_proto_rawDescGZIP() []byte {
	file_temporal_v1_temporal_proto_rawDescOnce.Do(func() {
		file_temporal_v1_temporal_proto_rawDescData = protoimpl.X.CompressGZIP(file_temporal_v1_temporal_proto_rawDescData)
	})
	return file_temporal_v1_temporal_proto_rawDescData
}

var file_temporal_v1_temporal_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_temporal_v1_temporal_proto_goTypes = []any{
	(*ActivityOptions)(nil),             // 0: temporal.v1.ActivityOptions
	(*WorkflowOptions)(nil),             // 1: temporal.v1.WorkflowOptions
	(*ServiceOptions)(nil),              // 2: temporal.v1.ServiceOptions
	(*RetryPolicy)(nil),                 // 3: temporal.v1.RetryPolicy
	(*SignalOptions)(nil),               // 4: temporal.v1.SignalOptions
	(*QueryOptions)(nil),                // 5: temporal.v1.QueryOptions
	(*wrapperspb.Int32Value)(nil),       // 6: google.protobuf.Int32Value
	(*wrapperspb.FloatValue)(nil),       // 7: google.protobuf.FloatValue
	(*wrapperspb.StringValue)(nil),      // 8: google.protobuf.StringValue
	(*descriptorpb.MethodOptions)(nil),  // 9: google.protobuf.MethodOptions
	(*descriptorpb.ServiceOptions)(nil), // 10: google.protobuf.ServiceOptions
}
var file_temporal_v1_temporal_proto_depIdxs = []int32{
	6,  // 0: temporal.v1.ActivityOptions.schedule_to_close_timeout:type_name -> google.protobuf.Int32Value
	6,  // 1: temporal.v1.ActivityOptions.start_to_close_timeout:type_name -> google.protobuf.Int32Value
	6,  // 2: temporal.v1.ActivityOptions.schedule_to_start_timeout:type_name -> google.protobuf.Int32Value
	3,  // 3: temporal.v1.ActivityOptions.retry_policy:type_name -> temporal.v1.RetryPolicy
	6,  // 4: temporal.v1.WorkflowOptions.workflow_execution_timeout:type_name -> google.protobuf.Int32Value
	6,  // 5: temporal.v1.WorkflowOptions.workflow_run_timeout:type_name -> google.protobuf.Int32Value
	6,  // 6: temporal.v1.WorkflowOptions.workflow_task_timeout:type_name -> google.protobuf.Int32Value
	3,  // 7: temporal.v1.WorkflowOptions.retry_policy:type_name -> temporal.v1.RetryPolicy
	1,  // 8: temporal.v1.ServiceOptions.default_workflow_options:type_name -> temporal.v1.WorkflowOptions
	0,  // 9: temporal.v1.ServiceOptions.default_activity_options:type_name -> temporal.v1.ActivityOptions
	6,  // 10: temporal.v1.RetryPolicy.initial_interval:type_name -> google.protobuf.Int32Value
	7,  // 11: temporal.v1.RetryPolicy.backoff_coefficient:type_name -> google.protobuf.FloatValue
	6,  // 12: temporal.v1.RetryPolicy.maximum_interval:type_name -> google.protobuf.Int32Value
	6,  // 13: temporal.v1.RetryPolicy.maximum_attempts:type_name -> google.protobuf.Int32Value
	8,  // 14: temporal.v1.RetryPolicy.non_retryable_error_types:type_name -> google.protobuf.StringValue
	9,  // 15: temporal.v1.activity:extendee -> google.protobuf.MethodOptions
	9,  // 16: temporal.v1.workflow:extendee -> google.protobuf.MethodOptions
	9,  // 17: temporal.v1.signal:extendee -> google.protobuf.MethodOptions
	9,  // 18: temporal.v1.query:extendee -> google.protobuf.MethodOptions
	10, // 19: temporal.v1.service:extendee -> google.protobuf.ServiceOptions
	0,  // 20: temporal.v1.activity:type_name -> temporal.v1.ActivityOptions
	1,  // 21: temporal.v1.workflow:type_name -> temporal.v1.WorkflowOptions
	4,  // 22: temporal.v1.signal:type_name -> temporal.v1.SignalOptions
	5,  // 23: temporal.v1.query:type_name -> temporal.v1.QueryOptions
	2,  // 24: temporal.v1.service:type_name -> temporal.v1.ServiceOptions
	25, // [25:25] is the sub-list for method output_type
	25, // [25:25] is the sub-list for method input_type
	20, // [20:25] is the sub-list for extension type_name
	15, // [15:20] is the sub-list for extension extendee
	0,  // [0:15] is the sub-list for field type_name
}

func init() { file_temporal_v1_temporal_proto_init() }
func file_temporal_v1_temporal_proto_init() {
	if File_temporal_v1_temporal_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_temporal_v1_temporal_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 5,
			NumServices:   0,
		},
		GoTypes:           file_temporal_v1_temporal_proto_goTypes,
		DependencyIndexes: file_temporal_v1_temporal_proto_depIdxs,
		MessageInfos:      file_temporal_v1_temporal_proto_msgTypes,
		ExtensionInfos:    file_temporal_v1_temporal_proto_extTypes,
	}.Build()
	File_temporal_v1_temporal_proto = out.File
	file_temporal_v1_temporal_proto_rawDesc = nil
	file_temporal_v1_temporal_proto_goTypes = nil
	file_temporal_v1_temporal_proto_depIdxs = nil
}
