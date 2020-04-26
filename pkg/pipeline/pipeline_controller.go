package pipeline

import (
	"github.com/querycap/pipeline/pkg/eventbus"
	"github.com/querycap/pipeline/pkg/storage"
)

func NewPipelineController(eventBus eventbus.EventBus, s storage.Storage, idGen IDGen, machineIdentifier MachineIdentifier) PipelineController {
	return &pipelineController{EventBus: eventBus, Storage: s, IDGen: idGen, MachineIdentifier: machineIdentifier}
}

type PipelineController interface {
	eventbus.EventBus
	storage.Storage
	IDGen
	MachineIdentifier

	WithScope(scope string) PipelineController
	Scope() string
}

type pipelineController struct {
	scope string
	eventbus.EventBus
	storage.Storage
	IDGen
	MachineIdentifier
}

func (p *pipelineController) Scope() string {
	return p.scope
}

func (p *pipelineController) WithScope(scope string) PipelineController {
	return &pipelineController{
		scope:             scope,
		IDGen:             p.IDGen,
		MachineIdentifier: p.MachineIdentifier,
		EventBus:          eventbus.EventBusWithPrefix(p.EventBus, scope),
		Storage:           storage.StorageWithBasePath(p.Storage, scope),
	}
}
