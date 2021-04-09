package hcl

import (
	"fmt"
	"reflect"
)

// Encoder has no documentation
type Encoder interface {
	Encode(key string, v interface{}, optional bool)
}

// NewEncoder has no documentation
func NewEncoder() Encoder {
	return &encoder{
		properties: map[string]interface{}{},
	}
}

type encoder struct {
	properties map[string]interface{}
}

var marshalerType = reflect.TypeOf((*Marshaler)(nil)).Elem()

func ind(n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result = result + "    "
	}
	return result
}

func unref(v reflect.Value, indent int) reflect.Value {
	// fmt.Println(fmt.Sprintf("%sunref(type: %v, kind: %v)", ind(indent), v.Type(), v.Kind()))
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			// nil pointers don't need to get dereferenced any futher
			// fmt.Println(fmt.Sprintf("%s-->%v", ind(indent), v.Type()))
			return v
		}
		typ := v.Type()
		// best case scenario: the value is a *Marshaler
		if typ.Implements(marshalerType) {
			// fmt.Println(fmt.Sprintf("%s-->%v", ind(indent), v.Type()))
			return v
		}
		switch typ.Elem().Kind() {
		case reflect.Struct:
			// if the given value points to a struct, but the pointer to the struct doesn't implement Marshaler, we don't dereference futher
			// otherwise we'd allocate a duplicate
			// fmt.Println(fmt.Sprintf("%s-->%v", ind(indent), v.Type()))
			return v
		default:
			result := unref(v.Elem(), indent+1)
			// fmt.Println(fmt.Sprintf("%s-->%v", ind(indent), result.Type()))
			return result
		}
	default:
		// fmt.Println(fmt.Sprintf("%s-->%v", ind(indent), v.Type()))
		return v
	}
}

func (e *encoder) Encode(key string, v interface{}, optional bool) {
	rv := unref(reflect.ValueOf(v), 0)
	v = rv.Interface()
	switch rv.Kind() {
	case reflect.String, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
		if !optional {
			e.properties[key] = v
		}
	case reflect.Interface:
		if m, ok := v.(Marshaler); ok {
			e2 := &encoder{properties: map[string]interface{}{}}
			if err := m.MarshalHCL(e2); err != nil {
				panic(err)
			}
			e.properties[key] = []interface{}{e2.properties}
		} else {
			panic("The interface passed is not a Marshaler")
		}
	case reflect.Slice:
		values := []interface{}{}
		len := rv.Len()
		for i := 0; i < len; i++ {
			vElem := rv.Index(i)
			if vElem.Kind() == reflect.Struct {
				vElem = vElem.Addr()
			}
			m := vElem.Interface().(Marshaler)
			e2 := &encoder{properties: map[string]interface{}{}}
			if err := m.MarshalHCL(e2); err != nil {
				panic(err)
			}
			values = append(values, e2.properties)
		}
		e.properties[key] = values
	case reflect.Array:
	case reflect.Map:
	case reflect.Ptr:
		if m, ok := v.(Marshaler); ok {
			if rv.IsNil() {
				if !optional {
					e.properties[key] = []interface{}{nil}
				}
			} else {
				e2 := &encoder{properties: map[string]interface{}{}}
				if err := m.MarshalHCL(e2); err != nil {
					panic(err)
				}
				e.properties[key] = []interface{}{e2.properties}
			}
		} else if rv.IsNil() {
			if !optional {
				e.properties[key] = nil
			}
		} else {
			panic(fmt.Sprintf("The interface passed for key '%s' is not a Marshaler (value: %v)", key, v))
		}
	case reflect.Struct:
		panic("Passing structs to the encoder is forbidden. Use a pointer to that struct instead.")
	default:
	}

}
