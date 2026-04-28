package sa

import "github.com/seaart/sa-go/internal/shared"

const (
	ErrAuth       = shared.ErrAuth
	ErrQuota      = shared.ErrQuota
	ErrTimeout    = shared.ErrTimeout
	ErrNetwork    = shared.ErrNetwork
	ErrTaskFailed = shared.ErrTaskFailed
	ErrGeneral    = shared.ErrGeneral
)

type Error = shared.Error

func newHTTPError(status int, message string) *Error {
	return shared.NewHTTPError(status, message)
}
