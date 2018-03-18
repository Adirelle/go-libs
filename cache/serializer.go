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

// Serialization adds a layer that (un)serialize keys and values to []byte.
// It can be useful for storages that requires keys and values to be of type []byte, like BoltStorage.
// Please note that []byte is not hashable, thus cannot be used with MemoryStorage.
func Serialization(key, value Serializer) Option {
	return func(c Cache) Cache {
		return &serializingCache{c, key, value}
	}
}

func (c *serializingCache) Put(key interface{}, value interface{}) (err error) {
	skey, err := c.KeySerializer.Serialize(key)
	if err != nil {
		return
	}
	svalue, err := c.ValueSerializer.Serialize(value)
	if err != nil {
		return
	}
	return c.Cache.Put(skey, svalue)
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
