package fs

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/querycap/pipeline/pipeline"
	"github.com/spf13/afero"
)

func NewFsStorage(fs afero.Fs) pipeline.Storage {
	return &FsStorage{
		fs: fs,
	}
}

type FsStorage struct {
	fs afero.Fs
}

func (f *FsStorage) Del(ctx context.Context, path string) error {
	return f.fs.Remove(path)
}

func (f *FsStorage) Read(ctx context.Context, path string) (io.ReadCloser, error) {
	return f.fs.Open(path)
}

func (f *FsStorage) Put(ctx context.Context, path string, writerTo io.WriterTo) error {
	dirname := filepath.Dir(path)

	if err := f.fs.MkdirAll(dirname, os.ModePerm); err != nil {
		return err
	}

	file, err := f.fs.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = writerTo.WriteTo(file)
	return err
}
