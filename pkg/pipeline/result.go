package pipeline

import "github.com/querycap/pipeline/pkg/operator"

type Result interface {
	Done() <-chan struct{}
	Err() error

	operator.Receiver
}

func newResult() *result {
	return &result{
		done: make(chan struct{}, 1),
	}
}

type result struct {
	done chan struct{}
	err  error
	operator.Receiver
}

func (r *result) Done() <-chan struct{} {
	return r.done
}

func (r *result) Err() error {
	return r.err
}

func (r *result) finish(receiver operator.Receiver, err error) {
	if err != nil {
		r.err = err
	} else {
		r.Receiver = receiver
	}
	r.done <- struct{}{}
}
