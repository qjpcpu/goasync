package goasync

// TaskHandler is user behave method
type TaskHandler func(Cb,...AsyncResult)

type Task struct {
	Dep     []string
	out     []string
	Handler TaskHandler
	visited bool
	name    string
	index   int // task index
}