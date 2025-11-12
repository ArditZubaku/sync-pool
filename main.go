package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	allocCount := 0
	pool := NewTypedPool(
		func() []byte {
			allocCount++
			fmt.Print(".")
			return make([]byte, 1024) // 1kB
		},
	)

	// simpleObjectReUse(&pool)

	var wg sync.WaitGroup

	for range 1000 {
		wg.Add(1)
		go func() {
			obj := pool.Get()
			fmt.Print("-")
			time.Sleep(100 * time.Millisecond)
			pool.Put(obj)
			wg.Done()
		}()
		time.Sleep(10 * time.Millisecond)
	}

	wg.Wait()

	fmt.Printf("\n Number of allocations: %d\n", allocCount)
}

func simpleObjectReUse[T ~[]E, E any](pool *TypedPool[T]) {
	// Get a new obj from the pool
	// This call will allocate since the pool is initially empty
	obj := pool.Get()
	fmt.Printf("Got object from pool, of length: %d\n", len(obj))

	// Put the object back in the pool
	pool.Put(obj)

	// Get the object again
	// This time it is reused from the pool
	reusedObj := pool.Get()
	fmt.Printf("Got reused object from pool, of length: %d\n", len(reusedObj))

	// Put the object back in the pool
	pool.Put(obj)
}
