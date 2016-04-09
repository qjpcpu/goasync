package goasync

import (
	"testing"
)

type MyStruct struct {
	name string
}

func TestAuto(t *testing.T) {
	graph := map[string]*Task{
		"b": &Task{
			Dep: []string{"a"},
			Handler: func(cb Cb, ar ...AsyncResult) {
				t.Log("task b")
				cb("b string", nil)
			},
		},
		"e": &Task{
			Dep: []string{"a", "b", "c"},
			Handler: func(cb Cb, ar ...AsyncResult) {
				var b string
				ar[1].Data(&b)
				t.Logf("task e got b's data:%s\n", b)
				var c MyStruct
				ar[2].Data(&c)
				t.Logf("task e got c's data:%+v\n", c)
				t.Log("task e done")
				tbl := map[string]MyStruct{
					"first": MyStruct{name: "inner"},
				}
				cb(tbl, nil)
			},
		},
		"f": &Task{
			Dep: []string{"e"},
			Handler: func(cb Cb, ar ...AsyncResult) {
				var tbl map[string]MyStruct
				ar[0].Data(&tbl)
				t.Logf("task f got e's data:+%v\n", tbl)
				cb(nil, nil)
			},
		},
		"c": &Task{
			Dep: []string{"a"},
			Handler: func(cb Cb, ar ...AsyncResult) {
				var data []string
				ar[0].Data(&data)
				t.Log("task c get a's data:", data)
				ms := &MyStruct{name: "from c"}
				cb(ms, nil)
			},
		},
		"a": &Task{
			Handler: func(cb Cb, ar ...AsyncResult) {
				t.Log("task a")
				d := []string{"bob", "foo"}
				cb(d, nil)
			},
		},
	}
	asy, _ := Auto(graph)
	asy.Run()
}

func TestParallel(t *testing.T) {
	asy, _ := Parallel(
		func(cb Cb, ar ...AsyncResult) {
			t.Log("aaa")
			cb("aaa", nil)
		},
		func(cb Cb, ar ...AsyncResult) {
			t.Log("bbb")
			cb("bbb", nil)
		},
	)
	asy.Run()
	names := asy.GetTaskNames()
	var s string
	asy.GetResults(names[1])[0].Data(&s)
	t.Log(s)
}
