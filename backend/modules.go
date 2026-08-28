package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/google/uuid"
	iam_handler "github.com/utmstack/utmstack/backend/modules/iam/handler"
	"github.com/utmstack/utmstack/backend/pkg/joblease"

	dash_usecase "github.com/utmstack/utmstack/backend/modules/dashboards/usecase"
	"github.com/utmstack/utmstack/backend/pkg/eventstore"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/internal/mail"
	"github.com/utmstack/utmstack/backend/modules/adaudit"
	"github.com/utmstack/utmstack/backend/modules/alerts"
	"github.com/utmstack/utmstack/backend/modules/alertscoring"
	"github.com/utmstack/utmstack/backend/modules/appconfig"
	appconfig_connectors "github.com/utmstack/utmstack/backend/modules/appconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/audit"
	"github.com/utmstack/utmstack/backend/modules/billing"
	"github.com/utmstack/utmstack/backend/modules/compliance"
	compliance_connectors "github.com/utmstack/utmstack/backend/modules/compliance/connectors"
	compliance_repository "github.com/utmstack/utmstack/backend/modules/compliance/repository"
	"github.com/utmstack/utmstack/backend/modules/dashboards"
	"github.com/utmstack/utmstack/backend/modules/datasources"
	ns_repository "github.com/utmstack/utmstack/backend/modules/datasources/repository"
	ns_usecase "github.com/utmstack/utmstack/backend/modules/datasources/usecase"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing"
	"github.com/utmstack/utmstack/backend/modules/iam"
	iam_repository "github.com/utmstack/utmstack/backend/modules/iam/repository"
	iam_usecase "github.com/utmstack/utmstack/backend/modules/iam/usecase"
	"github.com/utmstack/utmstack/backend/modules/incidents"
	"github.com/utmstack/utmstack/backend/modules/integrations"
	"github.com/utmstack/utmstack/backend/modules/loganalyzer"
	mcpmod "github.com/utmstack/utmstack/backend/modules/mcp"
	"github.com/utmstack/utmstack/backend/modules/notifications"
	"github.com/utmstack/utmstack/backend/modules/soar"
	"github.com/utmstack/utmstack/backend/modules/socai"
	socai_repository "github.com/utmstack/utmstack/backend/modules/socai/repository"
	"github.com/utmstack/utmstack/backend/modules/storage"
	"github.com/utmstack/utmstack/backend/modules/tenant"
	tenant_connectors "github.com/utmstack/utmstack/backend/modules/tenant/connectors"
	tenant_domain "github.com/utmstack/utmstack/backend/modules/tenant/domain"
	tenant_dto "github.com/utmstack/utmstack/backend/modules/tenant/dto"
	ep_repository "github.com/utmstack/utmstack/backend/modules/eventprocessing/repository"
	"github.com/utmstack/utmstack/backend/modules/threatintel"
	"github.com/utmstack/utmstack/backend/pkg/agentmanager"
	"github.com/utmstack/utmstack/backend/pkg/env"
	jwtpkg "github.com/utmstack/utmstack/backend/pkg/jwt"
	"github.com/utmstack/utmstack/backend/pkg/ratelimit"
	"github.com/utmstack/utmstack/backend/pkg/secret"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
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
	iam             *iam.Module
	audit           *audit.Module
	appconfig       *appconfig.Module
	billing         *billing.Module
	mail            *mail.Module
	compliance      *compliance.Module
	dashboards      *dashboards.Module
	loganalyzer     *loganalyzer.Module
	tenant          *tenant.Module
	alerts          *alerts.Module
	soar            *soar.Module
	datasources     *datasources.Module
	eventProcessing *eventprocessing.Module
	integrations    *integrations.Module
	incidents       *incidents.Module
	notifications   *notifications.Module
	socAI           *socai.Module
	adaudit         *adaudit.Module
	threatIntel     *threatintel.Module
	storage         *storage.Module
	mcp             *mcpmod.Module
	signer          *jwtpkg.Signer
}

func initModules(db *gorm.DB, cfg *config) *modules {
	if cfg.jwtSecret == "" {
		cfg.jwtSecret = cfg.internalKey
	}
	if cfg.jwtSecret == "" {
		_ = catcher.Error("JWT_SECRET (or INTERNAL_KEY) is not set — refusing to start", nil, nil)
		panic("JWT_SECRET or INTERNAL_KEY is required")
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
	auditMod := audit.NewModule(db, joblease.New(db), env.Int("AUDIT_RETENTION_DAYS", 365, false))

	events, err := eventstore.New()
	if err != nil {
		_ = catcher.Error("could not reach the event store; dashboards will not answer", err, nil)
	}
	// The dev licence is read only in dev mode, so a production build cannot be
	// talked into one by setting an environment variable.
	devLicense := ""
	if cfg.devMode {
		devLicense = env.String("DEV_LICENSE", "", false)
	}
	billingMod := billing.NewModule(env.String("UPDATES_DIR", "/updates", false), devLicense)

	if err := tenancy.Register(db, func() bool {
		return billingMod.License().Current().IsMSSP()
	}); err != nil {
		_ = catcher.Error("failed to register tenancy callbacks", err, nil)
		panic(err)
	}
	// ponytail: late-bound lister so configMod can be constructed before tenantMod exists;
	// tenantMod is always set before the first HTTP request reaches the handler.
	var tenantMod *tenant.Module
	tenantLister := func(ctx context.Context) ([]string, error) {
		tenants, _, err := tenantMod.GetTenantUsecase().List(ctx, tenant_dto.Filter{Size: 10000, Status: tenant_domain.StatusActive})
		if err != nil {
			return nil, err
		}
		ids := make([]string, 0, len(tenants))
		for _, t := range tenants {
			ids = append(ids, t.ID.String())
		}
		return ids, nil
	}
	var tenantListerForConfig func(context.Context) ([]string, error)
	configMod := appconfig.NewModule(db, cipher, cfg.uploadDir, func(ctx context.Context) ([]string, error) {
		return tenantListerForConfig(ctx)
	})
	mailMod := mail.NewModule(configMod.Store())
	configMod.SetMailer(mailMod.Service())
	// White-labeling renders only under an Enterprise license (resolved from billing).
	configMod.SetWhiteLabelEntitlement(func() bool {
		return billingMod.License().Current().IsEnterprise()
	})
	brand := configMod.Branding()
	var complianceEventReader compliance_repository.Reader
	if events != nil {
		complianceEventReader = events
	}
	complianceMod := compliance.NewModule(db, complianceEventReader, mailMod.Service(), complianceBranding{uc: brand, uploadDir: cfg.uploadDir},
		func() bool { return billingMod.License().Current().IsEnterprise() }, tenantLister)
	var eventReader dash_usecase.Reader
	if events != nil {
		eventReader = events
	}
	dashboardsMod := dashboards.NewModule(db, eventReader)
	loganalyzerMod := loganalyzer.NewModule(db, events)

	userRepo := iam_repository.NewUserRepository(db)
	rbacRepo := iam_repository.NewRBACRepository(db)
	refreshRepo := iam_repository.NewRefreshTokenRepository(db)
	challengeRepo := iam_repository.NewChallengeRepository(db)
	factorRepo := iam_repository.NewTfaFactorRepository(db, cipher)
	challengeMailer := iam_repository.NewChallengeMailer(mailMod.Service(), mailMod.ConfigRepo(), brand)

	tfaUsecase := iam_usecase.NewTfaUsecase(userRepo, refreshRepo, factorRepo, challengeRepo, challengeMailer, signer, preAuthSigner, refreshTokenTTL, brand)
	idpRepo := iam_repository.NewIdentityProviderRepository(db)
	federationUsecase := iam_usecase.NewFederationUsecase(idpRepo, userRepo, rbacRepo, refreshRepo, signer, cipher, refreshTokenTTL)
	authUsecase := iam_usecase.NewAuthUsecase(userRepo, rbacRepo, refreshRepo, challengeRepo, signer, limiter, refreshTokenTTL, challengeMailer, tfaUsecase, federationUsecase, preAuthSigner, cfg.tfaEnabled)
	userUsecase := iam_usecase.NewUserUsecase(userRepo, rbacRepo, challengeRepo, factorRepo, challengeMailer)
	roleUsecase := iam_usecase.NewRoleUsecase(rbacRepo)
	apiKeyRepo := iam_repository.NewAPIKeyRepository(db)
	apiKeyUsecase := iam_usecase.NewAPIKeyUsecase(apiKeyRepo, userRepo)
	idpUsecase := iam_usecase.NewIdentityProviderUsecase(idpRepo, rbacRepo, cipher)

	alertsMod := alerts.NewModule(db, events, alerts.NewAlertMailer(mailMod.Service(), configMod.Store()))

	agentClient, agentErr := agentmanager.NewClient()
	if agentErr != nil {
		_ = catcher.Error("agentmanager client init failed (alert response rules will not dispatch)", agentErr, nil)
		agentClient = nil
	}

	// SOC-AI client is built here so the SOAR LLM executors can share it.
	// socai.NewModule below reuses the same base URL/key.
	socAIClient := socai.NewSocAIClient(cfg.socAIBaseURL, cfg.internalKey)
	// Notifications module is built early so its usecase can back the SOAR
	// notify executor. Its own dependencies (db + audit logger + leases) are
	// already available at this point.
	notificationsMod := notifications.NewModule(db, auditMod.Logger(), joblease.New(db),
		env.Int("NOTIFICATIONS_READ_RETENTION_DAYS", 30, false),
		env.Int("NOTIFICATIONS_RETENTION_DAYS", 365, false))
	// Incidents module is built early so its usecase can back the SOAR incident
	// executor. Its own deps (db + mail + config + alerts + audit) are already
	// available at this point.
	incidentsMod := incidents.NewModule(
		db,
		incidents.NewIncidentMailer(mailMod.Service(), configMod.Store()),
		incidents.NewAlertsGatewayFromUsecase(alertsMod.GetAlertUsecase()),
		auditMod.Logger(),
	)
	soarMod := soar.NewModule(db, agentClient, signer, cipher, socAIClient, notificationsMod.Producer(), incidentsMod.GetIncidentUsecase(), tenantLister)
	eventProcessingMod := eventprocessing.NewModule(db, events, auditMod.Logger(), cfg.playgroundBaseURL, cfg.internalKey)

	alertsMod.SetCorrelationResolver(eventProcessingMod)

	dsRepo := ns_repository.NewDatasourceRepository(db)
	dsUC := ns_usecase.NewDatasourceUsecase(dsRepo, eventProcessingMod.GetAssetProjectionUsecase())
	// Discovery from ingestion needs the event store, not OpenSearch: the
	// statistics it reads moved there with the rest of the pipeline.
	var dsReconciler *ns_usecase.StatsReconciler
	if reader := ns_repository.NewStatsReader(eventConn(events)); reader != nil {
		dsReconciler = ns_usecase.NewStatsReconciler(dsRepo, reader, joblease.New(db))
	}
	datasourcesMod := datasources.NewModule(dsUC, dsReconciler, agentClient)

	iam_handler.AppBaseURL = env.String("APP_BASE_URL", "", false)

	// Extra purgers: ClickHouse rows and per-tenant filesystem folders.
	// Each subsystem contributes one; failures short-circuit before the SQL purge
	// so the tenant row survives an outage.
	var extraPurgers []tenant_connectors.TenantPurgeFunc
	if events != nil {
		extraPurgers = append(extraPurgers, func(ctx context.Context, id uuid.UUID) error {
			return events.PurgeTenant(ctx, id.String())
		})
	}
	rulesUserDir := filepath.Join(env.String(ep_repository.RulesDirEnv, ep_repository.DefaultRulesDir, false), ep_repository.UserSubdir)
	pipelinesUserDir := filepath.Join(env.String(ep_repository.PipelinesDirEnv, ep_repository.DefaultPipelinesDir, false), ep_repository.UserSubdir)
	extraPurgers = append(extraPurgers,
		func(_ context.Context, id uuid.UUID) error { return os.RemoveAll(filepath.Join(rulesUserDir, id.String())) },
		func(_ context.Context, id uuid.UUID) error {
			return os.RemoveAll(filepath.Join(pipelinesUserDir, id.String()))
		},
	)
	tenantMod = tenant.NewModule(db, userUsecase, extraPurgers...)
	tenantListerForConfig = tenantLister
	iamMod := iam.NewModule(authUsecase, userUsecase, roleUsecase, tfaUsecase, apiKeyUsecase, idpUsecase, federationUsecase, cfg.uploadDir, tenantLister)
	iamMod.SetSessionPurger(iam_usecase.NewSessionPurger(refreshRepo, joblease.New(db)))

	aiUsage := socai_repository.NewUsageRepo(db)
	aiQuota := &socai.AIQuota{
		LimitOf: func(ctx context.Context, tenantID string) (int, error) {
			id, err := uuid.Parse(tenantID)
			if err != nil {
				return 0, err
			}
			t, err := tenantMod.GetTenantUsecase().GetByID(ctx, id)
			if err != nil {
				return 0, err
			}
			return t.Limits.AllowanceOf(cfg.aiRequestLimit), nil
		},
		Consume: aiUsage.Consume,
		Used:    aiUsage.UsedToday,
	}

	socAIMod := socai.NewModule(cfg.socAIBaseURL, cfg.internalKey, cipher,
		env.String("INTEGRATIONS_CONFIG_DIR", "/workdir/pipeline", false),
		env.String("UPDATES_DIR", "/updates", false), aiQuota, joblease.New(db))
	adauditMod := adaudit.NewModule(db)
	storageMod := storage.NewModule(events, env.String("CLICKHOUSE_CONFIG_DIR", "/clickhouse-conf", false))
	threatintelMod := threatintel.NewModule(
		env.String("UPDATES_DIR", "/updates", false),
		env.String("INTEGRATIONS_CONFIG_DIR", "/workdir/pipeline", false),
		cipher,
	)

	integrationsMod := integrations.NewModule(db, cipher,
		env.String("INTEGRATIONS_CONFIG_DIR", "/workdir/pipeline", false),
		dsUC,
		agentClient,
	)

	var mcpModule *mcpmod.Module
	if cfg.mcpEnabled {
		mcpModule = mcpmod.NewModule(&mcpmod.Deps{
			IAM:             iamMod,
			Alerts:          alertsMod,
			AlertScoring:    alertscoring.NewModule(events, agentClient, datasourcesMod.GetDatasourceUsecase()),
			Incidents:       incidentsMod,
			SOAR:            soarMod,
			Compliance:      complianceMod,
			Audit:           auditMod,
			Dashboards:      dashboardsMod,
			LogAnalyzer:     loganalyzerMod,
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
		iam:             iamMod,
		audit:           auditMod,
		appconfig:       configMod,
		billing:         billingMod,
		mail:            mailMod,
		compliance:      complianceMod,
		dashboards:      dashboardsMod,
		loganalyzer:     loganalyzerMod,
		tenant:          tenantMod,
		alerts:          alertsMod,
		soar:            soarMod,
		datasources:     datasourcesMod,
		eventProcessing: eventProcessingMod,
		integrations:    integrationsMod,
		socAI:           socAIMod,
		incidents:       incidentsMod,
		notifications:   notificationsMod,
		adaudit:         adauditMod,
		threatIntel:     threatintelMod,
		storage:         storageMod,
		mcp:             mcpModule,
		signer:          signer,
	}
}

type complianceBranding struct {
	uc        appconfig_connectors.BrandingUsecase
	uploadDir string
}

func (a complianceBranding) ReportBrand(ctx context.Context) compliance_connectors.ReportBrand {
	b, err := a.uc.Get(ctx)
	if err != nil || b == nil || !b.Enabled {
		return compliance_connectors.ReportBrand{}
	}
	logoURL := b.ReportLogoURL
	if logoURL == "" {
		logoURL = b.LogoURL
	}
	return compliance_connectors.ReportBrand{
		Name:      b.ProductName,
		LogoPath:  uploadFilePath(a.uploadDir, logoURL),
		CoverPath: uploadFilePath(a.uploadDir, b.ReportCoverURL),
		AccentHex: b.AccentColor,
	}
}

func uploadFilePath(dir, url string) string {
	if url == "" || dir == "" {
		return ""
	}
	u := strings.TrimPrefix(url, "/")
	u = strings.TrimPrefix(u, "uploads/")
	return filepath.Join(dir, u)
}

// eventConn is the store's connection, or nil when there is no store. The
// reconciler is then not built and datasources simply stop being discovered
// from ingestion, which an install without ClickHouse cannot do anyway.
func eventConn(s *eventstore.Store) driver.Conn {
	if s == nil {
		return nil
	}
	return s.Conn
}
