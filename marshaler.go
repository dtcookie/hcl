package hcl

// Marshaler has no documentation
type Marshaler interface {
	MarshalHCL() (map[string]interface{}, error)
}
