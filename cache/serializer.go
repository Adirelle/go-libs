package cache

// Serializer is used to (un)serialize keys and values.
type Serializer interface {
	Serialize(interface{}) ([]byte, error)
	Unserialize([]byte) (interface{}, error)
}

// serializingCache stores serialized keys and values in its backend.
type serializingCache struct {
	Cache
	KeySerializer   Serializer
	ValueSerializer Serializer
}

// Serialization creates an Option to (un)serialize keys and values.
func Serialization(key, value Serializer) Option {
	return func(c Cache) Cache {
		return &serializingCache{c, key, value}
	}
}

// Set serializes both key and value before storing it into its backend.
func (c *serializingCache) Set(key interface{}, value interface{}) (err error) {
	skey, err := c.KeySerializer.Serialize(key)
	if err != nil {
		return
	}
	svalue, err := c.ValueSerializer.Serialize(value)
	if err != nil {
		return
	}
	return c.Cache.Set(skey, svalue)
}

// Get serializes the key, tries and fetchs the value from its backend and deserializes it.
func (c *serializingCache) Get(key interface{}) (value interface{}, err error) {
	skey, err := c.KeySerializer.Serialize(key)
	if err != nil {
		return
	}
	svalue, err := c.Cache.Get(skey)
	if err != nil {
		return
	}
	return c.ValueSerializer.Unserialize(svalue.([]byte))
}

// GetIFPresent serializes the key, tries and fetchs the value from its backend and deserializes it.
func (c *serializingCache) GetIFPresent(key interface{}) (value interface{}, err error) {
	skey, err := c.KeySerializer.Serialize(key)
	if err != nil {
		return
	}
	svalue, err := c.Cache.GetIFPresent(skey)
	if err != nil {
		return
	}
	return c.ValueSerializer.Unserialize(svalue.([]byte))
}

// Remove serializes the key and removes the entry from its backend.
func (c *serializingCache) Remove(key interface{}) bool {
	skey, err := c.KeySerializer.Serialize(key)
	if err != nil {
		return false
	}
	return c.Cache.Remove(skey)
}
