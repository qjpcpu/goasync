package goasync

import (
	"testing"
	"fmt"
)

func TestParallel(t *testing.T) {
	graph := map[string]*Task{
		"b": &Task{
			Dep: []string{"a"},
			Handler: func(cb Cb,ar ...AsyncResult) {
				fmt.Println("task b")
				d := []string{"gods","tt"}
				cb(d, nil)
			},
		},
		"e": &Task{
			Dep: []string{"a", "b", "c"},
			Handler: func(cb Cb,ar ...AsyncResult) {
				fmt.Println("task e")
				cb("done", nil)
			},
		},
		"c": &Task{
			Dep: []string{"a"},
			Handler: func(cb Cb,ar ...AsyncResult) {
				data,_ := ar[0].Data()
				fmt.Println("task c get a's data:",data[0])
				cb("done", nil)
			},
		},
		"a": &Task{
			Handler: func(cb Cb,ar ...AsyncResult) {
				fmt.Println("task a")
				d := []string{"bob","foo"}
				cb(d, nil)
			},
		},
	}
	asy, _ := Auto(graph)
	asy.Run()
	fmt.Println("===OK", asy)
}
