package hcl

import (
	"fmt"
)

// Decoder has no documentation
type MinDecoder interface {
	GetOk(key string) (interface{}, bool)
	Get(key string) interface{}
	GetChange(key string) (interface{}, interface{})
	GetOkExists(key string) (interface{}, bool)
}

// Decoder has no documentation
type Decoder interface {
	GetOk(key string) (interface{}, bool)
	Get(key string) interface{}
	GetChange(key string) (interface{}, interface{})
	GetStringSet(key string) []string
	GetOkExists(key string) (interface{}, bool)
}

type mindecoder struct {
	parent MinDecoder
}

func DecoderFrom(m MinDecoder) Decoder {
	return &mindecoder{parent: m}
}

func (d *mindecoder) GetStringSet(key string) []string {
	result := []string{}
	if value, ok := d.GetOk(key); ok {
		if arr, ok := value.([]interface{}); ok {
			for _, elem := range arr {
				result = append(result, elem.(string))
			}
		} else if set, ok := value.(Set); ok {
			if set.Len() > 0 {
				for _, elem := range set.List() {
					result = append(result, elem.(string))
				}
			}
		}
	}
	return result
}

func (d *mindecoder) GetOk(key string) (interface{}, bool) {
	return d.parent.GetOk(key)
}

func (d *mindecoder) GetOkExists(key string) (interface{}, bool) {
	return d.parent.GetOkExists(key)
}

func (d *mindecoder) GetChange(key string) (interface{}, interface{}) {
	return d.parent.GetChange(key)
}

func (d *mindecoder) Get(key string) interface{} {
	return d.parent.Get(key)
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

func (d *decoder) GetStringSet(key string) []string {
	result := []string{}
	if value, ok := d.GetOk(key); ok {
		if arr, ok := value.([]interface{}); ok {
			for _, elem := range arr {
				result = append(result, elem.(string))
			}
		} else if set, ok := value.(Set); ok {
			if set.Len() > 0 {
				for _, elem := range set.List() {
					result = append(result, elem.(string))
				}
			}
		}
	}
	return result
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
	return d.parent.GetOk(d.address + "." + key)
}

func (d *decoder) GetOkExists(key string) (interface{}, bool) {
	if d.address == "" {
		return d.parent.GetOkExists(key)
	}
	return d.parent.GetOkExists(d.address + "." + key)
}

func (d *decoder) GetChange(key string) (interface{}, interface{}) {
	if d.address == "" {
		return d.parent.GetChange(key)
	}
	return d.parent.GetChange(d.address + "." + key)
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
	return d.parent.Get(d.address + "." + key)
}
