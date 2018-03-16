package cache

// Serializer is used to (un)serialize keys and values.
type Serializer interface {
	Serialize(interface{}) ([]byte, error)
	Unserialize([]byte) (interface{}, error)
}

// SerializingCache stores serialized keys and values in its backend.
type SerializingCache struct {
	Backend         Cache
	KeySerializer   Serializer
	ValueSerializer Serializer
}

// Set serializes both key and value before storing it into its backend.
func (c *SerializingCache) Set(key interface{}, value interface{}) (err error) {
	skey, err := c.KeySerializer.Serialize(key)
	if err != nil {
		return
	}
	svalue, err := c.ValueSerializer.Serialize(value)
	if err != nil {
		return
	}
	return c.Backend.Set(skey, svalue)
}

// Get serializes the key, tries and fetchs the value from its backend and deserializes it.
func (c *SerializingCache) Get(key interface{}) (value interface{}, err error) {
	skey, err := c.KeySerializer.Serialize(key)
	if err != nil {
		return
	}
	svalue, err := c.Backend.Get(skey)
	if err != nil {
		return
	}
	return c.ValueSerializer.Unserialize(svalue.([]byte))
}

// GetIFPresent serializes the key, tries and fetchs the value from its backend and deserializes it.
func (c *SerializingCache) GetIFPresent(key interface{}) (value interface{}, err error) {
	skey, err := c.KeySerializer.Serialize(key)
	if err != nil {
		return
	}
	svalue, err := c.Backend.GetIFPresent(skey)
	if err != nil {
		return
	}
	return c.ValueSerializer.Unserialize(svalue.([]byte))
}

// Remove serializes the key and removes the entry from its backend.
func (c *SerializingCache) Remove(key interface{}) bool {
	skey, err := c.KeySerializer.Serialize(key)
	if err != nil {
		return false
	}
	return c.Backend.Remove(skey)
}
