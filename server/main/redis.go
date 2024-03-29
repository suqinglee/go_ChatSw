package main
import (
	"redigo/redis"
	"time"
)

var pool *redis.Pool

func initPool(address string,maxIdle int,maxActive int,idleTimeout time.Duration) {

	pool = &redis.Pool {
		MaxIdle:maxIdle,
		MaxActive:maxActive,
		IdleTimeout:idleTimeout,
		Dial:func() (redis.Conn,error) {
			return redis.Dial("tcp",address)
		},
	}
}