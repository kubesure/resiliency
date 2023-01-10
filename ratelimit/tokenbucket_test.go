package ratelimit

import (
	"testing"

	"github.com/kubesure/resiliency"
)

func TestTokenBucketOneMinLimitBreach(t *testing.T) {
	for _, v := range []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12} {
		limit, err := checkBucket("MKT-SEARCH-V1")
		if err != nil && err.Code != resiliency.LimitExpired {
			t.Errorf("should not have error other than limit")
		} else if (v == 11 || v == 12) && limit.available {
			t.Errorf("should not be avaiable")
		}
	}

}
