package docker

import (
	"testing"

	"github.com/docker/docker/client"
	"github.com/go-courier/semver"
	"github.com/querycap/pipeline/spec"
	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func Test(t *testing.T) {
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	mgr := NewDockerOperatorMgr(cli, &ContainerConfig{})

	s := spec.Stage{}
	s.Uses.Name = "docker.io/library/nginx"
	s.Uses.Version = *semver.MustParseVersion("1.17.10")

	scope := "xxx"
	stage := "fetch-image"

	t.Run("start", func(t *testing.T) {
		if err := mgr.Start(scope, stage, s); err != nil {
			panic(err)
		}
	})

	t.Run("stop", func(t *testing.T) {
		if err := mgr.Stop(scope, stage); err != nil {
			panic(err)
		}
	})
}
