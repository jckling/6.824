package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"os"
	"strconv"
)

//
// example to show how to declare the arguments
// and reply for an RPC.
//

type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

// Add your RPC definitions here.
type TaskType int

const (
	Map TaskType = iota
	Reduce
	Done // There are no pending tasks
)

// No arguments to send the coordinator to ask for a task
type GetTaskArgs struct{}

// Note: RPC fields need to be capitalized in order to be sent!
type GetTaskReply struct {
	// type of task
	TaskType TaskType

	// task number of either map or reduce task
	TaskNum int

	// Map: which file to write
	NReduceTasks int

	// Map: which file to read
	MapFile string

	// Reduce: how many intermediate map files to read
	NMapTasks int
}

// sent from an idle worker to coordinator to indicate that a task has been completed
type FinishedTaskArgs struct {
	// type of task
	TaskType TaskType

	// which task
	TaskNum int
}

// workers don't need to get a reply
type FinishedTaskReply struct{}

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/824-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
