package util

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/querycap/pipeline/pkg/operator"
	"github.com/querycap/pipeline/pkg/pipeline"
	"github.com/querycap/pipeline/pkg/redisutil"
)

func ServeOperator(handler operator.OperatorHandlerFunc) {
	PIPELINE_REDIS := FromEnv("PIPELINE_REDIS", "tcp://redis:6379")
	PIPELINE_SCOPE := FromEnv("PIPELINE_SCOPE", "x")
	PIPELINE_STAGE := FromEnv("PIPELINE_STAGE", "x")

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)

	pool, err := redisutil.NewPool(PIPELINE_REDIS)
	if err != nil {
		panic(err)
	}

	pipelineController, err := NewRedisBasedController(pool)
	if err != nil {
		panic(err)
	}

	sub := pipeline.ServeOperator(pipelineController.WithScope(PIPELINE_SCOPE), PIPELINE_STAGE, handler)

	<-stopCh

	sub.Unsubscribe()
}

func FromEnv(k string, defaults string) string {
	v := os.Getenv(k)
	if v == "" {
		return defaults
	}
	return v
}
