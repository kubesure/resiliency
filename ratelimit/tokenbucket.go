package ratelimit

import (
	"fmt"
	"log"
	"os"

	"github.com/gomodule/redigo/redis"
	"github.com/kubesure/resiliency"
)

//Redis k8s service
var redissvc = os.Getenv("redissvc")

type limit struct {
	available   bool
	msRemaining int
}

//checks if endpoint has request limits available. Returns
func checkBucket(ep string) (*limit, *resiliency.Error) {
	c, err := connWrite()
	if err != nil {
		return nil, &resiliency.Error{Code: resiliency.InternalError, Message: resiliency.DBError}
	}
	defer c.Close()
	//return nil, &resiliency.Error{Code: resiliency.LimitExpired, Message: resiliency.LimitExpiredError}
	return &limit{available: true, msRemaining: 1}, nil
}

//gives back a connection to master for writing or loading premium matrix
func connWrite() (redis.Conn, error) {
	sc, err := redis.DialURL("redis://" + redissvc + ":26379/0")
	if err != nil {
		return nil, fmt.Errorf("Cannot connect to redis sentinel %v ", err)
	}
	defer sc.Close()

	minfo, err := redis.Strings(sc.Do("sentinel", "get-master-addr-by-name", "redis-premium-master"))
	log.Println(minfo)
	if err != nil {
		return nil, fmt.Errorf("Cannot find redis master %v ", err)
	}

	mc, err := redis.DialURL("redis://" + minfo[0] + ":6379/0")
	if err != nil {
		return nil, fmt.Errorf("Cannot connect to redis master %v ", err)
	}
	sc.Close()
	return mc, nil
}
