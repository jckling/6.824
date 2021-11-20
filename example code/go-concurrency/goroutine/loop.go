package main

import "sync"

func main() {
	var wg sync.WaitGroup	// 等待多个线程完成，效果同 channel/wait.go
	for i := 0; i < 5; i++ {
		wg.Add(1)	// 计数器加 1
		go func(x int) {	// 并行，注意传参
			sendRPC(x)
			wg.Done()	// 计数器减 1
		}(i)
	}
	wg.Wait()	// 计数器清零
}

func sendRPC(i int) {
	println(i)
}
