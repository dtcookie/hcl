package hcl

// Marshal has no documentation
func Marshal(v Marshaler, res ResourceAccessor) error {
	enc := &encoder{
		properties: map[string]interface{}{},
	}
	if err := v.MarshalHCL(enc); err != nil {
		return err
	}
	for key, value := range enc.properties {
		res.Set(key, value)
	}

	return nil
}
