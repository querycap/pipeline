package s3

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/querycap/pipeline/pkg/s3util"
	"github.com/querycap/pipeline/pkg/storage"
)

func TestS3Storage(t *testing.T) {
	s3, _ := s3util.NewS3("s3://minioadmin:minioadmin@127.0.0.1:9000")

	s, _ := NewS3Storage(s3, "tmp")

	for i := 0; i < 100; i++ {
		filename := fmt.Sprintf("file/%d", i)

		t.Run(filename, func(t *testing.T) {
			input := []byte(strings.Repeat(filename, 10000))

			err := storage.PutWithCost(s, filename, bytes.NewBuffer(input))
			NewWithT(t).Expect(err).To(BeNil())

			r, err := s.Read(context.Background(), filename)
			NewWithT(t).Expect(err).To(BeNil())

			defer r.Close()
			data, _ := ioutil.ReadAll(r)

			NewWithT(t).Expect(data).To(Equal(input))
		})
	}

}
