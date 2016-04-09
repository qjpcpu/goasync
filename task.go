package goasync

// TaskHandler is user behave method.
type TaskHandler func(Cb, ResultSet)

type Task struct {
	Dep     []string
	out     []string
	Handler TaskHandler
	done    bool
	name    string
}
