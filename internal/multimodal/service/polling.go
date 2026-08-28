package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	mmtypes "github.com/SeaArt-Infra/sea-sdk-go/internal/multimodal/types"
	"github.com/SeaArt-Infra/sea-sdk-go/internal/shared"
	"github.com/SeaArt-Infra/sea-sdk-go/internal/transport"
)

const (
	StatusCompleted  = "completed"
	StatusFailed     = "failed"
	StatusInProgress = "in_progress"

	pollRetryLimit = 3
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
	retries := 0

	for time.Now().Before(deadline) {
		task, err := GetTask(client, ctx, taskID, nil)
		if err != nil {
			sdkErr, ok := err.(*shared.Error)
			if shouldRetryPollError(sdkErr, retries) {
				retries++
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
		retries = 0

		status := strings.ToLower(task.Status)
		if cfg.OnUpdate != nil {
			cfg.OnUpdate(status, task.Progress)
		}

		switch status {
		case StatusCompleted:
			return task, nil
		case StatusFailed:
			message := "task failed"
			code := 0
			if task.Error != nil {
				detail := task.Error.ErrorMessage
				if detail == "" {
					detail = task.Error.Message
				}
				if detail != "" {
					message = "task failed: " + detail
				}
				code = task.Error.Code
			}
			return nil, &shared.Error{Kind: shared.ErrTaskFailed, Message: message, TaskID: taskID, Code: code}
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
		retries := 0

		for time.Now().Before(deadline) {
			task, err := GetTask(client, ctx, taskID, nil)
			if err != nil {
				sdkErr, ok := err.(*shared.Error)
				if shouldRetryPollError(sdkErr, retries) {
					retries++
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
			retries = 0

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
				code := 0
				if task.Error != nil {
					detail := task.Error.ErrorMessage
					if detail == "" {
						detail = task.Error.Message
					}
					if detail != "" {
						message = "task failed: " + detail
					}
					code = task.Error.Code
				}
				ch <- mmtypes.TaskEvent{
					Err: &shared.Error{Kind: shared.ErrTaskFailed, Message: message, TaskID: taskID, Code: code},
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

func shouldRetryPollError(err *shared.Error, retries int) bool {
	if err == nil || retries >= pollRetryLimit {
		return false
	}
	if err.Kind == shared.ErrNetwork {
		return true
	}
	return err.Status == 502 || err.Status == 503 || err.Status == 504
}
