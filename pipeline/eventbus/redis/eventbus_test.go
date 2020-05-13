package redis_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/querycap/pipeline/pipeline/eventbus/redis"
	"github.com/querycap/pipeline/pkg/redisutil"
	"github.com/sirupsen/logrus"
)

var pool, _ = redisutil.NewPool("tcp://127.0.0.1:6379")

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func TestRedisEventBus(t *testing.T) {
	s := redis.NewRedisEventBus(pool)

	sub := s.Subscribe("test", func(ctx context.Context, data []byte) {
		time.Sleep(80 * time.Millisecond)
		fmt.Println(string(data))
	})

	for i := 0; i < 10; i++ {
		catch(s.Publish(context.Background(), "test", []byte(strconv.Itoa(i))))
	}

	time.Sleep(100 * 10 * time.Millisecond)
	sub.Unsubscribe()
}

func catch(err error) {

}
