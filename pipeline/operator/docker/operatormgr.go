package docker

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/querycap/pipeline/spec"
	"github.com/sirupsen/logrus"
)

func NewDockerOperatorMgr(dockerClient *client.Client, commonContainerConfig *ContainerConfig) *DockerOperatorMgr {
	return &DockerOperatorMgr{
		c:                     NewDockerController(dockerClient),
		commonContainerConfig: commonContainerConfig,
	}
}

type DockerOperatorMgr struct {
	c                     *DockerController
	commonContainerConfig *ContainerConfig
}

func (d *DockerOperatorMgr) Start(scope string, stage string, step spec.Stage) error {
	fullImageRef := step.Uses.RefID()

	ref := strings.TrimPrefix(fullImageRef, "docker.io/library/")

	ctx := context.Background()

	if err := d.c.PullImageIfNotExists(ctx, ref); err != nil {
		return err
	}

	co := &ContainerOwner{
		Scope: scope,
		Stage: stage,
	}

	err := d.c.RunContainer(ctx, fullImageRef, MergeContainerConfig(d.commonContainerConfig, &ContainerConfig{
		ContainerOwner: *co,
		Envs: map[string]string{
			"PIPELINE_SCOPE": co.Scope,
			"PIPELINE_STAGE": co.Stage,
		},
	}))
	if err != nil {
		return err
	}

	return nil
}

func (d DockerOperatorMgr) Stop(scope string, stage string) error {
	co := &ContainerOwner{
		Scope: scope,
		Stage: stage,
	}

	ctx := context.Background()

	containerListFilters := filters.NewArgs()

	SetArgsFromContainerOwner(containerListFilters, co)

	list, err := d.c.ListRunningContainer(ctx, containerListFilters)
	if err != nil {
		return err
	}

	for i := range list {
		logrus.Debugf("kill %v", list[i].Labels)
		if err := d.c.KillContainer(ctx, list[i].ID); err != nil {
			return err
		}
	}

	return nil
}
