package ratelimit

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
	r "github.com/kubesure/resiliency"
)

//Redis service
var redisSvc, redisPort string

type tokenbucket struct {
	limitKey                    string
	limit, limitDurationSeconds int
}

func NewTokenBucketLimiter(config r.Config) r.RateLimiter {
	redisSvc = config.RedisSvc
	redisPort = config.RedisPort
	return &tokenbucket{limitKey: config.LimitKey, limit: config.Limit, limitDurationSeconds: config.LimitDurationSeconds}
}

//checks if endpoint has request limits available.
//Returns limit availabiliy or errors (limits threshold breach err and other erros)
func (rl *tokenbucket) CheckLimit() (*r.Limit, *r.Error) {
	logger := r.NewLogger()
	mkey := minuteKey(rl.limitKey)

	count, kerr := getLimit(mkey)
	if kerr != nil {
		return nil, kerr
	}

	c, cerr := connWrite()
	if cerr != nil {
		logger.LogInternalError(cerr.Error())
		return nil, &r.Error{Code: r.InternalError, Message: r.DBError}
	}
	defer c.Close()

	if *count == rl.limit {
		m := make(map[string]interface{})
		secRemaining, err := redis.Int64(c.Do("TTL", mkey))
		if err != nil {
			logger.LogInternalError(err.Error())
			return nil, &r.Error{Code: r.InternalError, Message: r.DBError}
		}

		m["limit-seconds-remaining"] = secRemaining
		return &r.Limit{Available: false}, &r.Error{Code: r.LimitExpired, Message: r.LimitExpiredError, Misc: m}
	}

	c.Send("MULTI")
	c.Send("INCR", mkey)
	c.Send("EXPIRE", mkey, rl.limitDurationSeconds, "NX")
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

	result, rerr := redis.String(c.Do("GET", limitKey))

	if rerr != nil && rerr != redis.ErrNil {
		logger.LogInternalError(rerr.Error())
		return nil, &r.Error{Code: r.InternalError, Message: r.DBError}
	}

	if rerr == redis.ErrNil {
		return countPtr(0), nil
	}

	//logger.LogInfo(fmt.Sprintf("key : %v value: %v", limitKey, result))
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

//gives back a read write connection
func connWrite() (redis.Conn, error) {
	redisurl := fmt.Sprintf("redis://%v:%v", redisSvc, redisPort)
	c, err := redis.DialURL(redisurl)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("cannot connect to redis %v ", err)
	}
	return c, nil
}
