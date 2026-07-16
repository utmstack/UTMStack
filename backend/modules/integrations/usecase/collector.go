package usecase

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strconv"

	"github.com/utmstack/utmstack/backend/modules/integrations/connectors"
	"github.com/utmstack/utmstack/backend/modules/integrations/domain"
	"github.com/utmstack/utmstack/backend/modules/integrations/dto"
	"github.com/utmstack/utmstack/backend/pkg/agentmanager/agent"
)

const forwarderSearchQuery = "module.Is=FORWARDER"
const forwarderListPageSize = 1000

var allowedAuthValues = map[string]bool{
	"":       true,
	"none":   true,
	"bearer": true,
	"hmac":   true,
}

type collectorUsecase struct {
	client connectors.AgentManagerCollectorClient
}

func NewCollectorUsecase(client connectors.AgentManagerCollectorClient) connectors.CollectorUsecase {
	return &collectorUsecase{client: client}
}

var _ connectors.CollectorUsecase = (*collectorUsecase)(nil)

func (u *collectorUsecase) ListOnlineForwarders(ctx context.Context) ([]dto.CollectorResponse, error) {
	if u.client == nil {
		return nil, domain.ErrAgentManagerUnavailable
	}

	resp, err := u.client.ListCollectors(ctx, forwarderSearchQuery, 1, forwarderListPageSize, "")
	if err != nil {
		return nil, err
	}

	out := make([]dto.CollectorResponse, 0, len(resp.GetRows()))
	for _, c := range resp.GetRows() {
		if c.GetStatus() != agent.Status_ONLINE {
			continue
		}
		out = append(out, dto.CollectorResponse{
			ID:       c.GetId(),
			Hostname: c.GetHostname(),
			IP:       c.GetIp(),
			Version:  c.GetVersion(),
			LastSeen: c.GetLastSeen(),
		})
	}
	return out, nil
}

func (u *collectorUsecase) SetDataTypeConfig(ctx context.Context, collectorID uint32, dataType string, req dto.SetDataTypeConfigRequest) (*dto.ConfigKnowledgeResponse, error) {
	if req.Enabled == nil || req.Proto == "" {
		return nil, domain.ErrInvalidCollectorConfig
	}
	if *req.Enabled && req.Port == "" {
		return nil, domain.ErrInvalidCollectorConfig
	}
	if !allowedAuthValues[req.Auth] {
		return nil, domain.ErrInvalidCollectorConfig
	}

	if u.client == nil {
		return nil, domain.ErrAgentManagerUnavailable
	}

	kv := map[string]string{
		"enabled": strconv.FormatBool(*req.Enabled),
		"proto":   req.Proto,
	}
	if req.Port != "" {
		kv["port"] = req.Port
	}
	if req.TLS != nil {
		kv["tls"] = strconv.FormatBool(*req.TLS)
	}
	if req.Auth != "" {
		kv["auth"] = req.Auth
	}
	if req.Path != "" {
		kv["path"] = req.Path
	}
	if req.SignatureHeader != "" {
		kv["signature_header"] = req.SignatureHeader
	}

	resp, err := u.client.SetCollectorIntegration(ctx, collectorID, dataType, kv)
	if err != nil {
		return nil, err
	}

	return &dto.ConfigKnowledgeResponse{
		Accepted:        resp.GetAccepted() == "true",
		RequestID:       resp.GetRequestId(),
		ErrorMessage:    resp.GetErrorMessage(),
		GeneratedSecret: resp.GetGeneratedSecret(),
	}, nil
}

func (u *collectorUsecase) GetDataTypeConfig(ctx context.Context, collectorID uint32, dataType string) (*dto.GetDataTypeConfigResponse, error) {
	if dataType == "" {
		return nil, domain.ErrInvalidCollectorConfig
	}
	if u.client == nil {
		return nil, domain.ErrAgentManagerUnavailable
	}

	resp, err := u.client.GetCollectorIntegrationState(ctx, collectorID, dataType)
	if err != nil {
		return nil, err
	}
	if !resp.GetConfigured() {
		return &dto.GetDataTypeConfigResponse{Configured: false}, nil
	}

	out := &dto.GetDataTypeConfigResponse{
		Configured:   true,
		ConfigStatus: resp.GetConfigStatus(),
		LastError:    resp.GetLastError(),
	}
	for _, kv := range resp.GetConfigurations() {
		switch kv.GetConfKey() {
		case "enabled":
			enabled := kv.GetConfValue() == "true"
			out.Enabled = &enabled
		case "proto":
			out.Proto = kv.GetConfValue()
		case "port":
			out.Port = kv.GetConfValue()
		case "tls":
			tls := kv.GetConfValue() == "true"
			out.TLS = &tls
		case "auth":
			out.Auth = kv.GetConfValue()
		case "path":
			out.Path = kv.GetConfValue()
		case "signature_header":
			out.SignatureHeader = kv.GetConfValue()
		}
	}
	return out, nil
}

func decodeCertPEMField(field *string, required bool) (string, error) {
	if field == nil || *field == "" {
		if required {
			return "", domain.ErrInvalidCollectorConfig
		}
		return "", nil
	}
	raw, err := base64.StdEncoding.DecodeString(*field)
	if err != nil {
		return "", domain.ErrInvalidCollectorConfig
	}
	if block, _ := pem.Decode(raw); block == nil {
		return "", domain.ErrInvalidCollectorConfig
	}
	return string(raw), nil
}

func (u *collectorUsecase) SetForwarderCertificates(ctx context.Context, collectorID uint32, req dto.SetForwarderCertificatesRequest) (*dto.ConfigKnowledgeResponse, error) {
	certPem, err := decodeCertPEMField(req.CertPem, true)
	if err != nil {
		return nil, err
	}
	keyPem, err := decodeCertPEMField(req.KeyPem, true)
	if err != nil {
		return nil, err
	}
	caPem, err := decodeCertPEMField(req.CaPem, false)
	if err != nil {
		return nil, err
	}

	if u.client == nil {
		return nil, domain.ErrAgentManagerUnavailable
	}

	resp, err := u.client.SetCollectorCertificates(ctx, collectorID, certPem, keyPem, caPem)
	if err != nil {
		return nil, err
	}

	return &dto.ConfigKnowledgeResponse{
		Accepted:     resp.GetAccepted() == "true",
		RequestID:    resp.GetRequestId(),
		ErrorMessage: resp.GetErrorMessage(),
	}, nil
}

func (u *collectorUsecase) GetTLSStatus(ctx context.Context, collectorID uint32) (*dto.TLSStatusResponse, error) {
	if u.client == nil {
		return nil, domain.ErrAgentManagerUnavailable
	}

	resp, err := u.client.GetCollectorCertificatesStatus(ctx, collectorID)
	if err != nil {
		return nil, err
	}
	if resp.GetAccepted() != "true" {
		return nil, fmt.Errorf("tls status query rejected: %s", resp.GetErrorMessage())
	}

	var st dto.TLSStatusResponse
	if err := json.Unmarshal([]byte(resp.GetStatusPayload()), &st); err != nil {
		return nil, fmt.Errorf("failed to parse tls status payload: %w", err)
	}
	return &st, nil
}
