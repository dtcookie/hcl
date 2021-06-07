package hcl_test

import (
	"testing"

	"github.com/dtcookie/hcl"
	"github.com/dtcookie/opt"
)

type StringEnum string

func (me StringEnum) Ref() *StringEnum {
	return &me
}

func TestMarshaller(t *testing.T) {
	{
		properties := hcl.Properties{}
		obj := struct{ EnumRef *StringEnum }{EnumRef: StringEnum("asdf").Ref()}
		if err := properties.Marshal(hcl.VoidDecoder(), "asdf", obj.EnumRef); err != nil {
			t.Error(err)
		}
		if len(properties) != 1 {
			t.Fail()
		}
		if stored, found := properties["asdf"]; !found {
			t.Fail()
		} else {
			switch tStored := stored.(type) {
			case string:
				if tStored != "asdf" {
					t.Fail()
				}
			default:
				t.Fail()
			}
		}

	}

	{
		properties := hcl.Properties{}
		obj := struct{ EnumRef *StringEnum }{EnumRef: nil}
		if err := properties.Marshal(hcl.VoidDecoder(), "asdf", obj.EnumRef); err != nil {
			t.Error(err)
		}
		if len(properties) != 0 {
			t.Fail()
		}

	}

	{
		properties := hcl.Properties{}
		obj := struct{ OptString *string }{OptString: opt.NewString("asdf")}
		if err := properties.Marshal(hcl.VoidDecoder(), "asdf", obj.OptString); err != nil {
			t.Error(err)
		}
		if len(properties) != 1 {
			t.Fail()
		}
		if stored, found := properties["asdf"]; !found {
			t.Fail()
		} else {
			switch tStored := stored.(type) {
			case string:
				if tStored != "asdf" {
					t.Fail()
				}
			default:
				t.Fail()
			}
		}

	}
	{
		properties := hcl.Properties{}
		obj := struct{ OptString *string }{OptString: nil}
		if err := properties.Marshal(hcl.VoidDecoder(), "asdf", obj.OptString); err != nil {
			t.Error(err)
		}
		if len(properties) != 0 {
			t.Fail()
		}
	}

	{
		properties := hcl.Properties{}
		obj := struct{ EnumRef StringEnum }{EnumRef: StringEnum("asdf")}
		if err := properties.Marshal(hcl.VoidDecoder(), "asdf", obj.EnumRef); err != nil {
			t.Error(err)
		}
		if len(properties) != 1 {
			t.Fail()
		}
		if stored, found := properties["asdf"]; !found {
			t.Fail()
		} else {
			switch tStored := stored.(type) {
			case string:
				if tStored != "asdf" {
					t.Fail()
				}
			default:
				t.Fail()
			}
		}

	}
}
