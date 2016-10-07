package main

import (
	"fmt"
	"github.com/90TechSAS/go-buffer-pool-recycler"
	"time"
)

var (
	tmpBuffer []byte // Temporary buffer used to fill the buffers
)

func main() {
	const NbIter = 5000                 // Test with # iterations
	const BufferSize = 1024 * 1024 * 10 // Size of buffers
	var timer time.Time                 // Timer

	// Fill the temporary buffer
	tmpBuffer = make([]byte, BufferSize)
	for i, _ := range tmpBuffer {
		tmpBuffer[i] = 42
	}

	pool := bpool.GetPool(BufferSize, 10) // Get a pool

	// Test without bpool
	fmt.Print("Running test without bpool: ")
	timer = time.Now()
	for i := 0; i < NbIter; i++ {
		var buffer []byte = make([]byte, BufferSize)
		FullBuffer(buffer)
	}
	fmt.Println(time.Since(timer))

	// Test with bpool
	fmt.Print("Running test with bpool: ")
	timer = time.Now()
	for i := 0; i < NbIter; i++ {
		var buffer []byte = pool.Get() // Get a buffer
		FullBuffer(buffer)
		pool.Put(buffer) // Put the buffer
	}
	fmt.Println(time.Since(timer))
}

/*
	This function fills the buffer
*/
func FullBuffer(b []byte) {
	copy(b, tmpBuffer)
}
