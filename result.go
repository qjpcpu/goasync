package goasync

import (
	"errors"
	"fmt"
	"reflect"
)

// Result store certain task's execution result.
type Result struct {
	name string
	err  error
	data interface{}
}

// Name return Result's name.
func (ar Result) Name() (name string) {
	return ar.name
}

// Data return Result's data of certain task.
func (ar Result) Data(data interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%v", r))
		}
	}()
	if ar.err != nil {
		return ar.err
	}
	vfrom := reflect.ValueOf(ar.data)
	if !vfrom.IsValid() {
		return nil
	}
	if vfrom.Kind() == reflect.Ptr && vfrom.IsNil() {
		return nil
	}
	vto := reflect.ValueOf(data)
	if vto.Kind() != reflect.Ptr {
		return errors.New("goasync: the parameter passed in must be pointer type")
	}
	pfrom := reflect.Indirect(vfrom)
	pto := reflect.Indirect(vto)
	if !pto.IsValid() {
		return errors.New("goasync: the parameter passed in is not valid point")
	}
	pto.Set(pfrom)
	return nil
}
