package main

import "sync"

func main() {
	var a string
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {	// 匿名函数/闭包
		a = "hello world"
		wg.Done()
	}()
	wg.Wait()
	println(a)
}
