package cache

import (
	"fmt"

	"github.com/boltdb/bolt"
)

// boltStorage stores the value in a bucket of a Bolt Database.
type boltStorage struct {
	db         *bolt.DB
	bucketName []byte
}

// NewBoltStorage creates a boltStorage from an open bolt.DB.
func NewBoltStorage(db *bolt.DB, bucketName []byte, opts ...Option) Cache {
	return options(opts).applyTo(&boltStorage{db, bucketName})
}

func (s *boltStorage) Put(key interface{}, value interface{}) error {
	return s.db.Update(func(tx *bolt.Tx) (err error) {
		b, err := tx.CreateBucketIfNotExists(s.bucketName)
		if err != nil {
			return
		}
		return b.Put(key.([]byte), value.([]byte))
	})
}

func (s *boltStorage) Get(key interface{}) (value interface{}, err error) {
	err = s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.bucketName)
		if b == nil {
			return ErrKeyNotFound
		}
		result := b.Get(key.([]byte))
		if result == nil {
			return ErrKeyNotFound
		}
		value = result[:]
		return nil
	})
	return
}

func (s *boltStorage) Remove(key interface{}) bool {
	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.bucketName)
		if b == nil || b.Get(key.([]byte)) == nil {
			return ErrKeyNotFound
		}
		return b.Delete(key.([]byte))
	})
	return (err == nil)
}

func (s *boltStorage) Flush() error {
	return s.db.Sync()
}

func (s *boltStorage) Len() (len int) {
	s.db.View(func(tx *bolt.Tx) (err error) {
		b := tx.Bucket(s.bucketName)
		if b == nil {
			return
		}
		len = b.Stats().KeyN
		return
	})
	return
}

func (s *boltStorage) String() string {
	return fmt.Sprintf("Bolt(%q,%q)", s.db.Path(), s.bucketName)
}
