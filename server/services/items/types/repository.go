package types

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"

	"time"

	"github.com/boltdb/bolt"
	"github.com/js13kgames/glitchd/server"
	"github.com/js13kgames/glitchd/server/services/items/metrics"
)

const TOKEN_BYTES = 8
const TOKEN_LENGTH = 2 * TOKEN_BYTES

// StoreRepository is *not* thread-safe.
// In our use case mutations only ever happen via administrative actions, at most a few times *daily*
// and realistically never concurrently. In the edge case a store gets deleted during a read,
// even if the read call came in before the delete call, the deletion *may* happen before, causing
// the read request to fail to find the store and return with an error. If the read happens before
// the deletion, the read call stack will continue to hold a pointer to the Store even if it meanwhile
// gets removed from the Repository (which is fine).
// The maps are allocated on the heap and will be touched by the GC. Realistically we expect at most
// a few dozen up to a max of a few hundred stores to be present, so as long as this assumption holds
// true GC pauses should not be impacted in a meaningful manner.
type StoreRepository struct {
	// Token -> Store (nearly all access is going to be reads identified by an access token, not an ID
	// even though we primarily use the ID internally instead, to avoid coupling persisted data to a token
	// which may change (as opposed to an ID which will not).
	Items        map[string]*Store `json:"items"`
	db           *bolt.DB          `json:"-"`
	sequentialId uint16            `json:"-"`
	bucketKey    []byte            `json:"-"`
	ids          map[uint16]string `json:"-"` // ID -> Access token
}

// LoadStoreRepository creates a StoreRepository and populates it from the given bucket identified
// by its key in the backing storage.
func LoadStoreRepository(db *bolt.DB, bucketKey []byte) (*StoreRepository, error) {
	repository := &StoreRepository{
		Items:     make(map[string]*Store),
		db:        db,
		ids:       make(map[uint16]string),
		bucketKey: bucketKey,
	}

	if err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(repository.bucketKey)
		if err != nil {
			return err
		}

		// Recreate all persisted Stores on the current database handle.
		cur := bucket.Cursor()
		for _, v := cur.First(); v != nil; _, v = cur.Next() {
			var persisted *Store

			// The structs are small and we really only expect a few hundred at most (see note on struct)
			// so with this being performed only on startup, there is no reason to introduce a dependency
			// on Msgp or Colfer. Not even a faster JSON serializer than std.
			if err := json.Unmarshal(v, &persisted); err != nil {
				return err
			}

			// Map out and allocate the store.
			// Note: We are constructing a new one even though the stack already contains the unserialized
			// store, because we need to pass in the db and have the store construct its bucket key
			// (which does not get marshalled and persisted).
			store := NewStore(persisted.Id, persisted.Token, db)
			storeItemsBucket, err := tx.CreateBucketIfNotExists(store.bucketKey)
			if err != nil {
				return err
			}

			repository.putMemMap(store)

			var length, size uint64

			storeItemsCur := storeItemsBucket.Cursor()
			for _, v := storeItemsCur.First(); v != nil; _, v = storeItemsCur.Next() {
				length++
				size += uint64(len(v))
			}

			store.metrics = metrics.NewStoreAggregator(length, size)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return repository, nil
}

func (repository *StoreRepository) RegisterMetricsTicks(tickManager server.TickManager) {
	tickManager.OnTickSecond(repository.onTickSecond)
	tickManager.OnTickMinute(repository.onTickMinute)
	tickManager.OnTickHour(repository.onTickHour)
}

//
//
//
func (repository *StoreRepository) GetById(id uint16) *Store {
	return repository.Items[repository.ids[id]]
}

//
//
//
func (repository *StoreRepository) Create(id uint16, token string) (*Store, error) {
	store := NewStore(id, token, repository.db)

	repository.assignKeysTo(store)

	if err := repository.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(store.bucketKey)
		if err != nil {
			return err
		}

		store.metrics = metrics.NewStoreAggregator(0, 0)

		return repository.write(store, tx)
	}); err != nil {
		return nil, err
	}

	repository.putMemMap(store)

	return store, nil
}

//
//
//
func (repository *StoreRepository) Save(store *Store) error {
	repository.assignKeysTo(store)

	if err := repository.db.Update(func(tx *bolt.Tx) error {
		return repository.write(store, tx)
	}); err != nil {
		return err
	}

	repository.putMemMap(store)

	return nil
}

//
//
//
func (repository *StoreRepository) Delete(store *Store) error {
	// Naive memory lookup to avoid hitting the backing store if we don't have the Store
	// in memory to begin with. No-op if the Store does not exist.
	if _, exists := repository.ids[store.Id]; !exists {
		return nil
	}

	if err := repository.db.Update(func(tx *bolt.Tx) error {
		tx.DeleteBucket(store.bucketKey)
		return tx.Bucket(repository.bucketKey).Delete(storeIdToKey(store.Id))
	}); err != nil {
		return err
	}

	repository.delMemMap(store)

	return nil
}

//
//
//
func (repository *StoreRepository) RotateToken(store *Store) error {
	var currentToken string = store.Token

	store.Token = repository.genToken()

	if err := repository.Save(store); err != nil {
		// Roll back the change.
		store.Token = currentToken
		return err
	}

	// Unmap the old named reference. The new token has been mapped in Save().
	delete(repository.Items, store.Token)

	return nil
}

//
//
//
func (repository *StoreRepository) write(store *Store, tx *bolt.Tx) error {
	bucket := tx.Bucket(repository.bucketKey)

	data, err := json.Marshal(store)
	if err != nil {
		return err
	}

	return bucket.Put(storeIdToKey(store.Id), data)
}

//
//
//
func (repository *StoreRepository) putMemMap(store *Store) {
	// Auto-incrementing and without re-use, so let's account for potential unordered imports
	// from external sources in the future.
	if store.Id > repository.sequentialId {
		repository.sequentialId = store.Id
	}

	repository.Items[store.Token] = store
	repository.ids[store.Id] = store.Token
}

//
//
//
func (repository *StoreRepository) delMemMap(store *Store) {
	delete(repository.ids, store.Id)
	delete(repository.Items, store.Token)
}

//
//
//
func (repository *StoreRepository) assignKeysTo(store *Store) {
	if store.Token == "" {
		store.Token = repository.genToken()
	}

	if store.Id == 0 {
		store.Id = repository.sequentialId + 1
	}
}

// genToken generates a random, base16-encoded string that is unique
// amongst the current tokens in this StoreRepository.
func (repository *StoreRepository) genToken() string {
	src := make([]byte, TOKEN_BYTES)
	rand.Read(src)
	dst := make([]byte, hex.EncodedLen(len(src)))

	hex.Encode(dst, src)

	token := string(dst)

	// On the off chance we get a collision, keep re-running until we get a unique.
	if _, exists := repository.Items[token]; exists {
		return repository.genToken()
	}

	return token
}

// @todo Obviously we don't want to iterate over a map each second. This will need restructuring
// in how we reference the Stores internally. The seemingly more obvious solution of registering
// the aggregators with a TickManager falters due to functions in Go not being comparable and thus
// not having a reliable way of removing them when a Store gets deleted at runtime. Not without
// using named references for all tickers anyways.
func (repository *StoreRepository) onTickSecond(tick time.Time) {
	for _, store := range repository.Items {
		store.metrics.OnTickSecond(tick)
	}
}

func (repository *StoreRepository) onTickMinute(tick time.Time) {
	for _, store := range repository.Items {
		store.metrics.OnTickMinute(tick)
	}
}

func (repository *StoreRepository) onTickHour(tick time.Time) {
	for _, store := range repository.Items {
		store.metrics.OnTickHour(tick)
	}
}

// storeIdToKey returns a BigEndian representation of the given storeId (uint16).
func storeIdToKey(v uint16) []byte {
	b := make([]byte, 2)
	b[0] = byte(v >> 8)
	b[1] = byte(v)
	return b
}
