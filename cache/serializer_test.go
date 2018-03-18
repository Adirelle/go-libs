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

	t.SkipNow()

	ser := testSerializer{}
	c := NewVoidStorage(Spy(t.Logf), Serialization(ser, ser), Spy(t.Logf))

	c.Set(50, 65)
	c.Get(50)
	c.GetIFPresent(50)
	c.Remove(50)

}
