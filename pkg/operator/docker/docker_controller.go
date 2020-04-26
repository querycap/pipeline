package docker

import (
	"context"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/sirupsen/logrus"
)

func NewDockerController(dockerClient *client.Client) *DockerController {
	return &DockerController{
		client: dockerClient,
	}
}

type DockerController struct {
	client *client.Client
}

type ContainerOwner struct {
	Scope string
	Stage string
}

func SetArgsFromContainerOwner(f filters.Args, containerOwner *ContainerOwner) {
	f.Add("label", "pipeline/scope="+containerOwner.Scope)
	f.Add("label", "pipeline/stage="+containerOwner.Stage)
}

type ContainerConfig struct {
	ContainerOwner
	Envs    map[string]string
	Volumes map[string]string
	Links   map[string]string
}

func MergeContainerConfig(configs ...*ContainerConfig) *ContainerConfig {
	cc := &ContainerConfig{
		Envs:    map[string]string{},
		Volumes: map[string]string{},
		Links:   map[string]string{},
	}

	for i := range configs {
		c := configs[i]

		if c.ContainerOwner.Scope != "" {
			cc.ContainerOwner.Scope = c.ContainerOwner.Scope
		}

		if c.ContainerOwner.Stage != "" {
			cc.ContainerOwner.Stage = c.ContainerOwner.Stage
		}

		for k, v := range c.Envs {
			cc.Envs[k] = v
		}

		for k, v := range c.Volumes {
			cc.Volumes[k] = v
		}

		for k, v := range c.Links {
			cc.Links[k] = v
		}
	}

	return cc
}

func (c *DockerController) RunContainer(ctx context.Context, imageRef string, config *ContainerConfig) error {
	containerConfig := &container.Config{
		Image: imageRef,
	}
	hostConfig := &container.HostConfig{}

	setContainerConfig(containerConfig, config)
	setHostContainer(hostConfig, config)

	created, err := c.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, "")
	if err != nil {
		return err
	}

	if err := c.client.ContainerStart(ctx, created.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	go func() {
		if !logrus.IsLevelEnabled(logrus.DebugLevel) {
			return
		}

		// clone logs
		r, err := c.client.ContainerLogs(ctx, created.ID, types.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     true,
		})
		if err != nil {
			return
		}
		defer r.Close()

		if _, err := stdcopy.StdCopy(os.Stdout, os.Stderr, r); err != nil && err != io.EOF {
			logrus.Fatal(err)
		}
	}()

	return nil
}

func setContainerConfig(c *container.Config, config *ContainerConfig) {
	if c.Labels == nil {
		c.Labels = map[string]string{}
	}

	c.Labels["pipeline/scope"] = config.Scope
	c.Labels["pipeline/stage"] = config.Stage

	for k := range config.Envs {
		c.Env = append(c.Env, k+"="+config.Envs[k])
	}
}

func setHostContainer(hostConfig *container.HostConfig, config *ContainerConfig) {
	for from, to := range config.Links {
		hostConfig.Links = append(hostConfig.Links, from+":"+to)
	}

	for from, to := range config.Volumes {
		hostConfig.Mounts = append(hostConfig.Mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: from,
			Target: to,
		})
	}
}

func (c *DockerController) ListRunningContainer(ctx context.Context, args filters.Args) ([]types.Container, error) {
	return c.client.ContainerList(ctx, types.ContainerListOptions{
		Filters: args,
		All:     false,
	})
}

func (c *DockerController) KillContainer(ctx context.Context, containerID string) error {
	return c.client.ContainerKill(ctx, containerID, "SIGKILL")
}

func (c *DockerController) PullImageIfNotExists(ctx context.Context, imageRef string) error {
	imageListFilters := filters.NewArgs()
	imageListFilters.Add("reference", imageRef)

	images, err := c.client.ImageList(ctx, types.ImageListOptions{Filters: imageListFilters})
	if err != nil {
		return err
	}

	imageExists := len(images) > 0

	if !imageExists {
		if _, err = c.client.ImagePull(ctx, imageRef, types.ImagePullOptions{}); err != nil {
			return err
		}
	}

	return nil
}
