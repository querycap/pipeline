package global

import (
	"github.com/docker/docker/client"
	"github.com/querycap/pipeline/pkg/operator/docker"
	"github.com/querycap/pipeline/pkg/pipeline"
	"github.com/querycap/pipeline/pkg/redisutil"
	"github.com/querycap/pipeline/util"
	"github.com/sirupsen/logrus"
)

var Pipeline *pipeline.Pipeline

func init() {
	logrus.SetLevel(logrus.DebugLevel)

	pipelineSpec, err := pipeline.PipelineFromYAML("./pipeline.yml")
	if err != nil {
		panic(err)
	}
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	PIPELINE_REDIS := util.FromEnv("PIPELINE_REDIS", "tcp://redis:6379")

	pool, err := redisutil.NewPool(PIPELINE_REDIS)
	if err != nil {
		panic(err)
	}

	mgr := docker.NewDockerOperatorMgr(cli, &docker.ContainerConfig{
		Envs: map[string]string{
			"PIPELINE_REDIS": PIPELINE_REDIS,
		},
		Links: map[string]string{
			"redis": "redis",
		},
	})

	pipelineController, err := util.NewRedisBasedController(pool)
	if err != nil {
		panic(err)
	}

	p, err := pipeline.NewPipelineMgr(mgr, pipelineController).NewPipeline(pipelineSpec)
	if err != nil {
		panic(err)
	}

	Pipeline = p
}
