package cache

import (
	"fmt"
	"strconv"
)

type printCache struct{}

func (printCache) Set(k, v interface{}) error {
	fmt.Println("Set", k, v)
	return nil
}

func (printCache) Get(k interface{}) (interface{}, error) {
	fmt.Println("Get", k)
	return k, nil
}

func (printCache) GetIFPresent(k interface{}) (interface{}, error) {
	fmt.Println("GetIFPresent", k)
	return k, nil
}

func (printCache) Remove(k interface{}) bool {
	fmt.Println("Remove", k)
	return false
}

type testSerializer struct{}

func (testSerializer) Serialize(v interface{}) ([]byte, error) {
	return []byte(strconv.Itoa(v.(int))), nil
}

func (testSerializer) Unserialize(d []byte) (interface{}, error) {
	return strconv.Atoi(string(d))
}

func ExampleSerializingCache() {

	ser := testSerializer{}
	b := printCache{}
	c := SerializingCache{b, ser, ser}

	fmt.Println(c.Set(50, 65))
	fmt.Println(c.Get(50))
	fmt.Println(c.GetIFPresent(50))
	fmt.Println(c.Remove(50))

	// Output:
	// Set [53 48] [54 53]
	// <nil>
	// Get [53 48]
	// 50 <nil>
	// GetIFPresent [53 48]
	// 50 <nil>
	// Remove [53 48]
	// false
}
