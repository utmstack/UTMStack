package main

import (
	"fmt"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	sdkos "github.com/threatwinds/go-sdk/os"
	"github.com/utmstack/utmstack/backend/internal/mail"
	"github.com/utmstack/utmstack/backend/modules/alerts"
	"github.com/utmstack/utmstack/backend/modules/appconfig"
	"github.com/utmstack/utmstack/backend/modules/audit"
	"github.com/utmstack/utmstack/backend/modules/collectors"
	"github.com/utmstack/utmstack/backend/modules/correlation"
	"github.com/utmstack/utmstack/backend/modules/datainput"
	"github.com/utmstack/utmstack/backend/modules/iam"
	iam_repository "github.com/utmstack/utmstack/backend/modules/iam/repository"
	iam_usecase "github.com/utmstack/utmstack/backend/modules/iam/usecase"
	"github.com/utmstack/utmstack/backend/modules/incidents"
	incidents_connectors "github.com/utmstack/utmstack/backend/modules/incidents/connectors"
	"github.com/utmstack/utmstack/backend/modules/indexpattern"
	"github.com/utmstack/utmstack/backend/modules/logstash"
	"github.com/utmstack/utmstack/backend/modules/notifications"
	opensearchgw "github.com/utmstack/utmstack/backend/modules/opensearch"
	"github.com/utmstack/utmstack/backend/modules/soar"
	"github.com/utmstack/utmstack/backend/modules/socai"
	"github.com/utmstack/utmstack/backend/pkg/agentmanager"
	"github.com/utmstack/utmstack/backend/pkg/env"
	jwtpkg "github.com/utmstack/utmstack/backend/pkg/jwt"
	"github.com/utmstack/utmstack/backend/pkg/ratelimit"
	"github.com/utmstack/utmstack/backend/pkg/secret"
	"gorm.io/gorm"
)

const (
	accessTokenTTL   = 1 * time.Hour
	refreshTokenTTL  = 30 * 24 * time.Hour
	loginMaxFailures = 9
	loginBlockTTL    = 10 * time.Minute
	loginWindowTTL   = 10 * time.Minute
	tfaPreAuthTTL    = 5 * time.Minute
	tfaChallengeTTL  = 10 * time.Minute
)

type modules struct {
	iam               *iam.Module
	audit             *audit.Module
	appconfig         *appconfig.Module
	mail              *mail.Module
	alerts            *alerts.Module
	soar              *soar.Module
	collectors        *collectors.Module
	correlation       *correlation.Module
	datainput         *datainput.Module
	logstash          *logstash.Module
	indexpattern      *indexpattern.Module
	opensearchGateway *opensearchgw.Module
	incidents         *incidents.Module
	notifications     *notifications.Module
	socAI             *socai.Module
	signer            *jwtpkg.Signer
}

func initModules(db *gorm.DB, cfg *config) *modules {
	if cfg.jwtSecret == "" {
		_ = catcher.Error("JWT_SECRET is not set — refusing to start", nil, nil)
		panic("JWT_SECRET is required")
	}
	if cfg.encryptionKey == "" {
		_ = catcher.Error("ENCRYPTION_KEY is not set — refusing to start", nil, nil)
		panic("ENCRYPTION_KEY is required")
	}
	cipher, err := secret.NewCipher(cfg.encryptionKey)
	if err != nil {
		_ = catcher.Error("failed to init cipher", err, nil)
		panic(err)
	}

	signer := jwtpkg.NewSigner(cfg.jwtSecret, accessTokenTTL)
	preAuthSigner := jwtpkg.NewPreAuthSigner(cfg.jwtSecret, tfaPreAuthTTL)
	limiter := ratelimit.NewLoginLimiter(loginMaxFailures, loginBlockTTL, loginWindowTTL)

	auditMod := audit.NewModule(db)
	configMod := appconfig.NewModule(db, cipher)
	mailMod := mail.NewModule(configMod.Store())
	configMod.SetMailer(mailMod.Service())

	userRepo := iam_repository.NewUserRepository(db)
	rbacRepo := iam_repository.NewRBACRepository(db)
	refreshRepo := iam_repository.NewRefreshTokenRepository(db)
	resetMailer := iam_repository.NewPasswordResetMailer(mailMod.Service(), mailMod.ConfigRepo())
	invitationMailer := iam_repository.NewUserInvitationMailer(mailMod.Service(), mailMod.ConfigRepo())
	tfaStateRepo := iam_repository.NewInMemoryTfaStateRepository(tfaChallengeTTL)
	tfaMailer := iam_repository.NewTfaMailer(mailMod.Service(), mailMod.ConfigRepo())

	tfaUsecase := iam_usecase.NewTfaUsecase(userRepo, refreshRepo, rbacRepo, tfaStateRepo, tfaMailer, signer, preAuthSigner, refreshTokenTTL, cfg.tfaEnabled)
	authUsecase := iam_usecase.NewAuthUsecase(userRepo, rbacRepo, refreshRepo, signer, limiter, refreshTokenTTL, resetMailer, tfaUsecase, preAuthSigner, cfg.tfaEnabled)
	userUsecase := iam_usecase.NewUserUsecase(userRepo, rbacRepo, invitationMailer)
	roleUsecase := iam_usecase.NewRoleUsecase(rbacRepo)
	apiKeyRepo := iam_repository.NewAPIKeyRepository(db)
	apiKeyUsecase := iam_usecase.NewAPIKeyUsecase(apiKeyRepo, userRepo)

	// Configure the go-sdk OpenSearch global client used by all modules.
	osURL := fmt.Sprintf("https://%s:%d", cfg.esHost, cfg.esPort)
	if err := sdkos.Connect([]string{osURL}, cfg.esUser, cfg.esPassword); err != nil {
		_ = catcher.Error("opensearch SDK connect failed", err, nil)
	}

	alertsMod := alerts.NewModule(db, env.Bool("ALERTS_SCHEDULER_ENABLED", false))

	agentClient, agentErr := agentmanager.NewClient()
	if agentErr != nil {
		_ = catcher.Error("agentmanager client init failed (alert response rules will not dispatch)", agentErr, nil)
		agentClient = nil
	}

	soarMod := soar.NewModule(db, agentClient, signer)
	collectorsMod := collectors.NewModule(db, agentClient)
	datainputMod := datainput.NewModule(db)
	correlationMod := correlation.NewModule(db, auditMod.Logger(), datainputMod.GetReader())
	logstashMod := logstash.NewModule(db, auditMod.Logger())
	indexpatternMod := indexpattern.NewModule(db, cfg.esHost != "")

	return &modules{
		iam:               iam.NewModule(authUsecase, userUsecase, roleUsecase, tfaUsecase, apiKeyUsecase, cfg.uploadDir),
		audit:             auditMod,
		appconfig:         configMod,
		mail:              mailMod,
		alerts:            alertsMod,
		soar:              soarMod,
		collectors:        collectorsMod,
		correlation:       correlationMod,
		datainput:         datainputMod,
		logstash:          logstashMod,
		indexpattern:      indexpatternMod,
		opensearchGateway: opensearchgw.NewModule(),
		socAI:             socai.NewModule(cfg.socAIBaseURL, cfg.internalKey),
		incidents: incidents.NewModule(
			db,
			incidents_connectors.NewNoopMailer(),
			incidents.NewAlertsGatewayFromUsecase(alertsMod.GetAlertUsecase()),
			incidents.NewIAMGatewayFromRepo(userRepo),
			auditMod.Logger(),
		),
		notifications: notifications.NewModule(db, auditMod.Logger()),
		signer:        signer,
	}
}
