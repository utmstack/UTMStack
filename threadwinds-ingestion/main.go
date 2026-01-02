package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/threadwinds-ingestion/config"
	"github.com/utmstack/UTMStack/threadwinds-ingestion/internal/client"
	"github.com/utmstack/UTMStack/threadwinds-ingestion/internal/scheduler"
	"github.com/utmstack/UTMStack/threadwinds-ingestion/utils"
)

func main() {
	catcher.Info("Starting ThreadWinds Ingestion Service", nil)

	cfg, err := config.GetTWConfig()
	if err != nil {
		catcher.Error("failed to load configuration", err, nil)
		os.Exit(1)
	}

	postgresClient, err := client.NewPostgresClient(cfg)
	if err != nil {
		catcher.Error("failed to initialize postgres client", err, nil)
		os.Exit(1)
	}

	ctx := context.Background()

	adminEmail, err := postgresClient.WaitForValidAdminEmail(ctx, 60*time.Minute)
	if err != nil {
		catcher.Error("cannot start ThreadWinds Ingestion without valid admin email", err, nil)
		os.Exit(1)
	}

	catcher.Info("Valid admin email obtained", map[string]any{
		"admin_email": adminEmail,
	})

	cmClient := client.NewCustomersManagerClient(cfg)
	backendClient := client.NewBackendClient(cfg)
	opensearchClient, err := client.NewOpenSearchClient(cfg)
	if err != nil {
		catcher.Error("failed to initialize opensearch client", err, nil)
		os.Exit(1)
	}

	threadwindsClient := client.NewThreadWindsClient(cfg)

	twConfig, err := backendClient.GetThreadWindsConfig(ctx)
	if err != nil {
		catcher.Error("failed to check ThreadWinds configuration", err, nil)
		os.Exit(1)
	}

	if twConfig.APIKey == "" || twConfig.APISecret == "" {
		catcher.Info("ThreadWinds not configured, will attempt registration with retry...", nil)

		var regResp *client.RegistrationResponse

		registerFunc := func() error {
			currentEmail, emailErr := postgresClient.GetAdminEmail(ctx)
			if emailErr != nil {
				return catcher.Error("failed to get current admin email", emailErr, nil)
			}

			catcher.Info("attempting ThreadWinds registration", map[string]any{
				"email": currentEmail.Email,
			})

			resp, err := cmClient.RegisterUserReporter(currentEmail.Email)
			if err != nil {
				return err
			}
			regResp = resp
			return nil
		}

		utils.InfiniteRetry(registerFunc, "ThreadWinds registration")

		catcher.Info("ThreadWinds registration successful", nil)

		err = backendClient.SaveThreadWindsCredentials(ctx,
			regResp.APIKey,
			regResp.APISecret,
			twConfig.KeyID,
			twConfig.SecretID)
		if err != nil {
			catcher.Error("failed to save ThreadWinds credentials", err, nil)
			os.Exit(1)
		}

		threadwindsClient.UpdateCredentials(regResp.APIKey, regResp.APISecret)
	} else {
		catcher.Info("ThreadWinds already configured", nil)
		threadwindsClient.UpdateCredentials(twConfig.APIKey, twConfig.APISecret)
	}

	ingestionScheduler := scheduler.NewIngestionScheduler(
		cfg,
		backendClient,
		opensearchClient,
		threadwindsClient,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go ingestionScheduler.Start(ctx)

	sig := <-sigChan
	catcher.Info("received shutdown signal, initiating graceful shutdown", map[string]any{
		"signal": sig.String(),
	})

	cancel()

	time.Sleep(5 * time.Second)

	catcher.Info("ThreadWinds Ingestion Service stopped", nil)
}
