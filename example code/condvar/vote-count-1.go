package main

import "time"
import "math/rand"

func main() {
	rand.Seed(time.Now().UnixNano())

	count := 0		// 同意的票数
	finished := 0	// 总票数

	for i := 0; i < 10; i++ {
		go func() {
			vote := requestVote()
			if vote {
				count++	// 共享变量没加锁
			}
			finished++	// 共享变量没加锁
		}()
	}

	for count < 5 && finished != 10 {
		// wait
	}
	if count >= 5 {
		println("received 5+ votes!")
	} else {
		println("lost")
	}
}

func requestVote() bool {
	time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
	return rand.Int() % 2 == 0
}
