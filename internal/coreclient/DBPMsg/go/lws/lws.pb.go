// Code generated by protoc-gen-go. DO NOT EDIT.
// source: lws.proto

package lws

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Transaction_CDestination_PREFIX int32

const (
	Transaction_CDestination_PREFIX_NULL     Transaction_CDestination_PREFIX = 0
	Transaction_CDestination_PREFIX_PUBKEY   Transaction_CDestination_PREFIX = 1
	Transaction_CDestination_PREFIX_TEMPLATE Transaction_CDestination_PREFIX = 2
)

var Transaction_CDestination_PREFIX_name = map[int32]string{
	0: "PREFIX_NULL",
	1: "PREFIX_PUBKEY",
	2: "PREFIX_TEMPLATE",
}
var Transaction_CDestination_PREFIX_value = map[string]int32{
	"PREFIX_NULL":     0,
	"PREFIX_PUBKEY":   1,
	"PREFIX_TEMPLATE": 2,
}

func (x Transaction_CDestination_PREFIX) String() string {
	return proto.EnumName(Transaction_CDestination_PREFIX_name, int32(x))
}
func (Transaction_CDestination_PREFIX) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_lws_91d0fee67590ed8d, []int{3, 1, 0}
}

type GetBlocksArg struct {
	Hash                 string   `protobuf:"bytes,1,opt,name=hash,proto3" json:"hash,omitempty"`
	Number               int32    `protobuf:"varint,2,opt,name=number,proto3" json:"number,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetBlocksArg) Reset()         { *m = GetBlocksArg{} }
func (m *GetBlocksArg) String() string { return proto.CompactTextString(m) }
func (*GetBlocksArg) ProtoMessage()    {}
func (*GetBlocksArg) Descriptor() ([]byte, []int) {
	return fileDescriptor_lws_91d0fee67590ed8d, []int{0}
}
func (m *GetBlocksArg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetBlocksArg.Unmarshal(m, b)
}
func (m *GetBlocksArg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetBlocksArg.Marshal(b, m, deterministic)
}
func (dst *GetBlocksArg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetBlocksArg.Merge(dst, src)
}
func (m *GetBlocksArg) XXX_Size() int {
	return xxx_messageInfo_GetBlocksArg.Size(m)
}
func (m *GetBlocksArg) XXX_DiscardUnknown() {
	xxx_messageInfo_GetBlocksArg.DiscardUnknown(m)
}

var xxx_messageInfo_GetBlocksArg proto.InternalMessageInfo

func (m *GetBlocksArg) GetHash() string {
	if m != nil {
		return m.Hash
	}
	return ""
}

func (m *GetBlocksArg) GetNumber() int32 {
	if m != nil {
		return m.Number
	}
	return 0
}

type GetTxArg struct {
	Hash                 string   `protobuf:"bytes,1,opt,name=hash,proto3" json:"hash,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetTxArg) Reset()         { *m = GetTxArg{} }
func (m *GetTxArg) String() string { return proto.CompactTextString(m) }
func (*GetTxArg) ProtoMessage()    {}
func (*GetTxArg) Descriptor() ([]byte, []int) {
	return fileDescriptor_lws_91d0fee67590ed8d, []int{1}
}
func (m *GetTxArg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetTxArg.Unmarshal(m, b)
}
func (m *GetTxArg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetTxArg.Marshal(b, m, deterministic)
}
func (dst *GetTxArg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetTxArg.Merge(dst, src)
}
func (m *GetTxArg) XXX_Size() int {
	return xxx_messageInfo_GetTxArg.Size(m)
}
func (m *GetTxArg) XXX_DiscardUnknown() {
	xxx_messageInfo_GetTxArg.DiscardUnknown(m)
}

var xxx_messageInfo_GetTxArg proto.InternalMessageInfo

func (m *GetTxArg) GetHash() string {
	if m != nil {
		return m.Hash
	}
	return ""
}

type SendTxArg struct {
	Hash                 string   `protobuf:"bytes,1,opt,name=hash,proto3" json:"hash,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SendTxArg) Reset()         { *m = SendTxArg{} }
func (m *SendTxArg) String() string { return proto.CompactTextString(m) }
func (*SendTxArg) ProtoMessage()    {}
func (*SendTxArg) Descriptor() ([]byte, []int) {
	return fileDescriptor_lws_91d0fee67590ed8d, []int{2}
}
func (m *SendTxArg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SendTxArg.Unmarshal(m, b)
}
func (m *SendTxArg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SendTxArg.Marshal(b, m, deterministic)
}
func (dst *SendTxArg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SendTxArg.Merge(dst, src)
}
func (m *SendTxArg) XXX_Size() int {
	return xxx_messageInfo_SendTxArg.Size(m)
}
func (m *SendTxArg) XXX_DiscardUnknown() {
	xxx_messageInfo_SendTxArg.DiscardUnknown(m)
}

var xxx_messageInfo_SendTxArg proto.InternalMessageInfo

func (m *SendTxArg) GetHash() string {
	if m != nil {
		return m.Hash
	}
	return ""
}

type Transaction struct {
	NVersion             uint32                    `protobuf:"varint,1,opt,name=nVersion,proto3" json:"nVersion,omitempty"`
	NType                uint32                    `protobuf:"varint,2,opt,name=nType,proto3" json:"nType,omitempty"`
	NLockUntil           uint32                    `protobuf:"varint,3,opt,name=nLockUntil,proto3" json:"nLockUntil,omitempty"`
	HashAnchor           []byte                    `protobuf:"bytes,4,opt,name=hashAnchor,proto3" json:"hashAnchor,omitempty"`
	VInput               []*Transaction_CTxIn      `protobuf:"bytes,5,rep,name=vInput,proto3" json:"vInput,omitempty"`
	CDestination         *Transaction_CDestination `protobuf:"bytes,6,opt,name=cDestination,proto3" json:"cDestination,omitempty"`
	NAmount              int64                     `protobuf:"varint,7,opt,name=nAmount,proto3" json:"nAmount,omitempty"`
	NTxFee               int64                     `protobuf:"varint,8,opt,name=nTxFee,proto3" json:"nTxFee,omitempty"`
	VchData              []byte                    `protobuf:"bytes,9,opt,name=vchData,proto3" json:"vchData,omitempty"`
	VchSig               []byte                    `protobuf:"bytes,10,opt,name=vchSig,proto3" json:"vchSig,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                  `json:"-"`
	XXX_unrecognized     []byte                    `json:"-"`
	XXX_sizecache        int32                     `json:"-"`
}

func (m *Transaction) Reset()         { *m = Transaction{} }
func (m *Transaction) String() string { return proto.CompactTextString(m) }
func (*Transaction) ProtoMessage()    {}
func (*Transaction) Descriptor() ([]byte, []int) {
	return fileDescriptor_lws_91d0fee67590ed8d, []int{3}
}
func (m *Transaction) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Transaction.Unmarshal(m, b)
}
func (m *Transaction) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Transaction.Marshal(b, m, deterministic)
}
func (dst *Transaction) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Transaction.Merge(dst, src)
}
func (m *Transaction) XXX_Size() int {
	return xxx_messageInfo_Transaction.Size(m)
}
func (m *Transaction) XXX_DiscardUnknown() {
	xxx_messageInfo_Transaction.DiscardUnknown(m)
}

var xxx_messageInfo_Transaction proto.InternalMessageInfo

func (m *Transaction) GetNVersion() uint32 {
	if m != nil {
		return m.NVersion
	}
	return 0
}

func (m *Transaction) GetNType() uint32 {
	if m != nil {
		return m.NType
	}
	return 0
}

func (m *Transaction) GetNLockUntil() uint32 {
	if m != nil {
		return m.NLockUntil
	}
	return 0
}

func (m *Transaction) GetHashAnchor() []byte {
	if m != nil {
		return m.HashAnchor
	}
	return nil
}

func (m *Transaction) GetVInput() []*Transaction_CTxIn {
	if m != nil {
		return m.VInput
	}
	return nil
}

func (m *Transaction) GetCDestination() *Transaction_CDestination {
	if m != nil {
		return m.CDestination
	}
	return nil
}

func (m *Transaction) GetNAmount() int64 {
	if m != nil {
		return m.NAmount
	}
	return 0
}

func (m *Transaction) GetNTxFee() int64 {
	if m != nil {
		return m.NTxFee
	}
	return 0
}

func (m *Transaction) GetVchData() []byte {
	if m != nil {
		return m.VchData
	}
	return nil
}

func (m *Transaction) GetVchSig() []byte {
	if m != nil {
		return m.VchSig
	}
	return nil
}

type Transaction_CTxIn struct {
	Hash                 []byte   `protobuf:"bytes,1,opt,name=hash,proto3" json:"hash,omitempty"`
	N                    uint32   `protobuf:"varint,2,opt,name=n,proto3" json:"n,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Transaction_CTxIn) Reset()         { *m = Transaction_CTxIn{} }
func (m *Transaction_CTxIn) String() string { return proto.CompactTextString(m) }
func (*Transaction_CTxIn) ProtoMessage()    {}
func (*Transaction_CTxIn) Descriptor() ([]byte, []int) {
	return fileDescriptor_lws_91d0fee67590ed8d, []int{3, 0}
}
func (m *Transaction_CTxIn) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Transaction_CTxIn.Unmarshal(m, b)
}
func (m *Transaction_CTxIn) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Transaction_CTxIn.Marshal(b, m, deterministic)
}
func (dst *Transaction_CTxIn) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Transaction_CTxIn.Merge(dst, src)
}
func (m *Transaction_CTxIn) XXX_Size() int {
	return xxx_messageInfo_Transaction_CTxIn.Size(m)
}
func (m *Transaction_CTxIn) XXX_DiscardUnknown() {
	xxx_messageInfo_Transaction_CTxIn.DiscardUnknown(m)
}

var xxx_messageInfo_Transaction_CTxIn proto.InternalMessageInfo

func (m *Transaction_CTxIn) GetHash() []byte {
	if m != nil {
		return m.Hash
	}
	return nil
}

func (m *Transaction_CTxIn) GetN() uint32 {
	if m != nil {
		return m.N
	}
	return 0
}

type Transaction_CDestination struct {
	Prefix               uint32   `protobuf:"varint,1,opt,name=prefix,proto3" json:"prefix,omitempty"`
	Data                 []byte   `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	Size                 uint32   `protobuf:"varint,3,opt,name=size,proto3" json:"size,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Transaction_CDestination) Reset()         { *m = Transaction_CDestination{} }
func (m *Transaction_CDestination) String() string { return proto.CompactTextString(m) }
func (*Transaction_CDestination) ProtoMessage()    {}
func (*Transaction_CDestination) Descriptor() ([]byte, []int) {
	return fileDescriptor_lws_91d0fee67590ed8d, []int{3, 1}
}
func (m *Transaction_CDestination) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Transaction_CDestination.Unmarshal(m, b)
}
func (m *Transaction_CDestination) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Transaction_CDestination.Marshal(b, m, deterministic)
}
func (dst *Transaction_CDestination) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Transaction_CDestination.Merge(dst, src)
}
func (m *Transaction_CDestination) XXX_Size() int {
	return xxx_messageInfo_Transaction_CDestination.Size(m)
}
func (m *Transaction_CDestination) XXX_DiscardUnknown() {
	xxx_messageInfo_Transaction_CDestination.DiscardUnknown(m)
}

var xxx_messageInfo_Transaction_CDestination proto.InternalMessageInfo

func (m *Transaction_CDestination) GetPrefix() uint32 {
	if m != nil {
		return m.Prefix
	}
	return 0
}

func (m *Transaction_CDestination) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func (m *Transaction_CDestination) GetSize() uint32 {
	if m != nil {
		return m.Size
	}
	return 0
}

type Block struct {
	NVersion             uint32         `protobuf:"varint,1,opt,name=nVersion,proto3" json:"nVersion,omitempty"`
	NType                uint32         `protobuf:"varint,2,opt,name=nType,proto3" json:"nType,omitempty"`
	NTimeStamp           uint32         `protobuf:"varint,3,opt,name=nTimeStamp,proto3" json:"nTimeStamp,omitempty"`
	HashPrev             []byte         `protobuf:"bytes,4,opt,name=hashPrev,proto3" json:"hashPrev,omitempty"`
	HashMerkle           []byte         `protobuf:"bytes,5,opt,name=hashMerkle,proto3" json:"hashMerkle,omitempty"`
	VchProof             []byte         `protobuf:"bytes,6,opt,name=vchProof,proto3" json:"vchProof,omitempty"`
	TxMint               *Transaction   `protobuf:"bytes,7,opt,name=txMint,proto3" json:"txMint,omitempty"`
	Vtx                  []*Transaction `protobuf:"bytes,8,rep,name=vtx,proto3" json:"vtx,omitempty"`
	VchSig               []byte         `protobuf:"bytes,9,opt,name=vchSig,proto3" json:"vchSig,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *Block) Reset()         { *m = Block{} }
func (m *Block) String() string { return proto.CompactTextString(m) }
func (*Block) ProtoMessage()    {}
func (*Block) Descriptor() ([]byte, []int) {
	return fileDescriptor_lws_91d0fee67590ed8d, []int{4}
}
func (m *Block) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Block.Unmarshal(m, b)
}
func (m *Block) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Block.Marshal(b, m, deterministic)
}
func (dst *Block) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Block.Merge(dst, src)
}
func (m *Block) XXX_Size() int {
	return xxx_messageInfo_Block.Size(m)
}
func (m *Block) XXX_DiscardUnknown() {
	xxx_messageInfo_Block.DiscardUnknown(m)
}

var xxx_messageInfo_Block proto.InternalMessageInfo

func (m *Block) GetNVersion() uint32 {
	if m != nil {
		return m.NVersion
	}
	return 0
}

func (m *Block) GetNType() uint32 {
	if m != nil {
		return m.NType
	}
	return 0
}

func (m *Block) GetNTimeStamp() uint32 {
	if m != nil {
		return m.NTimeStamp
	}
	return 0
}

func (m *Block) GetHashPrev() []byte {
	if m != nil {
		return m.HashPrev
	}
	return nil
}

func (m *Block) GetHashMerkle() []byte {
	if m != nil {
		return m.HashMerkle
	}
	return nil
}

func (m *Block) GetVchProof() []byte {
	if m != nil {
		return m.VchProof
	}
	return nil
}

func (m *Block) GetTxMint() *Transaction {
	if m != nil {
		return m.TxMint
	}
	return nil
}

func (m *Block) GetVtx() []*Transaction {
	if m != nil {
		return m.Vtx
	}
	return nil
}

func (m *Block) GetVchSig() []byte {
	if m != nil {
		return m.VchSig
	}
	return nil
}

type SendTxRet struct {
	Hash                 string   `protobuf:"bytes,1,opt,name=hash,proto3" json:"hash,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SendTxRet) Reset()         { *m = SendTxRet{} }
func (m *SendTxRet) String() string { return proto.CompactTextString(m) }
func (*SendTxRet) ProtoMessage()    {}
func (*SendTxRet) Descriptor() ([]byte, []int) {
	return fileDescriptor_lws_91d0fee67590ed8d, []int{5}
}
func (m *SendTxRet) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SendTxRet.Unmarshal(m, b)
}
func (m *SendTxRet) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SendTxRet.Marshal(b, m, deterministic)
}
func (dst *SendTxRet) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SendTxRet.Merge(dst, src)
}
func (m *SendTxRet) XXX_Size() int {
	return xxx_messageInfo_SendTxRet.Size(m)
}
func (m *SendTxRet) XXX_DiscardUnknown() {
	xxx_messageInfo_SendTxRet.DiscardUnknown(m)
}

var xxx_messageInfo_SendTxRet proto.InternalMessageInfo

func (m *SendTxRet) GetHash() string {
	if m != nil {
		return m.Hash
	}
	return ""
}

func init() {
	proto.RegisterType((*GetBlocksArg)(nil), "lws.GetBlocksArg")
	proto.RegisterType((*GetTxArg)(nil), "lws.GetTxArg")
	proto.RegisterType((*SendTxArg)(nil), "lws.SendTxArg")
	proto.RegisterType((*Transaction)(nil), "lws.Transaction")
	proto.RegisterType((*Transaction_CTxIn)(nil), "lws.Transaction.CTxIn")
	proto.RegisterType((*Transaction_CDestination)(nil), "lws.Transaction.CDestination")
	proto.RegisterType((*Block)(nil), "lws.Block")
	proto.RegisterType((*SendTxRet)(nil), "lws.SendTxRet")
	proto.RegisterEnum("lws.Transaction_CDestination_PREFIX", Transaction_CDestination_PREFIX_name, Transaction_CDestination_PREFIX_value)
}

func init() { proto.RegisterFile("lws.proto", fileDescriptor_lws_91d0fee67590ed8d) }

var fileDescriptor_lws_91d0fee67590ed8d = []byte{
	// 513 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x53, 0xc1, 0x6e, 0xda, 0x40,
	0x10, 0xad, 0x01, 0x3b, 0x30, 0x18, 0x85, 0x6e, 0xab, 0x68, 0x85, 0xd4, 0xd4, 0xf2, 0xc9, 0xbd,
	0x70, 0xa0, 0xb7, 0xde, 0x9c, 0x86, 0x44, 0xa8, 0x50, 0xa1, 0xc5, 0x54, 0xed, 0xa9, 0x72, 0x9c,
	0x4d, 0x6c, 0x01, 0xbb, 0x68, 0xbd, 0x38, 0x6e, 0xff, 0xa2, 0x87, 0xaa, 0xbf, 0x5b, 0xed, 0xb2,
	0x38, 0x4e, 0xd3, 0x5c, 0x72, 0x9b, 0xf7, 0xde, 0xcc, 0x78, 0x3c, 0xf3, 0x16, 0x3a, 0xeb, 0xbb,
	0x7c, 0xb8, 0x15, 0x5c, 0x72, 0xd4, 0x5c, 0xdf, 0xe5, 0xfe, 0x07, 0x70, 0x2f, 0xa9, 0x3c, 0x5b,
	0xf3, 0x64, 0x95, 0x87, 0xe2, 0x16, 0x21, 0x68, 0xa5, 0x71, 0x9e, 0x62, 0xcb, 0xb3, 0x82, 0x0e,
	0xd1, 0x31, 0x3a, 0x01, 0x87, 0xed, 0x36, 0x57, 0x54, 0xe0, 0x86, 0x67, 0x05, 0x36, 0x31, 0xc8,
	0x3f, 0x85, 0xf6, 0x25, 0x95, 0x51, 0xf9, 0x44, 0x9d, 0xff, 0x16, 0x3a, 0x0b, 0xca, 0xae, 0x9f,
	0x4e, 0xf8, 0xdd, 0x82, 0x6e, 0x24, 0x62, 0x96, 0xc7, 0x89, 0xcc, 0x38, 0x43, 0x03, 0x68, 0xb3,
	0x2f, 0x54, 0xe4, 0x19, 0x67, 0x3a, 0xaf, 0x47, 0x2a, 0x8c, 0x5e, 0x83, 0xcd, 0xa2, 0x1f, 0x5b,
	0xaa, 0x67, 0xe8, 0x91, 0x3d, 0x40, 0xa7, 0x00, 0x6c, 0xca, 0x93, 0xd5, 0x92, 0xc9, 0x6c, 0x8d,
	0x9b, 0x5a, 0xaa, 0x31, 0x4a, 0x57, 0x5f, 0x0a, 0x59, 0x92, 0x72, 0x81, 0x5b, 0x9e, 0x15, 0xb8,
	0xa4, 0xc6, 0xa0, 0x21, 0x38, 0xc5, 0x84, 0x6d, 0x77, 0x12, 0xdb, 0x5e, 0x33, 0xe8, 0x8e, 0x4e,
	0x86, 0x6a, 0x3f, 0xb5, 0x99, 0x86, 0x1f, 0xa3, 0x72, 0xc2, 0x88, 0xc9, 0x42, 0x21, 0xb8, 0xc9,
	0x39, 0xcd, 0x65, 0xc6, 0x62, 0xa5, 0x62, 0xc7, 0xb3, 0x82, 0xee, 0xe8, 0xcd, 0xe3, 0xaa, 0x5a,
	0x12, 0x79, 0x50, 0x82, 0x30, 0x1c, 0xb1, 0x70, 0xc3, 0x77, 0x4c, 0xe2, 0x23, 0xcf, 0x0a, 0x9a,
	0xe4, 0x00, 0xf5, 0x9e, 0xa3, 0xf2, 0x82, 0x52, 0xdc, 0xd6, 0x82, 0x41, 0xaa, 0xa2, 0x48, 0xd2,
	0xf3, 0x58, 0xc6, 0xb8, 0xa3, 0xff, 0xe0, 0x00, 0x55, 0x45, 0x91, 0xa4, 0x8b, 0xec, 0x16, 0x83,
	0x16, 0x0c, 0x1a, 0xbc, 0x03, 0x5b, 0xcf, 0xfd, 0x60, 0xeb, 0xae, 0x39, 0xa7, 0x0b, 0x16, 0x33,
	0x5b, 0xb4, 0xd8, 0xe0, 0x97, 0x05, 0x6e, 0x7d, 0x5a, 0xd5, 0x73, 0x2b, 0xe8, 0x4d, 0x56, 0x9a,
	0x13, 0x18, 0xa4, 0x5a, 0x5d, 0xab, 0x11, 0x1a, 0xfb, 0x56, 0x2a, 0x56, 0x5c, 0x9e, 0xfd, 0xa4,
	0x66, 0xf1, 0x3a, 0xf6, 0x43, 0x70, 0xe6, 0x64, 0x7c, 0x31, 0xf9, 0x8a, 0x8e, 0xa1, 0xbb, 0x8f,
	0xbe, 0x7f, 0x5e, 0x4e, 0xa7, 0xfd, 0x17, 0xe8, 0x25, 0xf4, 0x0c, 0x31, 0x5f, 0x9e, 0x7d, 0x1a,
	0x7f, 0xeb, 0x5b, 0xe8, 0x15, 0x1c, 0x1b, 0x2a, 0x1a, 0xcf, 0xe6, 0xd3, 0x30, 0x1a, 0xf7, 0x1b,
	0xfe, 0x9f, 0x06, 0xd8, 0xda, 0x92, 0xcf, 0x74, 0x44, 0x94, 0x6d, 0xe8, 0x42, 0xc6, 0x9b, 0x6d,
	0xe5, 0x88, 0x8a, 0x51, 0x1d, 0xd5, 0x16, 0xe6, 0x82, 0x16, 0xc6, 0x0f, 0x15, 0x3e, 0xb8, 0x65,
	0x46, 0xc5, 0x6a, 0x4d, 0xb1, 0x7d, 0xef, 0x96, 0x3d, 0xa3, 0x6a, 0x8b, 0x24, 0x9d, 0x0b, 0xce,
	0x6f, 0xf4, 0xe5, 0x5d, 0x52, 0x61, 0x14, 0x80, 0x23, 0xcb, 0x59, 0x66, 0xae, 0xda, 0x1d, 0xf5,
	0xff, 0xf5, 0x04, 0x31, 0x3a, 0xf2, 0xa1, 0x59, 0xc8, 0x12, 0xb7, 0xb5, 0xe1, 0x1e, 0xa7, 0x29,
	0xb1, 0x76, 0xd8, 0x4e, 0xfd, 0xb0, 0xf7, 0x4f, 0x8a, 0x50, 0xf9, 0xbf, 0x27, 0x75, 0xe5, 0xe8,
	0xb7, 0xfd, 0xfe, 0x6f, 0x00, 0x00, 0x00, 0xff, 0xff, 0x63, 0x06, 0x4e, 0x09, 0xe8, 0x03, 0x00,
	0x00,
}
