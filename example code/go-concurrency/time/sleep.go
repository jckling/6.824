package main

import "time"

func main() {
	time.Sleep(1 * time.Second)
	println("started")
	go periodic()
	time.Sleep(5 * time.Second) // wait for a while so we can observe what ticker does
}

func periodic() {	// 周期任务
	for {
		println("tick")
		time.Sleep(1 * time.Second)
	}
}
