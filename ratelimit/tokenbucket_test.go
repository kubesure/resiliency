package ratelimit

import (
	"log"
	"testing"

	"github.com/kubesure/resiliency"
)

func TestTokenBucketOneMinLimitBreach(t *testing.T) {
	config := resiliency.Config{
		RedisSvc:             "localhost",
		RedisPort:            "6379",
		LimitKey:             "MKT-SEARCH-V1",
		Limit:                90,
		LimitDurationSeconds: 59,
	}

	limiter := NewTokenBucketLimiter(config)

	droppedreq := make(map[int]int)

	for d := 91; d <= 100; d++ {
		droppedreq[d] = d
	}

	for i := 1; i <= 100; i++ {
		limit, err := limiter.CheckLimit()
		if err != nil && err.Code != resiliency.LimitExpired {
			t.Errorf("should not have error other than limit")
		} else if (i == droppedreq[i]) && (err != nil && err.Code != resiliency.LimitExpired) {
			t.Errorf("should not be avaiable")
		} else if (i == droppedreq[i]) && (err != nil && err.Code == resiliency.LimitExpired) {
			log.Printf("Limit available: %v seconds remaining: %v", limit.Available, err.Misc["limit-seconds-remaining"])
		}
	}

}
