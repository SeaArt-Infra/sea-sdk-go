package sa

import "github.com/SeaArt-Infra/sea-sdk-go/internal/transport"

// PassthroughService provides vendor-compatible passthrough APIs.
type PassthroughService struct {
	client *transport.Client
}
