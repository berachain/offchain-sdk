package types

// KVStore describes the basic interface for interacting with key-value stores.
type KVStore interface {
	// Get returns nil iff key doesn't exist. Errors on nil key.
	Get(key []byte) ([]byte, error)

	// Has checks if a key exists. Errors on nil key.
	Has(key []byte) (bool, error)

	// Set sets the key. Errors on nil key or value.
	Set(key, value []byte) error

	// Delete deletes the key. Errors on nil key.
	Delete(key []byte) error
}

// Store is a simple in-memory key-value store.
type store struct {
	kv map[string][]byte
}

// Get returns nil iff key doesn't exist. Errors on nil key.
func (s *store) Get(key []byte) ([]byte, error) {
	return s.kv[string(key)], nil
}

// Has checks if a key exists. Errors on nil key.
func (s *store) Has(key []byte) (bool, error) {
	_, ok := s.kv[string(key)]
	return ok, nil
}

// Set sets the key. Errors on nil key or value.
func (s *store) Set(key, value []byte) error {
	s.kv[string(key)] = value
	return nil
}

// Delete deletes the key. Errors on nil key.
func (s *store) Delete(key []byte) error {
	delete(s.kv, string(key))
	return nil
}

// NewKVStore creates a new KVstore.
func NewKVStore() KVStore {
	return &store{
		kv: make(map[string][]byte),
	}
}

// MultiStore is a simple in-memory multi key-value store.
type MultiStore interface {
	GetKVStore(key string) KVStore
}

// NewMultiStore creates a new MultiStore.
type multiStore struct {
	stores map[string]KVStore
}

// NewMultiStore creates a new MultiStore.
func NewMultiStore() *multiStore {
	return &multiStore{
		stores: make(map[string]KVStore),
	}
}

// GetKVStore gets a KVStore from the MultiStore.
func (ms *multiStore) GetKVStore(key string) KVStore {
	return ms.stores[key]
}
