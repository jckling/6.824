package main

import (
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	count := 0
	ch := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			ch <- requestVote() // 往通道发送 10 次数据
		}()
	}
	for i := 0; i < 10; i++ { // 接收 10 次数据，省略 finished 变量
		v := <-ch
		if v {
			count += 1
		}
	}
	if count >= 5 {
		println("received 5+ votes!")
	} else {
		println("lost")
	}
}

func requestVote() bool {
	time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
	return rand.Int()%2 == 0
}
