package rwstore

import (
	"sync"
	"testing"
)

func TestRWMap(t *testing.T) {
	rwMap := NewRWMap[int, string]()

	// Test Set and Get
	rwMap.Set(1, "one")
	if v, ok := rwMap.Get(1); !ok || v != "one" {
		t.Errorf("Set or Get is not working, expected 'one', got %v", v)
	}

	// Test Exists
	if !rwMap.Exists(1) {
		t.Errorf("Exists is not working, expected true, got false")
	}

	// Test Len
	if rwMap.Len() != 1 {
		t.Errorf("Len is not working, expected 1, got %d", rwMap.Len())
	}

	// Test Delete
	rwMap.Delete(1)
	if rwMap.Exists(1) {
		t.Errorf("Delete is not working, key 1 still exists")
	}

	// Test Iterate
	rwMap.Set(2, "two")
	rwMap.Set(3, "three")
	visited := make(map[int]bool)
	rwMap.Iterate(func(k int, _ string) bool {
		visited[k] = true
		return true
	})
	if len(visited) != 2 || !visited[2] || !visited[3] {
		t.Errorf("Iterate is not working correctly, visited map: %+v", visited)
	}
}

func TestRWMapConcurrentAccess(t *testing.T) {
	rwMap := NewRWMap[int, int]()
	var wg sync.WaitGroup

	// Perform concurrent writes
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			rwMap.Set(val, val)
		}(i)
	}

	// Perform concurrent reads
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			if v, ok := rwMap.Get(val); !ok || v != val {
				t.Errorf("Concurrent access failed, expected %d, got %d", val, v)
			}
		}(i)
	}

	wg.Wait()

	// Check final map length
	if l := rwMap.Len(); l != 100 {
		t.Errorf("Concurrent writes failed, expected map length 100, got %d", l)
	}
}
