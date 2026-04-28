package sa

import (
	"context"
	"time"

	mmservice "github.com/seaart/sa-go/internal/multimodal/service"
	mmtypes "github.com/seaart/sa-go/internal/multimodal/types"
	"github.com/seaart/sa-go/internal/transport"
)

type PollOption = mmtypes.PollOption

func WithPollInterval(d time.Duration) PollOption {
	return mmtypes.WithPollInterval(d)
}

func WithPollTimeout(d time.Duration) PollOption {
	return mmtypes.WithPollTimeout(d)
}

func WithPollCallback(fn func(status string, progress float64)) PollOption {
	return mmtypes.WithPollCallback(fn)
}

func (m *ModalService) Get(ctx context.Context, taskID string, opts ...RequestOption) (*Task, error) {
	return getTask(m.client, ctx, taskID, buildRequestOptions(opts).headers)
}

func (m *ModalService) Wait(ctx context.Context, taskID string, opts ...PollOption) (*Task, error) {
	return waitTask(m.client, ctx, taskID, opts...)
}

func (t *Task) Wait(ctx context.Context, opts ...PollOption) (*Task, error) {
	if t == nil {
		return nil, &Error{Kind: ErrGeneral, Message: "task is nil"}
	}
	if t.client == nil {
		return nil, &Error{Kind: ErrGeneral, Message: "task is detached from client"}
	}
	return waitTask(t.client, ctx, t.ID, opts...)
}

func waitTask(client *transport.Client, ctx context.Context, taskID string, opts ...PollOption) (*Task, error) {
	task, err := mmservice.WaitTask(client, ctx, taskID, opts...)
	if err != nil {
		return nil, err
	}
	return newTaskFromResponse(client, task), nil
}
