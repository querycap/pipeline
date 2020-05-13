package container

import (
	"testing"
	"time"

	"github.com/go-courier/semver"
	"github.com/querycap/pipeline/spec"
	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

var scope, stage = "xxx", "fetch-image"
var imageRegistry, _ = ParseImageRegistry("registry://docker.io/library/")

func Test(t *testing.T) {
	t.Run("docker", func(t *testing.T) {
		c, _ := NewDockerPodController(imageRegistry)
		RunAll(t, c)
	})
}

func RunAll(t *testing.T, c PodController) {
	s := spec.Stage{}
	s.Uses.Name = "nginx"
	s.Uses.Version = *semver.MustParseVersion("1.17.10")

	mgr := NewOperatorMgr(c, &Container{})

	t.Run("start", func(t *testing.T) {
		if err := mgr.Up(scope, stage, s, 3); err != nil {
			panic(err)
		}

		time.Sleep(1 * time.Second)

		if err := mgr.Up(scope, stage, s, 5); err != nil {
			panic(err)
		}

		time.Sleep(1 * time.Second)

		if err := mgr.Up(scope, stage, s, 3); err != nil {
			panic(err)
		}

		time.Sleep(1 * time.Second)

		if err := mgr.Up(scope, stage, s, 0); err != nil {
			panic(err)
		}

		time.Sleep(1 * time.Second)
	})

	t.Run("stop", func(t *testing.T) {
		if err := mgr.Destroy(scope, stage); err != nil {
			panic(err)
		}
	})
}
