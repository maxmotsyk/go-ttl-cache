package main

import (
	"fmt"
	"github.com/maxmotsyk/go-ttl-cache/internal/cache"
)

func main() {
	cache := cache.New()
	cache.Set("userId", 22)
	fmt.Println(cache.Get("userId"))
	cache.Deleat("userId")
	fmt.Println(cache.Get("userId"))

}
