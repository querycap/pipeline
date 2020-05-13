package container

import (
	"context"
	"strings"

	"github.com/querycap/pipeline/spec"
)

func PodNameByScopeAndStage(scope string, stage string) string {
	return strings.Replace(scope+"--"+stage, ":", "__", -1)
}

type PodController interface {
	Apply(ctx context.Context, name string, container *Container) error
	Kill(ctx context.Context, name string) error
}

type Container struct {
	spec.Container
	Image       string
	Replicas    int
	Annotations map[string]string
}
