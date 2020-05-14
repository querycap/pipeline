package container

import (
	"context"
	"io"
	"math/rand"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/sirupsen/logrus"
)

func NewDockerPodController(imageRegistry *ImageRegistry) (*DockerPodController, error) {
	c, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	return &DockerPodController{
		client:        c,
		imageRegistry: imageRegistry,
	}, nil
}

type DockerPodController struct {
	client        *client.Client
	imageRegistry *ImageRegistry
}

func (c *DockerPodController) Apply(ctx context.Context, name string, container *Container) error {
	imageRef := c.imageRegistry.Fix(container.Image)

	if err := c.pullImageIfNotExists(ctx, imageRef); err != nil {
		return err
	}

	container.Annotations = map[string]string{
		"pipeline": name,
	}

	list, err := c.listMatchedRunningContainer(ctx, name)
	if err != nil {
		return err
	}

	current := len(list)

	offset := current - int(container.Replicas)

	// scale up
	if offset < 0 {
		for i := 0; i < -offset; i++ {
			if err := c.runContainer(ctx, imageRef, container); err != nil {
				return err
			}
		}
		return nil
	}

	// scale down
	if offset > 0 {
		rand.Shuffle(len(list), func(i, j int) { list[i], list[j] = list[j], list[i] })

		for _, item := range list[0:offset] {
			if err := c.killContainer(ctx, item.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *DockerPodController) Kill(ctx context.Context, name string) error {
	list, err := c.listMatchedRunningContainer(ctx, name)
	if err != nil {
		return err
	}

	for i := range list {
		if err := c.killContainer(ctx, list[i].ID); err != nil {
			return err
		}
	}

	return nil
}

func (c *DockerPodController) listMatchedRunningContainer(ctx context.Context, name string) ([]types.Container, error) {
	containerListFilters := filters.NewArgs()
	containerListFilters.Add("label", "pipeline="+name)

	return c.listRunningContainer(ctx, containerListFilters)
}

func (c *DockerPodController) runContainer(ctx context.Context, image string, cc *Container) error {
	logrus.WithContext(ctx).Debugf("running from %s", image)

	containerConfig := &container.Config{
		Image:      image,
		Labels:     cc.Annotations,
		Entrypoint: cc.Command,
		Cmd:        cc.Args,
	}

	for k := range cc.Envs {
		containerConfig.Env = append(containerConfig.Env, k+"="+cc.Envs[k])
	}

	created, err := c.client.ContainerCreate(ctx, containerConfig, nil, nil, "")
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

func (c *DockerPodController) listRunningContainer(ctx context.Context, args filters.Args) ([]types.Container, error) {
	return c.client.ContainerList(ctx, types.ContainerListOptions{
		Filters: args,
		All:     false,
	})
}

func (c *DockerPodController) killContainer(ctx context.Context, containerID string) error {
	logrus.WithContext(ctx).Debugf("killing %s", containerID)

	return c.client.ContainerKill(ctx, containerID, "SIGKILL")
}

func (c *DockerPodController) pullImageIfNotExists(ctx context.Context, image string) error {
	imageListFilters := filters.NewArgs()
	imageListFilters.Add("reference", strings.Replace(image, "docker.io/library/", "", 1))

	images, err := c.client.ImageList(ctx, types.ImageListOptions{Filters: imageListFilters})
	if err != nil {
		return err
	}

	imageExists := len(images) > 0

	if !imageExists {
		if _, err = c.client.ImagePull(ctx, image, types.ImagePullOptions{
			RegistryAuth: c.imageRegistry.RegistryAuth(),
		}); err != nil {
			return err
		}
	}

	return nil
}
