package main

import (
	"context"
	"fmt"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	sdkos "github.com/threatwinds/go-sdk/os"
	"github.com/utmstack/utmstack/backend/internal/mail"
	"github.com/utmstack/utmstack/backend/modules/adaudit"
	"github.com/utmstack/utmstack/backend/modules/alerts"
	"github.com/utmstack/utmstack/backend/modules/appconfig"
	"github.com/utmstack/utmstack/backend/modules/audit"
	"github.com/utmstack/utmstack/backend/modules/billing"
	"github.com/utmstack/utmstack/backend/modules/compliance"
	"github.com/utmstack/utmstack/backend/modules/dashboards"
	"github.com/utmstack/utmstack/backend/modules/datasources"
	ns_repository "github.com/utmstack/utmstack/backend/modules/datasources/repository"
	ns_usecase "github.com/utmstack/utmstack/backend/modules/datasources/usecase"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing"
	"github.com/utmstack/utmstack/backend/modules/iam"
	iam_repository "github.com/utmstack/utmstack/backend/modules/iam/repository"
	iam_usecase "github.com/utmstack/utmstack/backend/modules/iam/usecase"
	"github.com/utmstack/utmstack/backend/modules/incidents"
	incidents_connectors "github.com/utmstack/utmstack/backend/modules/incidents/connectors"
	"github.com/utmstack/utmstack/backend/modules/integrations"
	"github.com/utmstack/utmstack/backend/modules/loganalyzer"
	mcpmod "github.com/utmstack/utmstack/backend/modules/mcp"
	"github.com/utmstack/utmstack/backend/modules/notifications"
	notifications_domain "github.com/utmstack/utmstack/backend/modules/notifications/domain"
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
	billing           *billing.Module
	mail              *mail.Module
	compliance        *compliance.Module
	dashboards        *dashboards.Module
	loganalyzer       *loganalyzer.Module
	alerts            *alerts.Module
	soar              *soar.Module
	datasources       *datasources.Module
	eventProcessing   *eventprocessing.Module
	integrations      *integrations.Module
	opensearchGateway *opensearchgw.Module
	incidents         *incidents.Module
	notifications     *notifications.Module
	socAI             *socai.Module
	adaudit           *adaudit.Module
	mcp               *mcpmod.Module
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
	billingMod := billing.NewModule(env.String("UPDATES_DIR", "/updates", false), 25)
	configMod := appconfig.NewModule(db, cipher, cfg.uploadDir)
	mailMod := mail.NewModule(configMod.Store())
	configMod.SetMailer(mailMod.Service())
	// White-labeling renders only under an MSSP license (resolved from billing).
	configMod.SetWhiteLabelEntitlement(func() bool {
		return billingMod.License().Current().IsMSSP()
	})
	brand := configMod.Branding()
	complianceMod := compliance.NewModule(db, mailMod.Service())
	dashboardsMod := dashboards.NewModule(db)
	loganalyzerMod := loganalyzer.NewModule(db)

	userRepo := iam_repository.NewUserRepository(db)
	rbacRepo := iam_repository.NewRBACRepository(db)
	refreshRepo := iam_repository.NewRefreshTokenRepository(db)
	resetMailer := iam_repository.NewPasswordResetMailer(mailMod.Service(), mailMod.ConfigRepo(), brand)
	invitationMailer := iam_repository.NewUserInvitationMailer(mailMod.Service(), mailMod.ConfigRepo(), brand)
	tfaStateRepo := iam_repository.NewInMemoryTfaStateRepository(tfaChallengeTTL)
	tfaMailer := iam_repository.NewTfaMailer(mailMod.Service(), mailMod.ConfigRepo())

	tfaUsecase := iam_usecase.NewTfaUsecase(userRepo, refreshRepo, rbacRepo, tfaStateRepo, tfaMailer, signer, preAuthSigner, refreshTokenTTL, cfg.tfaEnabled, brand)
	authUsecase := iam_usecase.NewAuthUsecase(userRepo, rbacRepo, refreshRepo, signer, limiter, refreshTokenTTL, resetMailer, tfaUsecase, preAuthSigner, cfg.tfaEnabled)
	userUsecase := iam_usecase.NewUserUsecase(userRepo, rbacRepo, invitationMailer)
	roleUsecase := iam_usecase.NewRoleUsecase(rbacRepo)
	apiKeyRepo := iam_repository.NewAPIKeyRepository(db)
	apiKeyUsecase := iam_usecase.NewAPIKeyUsecase(apiKeyRepo, userRepo)
	idpRepo := iam_repository.NewIdentityProviderRepository(db)
	idpUsecase := iam_usecase.NewIdentityProviderUsecase(idpRepo, cipher)
	samlUsecase := iam_usecase.NewSAMLUsecase(idpRepo, userRepo, refreshRepo, signer, cipher, refreshTokenTTL)

	// Configure the go-sdk OpenSearch global client used by all modules.
	osURL := fmt.Sprintf("https://%s:%d", cfg.esHost, cfg.esPort)
	if err := sdkos.Connect([]string{osURL}, cfg.esUser, cfg.esPassword); err != nil {
		_ = catcher.Error("opensearch SDK connect failed", err, nil)
	}

	alertsMod := alerts.NewModule(db)

	agentClient, agentErr := agentmanager.NewClient()
	if agentErr != nil {
		_ = catcher.Error("agentmanager client init failed (alert response rules will not dispatch)", agentErr, nil)
		agentClient = nil
	}

	soarMod := soar.NewModule(db, agentClient, signer, cipher)
	eventProcessingMod := eventprocessing.NewModule(db, auditMod.Logger())

	dsRepo := ns_repository.NewDatasourceRepository(db)
	dsGroupRepo := ns_repository.NewAssetGroupRepository(db)
	dsUC := ns_usecase.NewDatasourceUsecase(dsRepo)
	dsGroupUC := ns_usecase.NewAssetGroupUsecase(dsGroupRepo)
	var dsReconciler *ns_usecase.StatsReconciler
	if cfg.esHost != "" {
		dsReconciler = ns_usecase.NewStatsReconciler(dsRepo, ns_repository.NewStatsReader())
	}
	datasourcesMod := datasources.NewModule(dsUC, dsGroupUC, dsReconciler, billingMod.License(), agentClient)

	integrationsMod := integrations.NewModule(db, cipher,
		env.String("INTEGRATIONS_TENANT_DIR", "/workdir/pipeline", false),
		dsUC,
	)

	opensearchMod := opensearchgw.NewModule(db, cfg.esHost != "")
	notificationsMod := notifications.NewModule(db, auditMod.Logger())

	if cfg.esHost != "" && cfg.diskGuardEnabled {
		opensearchMod.SetSpaceGuard(
			func(ctx context.Context, critical bool, msg string) error {
				ntype := notifications_domain.TypeWarning
				if critical {
					ntype = notifications_domain.TypeError
				}
				return notificationsMod.Producer().Notify(ctx, notifications_domain.SourceSystem, ntype, msg)
			},
			cfg.diskWarnPercent, cfg.diskDeletePercent,
			time.Duration(cfg.diskGuardIntervalSec)*time.Second,
		)
	}

	iamMod := iam.NewModule(authUsecase, userUsecase, roleUsecase, tfaUsecase, apiKeyUsecase, idpUsecase, samlUsecase, cfg.uploadDir)
	socAIMod := socai.NewModule(cfg.socAIBaseURL, cfg.internalKey)
	incidentsMod := incidents.NewModule(
		db,
		incidents_connectors.NewNoopMailer(),
		incidents.NewAlertsGatewayFromUsecase(alertsMod.GetAlertUsecase()),
		incidents.NewIAMGatewayFromRepo(userRepo),
		auditMod.Logger(),
	)
	adauditMod := adaudit.NewModule(db)

	var mcpModule *mcpmod.Module
	if cfg.mcpEnabled {
		mcpModule = mcpmod.NewModule(&mcpmod.Deps{
			IAM:             iamMod,
			Alerts:          alertsMod,
			Incidents:       incidentsMod,
			SOAR:            soarMod,
			Compliance:      complianceMod,
			Audit:           auditMod,
			Dashboards:      dashboardsMod,
			LogAnalyzer:     loganalyzerMod,
			OpenSearch:      opensearchMod,
			EventProcessing: eventProcessingMod,
			Datasources:     datasourcesMod,
			Integrations:    integrationsMod,
			Notifications:   notificationsMod,
			ADAudit:         adauditMod,
			SOCAI:           socAIMod,
			Billing:         billingMod,
			AppConfig:       configMod,
			ServerName:      cfg.serverName,
			ServerVersion:   cfg.mcpVersion,
		})
	}

	return &modules{
		iam:               iamMod,
		audit:             auditMod,
		appconfig:         configMod,
		billing:           billingMod,
		mail:              mailMod,
		compliance:        complianceMod,
		dashboards:        dashboardsMod,
		loganalyzer:       loganalyzerMod,
		alerts:            alertsMod,
		soar:              soarMod,
		datasources:       datasourcesMod,
		eventProcessing:   eventProcessingMod,
		opensearchGateway: opensearchMod,
		integrations:      integrationsMod,
		socAI:             socAIMod,
		incidents:         incidentsMod,
		notifications:     notificationsMod,
		adaudit:           adauditMod,
		mcp:               mcpModule,
		signer:            signer,
	}
}
