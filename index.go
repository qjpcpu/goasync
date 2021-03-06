package goasync

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"
)

type ResultSet map[string]*Result

// Async handler.
type Async struct {
	results map[string]*Result // task execution results
	tasks   map[string]*Task
	signals chan Result
	Timeout time.Duration
	Debug   bool // print schedule info
}

// Get result by name
func (rs ResultSet) Get(taskName string) (ar *Result) {
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
		return nil, errors.New("goasync: no task handlers found!")
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
	if err = dfsScan(graph); err != nil {
		return
	}
	return
}

// Run async tasks.
func (async *Async) Run() error {
	// create results for task result storage
	async.results = make(map[string]*Result)
	async.signals = make(chan Result, 1)
	// set default timeout to 10 minutes
	if async.Timeout < time.Millisecond*1 {
		async.SetTimeout(time.Minute * 10)
	}
	schedule := func(ar *Result) error {
		wt, err := async.waitingTasks(ar)
		if err != nil {
			return err
		}
		// all done
		if len(wt) == 0 {
			return nil
		}
		if async.Debug {
			info := "[goasync]\tSchedule tasks: "
			for _, t := range wt {
				info += t.name + " "
			}
			log.Println(info)
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
			if async.Debug {
				log.Printf("[goasync]\t Got [%+v] from %s\n", msg, msg.name)
			}
			if _, exists := async.results[msg.name]; exists {
				if async.Debug {
					log.Printf("[goasync]\tCallback invoked multiple times in %s,exit!\n", msg.name)
				}
				return errors.New("goasync: callback invoked multiple times!")
			}
			// store task result
			async.results[msg.name] = &msg
			// tag task state as done
			async.tasks[msg.name].done = true
			// abort when error happens
			if msg.err != nil {
				if async.Debug {
					log.Printf("[goasync]\tError ocurrs in %s,exit!\n", msg.name)
				}
				return msg.err
			}
			if async.Debug {
				log.Printf("[goasync]\t%s finished.\n", msg.name)
			}
			// goasync thinks all tasks are done if get all results
			if len(async.tasks) == len(async.results) {
				if async.Debug {
					log.Printf("[goasync]\tAll goasync tasks done.\n")
				}
				return nil
			}
			if err := schedule(&msg); err != nil {
				return err
			}
		case <-time.After(async.Timeout):
			return errors.New("goasync: task timeout!")
		}
	}
}

func (async *Async) waitingTasks(ar *Result) ([]*Task, error) {
	var waiting []*Task
	// first schedule
	if ar == nil {
		for _, t := range async.tasks {
			if len(t.Dep) == 0 {
				waiting = append(waiting, t)
			}
		}
		if len(waiting) == 0 {
			return nil, errors.New("goasync: can't find any task to schedule!")
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
			if tmp.done {
				return nil, errors.New(fmt.Sprintf("goasync: fail to schedule task[%s], maybe there's circle dependency", tmp.name))
			}
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

// GetResult fetch task execution result by name.
func (async *Async) GetResult(name string) (ar *Result) {
	if val, ok := async.results[name]; ok {
		ar = val
	}
	return
}
