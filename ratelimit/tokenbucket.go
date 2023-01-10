package ratelimit

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/kubesure/resiliency"
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
func (rl *tokenbucket) Process(limitKey string) (*r.Limit, *r.Error) {

	count, err := getLimit(limitKey)
	if err != nil {
		return nil, err
	}

	if *count == 0 {
		//set new minutre key return key
	}
	//incr expire

	return &r.Limit{Available: true, MsRemaining: 1}, nil
}

func getLimit(limitKey string) (*int, *resiliency.Error) {
	logger := resiliency.NewLogger()
	c, err := connWrite()
	if err != nil {
		logger.LogInternalError(err.Error())
		return nil, &resiliency.Error{Code: resiliency.InternalError, Message: resiliency.DBError}
	}
	defer c.Close()

	_, min, _ := time.Now().Clock()

	key := fmt.Sprintf("%v:%v", limitKey, min)

	result, rerr := redis.String(c.Do("GET", key))

	if rerr != nil && rerr != redis.ErrNil {
		logger.LogInternalError(rerr.Error())
		return nil, &resiliency.Error{Code: resiliency.InternalError, Message: resiliency.DBError}
	}

	if rerr == redis.ErrNil {
		logger.LogInfo(fmt.Sprintf("key : %v not found", key))
		return countPtr(0), nil
	}

	logger.LogInfo(fmt.Sprintf("key : %v value: %v", key, result))
	count, _ := strconv.Atoi(result)
	return countPtr(count), nil

}

func countPtr(c int) *int {
	return &c
}

//gives back a connection for writing
func connWrite() (redis.Conn, error) {
	c, err := redis.DialURL("redis://" + "localhost" + ":6379")
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("Cannot connect to redis sentinel %v ", err)
	}
	return c, nil
}
