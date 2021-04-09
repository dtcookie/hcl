package hcl

type Schema struct {
	Type        ValueType
	Description string
	Optional    bool
	MaxItems    int
	MinItems    int
	Sensitive   bool
	Elem        interface{}
	Default     interface{}
	Required    bool
}

type ValueType int

const (
	TypeInvalid ValueType = iota
	TypeBool
	TypeInt
	TypeFloat
	TypeString
	TypeList
	TypeMap
	TypeSet
)
