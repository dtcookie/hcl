package hcl

import (
	"fmt"
	"log"
)

// Decoder has no documentation
type Decoder interface {
	GetOk(key string) (interface{}, bool)
	Get(key string) interface{}
	// Append(key string) Resource
}

// NewDecoder has no documentation
func NewDecoder(parent Decoder, address ...interface{}) Decoder {
	joined := ""
	sep := ""
	for _, part := range address {
		joined = joined + sep + fmt.Sprintf("%v", part)
		sep = "."
	}
	return &decoder{parent: parent, address: joined}
}

type decoder struct {
	parent  Decoder
	address string
}

// GetOk returns the data for the given key and whether or not the key
// has been set to a non-zero value at some point.
//
// The first result will not necessarilly be nil if the value doesn't exist.
// The second result should be checked to determine this information.
func (d *decoder) GetOk(key string) (interface{}, bool) {
	if d.address == "" {
		return d.parent.GetOk(key)
	}
	log.Println("GetOk", d.address+"."+key)
	return d.parent.GetOk(d.address + "." + key)
}

// Get returns the data for the given key, or nil if the key doesn't exist
// in the schema.
//
// If the key does exist in the schema but doesn't exist in the configuration,
// then the default value for that type will be returned. For strings, this is
// "", for numbers it is 0, etc.
//
// If you want to test if something is set at all in the configuration,
// use GetOk.
func (d *decoder) Get(key string) interface{} {
	if d.address == "" {
		return d.parent.Get(key)
	}
	log.Println("Get", d.address+"."+key)
	return d.parent.Get(d.address + "." + key)
}
