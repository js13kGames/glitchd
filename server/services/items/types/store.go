package types

import (
	"strconv"

	"github.com/boltdb/bolt"
	"github.com/js13kgames/glitchd/server/services/items/metrics"
)

type Store struct {
	Id           uint16 `json:"id"`
	Token        string `json:"token"`
	OwnerId      uint64 `json:"ownerId" binding:"required"`
	SubmissionId uint64 `json:"submissionId"`

	db        *bolt.DB                 `json:"-"`
	bucketKey []byte                   `json:"-"`
	metrics   *metrics.StoreAggregator `json:"-"`
}

func NewStore(id uint16, token string, db *bolt.DB) *Store {
	return &Store{
		Id:        id,
		Token:     token,
		db:        db,
		bucketKey: []byte("stores." + strconv.FormatUint(uint64(id), 10) + ".items"),
	}
}

//
//
//
func (store *Store) Get(key string) ([]byte, error) {
	// @todo Pre-allocate. We'll need item keys mapped in-memory to sizes first though.
	var value []byte

	// Increment metrics even if the given key ends up not being found.
	if store.metrics != nil {
		store.metrics.IncReads()
	}

	err := store.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(store.bucketKey)

		if v := bucket.Get([]byte(key)); v != nil {
			value = make([]byte, len(v))
			copy(value, v)
		}

		return nil
	})

	return value, err
}

// @todo Keep monitoring metrics and batch writes via a WAL if need be.
// BoltDB is slow on random writes which "should" not matter in our case, but "should"
// is not a confident assumption.
func (store *Store) Put(key string, value []byte) error {
	keyBytes := []byte(key)
	return store.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(store.bucketKey)

		// @todo With metrics on, this is an additional read per write hitting the backing store.
		// Benchmark doing the size and len counting on-demand for a relatively large dataset (100k keys at 32KiB)
		if store.metrics != nil {
			var (
				lengthDelta uint64
				sizeDelta   uint64
				sizeItemNew int = len(value)
			)

			if v := bucket.Get(keyBytes); v != nil {
				sizeItemOld := len(v)
				sizeDiff := sizeItemNew - sizeItemOld

				if sizeDiff >= 0 {
					sizeDelta = uint64(sizeDiff)
				} else {
					sizeDelta = ^uint64(sizeDiff - 1)
				}
			} else {
				lengthDelta = 1
				sizeDelta = uint64(sizeItemNew)
			}

			store.metrics.IncWrites(lengthDelta, sizeDelta)
		}

		return bucket.Put(keyBytes, value)
	})
}

//
//
//
func (store *Store) Delete(key string) error {
	keyBytes := []byte(key)

	return store.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(store.bucketKey)

		if store.metrics != nil {
			if v := bucket.Get(keyBytes); v != nil {
				// Decrement length by 1 and size by len(v).
				store.metrics.IncWrites(^uint64(0), ^uint64(len(v)-1))
			}
		}

		return bucket.Delete(keyBytes)
	})
}

//
//
//
func (store *Store) Metrics() *metrics.StoreSnapshot {
	if store.metrics != nil {
		return store.metrics.Collect()
	}

	return nil
}
