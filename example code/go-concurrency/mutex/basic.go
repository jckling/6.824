package main

import "sync"
import "time"

func main() {
	counter := 0
	var mu sync.Mutex	// 互斥锁
	for i := 0; i < 1000; i++ {
		go func() {
			mu.Lock()
			defer mu.Unlock()		// 函数结束前释放锁
			counter = counter + 1	// 临界区
		}()
	}

	time.Sleep(1 * time.Second)
	mu.Lock()
	println(counter)
	mu.Unlock()
}
