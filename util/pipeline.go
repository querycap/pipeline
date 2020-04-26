package util

import (
	"time"

	"github.com/go-courier/snowflakeid"
	"github.com/go-courier/snowflakeid/workeridutil"
	redisEnvBus "github.com/querycap/pipeline/pkg/eventbus/redis"
	"github.com/querycap/pipeline/pkg/pipeline"
	"github.com/querycap/pipeline/pkg/storage/fs"
	"github.com/spf13/afero"
)

var startTime, _ = time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")
var sff = snowflakeid.NewSnowflakeFactory(16, 8, 5, startTime)

func NewRedisBasedController(pool redisEnvBus.RedisPool) (pipeline.PipelineController, error) {
	ip := workeridutil.ResolveLocalIP()
	idGen, err := sff.NewSnowflake(workeridutil.WorkerIDFromIP(workeridutil.ResolveLocalIP()))
	if err != nil {
		return nil, err
	}

	machineIdentifier := &machineID{
		ip: ip.String(),
	}

	s := NewStorageWithIPMachineID(fs.NewFsStorage(afero.NewBasePathFs(afero.NewOsFs(), "/tmp")), machineIdentifier)

	go func() {
		s.Serve()
	}()

	pipelineController := pipeline.NewPipelineController(
		redisEnvBus.NewRedisEventBus(pool),
		s,
		idGen,
		machineIdentifier,
	)

	return pipelineController, nil
}

type machineID struct {
	ip string
}

func (m *machineID) MachineID() (string, error) {
	return m.ip, nil
}
