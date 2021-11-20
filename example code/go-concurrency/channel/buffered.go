package main

import "time"
import "fmt"

func main() {
	c := make(chan bool, 1)	// 缓冲大小为 1。不强制同时完成发送和接收
	go func() {
		time.Sleep(1 * time.Second)
		<-c	// 通道中没有数据时，接收者阻塞；没有可用缓冲区时，发送者阻塞
	}()
	start := time.Now()
	c <- true	// 0s
	fmt.Printf("send took %v\n", time.Since(start))

	start = time.Now()
	c <- true	// 1.00076s
	fmt.Printf("send took %v\n", time.Since(start))
}
