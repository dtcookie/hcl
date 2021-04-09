package hcl

// Marshaler has no documentation
type Marshaler interface {
	MarshalHCL(enc Encoder) error
}
