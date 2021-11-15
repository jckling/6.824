package main

import (
	"fmt"
	"sync"
)

//
// Several solutions to the crawler exercise from the Go tutorial
// https://tour.golang.org/concurrency/10
//

//
// Serial crawler
//

func Serial(url string, fetcher Fetcher, fetched map[string]bool) {
	if fetched[url] {
		return
	}
	fetched[url] = true
	urls, err := fetcher.Fetch(url)
	if err != nil {
		return
	}
	for _, u := range urls {
		Serial(u, fetcher, fetched)	// go Serial(u, fetcher, fetched) 无法正确并行，因为线程共享对象和锁
	}
	return
}

//
// Concurrent crawler with shared state and Mutex
//

type fetchState struct {
	mu      sync.Mutex
	fetched map[string]bool
}

func ConcurrentMutex(url string, fetcher Fetcher, f *fetchState) {
	f.mu.Lock()
	already := f.fetched[url]
	f.fetched[url] = true
	f.mu.Unlock()

	if already {
		return
	}

	urls, err := fetcher.Fetch(url)
	if err != nil {
		return
	}
	var done sync.WaitGroup	// 内部有互斥锁或类似机制
	for _, u := range urls {
		done.Add(1)
		//u2 := u
		//go func() {
		// defer done.Done()	// 总是在包含它的函数结束前调用 done.Done()
		// ConcurrentMutex(u2, fetcher, f)
		//}()
		go func(u string) {	// 闭包/匿名函数；参数 u 不能省略，因为 for 循环中的 u 会更新（指向不同的字符串），所以要让 goroutine 有自己的一份拷贝
			defer done.Done()
			ConcurrentMutex(u, fetcher, f)	// 闭包中使用的变量会被 go 放到堆中，当外部函数返回时，对象仍然在堆上，之后由垃圾回收负责释放
		}(u)
	}
	done.Wait()	// 等待最后一个 goroutine 结束（计数归零）
	return
}

func makeState() *fetchState {
	f := &fetchState{}
	f.fetched = make(map[string]bool)
	return f
}

//
// Concurrent crawler with channels
//

func worker(url string, ch chan []string, fetcher Fetcher) {
	urls, err := fetcher.Fetch(url)
	if err != nil {
		ch <- []string{}	// 向 coordinator 发送消息
	} else {
		ch <- urls			// 通道内部实现有一个互斥锁
	}
}

func coordinator(ch chan []string, fetcher Fetcher) {
	n := 1
	fetched := make(map[string]bool)	// 私有变量，记录哪些网页抓取过
	for urls := range ch {				// 从通道中读取数据（由 worker 发送）
		for _, u := range urls {
			if fetched[u] == false {
				fetched[u] = true
				n += 1
				go worker(u, ch, fetcher)	// 没有抓取，启动新的 worker 抓取
			}
		}
		n -= 1
		if n == 0 {
			break	// 所有 worker 执行完毕，结束等待（否则最外层 for 会一直等待通道上出现数据）
		}
	}
}

func ConcurrentChannel(url string, fetcher Fetcher) {
	ch := make(chan []string)
	go func() {
		ch <- []string{url}
	}()
	coordinator(ch, fetcher)
}

//
// main
//

func main() {
	fmt.Printf("=== Serial===\n")
	Serial("http://golang.org/", fetcher, make(map[string]bool))

	fmt.Printf("=== ConcurrentMutex ===\n")
	ConcurrentMutex("http://golang.org/", fetcher, makeState())

	fmt.Printf("=== ConcurrentChannel ===\n")
	ConcurrentChannel("http://golang.org/", fetcher)
}

//
// Fetcher
//

type Fetcher interface {
	// Fetch returns a slice of URLs found on the page.
	Fetch(url string) (urls []string, err error)
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) ([]string, error) {
	if res, ok := f[url]; ok {
		fmt.Printf("found:   %s\n", url)
		return res.urls, nil
	}
	fmt.Printf("missing: %s\n", url)
	return nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"http://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"http://golang.org/pkg/",
			"http://golang.org/cmd/",
		},
	},
	"http://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"http://golang.org/",
			"http://golang.org/cmd/",
			"http://golang.org/pkg/fmt/",
			"http://golang.org/pkg/os/",
		},
	},
	"http://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
	"http://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
}
