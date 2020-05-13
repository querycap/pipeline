package redis

import (
	"context"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/querycap/pipeline/pipeline"
	"github.com/sirupsen/logrus"
)

type Conn = redis.Conn

type RedisPool interface {
	Get() Conn
}

func NewRedisEventBus(pool RedisPool) pipeline.EventBus {
	return &RedisEventBus{pool: pool}
}

type RedisEventBus struct {
	pool RedisPool
}

func (r *RedisEventBus) Publish(ctx context.Context, topic string, data []byte) error {
	conn := r.pool.Get()
	defer conn.Close()

	ret, err := redis.Values(conn.Do("PUBSUB", "NUMSUB", subscription(topic)))
	if err != nil {
		return err
	}

	if ret[1].(int64) == 0 {
		return pipeline.ErrNoSubscriptionsForTopic
	}

	if _, err := conn.Do("RPUSH", topic, data); err != nil {
		logrus.WithContext(ctx).Error(err)
		return err
	}
	return nil
}

func (r *RedisEventBus) Subscribe(topic string, callback func(ctx context.Context, data []byte)) pipeline.Subscription {
	chStop := make(chan interface{})

	sub, err := subscribe(r.pool, topic)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			select {
			case <-chStop:
				return
			default:
				data, err := r.pick(topic)
				if err != nil {
					if err != redis.ErrNil {
						logrus.Error(err)
						// waiting for connection recovered
						time.Sleep(500 * time.Millisecond)

						s, err := subscribe(r.pool, topic)
						if err != nil {
							logrus.Error(err)
						}
						sub = s
					}
					continue
				}

				if data != nil {
					callback(context.Background(), data)
				}
			}
		}
	}()

	return pipeline.NewSubscription(func() {
		sub.Unsubscribe()
		close(chStop)
	})
}

func (r *RedisEventBus) pick(topic string) ([]byte, error) {
	conn := r.pool.Get()
	defer conn.Close()

	return redis.Bytes(conn.Do("LPOP", topic))
}

func subscribe(pool RedisPool, topic string) (pipeline.Subscription, error) {
	psc := redis.PubSubConn{Conn: pool.Get()}
	channel := subscription(topic)
	chClose := make(chan struct{})

	if err := psc.Subscribe(channel); err != nil {
		return nil, err
	}

	go func() {
		ticker := time.NewTicker(5 * time.Second)

		for {
			select {
			case <-ticker.C:
				if err := psc.Ping(""); err != nil {
					ticker.Stop()
					return
				}

				catch(expire(pool, topic))

			case <-chClose:
				ticker.Stop()
				return
			}
		}
	}()

	logrus.Debugf("subscribed %s", topic)

	return pipeline.NewSubscription(func() {
		chClose <- struct{}{}
		catch(psc.Close())
	}), nil
}

func subscription(topic string) string {
	return strings.Join([]string{topic, "subscription"}, ":")
}

func expire(pool RedisPool, topic string) error {
	conn := pool.Get()
	defer conn.Close()

	_, err := conn.Do("EXPIRE", topic, 10)
	return err
}

func catch(err error) {

}
