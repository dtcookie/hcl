package hcl_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/dtcookie/hcl"
	"github.com/dtcookie/jsonx"
	"github.com/dtcookie/opt"
)

type Foo struct {
	NonZeroString string `json:"NonZeroString"`
	Bar           Bar    `json:"Bar"`
	Bars          []Bar  `json:"Bars"`
}

func (f *Foo) MarshalHCL(enc hcl.Encoder) error {
	enc.Encode("non_zero_string", &f.NonZeroString, false)
	enc.Encode("bar", &f.Bar, false)
	enc.Encode("bars", &f.Bars, false)
	return nil
}

type Bar struct {
	NonZeroString string
}

func (b *Bar) MarshalHCL(enc hcl.Encoder) error {
	enc.Encode("non_zero_string", &b.NonZeroString, false)
	return nil
}

type MockResource struct {
	m map[string]interface{}
}

func (mr *MockResource) Set(key string, value interface{}) error {
	fmt.Println("Set", key, value)
	mr.m[key] = value
	return nil
}

type marshalRecord struct {
	Unknowns            jsonx.Unknowns
	NonZeroString       string
	NonZeroStringOpt    string
	ZeroString          string
	ZeroStringOpt       string
	NonZeroStringPtr    *string
	NonZeroStringPtrOpt *string
	ZeroStringPtr       *string
	ZeroStringPtrOpt    *string
	SubRecord           *marshalRecord
}

func (mr *marshalRecord) MarshalHCL(enc hcl.Encoder) error {
	enc.Encode("non_zero_string", &mr.NonZeroString, false)
	enc.Encode("non_zero_string_opt", &mr.NonZeroStringOpt, true)
	enc.Encode("zero_string", &mr.ZeroString, false)
	enc.Encode("zero_string_opt", &mr.ZeroStringOpt, true)
	enc.Encode("non_zero_string_ptr", &mr.NonZeroStringPtr, false)
	enc.Encode("non_zero_string_ptr_opt", &mr.NonZeroStringPtrOpt, true)
	enc.Encode("zero_string_ptr", &mr.ZeroStringPtr, false)
	enc.Encode("zero_string_ptr_opt", &mr.ZeroStringPtrOpt, true)
	enc.Encode("sub_record", &mr.SubRecord, true)
	return nil
}

func TestEncoder(t *testing.T) {
	subrecord := marshalRecord{
		NonZeroString:       "NonZeroString",
		NonZeroStringOpt:    "NonZeroString",
		ZeroString:          "",
		ZeroStringOpt:       "",
		NonZeroStringPtr:    opt.NewString("NonZeroStringPtr"),
		NonZeroStringPtrOpt: opt.NewString("NonZeroStringPtr"),
		ZeroStringPtr:       nil,
		ZeroStringPtrOpt:    nil,
	}
	mr := marshalRecord{
		NonZeroString:       "NonZeroString",
		NonZeroStringOpt:    "NonZeroString",
		ZeroString:          "",
		ZeroStringOpt:       "",
		NonZeroStringPtr:    opt.NewString("NonZeroStringPtr"),
		NonZeroStringPtrOpt: opt.NewString("NonZeroStringPtr"),
		ZeroStringPtr:       nil,
		ZeroStringPtrOpt:    nil,
		SubRecord:           &subrecord,
	}
	res := &MockResource{
		m: map[string]interface{}{},
	}
	hcl.Marshal(&mr, res)
	var err error
	var bytes []byte
	if bytes, err = json.MarshalIndent(res.m, "", "  "); err != nil {
		t.Error(err)
	}
	fmt.Println(string(bytes))
}
