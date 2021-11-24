package mr

import (
	_ "6.824/raft"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"sort"
)

//
// Map functions return a slice of KeyValue.
//
type KeyValue struct {
	Key   string
	Value string
}

//
// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
//
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

// for sorting by key
type ByKey []KeyValue

// for sorting by key
func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

// finalizeReduceFile atomically renames temporary reduce file to a completed reduce task file
func finalizeReduceFile(tmpFile string, taskN int) {
	finalFile := fmt.Sprintf("mr-out-%d", taskN)
	os.Rename(tmpFile, finalFile)
}

// get name of the intermediate file, given the map and reduce task numbers
func getIntermediateFile(mapTaskN int, reduceTaskN int) string {
	return fmt.Sprintf("mr-%d-%d", mapTaskN, reduceTaskN)
}

// finalizeIntermediateFile atomically renames temporary intermediate file to a completed intermediate file
func finalizeIntermediateFile(tmpFile string, mapTaskN int, reduceTaskN int) {
	finalFile := getIntermediateFile(mapTaskN, reduceTaskN)
	os.Rename(tmpFile, finalFile)
}

// Implementation of map task
func performMap(filename string, taskNum int, nReduceTasks int, mapf func(string, string) []KeyValue) {
	// read contents to map
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("cannot open %v", filename)
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalf("cannot read %v", filename)
	}
	file.Close()

	// apply map function to contents of file and collect the set of key-value pairs
	kva := mapf(filename, string(content))
	// create temporary files and encoders for each file
	tmpFiles := []*os.File{}
	tmpFilenames := []string{}
	encoders := []*json.Encoder{}
	for r := 0; r < nReduceTasks; r++ {
		tmpFile, err := ioutil.TempFile("", "")
		if err != nil {
			log.Fatalf("cannot open tmpfile")
		}
		tmpFiles = append(tmpFiles, tmpFile)
		tmpFilename := tmpFile.Name()
		tmpFilenames = append(tmpFilenames, tmpFilename)
		enc := json.NewEncoder(tmpFile)
		encoders = append(encoders, enc)
	}
	// write output keys to temporary intermediate files using the ihash function
	for _, kv := range kva {
		r := ihash(kv.Key) % nReduceTasks
		encoders[r].Encode(&kv)
	}
	for _, f := range tmpFiles {
		f.Close()
	}
	// atomically rename temporary files to final intermediate files
	for r := 0; r < nReduceTasks; r++ {
		finalizeIntermediateFile(tmpFilenames[r], taskNum, r)
	}
}

// Implementation of reduce task
func performReduce(taskNum int, nMapTasks int, reducef func(string, []string) string) {
	// get all intermediate files corresponding to reduce task, and collect corresponding key-value pairs
	kva := []KeyValue{}
	for m := 0; m < nMapTasks; m++ {
		iFilename := getIntermediateFile(m, taskNum)
		file, err := os.Open(iFilename)
		if err != nil {
			log.Fatalf("cannot open %v", iFilename)
		}
		dec := json.NewDecoder(file)
		for {
			var kv KeyValue
			if err := dec.Decode(&kv); err != nil {
				break
			}
			kva = append(kva, kv)
		}
		file.Close()
	}

	// sort the keys
	sort.Sort(ByKey(kva))

	// get temporary reduce file to write values
	tmpFile, err := ioutil.TempFile("", "")
	if err != nil {
		log.Fatalf("cannot open tmpfile")
	}
	tmpFilename := tmpFile.Name()

	// apply reduce function once to all values of the same key
	key_begin := 0
	for key_begin < len(kva) {
		key_end := key_begin + 1
		// finds all values with the same keys -- they are grouped together because the keys are sorted
		for key_end < len(kva) && kva[key_end].Key == kva[key_begin].Key {
			key_end++
		}
		values := []string{}
		for k := key_begin; k < key_end; k++ {
			values = append(values, kva[k].Value)
		}
		output := reducef(kva[key_begin].Key, values)

		// write output to temporary reduce file
		fmt.Fprintf(tmpFile, "%v %v\n", kva[key_begin].Key, output)

		// go to next key
		key_begin = key_end
	}

	// atomically rename reduce file to final reduce file
	finalizeReduceFile(tmpFilename, taskNum)
}

//
// main/mrworker.go calls this function.
//
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	// Your worker implementation here.
	for {
		args := GetTaskArgs{}
		reply := GetTaskReply{}
		// wait until get assigned a task
		call("Coordinator.HandleGetTask", &args, &reply)

		switch reply.TaskType {
		case Map:
			performMap(reply.MapFile, reply.TaskNum, reply.NReduceTasks, mapf)
		case Reduce:
			performReduce(reply.TaskNum, reply.NMapTasks, reducef)
		case Done:
			// no tasks remaining
			os.Exit(0)
		default:
			log.Fatal("Bad finished task? %s", reply.TaskType)
		}

		// tell coordinator that we're done
		finargs := FinishedTaskArgs{
			TaskType: reply.TaskType,
			TaskNum:  reply.TaskNum,
		}
		finreply := FinishedTaskReply{}
		call("Coordinator.HandleFinishedTask", &finargs, &finreply)
	}

	// uncomment to send the Example RPC to the coordinator.
	// CallExample()
}

//
// example function to show how to make an RPC call to the coordinator.
//
// the RPC argument and reply types are defined in rpc.go.
//
func CallExample() {

	// declare an argument structure.
	args := ExampleArgs{}

	// fill in the argument(s).
	args.X = 99

	// declare a reply structure.
	reply := ExampleReply{}

	// send the RPC request, wait for the reply.
	call("Coordinator.Example", &args, &reply)

	// reply.Y should be 100.
	fmt.Printf("reply.Y %v\n", reply.Y)
}

//
// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
//
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := coordinatorSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
