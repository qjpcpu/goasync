package goasync

// AsyncResult store certain task's execution result
type AsyncResult struct {
	name  string
	err error
	data  interface{}
}

// Name return AsyncResult's name
func (ar AsyncResult) Name() (name string){
	return ar.name
}

// Data return AsyncResult's data of certain task
func (ar AsyncResult) Data() (data interface{},err error){
	return ar.data,ar.err
}
