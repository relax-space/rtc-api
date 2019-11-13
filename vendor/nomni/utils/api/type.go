package api

type Result struct {
	Result  interface{} `json:"result"`
	Success bool        `json:"success"`
	Error   Error       `json:"error"`
}

type ArrayResult struct {
	Items      interface{} `json:"items"`
	TotalCount int64       `json:"totalCount"`
}

type Error struct {
	Code    int         `json:"code,omitempty"`
	Message string      `json:"message,omitempty"`
	Details interface{} `json:"details,omitempty"`
}

func (e Error) Error() string {
	return e.Message
}
