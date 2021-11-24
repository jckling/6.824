# 实验说明

> [6.824 Lab 1: MapReduce](http://nil.csail.mit.edu/6.824/2021/labs/lab-mr.html)

## 介绍

构建一个 MapReduce 系统：实现一个调用应用程序 Map 和 Reduce 函数并处理文件读写的 worker 进程，以及一个向 worker 分派任务并处理故障 worker 的协调者进程。

*注：这里用的是协调者（coordinator）而不是主控制器（master）*

## 环境

我个人使用的是 Ubuntu 20.02 Server

```bash
# 配置 golang 1.15 环境
wget -qO- https://golang.org/dl/go1.15.8.linux-amd64.tar.gz | sudo tar xz -C /usr/local
export PATH=$PATH:/usr/local/go/bin

# 下载实验代码
git clone git://g.csail.mit.edu/6.824-golabs-2021 6.824
cd 6.824
ls
```

src/main/mrsequential.go 中提供了一个简单的顺序 mapreduce 实现，在单进程中运行，每次只运行一个 Map 和 Reduce 。

src/mrapps 中提供了几个应用程序：
- wc.go 单词计数
- indexer.go 文本索引器

```bash
# 切换路径
cd ~/6.824
cd src/main

# 编译
# -race 数据竞争检测
# -buildmode=plugin 将指定的包及其导入的包构建到 go 插件中，即创建共享库
go build -race -buildmode=plugin ../mrapps/wc.go

# 编译并运行 go 包，使用上一步生成的 wc.so
# 将 pg*.txt 作为参数输入
go run -race mrsequential.go wc.so pg*.txt

# 查看输出结果
more mr-out-0

# 删除输出数据
rm mr-out*
```

## 目标

实现分布式 MapReduce，由两个程序组成：协调者（coordinator）和工作者（worker）。一个协调者进程，一个或多个并行执行的工作者进程。在一个真实的系统中，worker 会在一堆不同的机器上运行，但实验只在一台机器上运行。
工作者将通过 RPC 与协调者通信，每个工作进程将向协调者请求任务，从一个或多个文件中读取任务的输入，执行任务，并将任务的输出写入一个或多个文件。协调者应该注意到工作进程是否在合理的时间内完成任务（实验中规定 10 秒），如果没有则将相同的任务交给不同的工作进程。

协调者和工作者的 `main` 例程（routine）在 main/mrcoordinator.go 和 main/mrworker.go 中，无需更改。实现应该放在 mr/coordinator.go、mr/worker.go、mr/rpc.go 中。

在单词计数 MapReduce 应用程序上运行代码的方法：

首先确保插件是重新构建的

```bash
go build -race -buildmode=plugin ../mrapps/wc.go
```

在 main 目录下运行协调者
- `pg-*.txt` 参数是输入文件；每个文件对应一个拆分（split），并且是一个 Map 任务的输入
- `-race` 标志在运行 go 时启用数据竞争检测

```bash
# 删除旧的输出（如果有）
rm mr-out*

# 运行
go run -race mrcoordinator.go pg-*.txt
```

在一个或多个其它窗口中，运行一些工作者程序：

```bash
go run -race mrworker.go wc.so
```

当工作者和协调者都完成后，查看 `mr-out-*` 中的输出。完成试验后，输出文件的排序并集应该和顺序输出匹配，如下所示：

```bash
cat mr-out-* | sort | more
# A 509
# ABOUT 2
# ACT 8
# ...
```

提供了一个测试脚本 main/test-mr.sh，检查 wc 和 indexer。MapReduce 应用程序在给定 pg-xxx.txt 文件作为输入时是否产生正确输出，测试还会检查实现是否并行地运行了 Map 和 Reduce 任务，是否能够从运行任务时崩溃的工作者那里恢复过来。

如果现在运行测试脚本，它会挂起，因为协调者永远不会完成：

```bash
# 切换目录
cd ~/6.824/src/main

# 运行脚本
bash test-mr.sh
# *** Starting wc test.
```

可以在 mr/coordinator.go 中的 Done 函数中将 `ret := false` 改为 true，以便协调者立即退出，这样运行的结果将会是：

```bash
bash test-mr.sh
# *** Starting wc test.
# sort: No such file or directory
# cmp: EOF on mr-wc-all
# --- wc output is not the same as mr-correct-wc.txt
# --- wc test: FAIL
```

测试脚本预期在 mr-out-X 文件中检测到输出，每个 Reduce 任务都会生成一个。mr/coordinator.go、mr/worker.go 为空时不会生成这些文件，因此测试会失败：

```bash
bash test-mr.sh
# *** Starting wc test.
# --- wc test: PASS
# *** Starting indexer test.
# --- indexer test: PASS
# *** Starting map parallelism test.
# --- map parallelism test: PASS
# *** Starting reduce parallelism test.
# --- reduce parallelism test: PASS
# *** Starting crash test.
# --- crash test: PASS
# *** PASSED ALL TESTS
```

同时 Go RPC 包还会报错：

```
2019/12/16 13:27:09 rpc.Register: method "Done" has 1 input parameters; needs exactly three
```

忽略这些消息；将协调者注册为 RPC 服务器会检查其所有方法是否适用于 RPC（有 3 个输入）；函数 Done 不是通过 RPC 调用。


## 规则

map 阶段应该将中间键划分为 `nReduce` reduce 任务的桶，`nReduce` 是 main/mrcoordinator.go 传给 MakeCoordinator() 的参数

worker 实现应该把第 X 个 reduce 任务的输出写入 mr-out-X 文件。

mr-out-X 文件中的一行代表一个 Reduce 函数的输出。每行数据应该用 Go 的 `%v %v` 格式生成，使用键和值调用。可以看看 main/mrsequential.go 中注释为 `this is the correct format` 的那一行，如果实现过于偏离这个格式，将无法通过测试脚本。

可以修改 mr/wokrer.go、mr/coordinator.go、mr/rpc.go，也可以临时修改其他文件进行测试，最后请保证代码能在原始版本中运行。

工作者应该将中间的 Map 输出放在当前目录的文件中，之后可以读取并作为 Reduce 任务的输入。

main/mrcoordinator.go 预期 mr/coordinator.go 实现一个 Done() 方法，该方法在 MapReduce 作业全部完成时返回 true，此时 mrcoordinator.go 将退出。

当任务全部完成时，工作者进程应该退出。实现这一点的简单方法是使用 call() 的返回值：如果工作者无法联系协调者，可以假设协调者已退出，因为任务已经完成，所以工作者也可以终止。根据设计，可以让协调者给工作者分派 `please exit` 伪任务，这可能很有帮助。

## 提示

可以这么开始：修改 mr/work.go 的 Worker()，通过 RPC 向协调者请求任务；修改协调者，使其回应一个文件名作为尚未启动的 map 任务；修改工作者来读取该文件，并调用应用程序的 Map 函数，就像 mrsequential.go 中那样。

应用程序的 Map 和 Reduce 函数在运行时使用 Go 插件包加载，文件名以 `.so` 结尾。

如果修改了 mr/ 目录中的任何内容，可能需要重新构建任何使用的 MapReduce 插件，例如 `go build -race -buildmode=plugin ../mrapps/wc.go` 。

实验依赖工作者共享一个文件系统。当所有的工作者都在同一台机器上运行时，这很简单，但如果工作者在不同的机器上运行，就需要像 GFS 这样的全局文件系统。

一个合理的中间文件名约定是 `mr-X-Y`，其中 X 是 Map 任务编号，Y 是 Reduce 任务编号。

工作者的 map 任务代码需要一种方法存储中间键值对，以便在 reduce 任务期间可以正确读取。一个方法是使用 Go 的 `encoding/json` 包，将键/值对写入 JSON 文件：

```go
enc := json.NewEncoder(file)
for _, kv := ... {
  err := enc.Encode(&kv)
}
```

从文件中读取：

```go
dec := json.NewDecoder(file)
for {
  var kv KeyValue
  if err := dec.Decode(&kv); err != nil {
    break
  }
  kva = append(kva, kv)
}
```

工作者的 map 部分可以使用 `ihash(key)` 函数（在 worker.go 中）根据给定的键挑选 reduce 任务。

可以参考 mrsequential.go 的一些写法，读取 Map 输入文件、对 Map 和 Reduce 之间的中间键/值对进行排序，以及将 Reduce 输出写入文件。

协调者作为 RPC 服务器，是并发的；不要忘记锁定共享数据。

使用 Go 的竞争检测器，`go build -race` 和 `go run -race`，`test-mr.sh` 默认使用竞争检测运行测试。

工作者有时需要等待，例如，在最后一个 map 任务完成前，reduce 任务无法启动。一种可能是工作者定期向协调者请求工作，在每次请求之间使用 `time.Sleep()` 睡眠。另一种可能是协调者中的相关 RPC 处理程序有一个循环等待，使用 `time.Sleep()` 或 `sync.Cond`。Go 在自己的线程中为每个 RPC 运行处理程序，因此一个处理程序在等待时不会阻碍协调者处理其他 RPC。

协调者无法可靠地区分崩溃的工作者、存活但由于某种原因停止的工作者、正在执行但速度太慢的工作者。可以做的是让协调者等待一段时间，然后放弃并将任务重新发布给不同的工作者。在实验中，让协调者等待 10 秒，之后协调者应该假定工作者宕机（虽然可能没有）。

如果选择实现备份任务（第 3.6 节），测试代码在没有崩溃的工作者执行任务时不会安排无关的任务。备份任务应该只在某个相对较长的时间（例如 10 秒）后被分派。

测试崩溃恢复，可以使用 mrapps/crash.go 应用程序插件，它在 Map 和 Reduce 函数中随机退出。

为了确保在奔溃时没有工作者读取到部分写入的文件，MapReduce 论文提到了使用临时文件并在完全写入后使用原子重命名。可以使用 `ioutil.TempFile` 创建一个临时文件，使用 `os.Rename` 对其进行原子重命名。

test-mr.sh 运行子目录 mr-tmp 中的所有进程，出现问题之后可以在其中查看中间文件或输出文件。可以修改 test-mr.sh 使其在测试失败后退出，然后脚本就不会继续测试（并覆盖输出文件）。

test-mr-many.sh 提供了一个运行 test-mr.sh 的基本脚本，并带有超时功能。它将运行测试的次数作为参数，所以不应该同时运行几个test-mr.sh 实例，因为协调者会重复使用同一个套接字，从而导致冲突。

## 挑战

实现自己的 MapReduce 应用程序（参见 mrapps/* 中的示例），例如分布式 Grep（MapReduce 论文的第 2.3 节）。

让 MapReduce 协调者和工作者在不同的机器上运行，需要将 RPC 设置为通过 TCP/IP 而不是 Unix 套接字进行通信（参阅 `Coordinator.server()` 中注释掉的内容），并使用共享文件系统读写文件，例如 AFS、S3。
