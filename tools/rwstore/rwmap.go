package rwstore

import (
	"sync"
)

type RWMap[K comparable, V any] struct {
	m map[K]V
	sync.RWMutex
}

func NewRWMap[K comparable, V any]() *RWMap[K, V] {
	return &RWMap[K, V]{
		m: make(map[K]V),
	}
}

func (rw *RWMap[K, V]) Get(key K) (V, bool) {
	rw.RLock()
	value, ok := rw.m[key]
	rw.RUnlock()
	return value, ok
}

func (rw *RWMap[K, V]) Set(key K, value V) {
	rw.Lock()
	rw.m[key] = value
	rw.Unlock()
}

func (rw *RWMap[K, V]) Delete(key K) {
	rw.Lock()
	delete(rw.m, key)
	rw.Unlock()
}

func (rw *RWMap[K, V]) Exists(key K) bool {
	rw.RLock()
	_, exists := rw.m[key]
	rw.RUnlock()
	return exists
}

func (rw *RWMap[K, V]) Len() int {
	rw.RLock()
	length := len(rw.m)
	rw.RUnlock()
	return length
}

func (rw *RWMap[K, V]) Iterate(iter func(K, V) bool) {
	rw.RLock()
	defer rw.RUnlock()
	for k, v := range rw.m {
		if !iter(k, v) {
			break
		}
	}
}
