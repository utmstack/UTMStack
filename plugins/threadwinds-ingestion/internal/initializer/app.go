package initializer

import (
	"context"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/threadwinds-ingestion/config"
	"github.com/utmstack/UTMStack/threadwinds-ingestion/internal/client"
	"github.com/utmstack/UTMStack/threadwinds-ingestion/internal/scheduler"
	"golang.org/x/sync/errgroup"
)

type App struct {
	config    *config.TWConfig
	clients   *client.ClientDependencies
	scheduler *scheduler.IngestionScheduler
	eg        *errgroup.Group
	egCtx     context.Context
}

func NewApp(ctx context.Context) (*App, error) {
	app := &App{}

	if err := app.loadConfiguration(); err != nil {
		return nil, err
	}

	if err := app.initializeClients(); err != nil {
		return nil, err
	}

	if err := app.configureThreadWinds(ctx); err != nil {
		return nil, err
	}

	if err := app.buildProcessingPipeline(); err != nil {
		return nil, err
	}

	return app, nil
}

func (a *App) Run(ctx context.Context) error {
	a.eg, a.egCtx = errgroup.WithContext(ctx)
	a.eg.Go(func() error {
		a.scheduler.Start(a.egCtx)
		return nil
	})
	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	catcher.Info("shutting down application", nil)

	done := make(chan error, 1)
	go func() {
		done <- a.eg.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			catcher.Error("scheduler stopped with error", err, nil)
			return err
		}
		catcher.Info("scheduler stopped gracefully", nil)
		return nil
	case <-ctx.Done():
		catcher.Info("shutdown timeout exceeded, forcing shutdown", nil)
		return ctx.Err()
	}
}
