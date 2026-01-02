package client

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/threadwinds-ingestion/config"
)

type PostgresClient struct {
	db *sql.DB
}

type AdminEmailResult struct {
	Email          string
	IsConfigured   bool
	LastModified   time.Time
	LastModifiedBy string
}

func NewPostgresClient(cfg *config.TWConfig) (*PostgresClient, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(2)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	catcher.Info("PostgreSQL connection established via native SQL", map[string]any{
		"host": cfg.DBHost,
		"port": cfg.DBPort,
	})

	return &PostgresClient{db: db}, nil
}

func (c *PostgresClient) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

func (c *PostgresClient) GetAdminEmail(ctx context.Context) (*AdminEmailResult, error) {
	query := `
		SELECT email, last_modified_by, last_modified_date
		FROM jhi_user
		WHERE login = $1 AND created_by = $2
		LIMIT 1
	`

	var email, lastModifiedBy string
	var lastModifiedDate sql.NullTime

	err := c.db.QueryRowContext(ctx, query, "admin", "system").
		Scan(&email, &lastModifiedBy, &lastModifiedDate)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("admin user not found in database")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query admin user: %w", err)
	}

	result := &AdminEmailResult{
		Email:          email,
		IsConfigured:   email != "admin@localhost",
		LastModifiedBy: lastModifiedBy,
	}

	if lastModifiedDate.Valid {
		result.LastModified = lastModifiedDate.Time
	}

	catcher.Info("retrieved current admin email", map[string]any{
		"email":            email,
		"last_modified_by": lastModifiedBy,
	})

	return result, nil
}

func (c *PostgresClient) WaitForValidAdminEmail(ctx context.Context, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	retryInterval := 10 * time.Second
	attempt := 1

	catcher.Info("waiting for admin email configuration before starting service", map[string]any{
		"timeout_minutes": timeout.Minutes(),
		"retry_interval":  retryInterval.String(),
	})

	query := `
		SELECT email
		FROM jhi_user
		WHERE login = $1 AND created_by = $2 AND email != $3
		LIMIT 1
	`

	for time.Now().Before(deadline) {
		queryCtx, cancel := context.WithTimeout(ctx, 5*time.Second)

		var email string
		err := c.db.QueryRowContext(queryCtx, query, "admin", "system", "admin@localhost").
			Scan(&email)
		cancel()

		if err == nil {
			catcher.Info("Admin email configured successfully, starting ThreadWinds ingestion", map[string]any{
				"attempts": attempt,
				"waited":   time.Since(time.Now().Add(-timeout)).Round(time.Second).String(),
			})
			return email, nil
		}

		if err != sql.ErrNoRows {
			catcher.Error("database error while checking admin email", err, map[string]any{
				"attempt": attempt,
			})
		} else {
			remainingTime := time.Until(deadline).Round(time.Second)
			catcher.Info("Admin email not configured yet, waiting...", map[string]any{
				"attempt":        attempt,
				"next_retry_in":  retryInterval.String(),
				"remaining_time": remainingTime.String(),
				"message":        "Service will start automatically once admin configures their email",
			})
		}

		select {
		case <-ctx.Done():
			return "", fmt.Errorf("context cancelled while waiting for admin email: %w", ctx.Err())
		case <-time.After(retryInterval):
			attempt++
		}
	}

	return "", fmt.Errorf("timeout after %v waiting for admin email configuration (%d attempts). Admin must configure email in first login", timeout, attempt-1)
}
