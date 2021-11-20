package main

import "sync"

func main() {
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {		// 没有传参（i 会被外层改变）
			sendRPC(i)
			wg.Done()
		}()
	}
	wg.Wait()
}

func sendRPC(i int) {
	println(i)
}
