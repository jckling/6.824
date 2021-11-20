package main

func main() {	// 效果同 goroutinue/loop.go (WaitGroup)
	done := make(chan bool)
	for i := 0; i < 5; i++ {
		go func(x int) {
			sendRPC(x)
			done <- true	// 发送完成到通道
		}(i)
	}
	for i := 0; i < 5; i++ {
		<-done	// 接收完成
	}
}

func sendRPC(i int) {
	println(i)
}
