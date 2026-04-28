package shared

import "fmt"

const (
	ErrAuth       = "auth"
	ErrQuota      = "quota"
	ErrTimeout    = "timeout"
	ErrNetwork    = "network"
	ErrTaskFailed = "task_failed"
	ErrGeneral    = "general"
)

// Error is the shared SDK error type used across internal services and the
// public sa package.
type Error struct {
	Kind    string
	Message string
	Status  int
	TaskID  string
}

func (e *Error) Error() string {
	if e.TaskID != "" {
		return fmt.Sprintf("%s (task_id: %s)", e.Message, e.TaskID)
	}
	return e.Message
}

func NewHTTPError(status int, message string) *Error {
	kind := ErrGeneral
	switch {
	case status == 401 || status == 403:
		kind = ErrAuth
	case status == 429:
		kind = ErrQuota
	case status == 408 || status == 504:
		kind = ErrTimeout
	}
	return &Error{Kind: kind, Status: status, Message: message}
}
