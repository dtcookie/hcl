package hcl

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Marshal(ctx context.Context, v interface{}) (map[string]interface{}, error) {
	result := map[string]interface{}{}

	for _, handler := range evalHandlers(reflect.TypeOf(v), v) {
		fieldValue, err := encoder(handler).encode(ctx)
		if err != nil {
			return nil, fmt.Errorf("cannot serialize field '%s': %s", handler.Field.Name, err.Error())
		}
		if fieldValue != emptyValue {
			result[handler.Property] = fieldValue
		}
	}
	return result, nil
}

func Unmarshal(ctx context.Context, rd ResourceData, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return &InvalidUnmarshalError{reflect.TypeOf(v)}
	}
	for _, handler := range evalHandlers(reflect.TypeOf(v).Elem(), rv.Elem().Interface()) {
		fieldValue, err := decoder(handler).Decode(ctx, rd, handler.Field.Type)
		if err != nil {
			return fmt.Errorf("cannot deserialize field '%s': %s", handler.Field.Name, err.Error())
		}
		if fieldValue != nil {
			set(rv.Elem().FieldByIndex(handler.Field.Index), fieldValue)
		}
	}
	return nil
}

func Schema(v interface{}) map[string]*schema.Schema {
	result := map[string]*schema.Schema{}

	for _, handler := range evalHandlers(reflect.TypeOf(v), v) {
		fieldValue := schemer(handler).Schema()
		if fieldValue != nil {
			result[handler.Property] = fieldValue
		}
	}
	return result
}

type InvalidUnmarshalError struct {
	Type reflect.Type
}

func (e *InvalidUnmarshalError) Error() string {
	if e.Type == nil {
		return "hcl: DeSerialize(nil)"
	}

	if e.Type.Kind() != reflect.Pointer {
		return "hcl: DeSerialize(non-pointer " + e.Type.String() + ")"
	}
	return "hcl: DeSerialize(nil " + e.Type.String() + ")"
}

type ResourceData interface {
	GetOk(key string) (interface{}, bool)
}

type resourceData struct {
	parent ResourceData
	prefix string
}

func (me *resourceData) GetOk(key string) (interface{}, bool) {
	return me.parent.GetOk(fmt.Sprintf("%s.%s", me.prefix, key))
}

func set(target reflect.Value, v any) {
	if target.Type().Kind() == reflect.Pointer {
		newTarget := reflect.New(target.Type().Elem())
		set(newTarget.Elem(), v)
		target.Set(newTarget)
		return
	}
	target.Set(reflect.ValueOf(v))
}
