package bpool

import (
	"sync"
	"time"
)

/*
	Buffer struct
*/
type Buffer struct {
	when  time.Time // When the buffer was added in the pool
	slice []byte    // The slide used
}

/*
	Pool struct
*/
type Pool struct {
	buffers    chan Buffer   // The list of Buffer
	bufferSize int           // The fixed size of buffers
	length     int           // The length of the list
	expiration time.Duration // The expiration time (in second) of a buffer in the pool
	sync.Mutex               // A mutex to make the Pool threadsafe
}

/*
	Get a buffer in the pool

	If the buffer is empty, it create a new buffer
*/
func (p *Pool) Get() (b []byte) {
	p.Lock()         // Lock pool
	defer p.Unlock() // Unlock pool
	select {
	case buffer := <-p.buffers: // If a buffer is in the pool, return it
		b = buffer.slice // Attach buffer of the Buffer struct
		p.length--       // Decrement pool length
		return
	default: // If the pool is empty, create a new buffer and return it
		b = make([]byte, p.bufferSize) // Create buffer
		return
	}
	return
}

/*
	Put a buffer in the pool
*/
func (p *Pool) Put(b []byte) {
	var buffer Buffer

	// If buffer size is different of pool buffer size => dont keep it
	if cap(b) != p.bufferSize {
		return
	}

	// If buffer cap is different of buffer len => resize
	if cap(b) != len(b) {
		b = b[:cap(b)]
	}

	p.Lock()                 // Lock pool
	defer p.Unlock()         // Unlock pool
	buffer.slice = b         // Attach buffer in the Buffer struct
	buffer.when = time.Now() // Reset the expiration date
	p.buffers <- buffer      // Add the buffer in the pool
	p.length++               // Increment pool length
}

/*
	Create a new Pool
	It launch a garbage collector goroutine

	bufferSize: fixed size of buffers
	expiration: expiration time (in second) of a buffer in the pool
*/
func GetPool(bufferSize, expiration int) *Pool {
	var pool Pool

	pool.buffers = make(chan Buffer, 10000)                   // Make a huge channel to store buffers
	pool.bufferSize = bufferSize                              // Set the fixed size of buffers
	pool.expiration = time.Second * time.Duration(expiration) // Set the expiration time of buffers

	// Garbage collector goroutine
	go func() {
		for {
			tmpChan := make(chan Buffer, 10000) // Make a huge temporary channel to store buffers
			time.Sleep(time.Second)             // Sleep to avoid 100% CPU consumption
			pool.Lock()                         // Lock pool

			if pool.length == 0 {
				pool.Unlock() // Unlock pool
				continue
			}

			// GC loop
		GC:
			for {
				select {
				case buffer := <-pool.buffers: // If a buffer exist in the pool, check expiration date
					if buffer.when.Add(pool.expiration).Before(time.Now()) {
						// Expirate => Remove it
						pool.length--
					} else {
						// Not expirate => Keep it
						tmpChan <- buffer
					}
				default: // End of the pool
					pool.buffers = tmpChan // Keep temporary chan
					break GC
				}
			}

			pool.Unlock() // Unlock pool
		}
	}()
	return &pool
}
