package s3

import (
	"bytes"
	"context"
	"io"

	"github.com/minio/minio-go/v6"
	"github.com/querycap/pipeline/pipeline"
)

func NewS3Storage(minio *minio.Client, bucket string) (pipeline.Storage, error) {
	exists, err := minio.BucketExists(bucket)
	if err != nil {
		return nil, err
	}

	if !exists {
		if err := minio.MakeBucket(bucket, ""); err != nil {
			return nil, err
		}
	}

	return &S3Storage{
		minio:  minio,
		bucket: bucket,
	}, nil
}

type S3Storage struct {
	minio  *minio.Client
	bucket string
}

func (f *S3Storage) Del(ctx context.Context, path string) error {
	return f.minio.RemoveObject(f.bucket, path)
}

func (f *S3Storage) Read(ctx context.Context, path string) (io.ReadCloser, error) {
	return f.minio.GetObjectWithContext(ctx, f.bucket, path, minio.GetObjectOptions{})
}

func (f *S3Storage) Put(ctx context.Context, path string, writerTo io.WriterTo) error {
	contentType := ""

	if contentTypeDescriber, ok := writerTo.(pipeline.ContentTypeDescriber); ok {
		contentType = contentTypeDescriber.ContentType()
	}

	buf := bytes.NewBuffer(nil)

	n, err := writerTo.WriteTo(buf)
	if err != nil {
		return err
	}

	if _, err = f.minio.PutObjectWithContext(ctx, f.bucket, path, buf, n, minio.PutObjectOptions{ContentType: contentType}); err != nil {
		return err
	}

	return nil
}
