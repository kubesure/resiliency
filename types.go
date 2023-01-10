package resiliency

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
)

type ErrorMessage string

const (
	DBError           ErrorMessage = "Redis rrror"
	LimitExpiredError ErrorMessage = "Endpoint limit expired"
)

type ErroResponse struct {
	Code    EventCode    `json:"errorCode"`
	Message ErrorMessage `json:"errorMessage"`
}
