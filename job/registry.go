package job

// Registry is a registry of jobtypes to jobs.
type Registry struct {
	// mapRegistry is the underlying map registry of job types to
	mapRegistry *mapRegistry[string, HasProducer]
}

// RegisterJob registers a job type to the registry.
func (r *Registry) RegisterJob(id string, job HasProducer) {
	if err := r.mapRegistry.Register(id, job); err != nil {
		panic(err)
	}
}

func (r *Registry) Iterator() *RegistryIterator[string, HasProducer] {
	return NewRegistryIterator(r.mapRegistry)
}

// Iterator for the map.
type RegistryIterator[K comparable, T any] struct {
	m    *mapRegistry[K, T]
	keys []K
	idx  int
}

// NewMapIterator creates a new iterator for the given map.
func NewRegistryIterator[K comparable,
	T any](m *mapRegistry[K, T]) *RegistryIterator[K, T] {
	keys := make([]K, 0, len(m.items))
	for key := range m.items {
		keys = append(keys, key)
	}
	return &RegistryIterator[K, T]{
		m:    m,
		keys: keys,
		idx:  -1,
	}
}

// Next moves the iterator to the next element and returns whether it's valid.
func (it *RegistryIterator[K, T]) Next() bool {
	it.idx++
	return it.idx < len(it.keys)
}

// Key returns the key of the current element.
func (it *RegistryIterator[K, T]) Key() K {
	if it.idx < 0 || it.idx >= len(it.keys) {
		var zeroK K
		return zeroK
	}
	return it.keys[it.idx]
}

// Value returns the value of the current element.
func (it *RegistryIterator[K, T]) Value() T { // Change 'interface{}' to your desired value type
	var zeroK K
	var zeroT T
	key := it.Key()
	if key == zeroK {
		return zeroT
	}
	return it.m.Get(key)
}

/// TODO: put the mapregistry in a common lib folder in Berachain, extract from Polaris.

// mapRegistry is a simple implementation of `Registry` that uses a
// map as the underlying data structure.
type mapRegistry[K comparable, T any] struct {
	// items is the map of items in the registry.
	items map[K]T
}

// NewMap creates and returns a new `mapRegistry`.
//
//nolint:revive // only used as Registry interface.
func NewMap[K comparable, T any]() *mapRegistry[K, T] {
	return &mapRegistry[K, T]{
		items: make(map[K]T),
	}
}

// Get returns an item using its ID.
func (mr *mapRegistry[K, T]) Get(id K) T {
	return mr.items[id]
}

// Register adds an item to the registry.
func (mr *mapRegistry[K, T]) Register(id K, item T) error {
	mr.items[id] = item
	return nil
}

// Remove removes an item from the registry.
func (mr *mapRegistry[K, T]) Remove(id K) {
	delete(mr.items, id)
}

// Has returns true if the item exists in the registry.
func (mr *mapRegistry[K, T]) Has(id K) bool {
	_, ok := mr.items[id]
	return ok
}

// Iterate returns the underlying map.
func (mr *mapRegistry[K, T]) Iterate() map[K]T {
	return mr.items
}
