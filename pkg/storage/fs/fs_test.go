package fs

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/querycap/pipeline/pkg/storage"
	"github.com/spf13/afero"
)

func TestFsStorage(t *testing.T) {
	s := NewFsStorage(afero.NewBasePathFs(afero.NewOsFs(), "/tmp"))

	for i := 0; i < 10; i++ {
		v := i

		t.Run("", func(t *testing.T) {
			filename := fmt.Sprintf("file/%d", v)

			input := []byte(filename)

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
