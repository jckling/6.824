package main

func main() {
	// go func () {	// 有一个线程在运行，go run 不会检测到死锁
	// 	for {}
	// }()
	c := make(chan bool)
	c <- true	// 发送者等待通道上出现接收者
	<-c			// 接收者等待通道上出现数据
	// go run unbuffered-deadlock.go 检测到死锁
}
