package resiliency

import "github.com/sirupsen/logrus"

type RateLimiter interface {
	CheckLimit() (*Limit, *Error)
}

type Limit struct {
	Available bool
}

type Error struct {
	Code       EventCode
	Inner      error
	Message    ErrorMessage
	StackTrace string
	Misc       map[string]interface{}
}

type EventCode int

const (
	InternalError EventCode = iota
	LimitExpired
	Info
	Debug
)

type ErrorMessage string

const (
	DBError           ErrorMessage = "Redis error"
	LimitExpiredError ErrorMessage = "Endpoint limit expired"
)

// StandardLogger enforces specific log message formats
type StandardLogger struct {
	*logrus.Logger
}

//LogEvent stores log message
type LogEvent struct {
	id      EventCode
	message string
}

type Config struct {
	RedisSvc, RedisPort         string
	LimitKey                    string
	Limit, LimitDurationSeconds int
}
