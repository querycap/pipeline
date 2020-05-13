package redisutil

import (
	"net/url"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
)

// tcp://127.0.0.1:6379?db=10&connectTimeout=10s
func NewPool(uri string) (*redis.Pool, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	q := u.Query()

	fromQuery := func(k string, defaults interface{}, process func(v string) (interface{}, error)) interface{} {
		str := q.Get(k)

		if str == "" {
			return defaults
		}

		v, err := process(str)
		if err != nil {
			return defaults
		}
		return v
	}

	maxActive := fromQuery("maxActive", int64(10), func(s string) (interface{}, error) { return strconv.ParseInt(u.Query().Get("maxActive"), 10, 64) }).(int64)
	db := fromQuery("db", int64(7), func(s string) (interface{}, error) { return strconv.ParseInt(u.Query().Get("db"), 10, 64) }).(int64)
	connectTimeout := fromQuery("connectTimeout", 10*time.Second, func(s string) (interface{}, error) { return time.ParseDuration(s) }).(time.Duration)
	readTimeout := fromQuery("readTimeout", 10*time.Second, func(s string) (interface{}, error) { return time.ParseDuration(s) }).(time.Duration)
	writeTimeout := fromQuery("writeTimeout", 10*time.Second, func(s string) (interface{}, error) { return time.ParseDuration(s) }).(time.Duration)
	idleTimeout := fromQuery("idleTimeout", 10*time.Second, func(s string) (interface{}, error) { return time.ParseDuration(s) }).(time.Duration)

	dialFunc := func() (redis.Conn, error) {
		options := []redis.DialOption{
			redis.DialDatabase(int(db)),
			redis.DialConnectTimeout(connectTimeout),
			redis.DialWriteTimeout(writeTimeout),
			redis.DialReadTimeout(readTimeout),
		}

		if password, ok := u.User.Password(); ok {
			options = append(options, redis.DialPassword(password))
		}

		return redis.Dial(
			"tcp",
			u.Host,
			options...,
		)
	}

	pool := &redis.Pool{
		Dial:        dialFunc,
		MaxActive:   int(maxActive),
		MaxIdle:     int(maxActive) / 2,
		IdleTimeout: idleTimeout,
		Wait:        true,
	}

	return pool, Ping(pool)
}

type RedisPool interface {
	Get() redis.Conn
}

func Ping(r RedisPool) error {
	conn := r.Get()
	defer conn.Close()
	_, err := conn.Do("PING")
	return err
}
