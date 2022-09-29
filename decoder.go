package hcl

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type decoder handler

func (me decoder) Decode(rd ResourceData, t reflect.Type) (any, error) {
	return me.decode(rd, t)
}

func setTo(rv reflect.Value, v any) {
	if rv.Type().Kind() == reflect.Pointer {
		referencedValue := reflect.New(rv.Type().Elem()).Elem()
		rv.Set(referencedValue.Addr())
		setTo(referencedValue, v)
		return
	}
	rv.Set(reflect.ValueOf(v).Convert(rv.Type()))
}

func (me decoder) decodePointer(rd ResourceData, t reflect.Type) (any, error) {
	sch, err := me.decode(rd, unref(t))
	if err == nil {
		return sch, err
	}
	if _, ok := err.(UnsupportedTypeError); ok {
		return nil, UnsupportedTypeError{me.Field.Name, t}
	}
	return nil, err
}

func (me decoder) decodeStruct(rd ResourceData, t reflect.Type) (any, error) {
	if value, ok := rd.GetOk(me.Property + ".#"); !ok || value == nil || value.(int) == 0 {
		return nil, nil
	}
	targetStructPointer := reflect.New(t).Interface()
	if err := Unmarshal(&resourceData{parent: rd, prefix: me.Property + ".0"}, targetStructPointer); err != nil {
		return nil, err
	}
	return reflect.ValueOf(targetStructPointer).Elem().Interface(), nil
}

func (me decoder) decodePrimitiveSlice(rd ResourceData, t reflect.Type) (any, error) {
	value, ok := rd.GetOk(me.Property)
	if !ok {
		return nil, nil
	}
	targetSlice := reflect.New(t).Elem()
	switch typedValue := value.(type) {
	case []any:
		if me.Unordered {
			fmt.Println("data type is []any, but *schema.Set was expected")
		}
		for _, el := range typedValue {
			targetElem := reflect.New(t.Elem()).Elem()
			setTo(targetElem, el)
			targetSlice = reflect.Append(targetSlice, targetElem)
		}
	case *schema.Set:
		if !me.Unordered {
			fmt.Println("data type is *schema.Set, but []any was expected")
		}
		for _, el := range typedValue.List() {
			targetElem := reflect.New(t.Elem()).Elem()
			setTo(targetElem, el)
			targetSlice = reflect.Append(targetSlice, targetElem)
		}
	}

	return targetSlice.Interface(), nil
}

func (me decoder) decodeStructSlice(rd ResourceData, t reflect.Type) (any, error) {
	value, ok := rd.GetOk(me.Property + ".#")
	if !ok || value == nil || value.(int) == 0 {
		return nil, nil
	}
	targetSlice := reflect.New(t).Elem()
	if me.Unordered {
		untypedResourceSet, ok := rd.GetOk(me.Property)
		if !ok {
			return nil, errors.New("ok expected")
		}
		resourceSet := untypedResourceSet.(*schema.Set)
		for _, resource := range resourceSet.List() {
			targetStructPointer := reflect.New(unref(t.Elem())).Interface()
			if err := Unmarshal(&resourceData{parent: rd, prefix: fmt.Sprintf("%s.%d", me.Property, resourceSet.F(resource))}, targetStructPointer); err != nil {
				return nil, err
			}
			targetElem := reflect.New(t.Elem()).Elem()
			setTo(targetElem, reflect.ValueOf(targetStructPointer).Elem().Interface())
			targetSlice = reflect.Append(targetSlice, targetElem)
		}
	} else {
		for idx := 0; idx < value.(int); idx++ {
			targetStructPointer := reflect.New(unref(t.Elem())).Interface()
			if err := Unmarshal(&resourceData{parent: rd, prefix: fmt.Sprintf("%s.%d", me.Property, idx)}, targetStructPointer); err != nil {
				return nil, err
			}
			targetElem := reflect.New(t.Elem()).Elem()
			setTo(targetElem, reflect.ValueOf(targetStructPointer).Elem().Interface())
			targetSlice = reflect.Append(targetSlice, targetElem)
		}
	}
	return targetSlice.Interface(), nil
}

func (me decoder) decodeSlice(rd ResourceData, t reflect.Type) (any, error) {
	switch elemKind := unref(t.Elem()).Kind(); elemKind {
	case reflect.Float32, reflect.Float64, reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return me.decodePrimitiveSlice(rd, t)
	case reflect.Struct:
		return me.decodeStructSlice(rd, t)
	}
	return nil, UnsupportedTypeError{me.Field.Name, t}
}

func (me decoder) decodePrimitive(rd ResourceData, t reflect.Type) (any, error) {
	value, ok := rd.GetOk(me.Property)
	if ok {
		rv := reflect.New(t).Elem()
		rv.Set(reflect.ValueOf(value).Convert(t))
		return rv.Interface(), nil
	}
	return nil, nil
}

func (me decoder) decode(rd ResourceData, t reflect.Type) (any, error) {
	switch kind := t.Kind(); kind {
	case reflect.Map, reflect.Interface, reflect.Array, reflect.Uintptr, reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return nil, UnsupportedTypeError{me.Field.Name, t}
	case reflect.Struct:
		return me.decodeStruct(rd, t)
	case reflect.Pointer:
		return me.decodePointer(rd, t)
	case reflect.Slice:
		return me.decodeSlice(rd, t)
	case reflect.Float32, reflect.Float64, reflect.String, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return me.decodePrimitive(rd, t)
	default:
		return nil, UnsupportedTypeError{me.Field.Name, t}
	}
}
