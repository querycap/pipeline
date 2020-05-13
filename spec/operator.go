package spec

import "github.com/go-courier/semver"

type Project struct {
	Group       string         `json:"group" yaml:"group"`
	Name        string         `json:"name" yaml:"name"`
	Version     semver.Version `json:"version" yaml:"version"`
	Description string         `json:"description" yaml:"description"`
}

func (o Project) String() string {
	return o.RefID()
}

func (o Project) RefID() string {
	return o.Group + "/" + o.Name + ":" + o.Version.String()
}

type OperatorMeta struct {
	Inputs  DataSchema `json:"inputs" yaml:"inputs"`
	Outputs DataSchema `json:"outputs" yaml:"outputs"`
}

type Operator struct {
	Project      `json:"project" yaml:"project"`
	OperatorMeta `yaml:",inline"`
}
