package spec

import "github.com/go-courier/semver"

type DataSchema struct {
	ContentType string `json:"contentType" yaml:"contentType"`
	Schema      Schema `json:"schema,omitempty" yaml:"schema,omitempty"`
}

type Pipeline struct {
	Name    string         `json:"name" yaml:"name"`
	Version semver.Version `json:"version" yaml:"version"`

	PipelineFlow `yaml:",inline"`
}

type PipelineFlow struct {
	Stages map[string]Stage `json:"stages" yaml:"stages"`
	Starts string           `json:"starts" yaml:"starts"`
	Ends   string           `json:"ends" yaml:"ends"`
}

func (o Pipeline) String() string {
	return o.RefID()
}

func (o Pipeline) RefID() string {
	return o.Name + ":" + o.Version.String()
}

type Stage struct {
	Deps      []string `json:"deps" yaml:"deps"`
	Uses      Ref      `json:"uses" yaml:"uses"`
	Container `yaml:",inline"`
}

type Container struct {
	Command []string `json:"command,omitempty" yaml:"command,omitempty"`
	Args    []string `json:"args,omitempty" yaml:"args,omitempty"`
	Envs    Envs     `json:"envs,omitempty" yaml:"envs,omitempty"`
}

type Envs map[string]string

func (envs Envs) Merge(otherEnvs Envs) Envs {
	es := Envs{}
	for k, v := range envs {
		es[k] = v
	}
	for k, v := range otherEnvs {
		es[k] = v
	}
	return es
}
