package main

import "time"

func main() {
	counter := 0
	for i := 0; i < 1000; i++ {
		go func() {
			counter = counter + 1	// 并发 goroutine，读取到相同的 counter
		}()
	}

	time.Sleep(1 * time.Second)
	println(counter)
}
