package storage

import (
	"context"
	"io"
	"path/filepath"
)

type Storage interface {
	Read(ctx context.Context, path string) (io.ReadCloser, error)
	Put(ctx context.Context, path string, writerTo io.WriterTo) error
	Del(ctx context.Context, path string) error
}

type WithLen interface {
	Len() int
}

type ContentTypeDescriber interface {
	ContentType() string
}

func StorageWithBasePath(s Storage, basePath string) Storage {
	return &storageWithBathPath{
		basePath: basePath,
		s:        s,
	}
}

type storageWithBathPath struct {
	basePath string
	s        Storage
}

func (s *storageWithBathPath) Read(ctx context.Context, path string) (io.ReadCloser, error) {
	return s.s.Read(ctx, filepath.Join(s.basePath, path))
}

func (s *storageWithBathPath) Put(ctx context.Context, path string, writerTo io.WriterTo) error {
	return s.s.Put(ctx, filepath.Join(s.basePath, path), writerTo)
}

func (s *storageWithBathPath) Del(ctx context.Context, path string) error {
	return s.s.Del(ctx, filepath.Join(s.basePath, path))
}
