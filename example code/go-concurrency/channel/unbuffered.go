package main

import "time"
import "fmt"

func main() {
	c := make(chan bool)
	go func() {
		time.Sleep(1 * time.Second)
		<-c	// 接收者阻塞，直到通道上有数据；发送者阻塞，直到通道上有接收者
	}()
	start := time.Now()
	c <- true // 1.00088s blocks until other goroutine receives
	fmt.Printf("send took %v\n", time.Since(start))
}
