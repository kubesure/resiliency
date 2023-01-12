package ratelimit

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"

	r "github.com/kubesure/resiliency"
)

//Redis k8s service
//var redissvc = os.Getenv("redissvc")

type tokenbucket struct{}

func NewTokenBucketLimiter() r.RateLimiter {
	return &tokenbucket{}
}

//checks if endpoint has request limits available.
//Returns limit availabiliy or errors (limits threshold breach err and other erros)
func (rl *tokenbucket) CheckLimit(limitKey string, limit, minDuration int) (*r.Limit, *r.Error) {
	logger := r.NewLogger()
	count, kerr := getLimit(limitKey)
	if kerr != nil {
		return nil, kerr
	}

	if *count == limit {
		m := make(map[string]interface{})
		//TODO
		m["x-seconds-remaining"] = 10
		return nil, &r.Error{Code: r.LimitExpired, Message: r.LimitExpiredError, Misc: m}
	}

	c, cerr := connWrite()
	if cerr != nil {
		logger.LogInternalError(cerr.Error())
		return nil, &r.Error{Code: r.InternalError, Message: r.DBError}
	}
	defer c.Close()

	key := minuteKey(limitKey)

	c.Send("MULTI")
	c.Send("INCR", key)
	c.Send("EXPIRE", key, 59, "NX")
	_, exrr := c.Do("EXEC")

	if exrr != nil {
		logger.LogInternalError(exrr.Error())
		return nil, &r.Error{Code: r.InternalError, Message: r.DBError}
	}

	return &r.Limit{Available: true}, nil
}

func getLimit(limitKey string) (*int, *r.Error) {
	logger := r.NewLogger()
	c, err := connWrite()
	if err != nil {
		logger.LogInternalError(err.Error())
		return nil, &r.Error{Code: r.InternalError, Message: r.DBError}
	}
	defer c.Close()

	key := minuteKey(limitKey)

	result, rerr := redis.String(c.Do("GET", key))

	if rerr != nil && rerr != redis.ErrNil {
		logger.LogInternalError(rerr.Error())
		return nil, &r.Error{Code: r.InternalError, Message: r.DBError}
	}

	if rerr == redis.ErrNil {
		logger.LogInfo(fmt.Sprintf("key : %v not found", key))
		return countPtr(0), nil
	}

	logger.LogInfo(fmt.Sprintf("key : %v value: %v", key, result))
	count, _ := strconv.Atoi(result)
	return countPtr(count), nil

}

func minuteKey(limitKey string) string {
	_, min, _ := time.Now().Clock()
	key := fmt.Sprintf("%v:%v", limitKey, min)
	return key
}

func countPtr(c int) *int {
	return &c
}

//gives back a connection for writing
func connWrite() (redis.Conn, error) {
	c, err := redis.DialURL("redis://" + "localhost" + ":6379")
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("cannot connect to redis %v ", err)
	}
	return c, nil
}
