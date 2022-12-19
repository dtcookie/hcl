package hcl_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/dtcookie/hcl"
)

type record struct {
	Value string
}

func (me *record) UnmarshalHCL(decoder hcl.Decoder) error {
	if err := decoder.Decode("value", &me.Value); err != nil {
		return err
	}
	return nil
}

type records []*record

func (me *records) UnmarshalHCL(decoder hcl.Decoder) error {
	if err := decoder.DecodeSlice("records", me); err != nil {
		return err
	}
	return nil
}

type testDecoder struct {
	Values map[string]interface{}
}

func (me *testDecoder) Decode(key string, v interface{}) error {
	return fmt.Errorf("Decode(%v, %T)", key, v)
}

func (me *testDecoder) DecodeAll(map[string]interface{}) error {
	return fmt.Errorf("DecodeAll(%v)", "...")
}

func (me *testDecoder) DecodeSlice(key string, v interface{}) error {
	return fmt.Errorf("DecodeSlice(%v, %T)", key, v)
}

func (me *testDecoder) Get(key string) interface{} {
	return nil
}

func (me *testDecoder) GetChange(key string) (interface{}, interface{}) {
	return nil, nil
}

func (me *testDecoder) GetOkExists(key string) (interface{}, bool) {
	return nil, false
}

func (me *testDecoder) GetOk(key string) (interface{}, bool) {
	if value, found := me.Values[key]; found {
		// fmt.Printf("GetOk(%v) => %v\n", key, value)
		return value, true
	}
	// fmt.Printf("GetOk(%v) not found\n", key)
	return nil, false
}

func (me *testDecoder) GetStringSet(key string) []string {
	return nil
}

func (me *testDecoder) HasChange(key string) bool {
	return false
}

func (me *testDecoder) DecodeAny(map[string]interface{}) (interface{}, error) {
	return nil, nil
}

func (me *testDecoder) MarshalAll(items map[string]interface{}) (hcl.Properties, error) {
	return nil, fmt.Errorf("MarshalAll(%v)", "...")
}

func (me *testDecoder) Reader(unkowns ...map[string]json.RawMessage) hcl.Reader {
	return nil
}

func TestDecodeSlice(t *testing.T) {
	recs := records{}
	decoder := hcl.NewDecoder(&testDecoder{
		Values: map[string]interface{}{
			"rectangle":       2,
			"records.0.value": "value0",
			"records.1.value": "value1",
		},
	})

	if err := recs.UnmarshalHCL(decoder); err != nil {
		t.Error(err)
	}
	for idx, rec := range recs {
		fmt.Printf("%d: %v\n", idx, rec.Value)
	}

}
