package main

import "time"
import "sync"

var done bool		// 全局变量
var mu sync.Mutex	// 互斥锁

func main() {
	time.Sleep(1 * time.Second)
	println("started")
	go periodic()
	time.Sleep(5 * time.Second) // wait for a while so we can observe what ticker does
	mu.Lock()
	done = true
	mu.Unlock()
	println("cancelled")
	time.Sleep(3 * time.Second) // observe no output
}

func periodic() {
	for {
		println("tick")
		time.Sleep(1 * time.Second)
		mu.Lock()
		if done {	// 终止 goroutine
			// mu.Unlock()	// 返回前释放锁，可以使用 defer 关键字
			return
		}
		mu.Unlock()
	}
}
