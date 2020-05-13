package pipeline

type Result interface {
	Done() <-chan struct{}
	Err() error

	Receiver
}

func newResult() *result {
	return &result{
		done: make(chan struct{}, 1),
	}
}

type result struct {
	done chan struct{}
	err  error
	Receiver
}

func (r *result) Done() <-chan struct{} {
	return r.done
}

func (r *result) Err() error {
	return r.err
}

func (r *result) finish(receiver Receiver, err error) {
	if err != nil {
		r.err = err
	} else {
		r.Receiver = receiver
	}
	r.done <- struct{}{}
}
