package pipeline

func NewPipelineController(eventBus EventBus, s Storage, idGen IDGen, machineIdentifier MachineIdentifier) PipelineController {
	return &pipelineController{EventBus: eventBus, Storage: s, IDGen: idGen, MachineIdentifier: machineIdentifier}
}

type PipelineController interface {
	EventBus
	Storage
	IDGen
	MachineIdentifier

	WithScope(scope string) PipelineController
	Scope() string
}

type pipelineController struct {
	scope string
	EventBus
	Storage
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
		EventBus:          EventBusWithPrefix(p.EventBus, scope),
		Storage:           StorageWithBasePath(p.Storage, scope),
	}
}
