package main

import "sync"
import "time"
import "fmt"

func main() {
	alice := 10000
	bob := 10000
	var mu sync.Mutex

	total := alice + bob

	go func() {	// 整个操作分为两步，不是原子的，因此可以观察到冲突（改变了 total）
		for i := 0; i < 1000; i++ {
			mu.Lock()
			alice -= 1	// 原子
			mu.Unlock()
			mu.Lock()
			bob += 1	// 原子
			mu.Unlock()
		}
	}()
	go func() {
		for i := 0; i < 1000; i++ {
			mu.Lock()
			bob -= 1
			mu.Unlock()
			mu.Lock()
			alice += 1
			mu.Unlock()
		}
	}()

	start := time.Now()
	for time.Since(start) < 1*time.Second {
		mu.Lock()
		if alice+bob != total {	// 检查总数
			fmt.Printf("observed violation, alice = %v, bob = %v, sum = %v\n", alice, bob, alice+bob)
		}
		mu.Unlock()
	}
}
