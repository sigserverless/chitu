// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v3.15.8
// source: pdict.proto

package pdict

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

type PDictSet struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id     string  `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Key    string  `protobuf:"bytes,2,opt,name=key,proto3" json:"key,omitempty"`
	Values *Values `protobuf:"bytes,3,opt,name=values,proto3" json:"values,omitempty"`
}

func (x *PDictSet) Reset() {
	*x = PDictSet{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pdict_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PDictSet) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PDictSet) ProtoMessage() {}

func (x *PDictSet) ProtoReflect() protoreflect.Message {
	mi := &file_pdict_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PDictSet.ProtoReflect.Descriptor instead.
func (*PDictSet) Descriptor() ([]byte, []int) {
	return file_pdict_proto_rawDescGZIP(), []int{0}
}

func (x *PDictSet) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *PDictSet) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *PDictSet) GetValues() *Values {
	if x != nil {
		return x.Values
	}
	return nil
}

type Values struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Values []float64 `protobuf:"fixed64,1,rep,packed,name=values,proto3" json:"values,omitempty"`
}

func (x *Values) Reset() {
	*x = Values{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pdict_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Values) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Values) ProtoMessage() {}

func (x *Values) ProtoReflect() protoreflect.Message {
	mi := &file_pdict_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Values.ProtoReflect.Descriptor instead.
func (*Values) Descriptor() ([]byte, []int) {
	return file_pdict_proto_rawDescGZIP(), []int{1}
}

func (x *Values) GetValues() []float64 {
	if x != nil {
		return x.Values
	}
	return nil
}

type PDictVal struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Val map[string]*Values `protobuf:"bytes,1,rep,name=val,proto3" json:"val,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *PDictVal) Reset() {
	*x = PDictVal{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pdict_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PDictVal) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PDictVal) ProtoMessage() {}

func (x *PDictVal) ProtoReflect() protoreflect.Message {
	mi := &file_pdict_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PDictVal.ProtoReflect.Descriptor instead.
func (*PDictVal) Descriptor() ([]byte, []int) {
	return file_pdict_proto_rawDescGZIP(), []int{2}
}

func (x *PDictVal) GetVal() map[string]*Values {
	if x != nil {
		return x.Val
	}
	return nil
}

type DictSetOp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key   string    `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Value []float64 `protobuf:"fixed64,2,rep,packed,name=value,proto3" json:"value,omitempty"`
}

func (x *DictSetOp) Reset() {
	*x = DictSetOp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pdict_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DictSetOp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DictSetOp) ProtoMessage() {}

func (x *DictSetOp) ProtoReflect() protoreflect.Message {
	mi := &file_pdict_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DictSetOp.ProtoReflect.Descriptor instead.
func (*DictSetOp) Descriptor() ([]byte, []int) {
	return file_pdict_proto_rawDescGZIP(), []int{3}
}

func (x *DictSetOp) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *DictSetOp) GetValue() []float64 {
	if x != nil {
		return x.Value
	}
	return nil
}

var File_pdict_proto protoreflect.FileDescriptor

var file_pdict_proto_rawDesc = []byte{
	0x0a, 0x0b, 0x70, 0x64, 0x69, 0x63, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x70,
	0x64, 0x69, 0x63, 0x74, 0x22, 0x53, 0x0a, 0x08, 0x50, 0x44, 0x69, 0x63, 0x74, 0x53, 0x65, 0x74,
	0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64,
	0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b,
	0x65, 0x79, 0x12, 0x25, 0x0a, 0x06, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x70, 0x64, 0x69, 0x63, 0x74, 0x2e, 0x56, 0x61, 0x6c, 0x75, 0x65,
	0x73, 0x52, 0x06, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x22, 0x20, 0x0a, 0x06, 0x56, 0x61, 0x6c,
	0x75, 0x65, 0x73, 0x12, 0x16, 0x0a, 0x06, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x18, 0x01, 0x20,
	0x03, 0x28, 0x01, 0x52, 0x06, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x22, 0x7d, 0x0a, 0x08, 0x50,
	0x44, 0x69, 0x63, 0x74, 0x56, 0x61, 0x6c, 0x12, 0x2a, 0x0a, 0x03, 0x76, 0x61, 0x6c, 0x18, 0x01,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x70, 0x64, 0x69, 0x63, 0x74, 0x2e, 0x50, 0x44, 0x69,
	0x63, 0x74, 0x56, 0x61, 0x6c, 0x2e, 0x56, 0x61, 0x6c, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x03,
	0x76, 0x61, 0x6c, 0x1a, 0x45, 0x0a, 0x08, 0x56, 0x61, 0x6c, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12,
	0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65,
	0x79, 0x12, 0x23, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x0d, 0x2e, 0x70, 0x64, 0x69, 0x63, 0x74, 0x2e, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x52,
	0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x33, 0x0a, 0x09, 0x44, 0x69,
	0x63, 0x74, 0x53, 0x65, 0x74, 0x4f, 0x70, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x18, 0x02, 0x20, 0x03, 0x28, 0x01, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x42,
	0x08, 0x5a, 0x06, 0x2f, 0x70, 0x64, 0x69, 0x63, 0x74, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_pdict_proto_rawDescOnce sync.Once
	file_pdict_proto_rawDescData = file_pdict_proto_rawDesc
)

func file_pdict_proto_rawDescGZIP() []byte {
	file_pdict_proto_rawDescOnce.Do(func() {
		file_pdict_proto_rawDescData = protoimpl.X.CompressGZIP(file_pdict_proto_rawDescData)
	})
	return file_pdict_proto_rawDescData
}

var file_pdict_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_pdict_proto_goTypes = []interface{}{
	(*PDictSet)(nil),  // 0: pdict.PDictSet
	(*Values)(nil),    // 1: pdict.Values
	(*PDictVal)(nil),  // 2: pdict.PDictVal
	(*DictSetOp)(nil), // 3: pdict.DictSetOp
	nil,               // 4: pdict.PDictVal.ValEntry
}
var file_pdict_proto_depIdxs = []int32{
	1, // 0: pdict.PDictSet.values:type_name -> pdict.Values
	4, // 1: pdict.PDictVal.val:type_name -> pdict.PDictVal.ValEntry
	1, // 2: pdict.PDictVal.ValEntry.value:type_name -> pdict.Values
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_pdict_proto_init() }
func file_pdict_proto_init() {
	if File_pdict_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pdict_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PDictSet); i {
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
		file_pdict_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Values); i {
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
		file_pdict_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PDictVal); i {
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
		file_pdict_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DictSetOp); i {
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
			RawDescriptor: file_pdict_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_pdict_proto_goTypes,
		DependencyIndexes: file_pdict_proto_depIdxs,
		MessageInfos:      file_pdict_proto_msgTypes,
	}.Build()
	File_pdict_proto = out.File
	file_pdict_proto_rawDesc = nil
	file_pdict_proto_goTypes = nil
	file_pdict_proto_depIdxs = nil
}
