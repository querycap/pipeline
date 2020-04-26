package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/querycap/pipeline/pkg/operator"
	"github.com/querycap/pipeline/pkg/pipeline"
	"github.com/querycap/pipeline/pkg/storage"
)

type FetchImage struct {
	URI string `json:"uri"`
}

func Handler(t operator.Transfer) error {
	input := &FetchImage{}

	for t.Scan() {
		if err := pipeline.ReadNext(t, func(r io.Reader) error {
			return json.NewDecoder(r).Decode(input)
		}); err != nil {
			return err
		}
	}

	resp, err := http.Get(input.URI)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return t.Put(storage.WithContentType(resp.Header.Get("Content-Type"))(storage.AsWriterTo(resp.Body)))
	}

	return fmt.Errorf("can't fetch from %s", input.URI)
}
