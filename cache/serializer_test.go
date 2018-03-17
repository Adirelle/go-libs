package cache

import (
	"fmt"
	"strconv"
)

type testSerializer struct{}

func (testSerializer) Serialize(v interface{}) ([]byte, error) {
	return []byte(strconv.Itoa(v.(int))), nil
}

func (testSerializer) Unserialize(d []byte) (interface{}, error) {
	return strconv.Atoi(string(d))
}

func ExampleSerializingCache() {

	ser := testSerializer{}
	c := NewVoidStorage(Serialization(ser, ser), Spy(fmt.Printf))

	c.Set(50, 65)
	c.Get(50)
	c.GetIFPresent(50)
	c.Remove(50)

	// Output:
	// Set([53 48], [54 53]) -> <nil>
	// Get([53 48]) -> <nil>, Key not found
	// GetIFPresent([53 48]) -> <nil>, Key not found
	// Remove([53 48]) -> false
}
