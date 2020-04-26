package operator

import (
	"context"
	"io"
)

type ReadFrom = func(r io.Reader) error

type Receiver interface {
	Scan() bool
	Next() (io.ReadCloser, error)
}

type Sender interface {
	Put(writerTo io.WriterTo) error
	Send() error
}

type Transfer interface {
	Receiver
	Sender
	Context() context.Context
}

type OperatorHandlerFunc = func(d Transfer) error
