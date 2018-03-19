package cache

import (
	"bytes"
	"encoding/binary"
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

func TestStringSerializer(t *testing.T) {

	s := StringSerializer{}

	if b, err := s.Serialize("foobar"); !bytes.Equal(b, []byte("foobar")) || err != nil {
		t.Errorf("Expected %v, got %v, %v", []byte("a"), b, err)
	}

	if b, err := s.Serialize(5); err == nil {
		t.Errorf("Expected an error, got %v, %v", b, err)
	}

	if s, err := s.Unserialize([]byte("foobar")); s != "foobar" || err != nil {
		t.Errorf("Expected %v, got %v, %v", "foobar", s, err)
	}

}

type serializableTest struct {
	X uint32
	Y string
}

func (s *serializableTest) MarshalBinary() ([]byte, error) {
	b := &bytes.Buffer{}
	binary.Write(b, binary.LittleEndian, s.X)
	b.WriteString(s.Y)
	return b.Bytes(), nil
}

func (s *serializableTest) UnmarshalBinary(data []byte) (err error) {
	b := bytes.NewBuffer(data)
	err = binary.Read(b, binary.LittleEndian, &s.X)
	if err != nil {
		return
	}
	s.Y, err = b.ReadString(0)
	return nil
}

func TestBinarySerializer(t *testing.T) {
	s := NewBinarySerializer(&serializableTest{})

	b, err := s.Serialize(&serializableTest{5, "foobar"})
	t.Logf("Serialize, got: %#v, %v", b, err)
	if err != nil {
		t.Fatal("Unexpected error")
	}

	y, err := s.Unserialize(b)
	t.Logf("Unserialize, got: %#v, %v", y, err)
	if err != nil {
		t.Fatal("Unexpected error")
	}
	if z, ok := y.(*serializableTest); !ok || z.X != 5 || z.Y != "foobar" {
		t.Fatal("Unexpected value")
	}
}
