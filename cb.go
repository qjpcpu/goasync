package goasync

// Cb is task callback, must be called in user's TaskHandler.
type Cb func(interface{}, error)

func (asy *Async) makeCb(taskName string) Cb {
	return func(data interface{}, err error) {
		ar := Result{
			name: taskName,
		}
		if err != nil {
			ar.err = err
			ar.data = nil
		} else {
			ar.err = nil
			ar.data = data
		}
		asy.signals <- ar
	}
}
