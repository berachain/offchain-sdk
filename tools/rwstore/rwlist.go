package rwstore

import "sync"

type RWList[V any] struct {
	a []V
	sync.RWMutex
}

func NewRWList[V any]() *RWList[V] {
	return &RWList[V]{
		a: make([]V, 0),
	}
}

func (rw *RWList[V]) Get(index int) (V, bool) {
	rw.RLock()
	defer rw.RUnlock()
	if index < 0 || index >= len(rw.a) {
		var zero V
		return zero, false
	}
	return rw.a[index], true
}

func (rw *RWList[V]) Replace(newList []V) {
	rw.Lock()
	rw.a = newList
	rw.Unlock()
}

func (rw *RWList[V]) Append(value V) {
	rw.Lock()
	rw.a = append(rw.a, value)
	rw.Unlock()
}

func (rw *RWList[V]) Copy() []V {
	rw.RLock()
	defer rw.RUnlock()
	return append([]V(nil), rw.a...)
}
