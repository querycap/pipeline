package pipeline

import (
	"context"
	"io"
	"time"

	"github.com/sirupsen/logrus"
)

func WriteTo(writeTo func(w io.Writer) (n int64, err error)) io.WriterTo {
	return &writerTo{
		writeTo: writeTo,
	}
}

func AsWriterTo(r io.Reader) io.WriterTo {
	return &writerTo{
		writeTo: func(w io.Writer) (n int64, err error) {
			return io.Copy(w, r)
		},
	}
}

type writerTo struct {
	writeTo func(w io.Writer) (n int64, err error)
}

func (w2 *writerTo) WriteTo(w io.Writer) (n int64, err error) {
	return w2.writeTo(w)
}

func PutWithCost(s Storage, path string, w io.WriterTo) error {
	starts := time.Now()
	defer func() {
		logrus.Infof("write %s cost %s", path, time.Since(starts))
	}()

	return s.Put(context.Background(), path, w)
}

func WithContentType(contentType string) func(writerTo io.WriterTo) io.WriterTo {
	return func(writerTo io.WriterTo) io.WriterTo {
		return &writerToWithContentType{WriterTo: writerTo, contentType: contentType}
	}
}

type writerToWithContentType struct {
	io.WriterTo
	contentType string
}

func (w *writerToWithContentType) ContentType() string {
	return w.contentType
}

func ReadNext(r Receiver, readFrom ReadFrom) error {
	f, err := r.Next()
	if err != nil {
		return err
	}
	defer f.Close()

	if err := readFrom(f); err != nil {
		return err
	}

	return nil
}

func SendByReader(s Sender, input io.Reader) error {
	if err := s.Put(AsWriterTo(input)); err != nil {
		return err
	}
	if err := s.Send(); err != nil {
		return err
	}
	return nil
}
