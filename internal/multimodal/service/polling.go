package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	mmtypes "github.com/seaart/sa-go/internal/multimodal/types"
	"github.com/seaart/sa-go/internal/shared"
	"github.com/seaart/sa-go/internal/transport"
)

const (
	StatusCompleted  = "completed"
	StatusFailed     = "failed"
	StatusInProgress = "in_progress"

	pollNetworkRetryLimit = 3
)

func GenerateAndWait(client *transport.Client, ctx context.Context, req *mmtypes.GenerateRequest, opts ...mmtypes.PollOption) (*mmtypes.TaskResponse, error) {
	gen, err := CreateTask(client, ctx, req, nil)
	if err != nil {
		return nil, err
	}
	return WaitTask(client, ctx, gen.ID, opts...)
}

func WaitTask(client *transport.Client, ctx context.Context, taskID string, opts ...mmtypes.PollOption) (*mmtypes.TaskResponse, error) {
	cfg := mmtypes.ApplyPollOptions(opts...)
	deadline := time.Now().Add(cfg.Timeout)
	networkErrors := 0

	for time.Now().Before(deadline) {
		task, err := GetTask(client, ctx, taskID, nil)
		if err != nil {
			sdkErr, ok := err.(*shared.Error)
			if ok && sdkErr.Kind == shared.ErrNetwork && networkErrors < pollNetworkRetryLimit {
				networkErrors++
				select {
				case <-ctx.Done():
					return nil, &shared.Error{Kind: shared.ErrNetwork, Message: "context cancelled", TaskID: taskID}
				case <-time.After(cfg.Interval):
					continue
				}
			}
			if ok {
				sdkErr.TaskID = taskID
			}
			return nil, err
		}
		networkErrors = 0

		status := strings.ToLower(task.Status)
		if cfg.OnUpdate != nil {
			cfg.OnUpdate(status, task.Progress)
		}

		switch status {
		case StatusCompleted:
			return task, nil
		case StatusFailed:
			message := "task failed"
			if task.Error != nil && task.Error.ErrorMessage != "" {
				message = "task failed: " + task.Error.ErrorMessage
			}
			return nil, &shared.Error{Kind: shared.ErrTaskFailed, Message: message, TaskID: taskID}
		}

		select {
		case <-ctx.Done():
			return nil, &shared.Error{Kind: shared.ErrTimeout, Message: "context cancelled", TaskID: taskID}
		case <-time.After(cfg.Interval):
		}
	}

	return nil, &shared.Error{
		Kind:    shared.ErrTimeout,
		Message: fmt.Sprintf("task timed out after %s", cfg.Timeout),
		TaskID:  taskID,
	}
}

func PollTaskAsync(client *transport.Client, ctx context.Context, taskID string, opts ...mmtypes.PollOption) <-chan mmtypes.TaskEvent {
	ch := make(chan mmtypes.TaskEvent, 8)

	go func() {
		defer close(ch)

		cfg := mmtypes.ApplyPollOptions(opts...)
		deadline := time.Now().Add(cfg.Timeout)
		networkErrors := 0

		for time.Now().Before(deadline) {
			task, err := GetTask(client, ctx, taskID, nil)
			if err != nil {
				sdkErr, ok := err.(*shared.Error)
				if ok && sdkErr.Kind == shared.ErrNetwork && networkErrors < pollNetworkRetryLimit {
					networkErrors++
					select {
					case <-ctx.Done():
						ch <- mmtypes.TaskEvent{Err: &shared.Error{Kind: shared.ErrNetwork, Message: "context cancelled", TaskID: taskID}}
						return
					case <-time.After(cfg.Interval):
						continue
					}
				}
				if ok {
					sdkErr.TaskID = taskID
				}
				ch <- mmtypes.TaskEvent{Err: err}
				return
			}
			networkErrors = 0

			event := mmtypes.TaskEvent{
				Status:   strings.ToLower(task.Status),
				Progress: task.Progress,
			}

			switch event.Status {
			case StatusCompleted:
				event.Task = task
				ch <- event
				return
			case StatusFailed:
				message := "task failed"
				if task.Error != nil && task.Error.ErrorMessage != "" {
					message = "task failed: " + task.Error.ErrorMessage
				}
				ch <- mmtypes.TaskEvent{
					Err: &shared.Error{Kind: shared.ErrTaskFailed, Message: message, TaskID: taskID},
				}
				return
			}

			ch <- event

			select {
			case <-ctx.Done():
				ch <- mmtypes.TaskEvent{Err: &shared.Error{Kind: shared.ErrTimeout, Message: "context cancelled", TaskID: taskID}}
				return
			case <-time.After(cfg.Interval):
			}
		}

		ch <- mmtypes.TaskEvent{
			Err: &shared.Error{
				Kind:    shared.ErrTimeout,
				Message: fmt.Sprintf("task timed out after %s", cfg.Timeout),
				TaskID:  taskID,
			},
		}
	}()

	return ch
}
