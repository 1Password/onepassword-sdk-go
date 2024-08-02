package onepassword

import (
	"errors"
	"sync"
)

var (
	ErrorIteratorDone = errors.New("end of iterator")
)

// Iterator defines a generic iterator
type Iterator[T any] struct {
	items        []T
	currentIndex int
	mutex        *sync.Mutex
}

// NewIterator creates a new iterator for the given slice
func NewIterator[T any](items []T) *Iterator[T] {
	return &Iterator[T]{
		items:        items,
		currentIndex: 0,
		mutex:        &sync.Mutex{},
	}
}

// Next returns the next item from the iterator
func (it *Iterator[T]) Next() (*T, error) {
	it.mutex.Lock()
	defer it.mutex.Unlock()

	if it.currentIndex >= len(it.items) {
		return nil, ErrorIteratorDone
	}

	item := it.items[it.currentIndex]
	it.currentIndex++
	return &item, nil
}
