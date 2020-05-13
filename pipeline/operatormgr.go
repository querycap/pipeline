package pipeline

import (
	"github.com/querycap/pipeline/spec"
)

type OperatorMgr interface {
	Start(scope string, name string, step spec.Stage) error
	Stop(scope string, name string) error
}
