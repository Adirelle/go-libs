package cache

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/boltdb/bolt"
)

func TestBolt(t *testing.T) {

	dbName := fmt.Sprintf("test%d.db", os.Getpid())
	db, err := bolt.Open(dbName, 0666, nil)
	if err != nil {
		t.Fatal("Unexpected error", err)
	}
	defer func() {
		db.Close()
		os.Remove(dbName)
	}()

	c := NewBoltStorage(db, []byte("MY"), Spy(t.Logf))

	if err := c.Put([]byte("foo"), []byte("bar")); err != nil {
		t.Fatal("Unexpected error", err)
	}

	c.Flush()
	db.Close()

	db, err = bolt.Open(dbName, 0666, nil)
	if err != nil {
		t.Fatal("Unexpected error", err)
	}

	c2 := NewBoltStorage(db, []byte("MY"), Spy(t.Logf))

	if value, err := c2.Get([]byte("foo")); !bytes.Equal(value.([]byte), []byte("bar")) || err != nil {
		t.Fatalf("Unexpected result: %v, %v", value, err)
	}

	if value, err := c2.Get([]byte("bar")); value != nil || err != ErrKeyNotFound {
		t.Fatalf("Unexpected result: %v, %v", value, err)
	}

	if len := c2.Len(); len != 1 {
		t.Fatalf("Unexpected result: %v", len)
	}

	if removed := c2.Remove([]byte("foo")); !removed {
		t.Fatalf("Unexpected result: %v", removed)
	}

	if removed := c2.Remove([]byte("bar")); removed {
		t.Fatalf("Unexpected result: %v", removed)
	}
}
