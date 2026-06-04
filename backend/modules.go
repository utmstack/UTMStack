package main

import (
	"time"

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
	"github.com/utmstack/utmstack/backend/modules/modulesconfig"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/modulekinds"
	mcfg_repository "github.com/utmstack/utmstack/backend/modules/modulesconfig/repository"
	mcfg_usecase "github.com/utmstack/utmstack/backend/modules/modulesconfig/usecase"
	"github.com/utmstack/utmstack/backend/modules/notifications"
	opensearchgw "github.com/utmstack/utmstack/backend/modules/opensearch"
	"github.com/utmstack/utmstack/backend/modules/soar"
	"github.com/utmstack/utmstack/backend/modules/socai"
	"github.com/utmstack/utmstack/backend/pkg/agentmanager"
	"github.com/utmstack/utmstack/backend/pkg/env"
	"github.com/utmstack/utmstack/backend/pkg/eventprocessor"
	jwtpkg "github.com/utmstack/utmstack/backend/pkg/jwt"
	"github.com/utmstack/utmstack/backend/pkg/logger"
	ospkg "github.com/utmstack/utmstack/backend/pkg/opensearch"
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
	modulesconfig     *modulesconfig.Module
	opensearchGateway *opensearchgw.Module
	incidents         *incidents.Module
	notifications     *notifications.Module
	socAI             *socai.Module
	signer            *jwtpkg.Signer
}

func initModules(db *gorm.DB, cfg *config) *modules {
	if cfg.jwtSecret == "" {
		logger.Critical("JWT_SECRET is not set — refusing to start")
		panic("JWT_SECRET is required")
	}
	if cfg.encryptionKey == "" {
		logger.Critical("ENCRYPTION_KEY is not set — refusing to start")
		panic("ENCRYPTION_KEY is required")
	}
	cipher, err := secret.NewCipher(cfg.encryptionKey)
	if err != nil {
		logger.Critical("failed to init cipher: " + err.Error())
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

	osCfg := ospkg.Config{
		Host:               cfg.esHost,
		Port:               cfg.esPort,
		User:               cfg.esUser,
		Password:           cfg.esPassword,
		UseTLS:             true,
		InsecureSkipVerify: env.Bool("OPENSEARCH_TLS_SKIP_VERIFY", false),
	}
	osClient, osErr := ospkg.NewClient(osCfg)
	if osErr != nil {
		logger.Error("opensearch client init failed: " + osErr.Error())
		osClient = nil
	}

	alertsMod := alerts.NewModule(db, osClient, env.Bool("ALERTS_SCHEDULER_ENABLED", false))

	agentClient, agentErr := agentmanager.NewClient()
	if agentErr != nil {
		logger.Error("agentmanager client init failed (alert response rules will not dispatch): " + agentErr.Error())
		agentClient = nil
	}

	soarMod := soar.NewModule(db, agentClient, signer)
	collectorsMod := collectors.NewModule(db, agentClient)
	datainputMod := datainput.NewModule(db)
	correlationMod := correlation.NewModule(db, auditMod.Logger(), datainputMod.GetReader())
	logstashMod := logstash.NewModule(db, osClient, auditMod.Logger())
	indexpatternMod := indexpattern.NewModule(db, osCfg)

	eventProcClient := eventprocessor.NewClient(cfg.eventProcessorHost, cfg.eventProcessorPort, cfg.internalKey)
	mcfgModuleRepo := mcfg_repository.NewModuleRepository(db)
	mcfgGroupRepo := mcfg_repository.NewGroupRepository(db)
	mcfgConfigRepo := mcfg_repository.NewConfigRepository(db)
	mcfgIdxToggler := mcfg_repository.NewIndexPatternToggler(db)
	mcfgLsToggler := mcfg_repository.NewLogstashFilterToggler(db)
	mcfgMenuToggler := mcfg_repository.NewNoopMenuToggler()
	mcfgFactory := mcfg_usecase.NewModuleFactory()
	modulekinds.RegisterAll(mcfgFactory)
	mcfgConfigUC := mcfg_usecase.NewConfigUsecase(mcfgConfigRepo, mcfgModuleRepo, mcfgFactory, cipher, eventProcClient)
	mcfgGroupUC := mcfg_usecase.NewGroupUsecase(mcfgGroupRepo, mcfgConfigRepo, mcfgModuleRepo, mcfgFactory, cipher, eventProcClient)
	mcfgModuleUC := mcfg_usecase.NewModuleUsecase(mcfgModuleRepo, mcfgFactory, mcfgIdxToggler, mcfgLsToggler, mcfgMenuToggler, eventProcClient)
	modulesconfigMod := modulesconfig.NewModule(mcfgModuleUC, mcfgGroupUC, mcfgConfigUC, mcfgFactory)

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
		modulesconfig:     modulesconfigMod,
		opensearchGateway: opensearchgw.NewModule(osClient),
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
