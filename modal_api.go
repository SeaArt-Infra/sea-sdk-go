package sa

import (
	"context"
	"net/http"

	mmservice "github.com/seaart/sa-go/internal/multimodal/service"
	mmtypes "github.com/seaart/sa-go/internal/multimodal/types"
	"github.com/seaart/sa-go/internal/transport"
)

func (m *ModalService) Create(ctx context.Context, body JSONMap, opts ...RequestOption) (*Task, error) {
	resp, err := mmservice.CreateTask(m.client, ctx, body, buildRequestOptions(opts).headers)
	if err != nil {
		return nil, err
	}
	return &Task{
		ID:     resp.ID,
		Status: resp.Status,
		Model:  resp.Model,
		Error:  resp.Error,
		client: m.client,
	}, nil
}

func getTask(client *transport.Client, ctx context.Context, taskID string, headers http.Header) (*Task, error) {
	resp, err := mmservice.GetTask(client, ctx, taskID, headers)
	if err != nil {
		return nil, err
	}
	return newTaskFromResponse(client, resp), nil
}

func newTaskFromResponse(client *transport.Client, resp *mmtypes.TaskResponse) *Task {
	if resp == nil {
		return nil
	}
	return &Task{
		ID:       resp.ID,
		Status:   resp.Status,
		Model:    resp.Model,
		Progress: resp.Progress,
		Output:   resp.Output,
		Usage:    resp.Usage,
		Error:    resp.Error,
		client:   client,
	}
}
