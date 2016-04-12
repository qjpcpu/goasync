package goasync

import (
	"errors"
)

type visitState int

const (
	unvisited = iota
	visiting
	visited
)

func dfsScan(graph map[string]*Task) error {
	visitList := make(map[string]visitState)
	for name, t := range graph {
		if len(t.out) > 0 {
			continue
		}
		if visitList[name] == unvisited {
			if err := visit(graph, name, visitList); err != nil {
				return err
			}
		}
	}

	if len(visitList) != len(graph) {
		return errors.New("goasync: circle dependency detected,isolated task circle.")
	}
	return nil
}

func visit(graph map[string]*Task, name string, visitList map[string]visitState) error {
	current := graph[name]
	visitList[name] = visiting
	for _, upstream := range current.Dep {
		if visitList[upstream] == unvisited {
			if err := visit(graph, upstream, visitList); err != nil {
				return err
			}
		} else if visitList[upstream] == visiting {
			return errors.New("goasync: circle dependency detected.")
		}
	}
	visitList[name] = visited
	return nil
}
