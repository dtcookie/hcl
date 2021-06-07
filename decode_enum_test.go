package hcl_test

import (
	"testing"

	"github.com/dtcookie/hcl"
)

type Enum string

type EnumContainer struct {
	Enum    Enum
	OptEnum *Enum
}

func (me *EnumContainer) UnmarshalHCL(decoder hcl.Decoder) error {
	if err := decoder.Decode("enum", &me.Enum); err != nil {
		return err
	}
	if err := decoder.Decode("opt_enum", &me.OptEnum); err != nil {
		return err
	}
	return nil
}

func TestDecodeEnum(t *testing.T) {
	decoder := hcl.NewDecoder(&testDecoder{
		Values: map[string]interface{}{
			"enum":     "Test",
			"opt_enum": "OptTest",
		},
	})
	ec := &EnumContainer{}
	if err := ec.UnmarshalHCL(decoder); err != nil {
		t.Error(err)
	}
	if string(ec.Enum) != "Test" {
		t.Errorf("expected: %v, actual: %v", "Test", ec.Enum)
	}
	if string(*ec.OptEnum) != "OptTest" {
		t.Errorf("expected: %v, actual: %v", "OptTest", ec.OptEnum)
	}
}
