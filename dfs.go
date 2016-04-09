package goasync

//import (
// 	"fmt"
// )

func initGraph(graph map[string]*Task) {
	for name, t := range graph {
		t.name = name
		t.index = 0
		for _, dep := range t.Dep {
			graph[dep].out = append(graph[dep].out, name)
		}
	}
}
func (asy *Async) dfsSort(graph map[string]*Task) {
	initGraph(graph)
	for name := range graph {
		if !graph[name].visited {
			asy.visit(graph, name)
		}
	}
	// generate task index
	indexArr := make(map[string]int)
	for _, t := range asy.taskList {
		if len(t.Dep) == 0 {
			t.index = 0
			indexArr[t.name] = 0
			continue
		}
		max := 0
		for _, dep := range t.Dep {
			if val, ok := indexArr[dep]; ok && val > max {
				max = val
			}
		}
		indexArr[t.name] = max + 1
	}
	for name, id := range indexArr {
		graph[name].index = id
	}
}

func (asy *Async) visit(graph map[string]*Task, name string) {
	current := graph[name]
	current.visited = true
	current.name = name
	for _, upstream := range current.Dep {
		if !graph[upstream].visited {
			asy.visit(graph, upstream)
		}
	}
	asy.taskList = append(asy.taskList, current)
}
