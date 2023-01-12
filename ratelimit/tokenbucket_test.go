package ratelimit

import (
	"testing"

	"github.com/kubesure/resiliency"
)

func TestTokenBucketOneMinLimitBreach(t *testing.T) {
	limiter := NewTokenBucketLimiter()
	droppedreq := make(map[int]int)

	for d := 91; d <= 100; d++ {
		droppedreq[d] = d
	}

	for i := 1; i <= 100; i++ {
		_, err := limiter.CheckLimit("MKT-SEARCH-V1", 90, 1)
		if err != nil && err.Code != resiliency.LimitExpired {
			t.Errorf("should not have error other than limit")
		} else if (i == droppedreq[i]) && (err != nil && err.Code != resiliency.LimitExpired) {
			t.Errorf("should not be avaiable")
		}
	}

}
