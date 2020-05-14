package pipeline

import (
	"github.com/querycap/pipeline/spec"
)

type OperatorMgr interface {
	Up(scope string, name string, step spec.Stage, replicas int32) error
	Destroy(scope string, name string) error
}
