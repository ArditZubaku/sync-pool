package main

import "sync"

// TypedPool wraps sync.Pool with a generic type
type TypedPool[T any] struct {
	pool sync.Pool
}

// NewTypedPool creates a new TypedPool using the provided constructor.
func NewTypedPool[T any](newFn func() T) *TypedPool[T] {
	return &TypedPool[T]{
		pool: sync.Pool{
			New: func() any {
				return newFn()
			},
		},
	}
}

// Get retrieves an item from the pool (properly typed).
func (tp *TypedPool[T]) Get() T {
	return tp.pool.Get().(T)
}

// Put returns an item back to the pool.
func (tp *TypedPool[T]) Put(v T) {
	tp.pool.Put(v)
}
