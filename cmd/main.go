package main

import (
	"fmt"
	"github.com/maxmotsyk/go-ttl-cache/internal/cache"
	"sync"
	"time"
)

func main() {
	c := cache.New()
	var wg sync.WaitGroup

	// паралельно пишемо
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", n)
			for j := 0; j < 1000; j++ {
				_ = c.Set(key, j)
			}
		}(i)
	}

	// паралельно читаємо
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", n)
			for j := 0; j < 1000; j++ {
				_, _ = c.Get(key)
			}
		}(i)
	}

	// паралельно видаляємо
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", n)
			for j := 0; j < 500; j++ {
				_ = c.Deleat(key)
			}
		}(i)
	}

	wg.Wait()
	fmt.Println("Test finished")

	// залишаємо час, щоб подивитися в race detector
	time.Sleep(1 * time.Second)
}
