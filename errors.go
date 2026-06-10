package sa

import "github.com/SeaArt-Infra/sea-sdk-go/internal/shared"

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
