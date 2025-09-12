package main

import (
	"context"
	"fmt"
	"github.com/maxmotsyk/go-ttl-cache/internal/cache"
	"math/rand/v2"
	"sync"
	"sync/atomic"
	"time"
)

func main() {

	const (
		ttl         = 6 * time.Second
		nKeys       = 100
		nWorkers    = 32
		phaseWarmup = 2 * time.Second
		phaseLoad   = 3 * time.Second
		phaseDrain  = ttl + 3*time.Second
	)

	s := cache.New()

	keys := make([]string, nKeys)
	for i := range keys {
		keys[i] = fmt.Sprintf("k%03d", i)
	}

	var setsOK, setsErr, getsHit, getsMiss atomic.Int64

	// ---- PHASE 1: прогріваємо (тільки Set)
	fmt.Println("Warmup…")
	doTimed(phaseWarmup, func(ctx context.Context, wg *sync.WaitGroup) {
		for i := 0; i < nWorkers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for ctx.Err() == nil {
					k := keys[rand.IntN(nKeys)]
					v := rand.IntN(1000)
					if err := s.Set(k, v); err != nil {
						setsErr.Add(1)
					} else {
						setsOK.Add(1)
					}
					time.Sleep(time.Duration(10+rand.IntN(20)) * time.Millisecond)
				}
			}()
		}
	})

	// ---- PHASE 2: мікс Set/Get ----
	fmt.Println("Load…")
	doTimed(phaseLoad, func(ctx context.Context, wg *sync.WaitGroup) {
		for i := 0; i < nWorkers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for ctx.Err() == nil {
					k := keys[rand.IntN(nKeys)]
					switch rand.IntN(3) { // ~33% get, ~66% set
					case 0:
						if _, err := s.Get(k); err != nil {
							getsMiss.Add(1)
						} else {
							getsHit.Add(1)
						}
					default:
						v := rand.IntN(1000)
						if err := s.Set(k, v); err != nil {
							setsErr.Add(1)
						} else {
							setsOK.Add(1)
						}
					}
					time.Sleep(time.Duration(5+rand.IntN(10)) * time.Millisecond)
				}
			}()
		}
	})

	// ---- PHASE 3: чекаємо, поки все протухне ----
	fmt.Println("Drain… (no writes, waiting for TTL to expire)")
	time.Sleep(phaseDrain)

	var finalHit, finalMiss int
	for _, k := range keys {
		if _, err := s.Get(k); err != nil {
			finalMiss++
		} else {
			finalHit++
		}
	}

	fmt.Println("---- STATS ----")
	fmt.Printf("setsOK=%d  setsErr=%d  getsHit=%d  getsMiss=%d\n",
		setsOK.Load(), setsErr.Load(), getsHit.Load(), getsMiss.Load())
	fmt.Printf("After drain: alive=%d  expired=%d  (ttl=%s)\n", finalHit, finalMiss, ttl)
	fmt.Println("Tip: run with:  go run -race .")
}

// helper для запуску воркерів на час d
func doTimed(d time.Duration, spawn func(ctx context.Context, wg *sync.WaitGroup)) {
	ctx, cancel := context.WithTimeout(context.Background(), d)
	var wg sync.WaitGroup
	spawn(ctx, &wg)
	wg.Wait()
	cancel()
}
