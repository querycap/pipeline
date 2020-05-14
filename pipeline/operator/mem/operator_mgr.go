package mem

import (
	"fmt"
	"sync"

	"github.com/querycap/pipeline/pipeline"
	"github.com/querycap/pipeline/spec"
)

func NewMemOperatorMgr(pipelineController pipeline.PipelineController) *MemOperatorMgr {
	return &MemOperatorMgr{
		pipelineController: pipelineController,
	}
}

var _ pipeline.OperatorMgr = (*MemOperatorMgr)(nil)

type MemOperatorMgr struct {
	pipelineController pipeline.PipelineController
	handlerFuncs       sync.Map
	instances          sync.Map
}

func (m *MemOperatorMgr) Register(ref pipeline.WithRefID, handlerFunc pipeline.OperatorHandlerFunc) error {
	m.handlerFuncs.Store(ref.RefID(), handlerFunc)
	return nil
}

func (m *MemOperatorMgr) Up(scope string, name string, step spec.Stage, replicas int32) error {
	v, ok := m.handlerFuncs.Load(step.Uses.RefID())
	if !ok {
		return fmt.Errorf("%s not found", step)
	}

	subscription := pipeline.ServeOperator(m.pipelineController.WithScope(scope), name, v.(pipeline.OperatorHandlerFunc))

	m.instances.Store(scope+"/"+name, subscription)
	return nil
}

func (m *MemOperatorMgr) Destroy(scope string, name string) error {
	instanceID := scope + "/" + name

	v, ok := m.instances.Load(instanceID)
	if ok {
		v.(pipeline.Subscription).Unsubscribe()
		m.instances.Delete(instanceID)
	}

	return nil
}
