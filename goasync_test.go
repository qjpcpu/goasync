package goasync

import (
	"errors"
	"testing"
)

type MyStruct struct {
	name string
}

func TestAuto(t *testing.T) {
	graph := map[string]*Task{
		"b": &Task{
			Dep: []string{"a"},
			Handler: func(cb Cb, ar ResultSet) {
				t.Log("task b")
				cb("b string", nil)
			},
		},
		"e": &Task{
			Dep: []string{"a", "b", "c"},
			Handler: func(cb Cb, ar ResultSet) {
				var b string
				ar["b"].Data(&b)
				if b != "b string" {
					t.Error("should be 'b string'")
				}
				var c MyStruct
				ar["c"].Data(&c)
				if c.name != "from c" {
					t.Error("should be 'from c'")
				}
				tbl := map[string]MyStruct{
					"first": MyStruct{name: "inner"},
				}
				cb(tbl, nil)
			},
		},
		"f": &Task{
			Dep: []string{"e"},
			Handler: func(cb Cb, ar ResultSet) {
				var tbl map[string]MyStruct
				ar["e"].Data(&tbl)
				if tbl["first"].name != "inner" {
					t.Error("should be 'inner'")
				}
				cb(nil, nil)
			},
		},
		"c": &Task{
			Dep: []string{"a"},
			Handler: func(cb Cb, ar ResultSet) {
				var data []string
				ar["a"].Data(&data)
				t.Log("task c get a's data:", data)
				ms := &MyStruct{name: "from c"}
				cb(ms, nil)
			},
		},
		"a": &Task{
			Handler: func(cb Cb, ar ResultSet) {
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
		func(cb Cb, ar ResultSet) {
			t.Log("aaa")
			cb(0, nil)
		},
		func(cb Cb, ar ResultSet) {
			t.Log("bbb")
			cb("", nil)
		},
	)
	err := asy.Run()
	if err != nil {
		t.Error("should no error")
	}
	var s int = 2
	err = asy.GetResult("1").Data(&s)
	if err == nil {
		t.Error("should an error when get data")
	}
	var str string = "xyz"
	asy.GetResult("1").Data(&str)
	if str != "" {
		t.Error("should be empty")
	}
}
func TestAutoErr(t *testing.T) {
	graph := map[string]*Task{
		"b": &Task{
			Dep: []string{"a"},
			Handler: func(cb Cb, ar ResultSet) {
				t.Log("task b")
				cb("b string", nil)
			},
		},
		"a": &Task{
			Handler: func(cb Cb, ar ResultSet) {
				t.Log("task a")
				d := []string{"bob", "foo"}
				cb(d, errors.New("error happens in a"))
			},
		},
	}
	asy, _ := Auto(graph)
	err := asy.Run()
	if err == nil {
		t.Error("should get an error")
	}
	ar := asy.GetResult("a")
	if ar.err == nil {
		t.Error("should be error")
	}
}

func TestAutoSequence(t *testing.T) {
	graph := map[string]*Task{
		"a": &Task{
			Handler: func(cb Cb, ar ResultSet) {
				cb(1, nil)
			},
		},
		"b": &Task{
			Handler: func(cb Cb, ar ResultSet) {
				cb(2, nil)
			},
		},
		"c": &Task{
			Dep: []string{"a", "b"},
			Handler: func(cb Cb, ar ResultSet) {
				var a int
				ar.Get("a").Data(&a)
				var b int
				ar.Get("b").Data(&b)
				if a != 1 {
					t.Error("task a result should be 1")
				}
				if b != 2 {
					t.Error("task b result should be 2")
				}
				cb(a+b, nil)
			},
		},
		"d": &Task{
			Dep: []string{"c"},
			Handler: func(cb Cb, ar ResultSet) {
				var c int
				ar.Get("c").Data(&c)
				if c != 3 {
					t.Error("task c result should be 3")
				}
				cb(nil, nil)
			},
		},
		"e": &Task{
			Dep: []string{"c"},
			Handler: func(cb Cb, ar ResultSet) {
				var c int
				ar.Get("c").Data(&c)
				if c != 3 {
					t.Error("task c result should be 3")
				}
				cb(nil, nil)
			},
		},
	}
	asy, _ := Auto(graph)
	err := asy.Run()
	if err != nil {
		t.Error("should get no error")
	}
}

func TestAutoErr1(t *testing.T) {
	graph := map[string]*Task{
		"a": &Task{
			Handler: func(cb Cb, ar ResultSet) {
				cb("text", nil)
			},
		},
		"b": &Task{
			Dep: []string{"a"},
			Handler: func(cb Cb, ar ResultSet) {
				cb(nil, nil)
			},
		},
	}
	asy, _ := Auto(graph)
	err := asy.Run()
	if err != nil {
		t.Error("should get no error")
	}
	ar := asy.GetResult("a")
	var g int
	if err = ar.Data(g); err == nil {
		t.Error("should be an error")
	}
}
