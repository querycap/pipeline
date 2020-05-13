package container

import (
	"context"

	"github.com/querycap/pipeline/pipeline"
	"github.com/querycap/pipeline/spec"
)

func NewOperatorMgr(podController PodController, container *Container) pipeline.OperatorMgr {
	return &operatorMgr{
		podController: podController,
		container:     container,
	}
}

type operatorMgr struct {
	podController PodController
	container     *Container
}

func (d *operatorMgr) Up(scope string, stage string, step spec.Stage, replicas int) error {
	c := Container{
		Container: step.Container,
		Image:     step.Uses.RefID(),
		Replicas:  replicas,
	}

	c.Envs = c.Envs.Merge(d.container.Envs)

	c.Envs["PIPELINE_SCOPE"] = scope
	c.Envs["PIPELINE_STAGE"] = stage

	return d.podController.Apply(context.Background(), PodNameByScopeAndStage(scope, stage), &c)
}

func (d operatorMgr) Destroy(scope string, stage string) error {
	return d.podController.Kill(context.Background(), PodNameByScopeAndStage(scope, stage))
}
