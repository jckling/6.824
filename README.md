# 6.824

个人笔记：https://jck.notion.site/6-824-2221ab969f6644349f1b323e431939c9 如有错误，麻烦提 issue 指正~


Schedule & Video & Reference
- [6.824 Schedule: Spring 2021](http://nil.csail.mit.edu/6.824/2021/schedule.html)
- [6.824 Schedule: Spring 2020](http://nil.csail.mit.edu/6.824/2020/schedule.html)
- [2020 MIT 6.824 分布式系统](https://www.bilibili.com/video/BV1R7411t71W)
- [6.824 / Spring 2021 [麻省理工分布式系统 - 2021年春季]](https://www.bilibili.com/video/BV16f4y1z7kn)
- [chaozh/MIT-6.824](https://github.com/chaozh/MIT-6.824)


注释：
- 2020-LEC5 和 2021-LEC10 都是 Go 语言相关的内容；
- 2021-LEC6 是 Lab1 Q&A；2021-LEC8 是 Lab2 A+B Q&A；
- 2020-LEC9 关于 CRAQ 内容太少，建议补充 2021-LEC11 关于 CR 的内容； 


---


Papers
- [MapReduce: Simplified Data Processing on Large Clusters](https://pdos.csail.mit.edu/6.824/papers/mapreduce.pdf)
- [The Google File System](http://nil.csail.mit.edu/6.824/2021/papers/gfs.pdf)
  - [Case Study - GFS: Evolution on Fast-forward](https://queue.acm.org/detail.cfm?id=1594206)
- [The Design of a Practical System for Fault-Tolerant Virtual Machines](http://nil.csail.mit.edu/6.824/2021/papers/vm-ft.pdf)
- [In Search of an Understandable Consensus Algorithm (Extended Version)](http://nil.csail.mit.edu/6.824/2021/papers/raft-extended.pdf)
- [ZooKeeper: Wait-free coordination for Internet-scale systems](http://nil.csail.mit.edu/6.824/2021/papers/zookeeper.pdf)
- [Chain Replication for Supporting High Throughput and Availability](http://nil.csail.mit.edu/6.824/2021/papers/cr-osdi04.pdf)
- [Object Storage on CRAQ: High-throughput chain replication for read-mostly workloads](http://nil.csail.mit.edu/6.824/2020/papers/craq.pdf)
- [Amazon Aurora: Design Considerations for High Throughput Cloud-Native Relational Databases](http://nil.csail.mit.edu/6.824/2020/papers/aurora.pdf)
- [Frangipani: A Scalable Distributed File System](http://nil.csail.mit.edu/6.824/2020/papers/thekkath-frangipani.pdf)
- [Chapter 9: Atomicity: All-or-nothing and Before-or-after](https://ocw.mit.edu/resources/res-6-004-principles-of-computer-system-design-an-introduction-spring-2009/online-textbook/)
  - 9.1.5、9.1.6、9.5.2、9.5.3、9.6.3


---


Labs                                                                                                   | Task
-------------------------------------------------------------------------------------------------------|--------------------------
[Lab 1: MapReduce](http://nil.csail.mit.edu/6.824/2021/labs/lab-mr.html)                               | 1 part
[Lab 2: Raft](http://nil.csail.mit.edu/6.824/2021/labs/lab-raft.html)                                  | 4 parts
[Lab 3: Fault-tolerant Key/Value Service](http://nil.csail.mit.edu/6.824/2021/labs/lab-kvraft.html)    | 2 parts
[Lab 4: Sharded Key/Value Service](http://nil.csail.mit.edu/6.824/2021/labs/lab-shard.html)            | 2 parts
