package pipeline

import (
	"io"
	"os"

	"github.com/querycap/pipeline/pkg/operator"
	"github.com/querycap/pipeline/pkg/storage"
	"github.com/querycap/pipeline/spec"
	"gopkg.in/yaml.v2"
)

func UnmarshalYAMLFile(filename string, out interface{}) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return yaml.NewDecoder(f).Decode(out)
}

func PipelineFromYAML(filename string) (*spec.Pipeline, error) {
	p := spec.Pipeline{}

	if err := UnmarshalYAMLFile(filename, &p); err != nil {
		return nil, err
	}

	return &p, nil
}

func OperatorFromYAML(filename string) (*spec.Operator, error) {
	op := spec.Operator{}

	if err := UnmarshalYAMLFile(filename, &op); err != nil {
		return nil, err
	}

	return &op, nil
}

func ReadNext(r operator.Receiver, readFrom operator.ReadFrom) error {
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

func SendByReader(s operator.Sender, input io.Reader) error {
	if err := s.Put(storage.AsWriterTo(input)); err != nil {
		return err
	}
	if err := s.Send(); err != nil {
		return err
	}
	return nil
}
