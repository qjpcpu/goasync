package goasync

import (
	"errors"
	"strconv"
)

// Async handler
type Async struct {
	results  map[string]*AsyncResult // task execution results
	taskList []*Task
	signals  chan AsyncResult
}

// Parallel generate an async handler for parallel execution
func Parallel(functions ...TaskHandler) (async *Async, err error) {
	if len(functions) == 0 {
		return nil, errors.New("No task handlers found!")
	}
	graph := make(map[string]*Task)
	for i, th := range functions {
		graph[strconv.Itoa(i)] = &Task{
			Handler: th,
		}
	}
	return Auto(graph)
}

// Auto generate an async handler and auto parse task flow
func Auto(graph map[string]*Task) (async *Async, err error) {
	// build DAG
	async = &Async{}
	async.dfsSort(graph)
	return
}

// Run async tasks
func (async *Async) Run() error {
	async.results = make(map[string]*AsyncResult)
	async.signals = make(chan AsyncResult)
	workerIndex, workerCnt := 0, 0
	schedule := func() {
		for _, t := range async.taskList {
			if t.index == workerIndex {
				workerCnt += 1
				go t.Handler(async.makeCb(t.name), async.GetResults(t.Dep...)...)
			}
		}
	}
	schedule()
	for {
		msg := <-async.signals
		async.results[msg.name] = &msg
		if msg.err != nil {
			return msg.err
		}
		if len(async.results) == len(async.taskList) {
			return nil
		}
		workerCnt -= 1
		if workerCnt == 0 {
			workerIndex += 1
			schedule()
			if workerCnt == 0 {
				return nil
			}
		}
	}
}

// GetResults fetch task execution results list by names
func (async *Async) GetResults(names ...string) (arr []AsyncResult) {
	if len(names) == 0 || len(async.results) == 0 {
		return
	}
	for _, name := range names {
		if val, ok := async.results[name]; ok {
			arr = append(arr, *val)
		}
	}
	return
}
