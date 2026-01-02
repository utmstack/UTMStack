package client

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/utils"
	"github.com/utmstack/UTMStack/threadwinds-ingestion/config"
)

type CustomersManagerClient struct {
	cmURL string
}

func NewCustomersManagerClient(cfg *config.TWConfig) *CustomersManagerClient {
	return &CustomersManagerClient{
		cmURL: cfg.CustomersManagerURL,
	}
}

type RegistrationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type RegistrationResponse struct {
	APIKey    string `json:"api_key"`
	APISecret string `json:"api_secret"`
}

func (c *CustomersManagerClient) RegisterUserReporter(email string) (*RegistrationResponse, error) {
	endpoint := fmt.Sprintf("%s/api/v1/intelligence/register", c.cmURL)

	reqBody := RegistrationRequest{
		Email: email,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal registration request: %w", err)
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	credentials, statusCode, err := utils.DoReq[RegistrationResponse](
		endpoint,
		jsonData,
		http.MethodPost,
		headers,
	)

	if err != nil {
		switch statusCode {
		case http.StatusBadRequest:
			return nil, catcher.Error("invalid registration data", err, map[string]any{
				"status": statusCode,
				"email":  email,
			})
		case http.StatusInternalServerError:
			return nil, catcher.Error("registration service error", err, map[string]any{
				"status": statusCode,
				"email":  email,
			})
		default:
			return nil, catcher.Error("registration failed", err, map[string]any{
				"status": statusCode,
				"email":  email,
			})
		}
	}

	catcher.Info("Successfully registered ThreadWinds intelligence reporter", map[string]any{
		"email": email,
	})

	return &credentials, nil
}
