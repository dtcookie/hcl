package hcl

import (
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type schemer handler

func unref(t reflect.Type) reflect.Type {
	if t == nil {
		return nil
	}
	switch kind := t.Kind(); kind {
	case reflect.Pointer:
		return unref(t.Elem())
	}
	return t
}

func (me schemer) Schema() (*schema.Schema, error) {
	return me.schema(me.Field.Type)
}

func (me schemer) schema(t reflect.Type) (*schema.Schema, error) {
	switch kind := t.Kind(); kind {
	case reflect.Map, reflect.Interface, reflect.Array, reflect.Uintptr, reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return nil, UnsupportedTypeError{me.Field.Name, t}
	case reflect.Struct:
		structSchema, err := Schema(reflect.New(t).Elem().Interface())
		if err != nil {
			return nil, err
		}
		return &schema.Schema{
			Type:        schema.TypeList,
			Description: me.Documentation,
			MaxItems:    1,
			MinItems:    1,
			Required:    !me.OmitEmpty,
			Optional:    me.OmitEmpty,
			Elem:        &schema.Resource{Schema: structSchema},
		}, nil
	case reflect.Pointer:
		sch, err := me.schema(unref(t))
		if err == nil {
			return sch, nil
		}
		if _, ok := err.(UnsupportedTypeError); ok {
			return nil, UnsupportedTypeError{me.Field.Name, t}
		}
		return nil, err
	case reflect.Slice:
		schemaType := schema.TypeList
		if me.Unordered {
			schemaType = schema.TypeSet
		}
		switch elemKind := unref(t.Elem()).Kind(); elemKind {
		case reflect.String:
			return &schema.Schema{
				Type:        schemaType,
				Description: me.Documentation,
				MinItems:    1,
				Required:    !me.OmitEmpty,
				Optional:    me.OmitEmpty,
				Elem:        &schema.Schema{Type: schema.TypeString},
			}, nil
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return &schema.Schema{
				Type:        schemaType,
				Description: me.Documentation,
				MinItems:    1,
				Required:    !me.OmitEmpty,
				Optional:    me.OmitEmpty,
				Elem:        &schema.Schema{Type: schema.TypeInt},
			}, nil
		case reflect.Float32, reflect.Float64:
			return &schema.Schema{
				Type:        schemaType,
				Description: me.Documentation,
				MinItems:    1,
				Required:    !me.OmitEmpty,
				Optional:    me.OmitEmpty,
				Elem:        &schema.Schema{Type: schema.TypeFloat},
			}, nil
		case reflect.Struct:
			structSchema, err := Schema(reflect.New(unref(t.Elem())).Elem().Interface())
			if err != nil {
				return nil, err
			}
			res := &schema.Schema{
				Type:        schemaType,
				Description: me.Documentation,
				MinItems:    1,
				Required:    !me.OmitEmpty,
				Optional:    me.OmitEmpty,
				Elem:        &schema.Resource{Schema: structSchema},
			}
			if len(me.Elem) > 0 {
				res = &schema.Schema{
					Type:        schema.TypeList,
					Description: me.Documentation,
					MinItems:    1,
					MaxItems:    1,
					Required:    !me.OmitEmpty,
					Optional:    me.OmitEmpty,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							me.Elem: {
								Type:        schemaType,
								Description: me.Documentation,
								MinItems:    1,
								Required:    true,
								Elem:        &schema.Resource{Schema: structSchema},
							},
						},
					},
				}
			}
			return res, nil
		}
		return nil, UnsupportedTypeError{me.Field.Name, t}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &schema.Schema{
			Type:        schema.TypeInt,
			Description: me.Documentation,
			Required:    !me.OmitEmpty,
			Optional:    me.OmitEmpty,
		}, nil
	case reflect.Bool:
		return &schema.Schema{
			Type:        schema.TypeBool,
			Description: me.Documentation,
			Required:    !me.OmitEmpty,
			Optional:    me.OmitEmpty,
		}, nil
	case reflect.String:
		return &schema.Schema{
			Type:        schema.TypeString,
			Description: me.Documentation,
			Required:    !me.OmitEmpty,
			Optional:    me.OmitEmpty,
		}, nil
	case reflect.Float32, reflect.Float64:
		return &schema.Schema{
			Type:        schema.TypeFloat,
			Description: me.Documentation,
			Required:    !me.OmitEmpty,
			Optional:    me.OmitEmpty,
		}, nil
	default:
		return nil, UnsupportedTypeError{me.Field.Name, t}
	}
}

// func schemaFieldHandlersFor(refType reflect.Type) []*schemaFieldHandler {
// 	handlers := []*schemaFieldHandler{}
// 	for idx := 0; idx < refType.NumField(); idx++ {
// 		field := refType.Field(idx)
// 		if field.Anonymous {
// 			handlers = append(handlers, schemaFieldHandlersFor(field.Type)...)
// 			continue
// 		}
// 		if !field.IsExported() {
// 			continue
// 		}
// 		documentation := NoDocumentationAvailable
// 		tag := field.Tag
// 		if tagValue := tag.Get("doc"); len(tagValue) > 0 {
// 			documentation = tagValue
// 		}
// 		if tagValue := tag.Get("hcl"); len(tagValue) > 0 {
// 			tagValues := strings.Split(tagValue, ",")
// 			tv := tagValues[0]
// 			if tv == "-" {
// 				continue
// 			} else if tv == "" || strings.Contains(tv, "=") {
// 				tv = camel.Strip(field.Name)
// 			}
// 			handler := &schemaFieldHandler{Field: field, OmitEmpty: false, Property: tv, Documentation: documentation}
// 			for _, tagValue := range tagValues {
// 				if tagValue == "omitempty" {
// 					handler.OmitEmpty = true
// 				}
// 				if tagValue == "unordered" {
// 					handler.Unordered = true
// 				}
// 				if strings.HasPrefix(tagValue, "elem=") {
// 					handler.Elem = strings.TrimSpace(strings.TrimPrefix(tagValue, "elem="))
// 				}
// 			}
// 			handlers = append(handlers, handler)
// 		} else if tagValue := tag.Get("json"); len(tagValue) > 0 {
// 			tagValues := strings.Split(tagValue, ",")
// 			tv := tagValues[0]
// 			if tv == "-" {
// 				continue
// 			} else if tv == "" {
// 				tv = field.Name
// 			}
// 			handler := &schemaFieldHandler{Field: field, OmitEmpty: false, Property: camel.Strip(tv), Documentation: documentation}
// 			for _, tagValue := range tagValues {
// 				if tagValue == "omitempty" {
// 					handler.OmitEmpty = true
// 				}
// 			}
// 			handlers = append(handlers, handler)
// 		} else {
// 			handler := &schemaFieldHandler{Field: field, OmitEmpty: false, Property: camel.Strip(field.Name), Documentation: documentation}
// 			handlers = append(handlers, handler)
// 		}
// 	}
// 	return handlers
// }
