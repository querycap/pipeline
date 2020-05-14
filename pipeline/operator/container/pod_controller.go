package container

import (
	"context"
	"strings"

	"github.com/querycap/pipeline/spec"
)

func PodNameByScopeAndStage(scope string, stage string) string {
	parts := strings.Split(scope, "/")
	return "p" + parts[len(parts)-1] + "-" + stage
}

type PodController interface {
	Apply(ctx context.Context, name string, container *Container) error
	Kill(ctx context.Context, name string) error
}

type Container struct {
	spec.Container
	Image       string
	Replicas    int32
	Annotations map[string]string
}
