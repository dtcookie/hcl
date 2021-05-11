package hcl

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
)

type exportEntry interface {
	Write(w io.Writer, indent string) error
}

type exportEntries []exportEntry

func (e *exportEntries) eval(key string, value interface{}) {
	switch v := value.(type) {
	case string, bool, int, int32, int64, int8, int16, uint, uint32, uint64, uint8, uint16, float32, float64:
		entry := &primitiveEntry{Key: key, Value: value}
		*e = append(*e, entry)
	case *string, *bool, *int, *int32, *int64, *int8, *int16, *uint, *uint32, *uint64, *uint8, *uint16, *float32, *float64:
		if v == nil {
			return
		}
		entry := &primitiveEntry{Key: key, Value: v}
		*e = append(*e, entry)
	case []interface{}:
		if len(v) == 0 {
			return
		}
		switch typedElem := v[0].(type) {
		case map[string]interface{}:
			for _, elem := range v {
				entry := &resourceEntry{Key: key, Entries: exportEntries{}}
				entry.Entries.handle(elem.(map[string]interface{}))
				*e = append(*e, entry)
			}
		case string, bool, int, int32, int64, int8, int16, uint, uint32, uint64, uint8, uint16, float32, float64:
			entry := &primitiveEntry{Key: key, Value: value}
			*e = append(*e, entry)
		default:
			panic(fmt.Sprintf("unsupported elem type %T", typedElem))
		}
	case []string:
		if len(v) == 0 {
			return
		}
		entry := &primitiveEntry{Key: key, Value: value}
		*e = append(*e, entry)
	default:
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.String:
			e.eval(key, fmt.Sprintf("%v", v))
		default:
			panic(fmt.Sprintf(">>>>> [%v] type %T not supported yet\n", key, v))
		}

	}
}

func (e *exportEntries) handle(m map[string]interface{}) {
	for k, v := range m {
		e.eval(k, v)
	}
}

func Export(marshaler Marshaler, w io.Writer) error {
	var m map[string]interface{}
	var err error
	if m, err = marshaler.MarshalHCL(); err != nil {
		return err
	}
	ents := exportEntries{}
	ents.handle(m)
	for _, entry := range ents {
		if err := entry.Write(w, "  "); err != nil {
			return err
		}
		if _, err := w.Write([]byte("\n")); err != nil {
			return err
		}
	}
	return err
}

type primitiveEntry struct {
	Indent string
	Key    string
	Value  interface{}
}

func jsonenc(v interface{}) string {
	bytes, _ := json.Marshal(v)
	return string(bytes)
}

func (pe *primitiveEntry) Write(w io.Writer, indent string) error {
	_, err := w.Write([]byte(fmt.Sprintf("%s%v = %v", indent, pe.Key, jsonenc(pe.Value))))
	return err
}

type resourceEntry struct {
	Indent  string
	Key     string
	Entries exportEntries
}

func (re *resourceEntry) Write(w io.Writer, indent string) error {
	s := fmt.Sprintf("%s%v {\n", indent, re.Key)
	if _, err := w.Write([]byte(s)); err != nil {
		return err
	}
	for _, entry := range re.Entries {
		if err := entry.Write(w, indent+"  "); err != nil {
			return err
		}
		if _, err := w.Write([]byte("\n")); err != nil {
			return err
		}
	}
	if _, err := w.Write([]byte(indent + "}")); err != nil {
		return err
	}
	return nil
}
