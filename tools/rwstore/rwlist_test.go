package rwstore

import (
	"testing"
)

func TestNewRWList(t *testing.T) {
	list := NewRWList[int]()
	if list == nil {
		t.Error("NewRWList did not create a new list.")
	}
}

func TestGet(t *testing.T) {
	list := NewRWList[int]()
	list.Append(10)
	value, ok := list.Get(0)
	if !ok || value != 10 {
		t.Errorf("Get failed, expected 10 got %v", value)
	}
	_, ok = list.Get(-1)
	if ok {
		t.Error("Get should return false for negative index")
	}
	_, ok = list.Get(1)
	if ok {
		t.Error("Get should return false for out-of-range index")
	}
}

func TestReplace(t *testing.T) {
	list := NewRWList[int]()
	list.Append(10)
	list.Replace([]int{20, 30})
	if len(list.Copy()) != 2 {
		t.Error("Replace did not work as expected")
	}
}

func TestAppend(t *testing.T) {
	list := NewRWList[int]()
	list.Append(10)
	if len(list.Copy()) != 1 {
		t.Error("Append did not increase the length of the list")
	}
	if list.a[0] != 10 {
		t.Errorf("Append did not append the correct value, got %v", list.a[0])
	}
}

func TestCopy(t *testing.T) {
	list := NewRWList[int]()
	list.Append(10)
	copiedList := list.Copy()
	if len(copiedList) != 1 || copiedList[0] != 10 {
		t.Error("Copy did not copy the list correctly")
	}
	// Modify original list and ensure copy is unaffected
	list.Append(20)
	if len(copiedList) != 1 {
		t.Error("Copy was affected by changes to the original list")
	}
}
