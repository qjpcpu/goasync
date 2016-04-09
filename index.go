package goasync

import (
	"errors"
	"strconv"
	"time"
)

type ResultSet map[string]*AsyncResult

// Async handler.
type Async struct {
	results map[string]*AsyncResult // task execution results
	tasks   map[string]*Task
	signals chan AsyncResult
	Timeout time.Duration
}

// Get result by name
func (rs ResultSet) Get(taskName string) (ar *AsyncResult) {
	return rs[taskName]
}

func (asy *Async) init(graph map[string]*Task) {
	for name, t := range graph {
		t.name = name
		t.done = false
		for _, dep := range t.Dep {
			graph[dep].out = append(graph[dep].out, name)
		}
	}
	asy.tasks = graph
}

// Parallel generate an async handler for parallel execution.
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

// Auto generate an async handler and auto parse task flow.
func Auto(graph map[string]*Task) (async *Async, err error) {
	// build DAG
	async = &Async{}
	async.init(graph)
	return
}

// Run async tasks.
func (async *Async) Run() error {
	// create results for task result storage
	async.results = make(map[string]*AsyncResult)
	async.signals = make(chan AsyncResult, 1)
	// set default timeout to 10 minutes
	if async.Timeout < time.Millisecond*1 {
		async.SetTimeout(time.Minute * 10)
	}
	schedule := func(ar *AsyncResult) error {
		wt, err := async.waitingTasks(ar)
		if err != nil {
			return err
		}
		// all done
		if len(wt) == 0 {
			return nil
		}
		for _, t := range wt {
			go t.Handler(async.makeCb(t.name), async.GetResults(t.Dep...))
		}
		return nil
	}
	if err := schedule(nil); err != nil {
		return err
	}
	for {
		select {
		case msg := <-async.signals:
			// store task result
			async.results[msg.name] = &msg
			// tag task state as done
			async.tasks[msg.name].done = true
			// abort when error happens
			if msg.err != nil {
				return msg.err
			}
			// goasync thinks all tasks are done if get all results
			if len(async.tasks) == len(async.results) {
				return nil
			}
			if err := schedule(&msg); err != nil {
				return err
			}
		case <-time.After(async.Timeout):
			return errors.New("async task timeout!")
		}
	}
}

func (async *Async) waitingTasks(ar *AsyncResult) ([]*Task, error) {
	var waiting []*Task
	// first schedule
	if ar == nil {
		for _, t := range async.tasks {
			if len(t.Dep) == 0 {
				waiting = append(waiting, t)
			}
		}
		if len(waiting) == 0 {
			return nil, errors.New("Can't find any task to schedule!")
		} else {
			return waiting, nil
		}
	}
	if ar.err != nil {
		return nil, ar.err
	}
	current := async.tasks[ar.name]
	for _, down := range current.out {
		tmp := async.tasks[down]
		ok := true
		for _, dep := range tmp.Dep {
			if !async.tasks[dep].done {
				ok = false
				break
			}
		}
		if ok {
			waiting = append(waiting, tmp)
		}
	}
	return waiting, nil
}

// SetTimeout of async tasks, default is 10 minutes.
func (async *Async) SetTimeout(duration time.Duration) {
	async.Timeout = duration
}

// GetTaskNames return all task names.
func (async *Async) GetTaskNames() (names []string) {
	for n, _ := range async.results {
		names = append(names, n)
	}
	return
}

// GetResults fetch task execution results list by names.
func (async *Async) GetResults(names ...string) (rs ResultSet) {
	rs = make(ResultSet)
	if len(names) == 0 || len(async.results) == 0 {
		return
	}
	for _, name := range names {
		if val, ok := async.results[name]; ok {
			rs[name] = val
		}
	}
	return
}
