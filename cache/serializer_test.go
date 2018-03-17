package cache

import (
	"strconv"
	"testing"
)

type testSerializer struct{}

func (testSerializer) Serialize(v interface{}) ([]byte, error) {
	return []byte(strconv.Itoa(v.(int))), nil
}

func (testSerializer) Unserialize(d []byte) (interface{}, error) {
	return strconv.Atoi(string(d))
}

func TestSerializingCache(t *testing.T) {

	ser := testSerializer{}
	c := NewVoidStorage(Spy(t.Logf), Serialization(ser, ser), Spy(t.Logf))

	c.Set(50, 65)
	c.Get(50)
	c.GetIFPresent(50)
	c.Remove(50)

	t.Fail()

	// Output:
	// Set([53 48], [54 53]) -> <nil>
	// Get([53 48]) -> <nil>, Key not found
	// GetIFPresent([53 48]) -> <nil>, Key not found
	// Remove([53 48]) -> false
}
