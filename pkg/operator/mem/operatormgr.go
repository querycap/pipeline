package mem

import (
	"fmt"
	"sync"

	"github.com/querycap/pipeline/pkg/eventbus"
	"github.com/querycap/pipeline/pkg/operator"
	"github.com/querycap/pipeline/pkg/pipeline"
	"github.com/querycap/pipeline/spec"
)

func NewMemOperatorMgr(pipelineController pipeline.PipelineController) *MemOperatorMgr {
	return &MemOperatorMgr{
		pipelineController: pipelineController,
	}
}

type MemOperatorMgr struct {
	pipelineController pipeline.PipelineController
	handlerFuncs       sync.Map
	instances          sync.Map
}

func (m *MemOperatorMgr) Register(ref operator.WithRefID, handlerFunc operator.OperatorHandlerFunc) error {
	m.handlerFuncs.Store(ref.RefID(), handlerFunc)
	return nil
}

func (m *MemOperatorMgr) Start(scope string, name string, step spec.Stage) error {
	v, ok := m.handlerFuncs.Load(step.Uses.RefID())
	if !ok {
		return fmt.Errorf("%s not found", step)
	}

	subscription := pipeline.ServeOperator(m.pipelineController.WithScope(scope), name, v.(operator.OperatorHandlerFunc))

	m.instances.Store(scope+"/"+name, subscription)
	return nil
}

func (m *MemOperatorMgr) Stop(scope string, name string) error {
	instanceID := scope + "/" + name

	v, ok := m.instances.Load(instanceID)
	if ok {
		v.(eventbus.Subscription).Unsubscribe()
		m.instances.Delete(instanceID)
	}

	return nil
}
