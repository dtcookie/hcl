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

func (me schemer) Schema() *schema.Schema {
	return me.schema(me.Field.Type)
}

func (me schemer) schema(t reflect.Type) *schema.Schema {
	switch kind := t.Kind(); kind {
	case reflect.Map, reflect.Interface, reflect.Array, reflect.Uintptr, reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return nil
	case reflect.Struct:
		structSchema := Schema(reflect.New(t).Elem().Interface())
		return &schema.Schema{
			Type:        schema.TypeList,
			Description: me.Documentation,
			MaxItems:    1,
			MinItems:    1,
			Required:    !me.OmitEmpty,
			Optional:    me.OmitEmpty,
			Elem:        &schema.Resource{Schema: structSchema},
		}
	case reflect.Pointer:
		return me.schema(unref(t))
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
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return &schema.Schema{
				Type:        schemaType,
				Description: me.Documentation,
				MinItems:    1,
				Required:    !me.OmitEmpty,
				Optional:    me.OmitEmpty,
				Elem:        &schema.Schema{Type: schema.TypeInt},
			}
		case reflect.Float32, reflect.Float64:
			return &schema.Schema{
				Type:        schemaType,
				Description: me.Documentation,
				MinItems:    1,
				Required:    !me.OmitEmpty,
				Optional:    me.OmitEmpty,
				Elem:        &schema.Schema{Type: schema.TypeFloat},
			}
		case reflect.Struct:
			structSchema := Schema(reflect.New(unref(t.Elem())).Elem().Interface())
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
			return res
		}
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &schema.Schema{
			Type:        schema.TypeInt,
			Description: me.Documentation,
			Required:    !me.OmitEmpty,
			Optional:    me.OmitEmpty,
		}
	case reflect.Bool:
		return &schema.Schema{
			Type:        schema.TypeBool,
			Description: me.Documentation,
			Required:    !me.OmitEmpty,
			Optional:    me.OmitEmpty,
		}
	case reflect.String:
		return &schema.Schema{
			Type:        schema.TypeString,
			Description: me.Documentation,
			Required:    !me.OmitEmpty,
			Optional:    me.OmitEmpty,
		}
	case reflect.Float32, reflect.Float64:
		return &schema.Schema{
			Type:        schema.TypeFloat,
			Description: me.Documentation,
			Required:    !me.OmitEmpty,
			Optional:    me.OmitEmpty,
		}
	default:
		return nil
	}
}
