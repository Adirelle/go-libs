package cache

import "github.com/boltdb/bolt"

// boltStorage stores the value in a bucket of a Bolt Database.
type boltStorage struct {
	DB         *bolt.DB
	BucketName []byte
}

// NewBoltStorage creates a boltStorage from an open bolt.DB.
func NewBoltStorage(db *bolt.DB, bucketName []byte) Cache {
	return &boltStorage{db, bucketName}
}

func (s *boltStorage) Set(key interface{}, value interface{}) error {
	return s.DB.Update(func(tx *bolt.Tx) (err error) {
		b, err := tx.CreateBucketIfNotExists(s.BucketName)
		if err != nil {
			return
		}
		return b.Put(key.([]byte), value.([]byte))
	})
}

func (s *boltStorage) Get(key interface{}) (value interface{}, err error) {
	err = s.DB.View(func(tx *bolt.Tx) (err error) {
		b := tx.Bucket(s.BucketName)
		if b != nil {
			value = b.Get(key.([]byte))
		}
		return
	})
	return
}

func (s *boltStorage) GetIFPresent(key interface{}) (interface{}, error) {
	return s.Get(key)
}

func (s *boltStorage) Flush() error {
	return s.DB.Sync()
}

func (s *boltStorage) Remove(key interface{}) (found bool) {
	s.DB.Update(func(tx *bolt.Tx) (err error) {
		b := tx.Bucket(s.BucketName)
		if b == nil {
			return
		}
		found = b.Get(key.([]byte)) == nil
		if !found {
			return
		}
		return b.Delete(key.([]byte))
	})
	return
}
