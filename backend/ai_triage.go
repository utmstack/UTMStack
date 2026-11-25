package main

import (
	"context"

	"github.com/threatwinds/go-sdk/catcher"

	"github.com/utmstack/utmstack/backend/modules/alerts"
	"github.com/utmstack/utmstack/backend/modules/alertscoring"
	"github.com/utmstack/utmstack/backend/modules/socai"
	"github.com/utmstack/utmstack/backend/pkg/env"
	"github.com/utmstack/utmstack/backend/pkg/eventstore"
)

// enableAITriage sends each new alert to the SOC AI service over NATS, with its dossier already worked
// out. Without NATS_URL nothing is sent and the alerts are not analysed automatically; the chat and the
// manual analysis, which do not go through NATS, keep working. It is not worth stopping the backend
// for, unlike the integrations' configuration.
func enableAITriage(
	alertsMod *alerts.Module, socAIMod *socai.Module, quota *socai.AIQuota,
	scoring *alertscoring.Module, events *eventstore.Store,
) {
	url := env.String("NATS_URL", "", false)
	if url == "" {
		catcher.Info("soc-ai: NATS_URL is not set, new alerts will not be analysed automatically", nil)

		return
	}

	pub, err := socai.NewJetStreamPublisher(context.Background(), url, natsStartWait)
	if err != nil {
		_ = catcher.Error("soc-ai: cannot reach NATS, new alerts will not be analysed automatically", err, nil)

		return
	}

	score := func(ctx context.Context, alertID string) (any, error) {
		return scoring.GetScoringUsecase().ScoreAlert(ctx, alertID)
	}

	alertsMod.AddRaisedHook(socai.NewTriage(socAIMod.Config(), quota, score, events, pub).AlertRaised)

	// The service picks its model from the configuration of the tenant an alert belongs to, and reads it
	// from NATS: this keeps that copy current.
	mirror, err := socai.NewConfigMirror(context.Background(), url, socAIMod.ConfigStore(), socAIMod.Cipher(), natsStartWait)
	if err != nil {
		_ = catcher.Error("soc-ai: cannot publish the configuration to NATS; the service will not know the model to use", err, nil)

		return
	}

	go mirror.Run(context.Background())
}
