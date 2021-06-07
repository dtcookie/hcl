package hcl

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
)

// Decoder has no documentation
type MinDecoder interface {
	GetOk(key string) (interface{}, bool)
	Get(key string) interface{}
	GetChange(key string) (interface{}, interface{})
	GetOkExists(key string) (interface{}, bool)
	HasChange(key string) bool
}

// Decoder has no documentation
type Decoder interface {
	GetOk(key string) (interface{}, bool)
	Get(key string) interface{}
	GetChange(key string) (interface{}, interface{})
	GetStringSet(key string) []string
	GetOkExists(key string) (interface{}, bool)
	Reader(unkowns ...map[string]json.RawMessage) Reader
	HasChange(key string) bool
	MarshalAll(items map[string]interface{}) (Properties, error)

	Decode(key string, v interface{}) error
	DecodeAll(map[string]interface{}) error
	DecodeAny(map[string]interface{}) (interface{}, error)

	DecodeSlice(key string, v interface{}) error
}

type mindecoder struct {
	parent MinDecoder
}

func DecoderFrom(m MinDecoder) Decoder {
	return &mindecoder{parent: m}
}

func (d *mindecoder) Decode(key string, v interface{}) error {
	return DecoderFrom(d).Decode(key, v)
}

func (d *mindecoder) DecodeAll(m map[string]interface{}) error {
	return DecoderFrom(d).DecodeAll(m)
}

func (d *mindecoder) DecodeSlice(key string, v interface{}) error {
	return DecoderFrom(d).DecodeSlice(key, v)
}

func (d *mindecoder) DecodeAny(m map[string]interface{}) (interface{}, error) {
	return DecoderFrom(d).DecodeAny(m)
}

func (d *decoder) DecodeAny(m map[string]interface{}) (interface{}, error) {
	if len(m) == 0 {
		return nil, nil
	}
	for k, v := range m {
		found, err := d.decode(k, v)
		if err != nil {
			return nil, err
		}
		if found {
			return v, nil
		}
	}
	return nil, nil
}

func (d *decoder) DecodeAll(m map[string]interface{}) error {
	if len(m) == 0 {
		return nil
	}
	for k, v := range m {
		if err := d.Decode(k, v); err != nil {
			return err
		}
	}
	return nil
}

func (d *decoder) DecodeSlice(key string, v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Type().Kind() != reflect.Ptr || rv.Type().Elem().Kind() != reflect.Slice {
		return fmt.Errorf("decoding slices requires a pointer to a slice to be specified. %T doesn't qualify", v)
	}
	elemType := rv.Type().Elem().Elem()
	if !elemType.Implements(reflect.TypeOf((*Unmarshaler)(nil)).Elem()) {
		return fmt.Errorf("decoding slices requires a pointer to a slice of elements that implement hcl.Unmarshaler to be specified. %T doesn't qualify (%v is not implementing %v)", v, elemType, reflect.TypeOf((*Unmarshaler)(nil)).Elem())
	}
	vSlice := rv.Elem()
	if result, ok := d.GetOk(fmt.Sprintf("%v.#", key)); ok {
		for idx := 0; idx < result.(int); idx++ {
			entry := reflect.New(elemType.Elem()).Interface()
			if err := entry.(Unmarshaler).UnmarshalHCL(NewDecoder(d, key, idx)); err != nil {
				return err
			}
			vSlice.Set(reflect.Append(vSlice, reflect.ValueOf(entry)))
		}
	}

	return nil
}

func (d *decoder) Decode(key string, v interface{}) error {
	_, err := d.decode(key, v)
	return err
}

var stringType = reflect.TypeOf("")

func (d *decoder) decode(key string, v interface{}) (bool, error) {
	vTarget := reflect.ValueOf(v)
	if !vTarget.IsValid() || vTarget.IsNil() {
		return false, errors.New("passed an invalid target value to Decode()")
	}
	if unmarshaler, ok := v.(Unmarshaler); ok {
		if _, ok := d.GetOk(fmt.Sprintf("%v.#", key)); ok {
			if err := unmarshaler.UnmarshalHCL(NewDecoder(d, key, 0)); err != nil {
				return true, err
			}
			return true, nil
		}
	}
	if vTarget.Type().Kind() != reflect.Ptr {
		return false, fmt.Errorf("Decode (%v) requires a pointer to store results into", key)
	}
	if result, ok := d.GetOk(key); ok {
		vTarget := vTarget.Elem()
		vResult := reflect.ValueOf(result)
		tResult := vResult.Type()
		tTarget := vTarget.Type()
		if tResult == stringType {
			if tTarget.Kind() == reflect.String {
				if tTarget != stringType {
					vTarget.Set(vResult.Convert(tTarget))
					return true, nil
				}
			}
			if tTarget.Kind() == reflect.Ptr {
				tTarget = tTarget.Elem()
				if tTarget.Kind() == reflect.String {
					if tTarget != stringType {
						tEnum := reflect.ValueOf(v).Type().Elem().Elem()
						vEnumPtr := reflect.New(tEnum)
						vEnum := vEnumPtr.Elem()
						vEnum.Set(vResult.Convert(tEnum))
						vTarget.Set(vEnumPtr)
						return true, nil
					}
				}
			}
		}
		if vResult.Type().AssignableTo(vTarget.Type()) {
			vTarget.Set(vResult)
		}
		log.Printf("cannot assign type %v to values of type %v", vResult.Type(), vTarget.Type())
		return true, nil
	}
	return false, nil
}

func (d *mindecoder) GetStringSet(key string) []string {
	result := []string{}
	if value, ok := d.GetOk(key); ok {
		if arr, ok := value.([]interface{}); ok {
			for _, elem := range arr {
				result = append(result, elem.(string))
			}
		} else if set, ok := value.(Set); ok {
			if set.Len() > 0 {
				for _, elem := range set.List() {
					result = append(result, elem.(string))
				}
			}
		}
	}
	return result
}

func (d *mindecoder) Reader(unkowns ...map[string]json.RawMessage) Reader {
	if len(unkowns) > 0 {
		return NewReader(d, unkowns[0])
	}
	return NewReader(d, nil)
}

func (d *mindecoder) MarshalAll(items map[string]interface{}) (Properties, error) {
	properties := Properties{}
	if err := properties.MarshalAll(d, items); err != nil {
		return nil, err
	}
	return properties, nil
}

func (d *mindecoder) GetOk(key string) (interface{}, bool) {
	return d.parent.GetOk(key)
}

func (d *mindecoder) HasChange(key string) bool {
	return d.parent.HasChange(key)
}

func (d *mindecoder) GetOkExists(key string) (interface{}, bool) {
	return d.parent.GetOkExists(key)
}

func (d *mindecoder) GetChange(key string) (interface{}, interface{}) {
	return d.parent.GetChange(key)
}

func (d *mindecoder) Get(key string) interface{} {
	return d.parent.Get(key)
}

// NewDecoder has no documentation
func NewDecoder(parent Decoder, address ...interface{}) Decoder {
	joined := ""
	sep := ""
	for _, part := range address {
		joined = joined + sep + fmt.Sprintf("%v", part)
		sep = "."
	}
	return &decoder{parent: parent, address: joined}
}

type decoder struct {
	parent  Decoder
	address string
}

func (d *decoder) Reader(unkowns ...map[string]json.RawMessage) Reader {
	if len(unkowns) > 0 {
		return NewReader(d, unkowns[0])
	}
	return NewReader(d, nil)
}

func (d *decoder) MarshalAll(items map[string]interface{}) (Properties, error) {
	properties := Properties{}
	if err := properties.MarshalAll(d, items); err != nil {
		return nil, err
	}
	return properties, nil
}

func (d *decoder) HasChange(key string) bool {
	if d.address == "" {
		return d.parent.HasChange(key)
	}
	return d.parent.HasChange(d.address + "." + key)
}

func (d *decoder) GetStringSet(key string) []string {
	result := []string{}
	if value, ok := d.GetOk(key); ok {
		if arr, ok := value.([]interface{}); ok {
			for _, elem := range arr {
				result = append(result, elem.(string))
			}
		} else if set, ok := value.(Set); ok {
			if set.Len() > 0 {
				for _, elem := range set.List() {
					result = append(result, elem.(string))
				}
			}
		}
	}
	return result
}

// GetOk returns the data for the given key and whether or not the key
// has been set to a non-zero value at some point.
//
// The first result will not necessarilly be nil if the value doesn't exist.
// The second result should be checked to determine this information.
func (d *decoder) GetOk(key string) (interface{}, bool) {
	if d.address == "" {
		return d.parent.GetOk(key)
	}
	return d.parent.GetOk(d.address + "." + key)
}

func (d *decoder) GetOkExists(key string) (interface{}, bool) {
	if d.address == "" {
		return d.parent.GetOkExists(key)
	}
	return d.parent.GetOkExists(d.address + "." + key)
}

func (d *decoder) GetChange(key string) (interface{}, interface{}) {
	if d.address == "" {
		return d.parent.GetChange(key)
	}
	return d.parent.GetChange(d.address + "." + key)
}

// Get returns the data for the given key, or nil if the key doesn't exist
// in the schema.
//
// If the key does exist in the schema but doesn't exist in the configuration,
// then the default value for that type will be returned. For strings, this is
// "", for numbers it is 0, etc.
//
// If you want to test if something is set at all in the configuration,
// use GetOk.
func (d *decoder) Get(key string) interface{} {
	if d.address == "" {
		return d.parent.Get(key)
	}
	return d.parent.Get(d.address + "." + key)
}

func VoidDecoder() Decoder {
	return &voidDecoder{}
}

type voidDecoder struct{}

func (d *voidDecoder) DecodeAny(m map[string]interface{}) (interface{}, error) {
	return nil, nil
}

func (vd *voidDecoder) GetOk(key string) (interface{}, bool) {
	return nil, false
}

func (vd *voidDecoder) Get(key string) interface{} {
	return nil
}

func (vd *voidDecoder) GetChange(key string) (interface{}, interface{}) {
	return nil, false
}

func (vd *voidDecoder) GetStringSet(key string) []string {
	return nil
}

func (vd *voidDecoder) GetOkExists(key string) (interface{}, bool) {
	return nil, false
}

func (vd *voidDecoder) Decode(key string, v interface{}) error {
	return nil
}

func (d *voidDecoder) DecodeAll(m map[string]interface{}) error {
	return nil
}

func (vd *voidDecoder) Reader(unkowns ...map[string]json.RawMessage) Reader {
	if len(unkowns) > 0 {
		return NewReader(vd, unkowns[0])
	}
	return NewReader(vd, nil)
}

func (vd *voidDecoder) HasChange(key string) bool {
	return false
}

func (vd *voidDecoder) MarshalAll(items map[string]interface{}) (Properties, error) {
	properties := Properties{}
	if err := properties.MarshalAll(vd, items); err != nil {
		return nil, err
	}
	return properties, nil
}

func (vd *voidDecoder) DecodeSlice(key string, v interface{}) error {
	return nil
}
