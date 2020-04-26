package spec

import "github.com/go-courier/semver"

type DataSchema struct {
	ContentType string `json:"contentType" yaml:"contentType"`
	Schema      Schema `json:"schema,omitempty" yaml:"schema,omitempty"`
}

type OperatorMeta struct {
	Inputs  DataSchema `json:"inputs" yaml:"inputs"`
	Outputs DataSchema `json:"outputs" yaml:"outputs"`
}

type Operator struct {
	Name         string         `json:"name" yaml:"name"`
	Version      semver.Version `json:"version" yaml:"version"`
	OperatorMeta `yaml:",inline"`
}

func (o Operator) String() string {
	return o.RefID()
}

func (o Operator) RefID() string {
	return o.Name + ":" + o.Version.String()
}

type Pipeline struct {
	Name    string           `json:"name" yaml:"name"`
	Version semver.Version   `json:"version" yaml:"version"`
	Stages  map[string]Stage `json:"stages" yaml:"stages"`
	Starts  string           `json:"starts" yaml:"starts"`
	Ends    string           `json:"ends" yaml:"ends"`
}

func (o Pipeline) String() string {
	return o.RefID()
}

func (o Pipeline) RefID() string {
	return o.Name + ":" + o.Version.String()
}

type Stage struct {
	Deps       []string `json:"deps" yaml:"deps"`
	Uses       Ref      `json:"uses" yaml:"uses"`
	StepOption `yaml:",inline"`
}

type StepOption struct {
	Env  map[string]string `json:"env,omitempty" yaml:"env,omitempty"`
	Args []string          `json:"args,omitempty" yaml:"args,omitempty"`
}
