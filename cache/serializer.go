package cache

import "fmt"

// Serializer is used to (un)serialize keys and values.
type Serializer interface {
	Serialize(interface{}) ([]byte, error)
	Unserialize([]byte) (interface{}, error)
}

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

func (c *serializingCache) Remove(key interface{}) bool {
	skey, err := c.KeySerializer.Serialize(key)
	if err != nil {
		return false
	}
	return c.Cache.Remove(skey)
}

func (c *serializingCache) String() string {
	return fmt.Sprintf("Serialized(%s,%v,%v)", c.Cache, c.KeySerializer, c.ValueSerializer)
}
