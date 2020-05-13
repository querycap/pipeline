package pipeline

import (
	"os"

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
