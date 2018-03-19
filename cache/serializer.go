package cache

import (
	"encoding"
	"fmt"
	"reflect"
)

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

// StringSerializer (un)serializes strings as-is.
type StringSerializer struct{}

// Serialize returns the string as a slice of bytes.
func (StringSerializer) Serialize(data interface{}) ([]byte, error) {
	if s, ok := data.(string); ok {
		return []byte(s), nil
	}
	return nil, fmt.Errorf("StringSerializer.Serialize, invalid type %T", data)
}

// Unserialize returns the slice of bytes as a string.
func (StringSerializer) Unserialize(data []byte) (interface{}, error) {
	return string(data), nil
}

// BinarySerializable combines encoding.BinaryMarshaler and encoding.BinaryUnmarshaler.
type BinarySerializable interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}

type binarySerializer struct {
	t reflect.Type
}

// NewBinarySerializer creates a Serializer for a specific type implementing BinarySerializable.
func NewBinarySerializer(sample BinarySerializable) Serializer {
	return &binarySerializer{reflect.TypeOf(sample).Elem()}
}

func (s *binarySerializer) Serialize(data interface{}) ([]byte, error) {
	bs, ok := data.(BinarySerializable)
	if !ok {
		return nil, fmt.Errorf("BinarySerializer.Serialize: unexpected value %v", data)
	}
	return bs.MarshalBinary()
}

func (s *binarySerializer) Unserialize(data []byte) (interface{}, error) {
	v := reflect.New(s.t).Interface()
	err := v.(BinarySerializable).UnmarshalBinary(data)
	return v, err
}
