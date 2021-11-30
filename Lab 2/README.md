# 实验说明

> [6.824 Lab 2: Raft](http://nil.csail.mit.edu/6.824/2021/labs/lab-raft.html)

## 介绍

建立一个容错的键/值存储系统。在这个实验中，你将实现 Raft，一个复制状态机协议；在下一个实验中，你将在 Raft 上构建一个键/值服务；然后你将在多个复制状态机上“分片”服务器，以获得更高的性能。

复制服务通过将其状态（即，数据）的完整副本存储在多个副本服务器上来实现容错。即使一些服务器出现故障（崩溃、网络崩溃、网络波动），复制也允许服务继续运行。挑战是失败可能导致副本服务器持有不同的数据副本。

Raft 将客户端请求组织成序列，称为日志（log），并确保所有副本服务器看到相同的日志。每个副本服务器按日志中的顺序执行客户端请求，将其应用于服务状态的本地副本。由于所有活动的副本服务器都看到相同的日志内容，都以相同的顺序执行相同的请求，因此会继续拥有相同的服务状态。如果一个服务器发生故障但后来恢复了，Raft 会负责更新其日志。只要至少大多数服务器都在活动，并且能够互相通信，Raft 就会继续运行。如果没有这样的大多数，Raft 将不会继续运行，只要有大多数可以相互通信，Raft 就会从其停止的地方继续运行。

在这个实验中，你将把 Raft 实现为一个带有相关方法的 Go 对象类型，之后用作更大服务中的模块。一组 Raft 实例通过 RPC 互相通信，维护复制的日志。你的 Raft 接口将支持无限的编号命令序列，也称为日志条目（log entry），日志条目使用索引号（index numbers）编号。具有特定索引的日志条目最终会被提交，此时你的 Raft 应该将日志条目发送到更大的服务，使其继续执行。

你应该遵循论文中的设计，特别注意图 2。你将实现论文中的大部分内容，包括持久保存状态，在节点故障重启后读取状态。你不用实现集群成员身份改变（第 6 节）。

你可能会发现这个 [指南](https://thesquareplanet.com/blog/students-guide-to-raft/) 很有用，还有这个关于 [加锁](http://nil.csail.mit.edu/6.824/2021/labs/raft-locking.txt) 和 [结构](http://nil.csail.mit.edu/6.824/2021/labs/raft-structure.txt) 的建议。为了获得更广泛的视角，可以看看 Paxos、Chubby、Paxos Made Live、Spanner、Zookeeper、Harp、Viewstamped Replication、[Bolosky](http://static.usenix.org/event/nsdi11/tech/full_papers/Bolosky.pdf) 等等。

[Raft 交互图](http://nil.csail.mit.edu/6.824/2021/notes/raft_diagram.pdf)，可以帮助阐明你的 Raft 代码如何与其上层进行交互。

本实验分为四个部分。

## 开始工作

如果你已经完成了 Lab 1，那么你已经拥有实验源码的副本。如果没有，可以在 Lab 1 的说明中找到通过 Git 获得源码的方法。

我们提供骨架代码 src/raft/raft.go，以及一系列测试 src/raft/test_test.go，你应该使用它们推动你的实现，我们会用它来给提交的代码评分。

要启动和运行，请执行以下命令。不要忘记用 `git pull` 来获取最新的代码。

```bash
# 更新代码
cd ~/6.824
git pull

# 运行测试
cd src/raft
go test -race
```

## 代码

通过向 raft/raft.go 添加代码来实现 Raft。在该文件中，你会发现骨架代码，以及如何发送和接收 RPC 的示例。

你的实现必须支持以下接口，测试器和（最终）的键/值服务器将使用。可以在 raft.go 的注释中找到更多详细信息。

```go
// create a new Raft server instance:
rf := Make(peers, me, persister, applyCh)

// start agreement on a new log entry:
rf.Start(command interface{}) (index, term, isleader)

// ask a Raft for its current term, and whether it thinks it is leader
rf.GetState() (term, isLeader)

// each time a new entry is committed to the log, each Raft peer
// should send an ApplyMsg to the service (or tester).
type ApplyMsg
```

服务调用 `Make(peers, me, ...)` 来创建一个 Raft 对等体（peer），对等体参数是 Raft 对等体的网络标识符数组，和 RPC 一起使用。`me` 参数是该对等体在对等体数组中的索引。`Start(command)` 要求 Raft 开始处理，将该命令添加到复制日志中。`Start()` 应该立即返回，无需等待日志追加完成。该服务希望你的实现能够为每个新提交的日志条目发送 `ApplyMsg` 到 `Make()` 的 `appCh` 通道参数。

raft.go 包含发送 RPC（`sendRequestVote()`）和处理传入 RPC（`RequestVote()`）的示例代码。你的 Raft 对等体应该使用 labrpc Go 包（源码在 src/labrpc 中）交换 RPC。测试器可以告诉 labrpc 延迟 RPC，重新排序，丢弃，以模拟各种网络故障。虽然你可以临时修改 labrpc，但要确保你的 Raft 和原始的 labrpc 能够一起使用，因为我们会用它来对你的代码进行测试和评分。你的 Raft 实例只能用 RPC 交互；例如，不允许使用共享的 Go 变量或文件进行通信。

后续的实验将以本实验为基础，因此花足够的时间来编写可靠的代码非常重要。

## Part 2A: leader election

### Task

实现 Raft 领导者选举和心跳（没有日志条目的 `AppendEntries` RPC）。第 2A 部分的目标是选出一个领导者，如果没有失败，原来的领导者仍然担任领导者；如果原来的领导者失败或从/去原来的领导者的数据包丢失，则由新的领导者接管。运行 `go test -run 2A -race` 来测试 2A 代码。

### Hints

- 你不能简单地直接运行你的 Raft 实现；相反，应该通过测试器来运行：`go test -run 2A -race`。
- 按照论文的图 2，此时你关心的是发送和接收 `RequestVote` 的 RPC、与选举有关的服务器规则（Rule）、以及与领导选举相关的状态（State）。
- 在 raft.go 的 Raft 结构体中添加图 2 中的领导者选举状态。你还需要定义一个结构来保存每个日志条目的信息。
- 填写 `RequestVoteArgs` 和 `RequestVoteReply` 的结构。修改 `Make()` 创建后台 goroutine，当它有一段时间没有收到另一个对等体的消息时，它将通过发送 `RequestVote` RPC 来定期启动领导者选举。通过这种方式，如果已经有了领导者，对等体将知道谁是领导者，或自己成为领导者。实现 `RequestVote` RPC 处理程序（handler），以便服务器相互投票。
- 为了实现心跳，定义一个 `AppendEntries` RPC 结构（尽管你可能还不需要所有参数），并让领导者定期发送。编写一个 `AppendEntries` RPC 处理程序，重置选举超时，以便在一个服务器已经当选的情况下，其他服务器不会再成为领导者。
- 确保不同对等体的选举超时不会总是同时触发，否则所有对等体只会给自己投票，没有人会成为领导者。
- 测试器要求领导者每秒发送心跳 RPC 不超过 10 次。
- 测试器要求你的 Raft 在原来的领导者失败后的 5 秒内选举出一个新的领导者（如果大多数对等体仍能通信）。但是请记住，如果出现分裂投票（如果数据包丢失或候选者不幸选择相同的随机后退时间，可能会发生这种情况），可能需要进行多轮领导者选举。你必须选择足够短的选举超时（以及心跳间隔），即使需要多轮选举，也很可能在不到 5 秒的时间内完成。
- 论文的第 5.2 节提到选举超时的范围是 150 到 300 毫秒。只有当领导者发送心跳的频率远高于每 150 毫秒一次时，这个范围才有意义。因为测试器会限制每秒 10 次心跳，所以你必须使用更大的选举超时，但也不能太大，否则可能无法在 5 秒内选举出领导者。
- 你可能会发现 Go 的 [rand](https://golang.org/pkg/math/rand/) 很有用。
- 你需要编写代码定期或延迟采取行动。最简单的方法是创建带循环调用 [time.Sleep()](https://golang.org/pkg/time/#Sleep) 的 goroutine；（参阅 `Make()` 为此目的创建的 `ticker()` goroutinue）。不要使用 Go 的 `time.Timer` 或 `time.Ticker`，它们很难正确使用。
- [指导页面](http://nil.csail.mit.edu/6.824/2021/labs/guidance.html) 提供了一些关于如何开发和调试代码的提示。
- 如果你的代码无法通过测试，请再次阅读论文的图 2；领导者选举的完整逻辑分布在图中的多个部分。
- 不要忘记实现 `GetState()`。
- 测试器在永久关闭一个实例时，会调用你实现的 Raft 的 `rf.Kill()` 。你可以使用 `rf.killed()` 检查 `Kill()` 是否被调用。你可能希望在所有的循环中都这么做，避免结束的 Raft 实例打印出混乱的信息。
- Go RPC 只发送名称以大写字母开头的结构体字段。子结构也必须有大写的字段名（例如，数组中的日志记录字段）。labgob 包会警告你这一点；不要忽视警告。

### Submit

确保在提交前通过 2A 测试，测试通过将看到类似这样的内容：

```bash
go test -run 2A -race
# Test (2A): initial election ...
#   ... Passed --   4.0  3   32    9170    0
# Test (2A): election after network failure ...
#   ... Passed --   6.1  3   70   13895    0
# PASS
# ok      raft    10.187s
```

每个 `Passed` 包含五个数字：测试所用的时间（秒）、Raft 对等体的数量（通常为 3 或 5 个）、测试期间发送的 RPC 数量、RPC 消息中的总字节数、以及 Raft 报告提交的日志条目数量。

## Part 2B: log

### Task

实现领导者和追随者代码，追加新的日志条目，以便通过 `go test -run 2B -race` 测试。

### Hints

- 运行 `git pull` 获取最新的代码。
- 你的第一个目标应该是通过 `TestBasicAgree2B()`。从实现 `Start()` 开始，然后编写代码通过 `AppendEntries` RPC 发送和接收新的日志条目，如图 2 所示。
- 你将要实现选举限制（论文的第 5.4.1 节）。
- 在早期的 Lab 2B 测试中，无法达成协议的一种方法是：即使领导者还在活动，也要重复进行选举。寻找选举计时器管理中的错误，或者在赢得选举后没有立即发送心跳的问题。
- 你的代码可能有重复检查某些事件的循环。不要让这些循环在没有暂停的情况下连续执行，因为这会减慢执行速度从而导致无法通过测试。使用 Go 的 [条件变量](https://golang.org/pkg/sync/#Cond)，或在每个循环迭代中插入 `time.Sleep(10 * time.Millisecond)`。
- 为之后的实验做准备，编写简洁的代码。更多信息请重新访问 [指导页面](http://nil.csail.mit.edu/6.824/2021/labs/guidance.html)，了解如何开发和调试代码的技巧。
- 如果测试失败，请查看 config.go 和 test_test.go 中的测试代码，以便更好地了解测试的内容。config.go 还说明了测试器如何调用 Raft API。

### Submit

如果代码运行速度太慢，即将进行的实验测试可能会失败。可以使用 `time` 命令检查你的实现所花费的实时时间和 CPU 时间。以下是一个典型的输出：

```bash
time go test -run 2B
# Test (2B): basic agreement ...
#   ... Passed --   1.6  3   18    5158    3
# Test (2B): RPC byte count ...
#   ... Passed --   3.3  3   50  115122   11
# Test (2B): agreement despite follower disconnection ...
#   ... Passed --   6.3  3   64   17489    7
# Test (2B): no agreement if too many followers disconnect ...
#   ... Passed --   4.9  5  116   27838    3
# Test (2B): concurrent Start()s ...
#   ... Passed --   2.1  3   16    4648    6
# Test (2B): rejoin of partitioned leader ...
#   ... Passed --   8.1  3  111   26996    4
# Test (2B): leader backs up quickly over incorrect follower logs ...
#   ... Passed --  28.6  5 1342  953354  102
# Test (2B): RPC counts aren't too high ...
#   ... Passed --   3.4  3   30    9050   12
# PASS
# ok      raft    58.142s

# real    0m58.475s
# user    0m2.477s
# sys     0m1.406s
```

`ok raft 58.142s` 表示 Go 测量到的 2B 测试时间是 58.142 秒的实际（挂钟）时间。`user 0m2.477s` 表示代码消耗了 2.477 秒的 CPU 时间，或实际执行指令的时间（而不是等待或睡眠）。如果你的实现在 2B 测试中使用的实际时间远远超过 1 分钟，或者远超过 5 秒的 CPU 时间，之后可能遇到麻烦。查找睡眠或等待 RPC 超时所花费的时间、在没有睡眠或等待条件或通道消息的情况下运行循环、或发送大量 RPC 等问题。

## Part 2C: persistence

如果 Raft 的服务器重新启动，它应该从停止的地方恢复服务。这要求 Raft 能够持久保存状态。论文的图 2 提到了哪些状态应该是持久的。

真正的实现会在 Raft 每次改变状态时将其写入磁盘，并在重启后从磁盘读取状态。你的实现不会使用磁盘；相反，它将从一个 `Persister` 对象保存和恢复持久状态（参阅 perister.go）。任何 `Raft.Make()` 调用都会提供一个 `Persister`，最初保存 Raft 最近的持久状态（如果有的话）。Raft 应该从该 `Persister` 中初始化它的状态，并且应该在每次状态改变时用它来保存持久状态。使用 `Persister` 的 `ReadRaftState()` 和 `SaveRaftState()` 方法。

### Task

通过添加保存和恢复持久状态的代码，实现 raft.go 中的 `persist()` 和 `readPersist()` 函数。你需要把状态编码（序列化）成字节数组，以便传递给 `Persister`。使用 labgob 编码器；参阅 `persist()` 和 `readPersist()` 中的注释。labgob 类似于 Go 的 gob 编码器，如果用小写的字段名称对结构进行编码，它会打印出错误信息。

在更改持久状态的地方插入对 `persist()` 的调用。完成之后应该就能通过测试。

### Note

为了避免内存不足，Raft 必须定期丢弃旧的日志条目，在下一个实验之前不用担心这点。

### Hints

- 运行 `git pull` 获取最新的代码。
- 第 2C 部分的许多测试涉及到服务器故障和网络丢失 RPC 请求/回复。这些事件是不确定的，即使代码有问题也可能通过测试。通常情况下，多次运行测试就会暴露这些问题。
- 你可能需要通过一次备份多个条目实现优化。查看论文从第 7 页底部和第 8 页顶部开始的内容（用灰线标记）。论文中的细节很模糊；可以借助课程填补这些空白。
- 第 2C 部分只要求你实现持久性和快速日志回溯，但测试失败可能与之前的实现有关。即使你通过了 2A 和 2B 测试，仍然可能会有选举或日志错误，然后在 2C 测试中暴露。

### Submit

你的代码应该通过所有 2C 测试（如下），以及 2A 和 2B 测试。

```bash
go test -run 2C -race
# Test (2C): basic persistence ...
#   ... Passed --   7.2  3  206   42208    6
# Test (2C): more persistence ...
#   ... Passed --  23.2  5 1194  198270   16
# Test (2C): partitioned leader and one follower crash, leader restarts ...
#   ... Passed --   3.2  3   46   10638    4
# Test (2C): Figure 8 ...
#   ... Passed --  35.1  5 9395 1939183   25
# Test (2C): unreliable agreement ...
#   ... Passed --   4.2  5  244   85259  246
# Test (2C): Figure 8 (unreliable) ...
#   ... Passed --  36.3  5 1948 4175577  216
# Test (2C): churn ...
#   ... Passed --  16.6  5 4402 2220926 1766
# Test (2C): unreliable churn ...
#   ... Passed --  16.5  5  781  539084  221
# PASS
# ok      raft    142.357s
```

在提交之前最好多次运行测试，检查每次运行是否都打印 `PASS`。

```bash
for i in {0..10}; do go test; done
```

## Part 2D: log compaction

按照现在的代码，重启服务会复制完整的 Raft 日志以恢复状态。然而，对于一个长期运行的服务来说，永远记住完整的 Raft 日志是不现实的。相反，你将修改 Raft，使其合作以节省空间：一个服务会不时地持久存储当前状态的快照（snapshot），Raft 会丢弃快照之前的日志条目。当一个服务远远落后于领导者且必须赶上时，该服务将首先安装快照，然后从创建快照之后的点重放日志条目。论文的第 7 节概述了该方案，你将设计细节。

你可能会发现参考 [Raft 交互图](http://nil.csail.mit.edu/6.824/2021/notes/raft_diagram.pdf) 对理解复制服务和 Raft 通信方式会很有帮助。

为了支持快照，需要一个服务和 Raft 库之间的接口。Raft 论文并没有指定这个接口，有几种可能的设计。为了实现简单，我们决定在服务和 Raft 之间使用以下接口：
- `Snapshot(index int, snapshot []byte)`
- `CondInstallSnapshot(lastIncludedTerm int, lastIncludedIndex int, snapshot []byte) bool`

一个服务调用 `Snapshot()` 向 Raft 传递其状态的快照。快照包括截至并包括索引的所有信息。这意味着相应的 Raft 对等体不再需要到（并包含）索引的日志。你的 Raft 实现应该尽可能减小日志。你必须修改 Raft 代码，以便在仅存储日志尾部的同时进行操作。

正如论文中所讨论的，Raft 的领导者有时必须告诉落后的 Raft 对等体通过安装快照来更新状态。出现这种情况时，你要实现 `InstallSnapshot` RPC 发送和处理程序，用于安装快照。这与 `AppendEntries` 形成对比，后者发送日志条目，然后由服务逐一应用。

注意 `InstallSnapshot` RPC 在 Raft 对等体之间发送，而服务使用提供的骨架函数 `Snapshot/CondInstallSnapshot` 与 Raft 通信。

当追随者接收并处理 `InstallSnapshot` RPC 时，它必须使用 Raft 将包含的快照发送给服务。`InstallSnapshot` 处理程序可以使用 `applyCh` 将快照发送给服务，方法是将快照放在 `ApplyMsg` 中。服务从 `applyCh` 中读取，并使用快照调用 `CondInstallSnapshot`，告诉 Raft 该服务正在切换到传入的快照状态，同时 Raft 应该更新它的日志。（参阅 config.go 中的 `applierSnap()` 了解测试器如何执行该操作）

如果快照是旧的（即，如果 Raft 在快照的 `lastIncludedTerm/lastIncludedIndex` 之后处理了条目），`CondInstallSnapshot` 应该拒绝安装该快照。因为 Raft 可能会在处理 `InstallSnapshot` RPC 之后，在服务调用 `CondInstallSnapshot` 之前，处理其他 RPC 并在 `applyCh` 上发送消息。Raft 不允许回到旧快照，所以必须拒绝旧的快照。当你的实现拒绝快照时，`CondInstallSnapshot` 应该只返回 `false`，让服务知道它不应该切换到快照。

如果快照是最近的，那么 Raft 应该修剪它的日志，保持新状态，返回 `true`，并且服务应该在处理 `applyCh` 上的下一条消息之前切换到快照。

`CondInstallSnapshot` 是更新 Raft 和服务状态的一种方式；服务和 Raft 之间的其他接口也是可能的。这种特殊设计允许你的实现检查快照是否必须安装，并原子地将服务和 Raft 切换到快照。你可以自由地实现 Raft，以 `CondInstallSnapshot` 始终返回 `true` 的方式实现。

### Task

修改你的 Raft 代码以支持快照：实现 `Snapshot`、`CondInstallSnapshot` 和 `InstallSnapshot` RPC，以及对 Raft 修改以支持这些（例如，继续使用修剪过的日志进行操作）。当你的实现通过 2D 测试和所有 Lab 2 的测试时，整个实验就完成了。（注意，Lab 3 比 Lab 2 更彻底地测试快照，因为 Lab 3 有一个真正的服务来对 Raft 的快照进行压力测试）

### Hints

- 在单个 `InstallSnapshot` RPC 中发送整个快照。不要实现图 13 的用于拆分快照的偏移机制（`offset`）。
- Raft 必须以允许 Go 垃圾收集器释放和重用内存的方式丢弃旧的日志条目；这要求对丢弃的日志条目没有可访问的引用（指针）。
- Raft 日志不能再使用日志条目的位置或日志的长度来确定日志条目索引；你将要使用独立于日志位置的索引方案。
- 即使日志被修剪，你的实现仍然需要在 `AppendEntries` RPC 中的新条目之前正确发送条目的任期（term）和索引；这可能需要保存和引用最新快照的 `lastIncludedTerm/lastIncludedIndex`（考虑是否应该保留）。
- Raft 必须使用 `SaveStateAndSnapshot()` 将每个快照存储在 `Persister` 对象中。
- 完整的 Lab 2 测试集（2A+2B+2C+2D）的合理耗时是 8 分钟的真实时间和 1.5 分钟的 CPU 时间。
