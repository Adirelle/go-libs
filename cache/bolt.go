package cache

import "github.com/boltdb/bolt"

// BoltStorage stores the value in a bucket of a Bolt Database.
type BoltStorage struct {
	DB         *bolt.DB
	BucketName []byte
}

// Set writes the entry in the bucket.
func (s *BoltStorage) Set(key interface{}, value interface{}) error {
	return s.DB.Update(func(tx *bolt.Tx) (err error) {
		b, err := tx.CreateBucketIfNotExists(s.BucketName)
		if err != nil {
			return
		}
		return b.Put(key.([]byte), value.([]byte))
	})
}

// Get fetchs an entry from the bucket.
func (s *BoltStorage) Get(key interface{}) (value interface{}, err error) {
	err = s.DB.View(func(tx *bolt.Tx) (err error) {
		b := tx.Bucket(s.BucketName)
		if b != nil {
			value = b.Get(key.([]byte))
		}
		return
	})
	return
}

// GetIFPresent is a synonym to Get.
func (s *BoltStorage) GetIFPresent(key interface{}) (interface{}, error) {
	return s.Get(key)
}

// Remove removes an entry from the bucket.
func (s *BoltStorage) Remove(key interface{}) (found bool) {
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
