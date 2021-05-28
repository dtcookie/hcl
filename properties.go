package hcl

import (
	"fmt"
	"reflect"
)

type Properties map[string]interface{}

func (me Properties) MarshalAll(decoder Decoder, items map[string]interface{}) error {
	if items == nil {
		return nil
	}
	for k, v := range items {
		if err := me.Marshal(decoder, k, v); err != nil {
			return err
		}
	}
	return nil
}

func (me Properties) Marshal(decoder Decoder, key string, v interface{}) error {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case string:
		me[key] = t
	case int:
		me[key] = t
	case float64:
		me[key] = t
	case bool:
		me[key] = t
	case int8:
		me[key] = int(t)
	case int16:
		me[key] = int(t)
	case int32:
		me[key] = int(t)
	case int64:
		me[key] = int(t)
	case uint:
		me[key] = int(t)
	case uint8:
		me[key] = int(t)
	case uint16:
		me[key] = int(t)
	case uint32:
		me[key] = int(t)
	case uint64:
		me[key] = int(t)
	case float32:
		me[key] = float64(t)
	default:
		if marshaller, ok := v.(ExtMarshaler); ok {
			if marshalled, err := marshaller.MarshalHCL(NewDecoder(decoder, key, 0)); err == nil {
				me[key] = []interface{}{marshalled}
			} else {
				return err
			}
		}
		switch reflect.TypeOf(v).Kind() {
		case reflect.String:
			me[key] = fmt.Sprintf("%v", v)
		default:
			panic(fmt.Sprintf("unsupported type %T", v))
		}
	}
	return nil
}
