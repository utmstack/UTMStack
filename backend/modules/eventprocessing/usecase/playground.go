package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/utmstack/utmstack/backend/modules/eventprocessing/domain"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/dto"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/infrastructure"
)

type PlaygroundUsecase struct {
	client *infrastructure.PlaygroundClient
	sem    chan struct{}
}

func NewPlaygroundUsecase(client *infrastructure.PlaygroundClient) *PlaygroundUsecase {
	uc := &PlaygroundUsecase{
		client: client,
		sem:    make(chan struct{}, 1),
	}
	uc.sem <- struct{}{}
	return uc
}

func (uc *PlaygroundUsecase) acquire(ctx context.Context) error {
	timer := time.NewTimer(SemaphoreAcquireTimeout)
	defer timer.Stop()

	select {
	case <-uc.sem:
		return nil
	case <-timer.C:
		return domain.ErrPlaygroundBusy
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (uc *PlaygroundUsecase) release() { uc.sem <- struct{}{} }

func (uc *PlaygroundUsecase) TestPipeline(ctx context.Context, req dto.TestPipelineRequest) (*dto.PlaygroundResponse, error) {
	if err := uc.acquire(ctx); err != nil {
		return nil, err
	}
	defer uc.release()

	if status, body, err := uc.client.DeleteRules(ctx); err != nil || status >= 400 {
		return nil, mapPluginError(status, body, err)
	}

	if status, body, err := uc.client.DeleteFilters(ctx); err != nil || status >= 400 {
		return nil, mapPluginError(status, body, err)
	}

	if req.Pipeline == nil {
		if status, body, err := uc.client.CopyFilters(ctx); err != nil || status >= 400 {
			return nil, mapPluginError(status, body, err)
		}
	} else {
		if status, body, err := uc.client.WriteFilter(ctx, PlaygroundUserFilename, req.Pipeline.Content); err != nil || status >= 400 {
			return nil, mapPluginError(status, body, err)
		}
	}
	return uc.run(ctx, req.Log)
}

func (uc *PlaygroundUsecase) TestRule(ctx context.Context, req dto.TestRuleRequest) (*dto.PlaygroundResponse, error) {
	if err := uc.acquire(ctx); err != nil {
		return nil, err
	}
	defer uc.release()

	if status, body, err := uc.client.DeleteRules(ctx); err != nil || status >= 400 {
		return nil, mapPluginError(status, body, err)
	}

	if status, body, err := uc.client.DeleteFilters(ctx); err != nil || status >= 400 {
		return nil, mapPluginError(status, body, err)
	}

	if status, body, err := uc.client.CopyFilters(ctx); err != nil || status >= 400 {
		return nil, mapPluginError(status, body, err)
	}

	if req.Rule == nil {
		if status, body, err := uc.client.CopyRules(ctx); err != nil || status >= 400 {
			return nil, mapPluginError(status, body, err)
		}
	} else {
		if status, body, err := uc.client.WriteRule(ctx, PlaygroundUserFilename, req.Rule.Content); err != nil || status >= 400 {
			return nil, mapPluginError(status, body, err)
		}
	}
	return uc.run(ctx, req.Log)
}

func (uc *PlaygroundUsecase) run(ctx context.Context, log json.RawMessage) (*dto.PlaygroundResponse, error) {
	status, body, err := uc.client.Run(ctx, log)
	if err != nil || status >= 400 {
		return nil, mapPluginError(status, body, err)
	}

	var envelope struct {
		Status  string          `json:"status"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("%w: malformed plugin envelope", domain.ErrPlaygroundInfra)
	}
	if envelope.Status == "error" {
		return nil, fmt.Errorf("%w: plugin signalled error after HTTP 200: %s", domain.ErrPlaygroundInfra, envelope.Message)
	}
	if envelope.Data == nil {
		return nil, fmt.Errorf("%w: plugin returned null data", domain.ErrPlaygroundInfra)
	}

	var result dto.PlaygroundResponse
	if err := json.Unmarshal(envelope.Data, &result); err != nil {
		return nil, fmt.Errorf("%w: unmarshal RunResult: %s", domain.ErrPlaygroundInfra, err.Error())
	}
	if result.Alerts == nil {
		result.Alerts = []json.RawMessage{}
	}
	return &result, nil
}

func mapPluginError(status int, body []byte, err error) error {
	if err != nil {
		return fmt.Errorf("%w: %s", domain.ErrPlaygroundInfra, err.Error())
	}
	switch {
	case status == 400:
		return fmt.Errorf("%w: %s", domain.ErrPlaygroundBadInput, extractPluginMessage(body))
	case status == 401:
		return domain.ErrPlaygroundMisconfigured
	case status >= 500:
		return fmt.Errorf("%w: plugin returned %d", domain.ErrPlaygroundInfra, status)
	default:
		return fmt.Errorf("%w: unexpected plugin status %d", domain.ErrPlaygroundInfra, status)
	}
}

func extractPluginMessage(body []byte) string {
	var env struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &env); err != nil || env.Message == "" {
		return "invalid input"
	}
	return env.Message
}
