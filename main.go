package main

import (
	"fmt"
	"sync"
)

func main() {
	pool := sync.Pool{
		New: func() any {
			fmt.Println("Allocating new byte slice")
			return make([]byte, 1024) // 1kB
		},
	}

	// Get a new obj from the pool
	// This call will allocate since the pool is empty
	obj := pool.Get().([]byte)
	fmt.Printf("Got object from pool, of length: %d\n", len(obj))

	// Put the object back in the pool
	pool.Put(obj)

	// Get the object again
	// This time it is reused from the pool
	reusedObj := pool.Get().([]byte)
	fmt.Printf("Got reused object from pool, of length: %d\n", len(reusedObj))

	// Put the object back in the pool
	pool.Put(obj)
}
