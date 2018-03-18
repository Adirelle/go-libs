package cache

import (
	"bytes"
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
	ch := make(chan Event, 1)
	c := NewVoidStorage(Serialization(ser, ser), Emitter(ch), Spy(t.Logf))

	c.Put(50, 60)
	if e := <-ch; !bytes.Equal(e.Key.([]byte), []byte("50")) || !bytes.Equal(e.Value.([]byte), []byte("60")) {
		t.Error(`Expected [53 48] and [54 48]`)
	}

	c.Get(50)
	if e := <-ch; !bytes.Equal(e.Key.([]byte), []byte("50")) {
		t.Error(`Expected [53 48]`)
	}

	c.Remove(50)
	if e := <-ch; !bytes.Equal(e.Key.([]byte), []byte("50")) {
		t.Error(`Expected [53 48]`)
	}
}
