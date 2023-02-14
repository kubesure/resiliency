package ratelimit

import (
	"log"
	"testing"

	"github.com/kubesure/resiliency"
)

func TestTokenBucketOneMinLimitBreach(t *testing.T) {
	limiter := newOneMinLimiter()
	droppedreq := droppedRequests()

	for i := 1; i <= 100; i++ {
		limit, err := limiter.CheckLimit()
		if err != nil && err.Code != resiliency.LimitExpired {
			t.Errorf("should not have error other than limit")
		} else if (i == droppedreq[i]) && (err != nil && err.Code != resiliency.LimitExpired) {
			t.Errorf("should not be available")
		} else if (i == droppedreq[i]) && (err != nil && err.Code == resiliency.LimitExpired) {
			log.Printf("Limit available: %v seconds remaining: %v", limit.Available, err.Misc["limit-seconds-remaining"])
		}
	}
}

func TestTokenBucketFailOpen(t *testing.T) {
	limiter := newLimiterRedisUnavailable()
	_, err := limiter.CheckLimit()
	if err != nil && err.Code != resiliency.InternalError {
		t.Errorf("Redis is not available this case should pass")
	}
}

func droppedRequests() map[int]int {
	droppedreq := make(map[int]int)

	for d := 91; d <= 100; d++ {
		droppedreq[d] = d
	}
	return droppedreq
}

func newOneMinLimiter() resiliency.RateLimiter {
	config := resiliency.Config{
		RedisSvc:             "localhost",
		RedisPort:            "6379",
		LimitKey:             "MKT-SEARCH-V1",
		Limit:                90,
		LimitDurationSeconds: 59,
	}

	limiter := NewTokenBucketLimiter(config)
	return limiter
}

func newLimiterRedisUnavailable() resiliency.RateLimiter {
	config := resiliency.Config{
		RedisSvc:             "localGhooooost",
		RedisPort:            "6379",
		LimitKey:             "MKT-SEARCH-V1",
		Limit:                90,
		LimitDurationSeconds: 59,
	}

	limiter := NewTokenBucketLimiter(config)
	return limiter
}
