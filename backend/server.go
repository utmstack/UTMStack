package main

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/adaudit"
	"github.com/utmstack/utmstack/backend/modules/alerts"
	"github.com/utmstack/utmstack/backend/modules/appconfig"
	"github.com/utmstack/utmstack/backend/modules/audit"
	"github.com/utmstack/utmstack/backend/modules/billing"
	"github.com/utmstack/utmstack/backend/modules/compliance"
	"github.com/utmstack/utmstack/backend/modules/dashboards"
	datasources "github.com/utmstack/utmstack/backend/modules/datasources"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing"
	"github.com/utmstack/utmstack/backend/modules/iam"
	"github.com/utmstack/utmstack/backend/modules/incidents"
	"github.com/utmstack/utmstack/backend/modules/integrations"
	"github.com/utmstack/utmstack/backend/modules/loganalyzer"
	mcpmod "github.com/utmstack/utmstack/backend/modules/mcp"
	"github.com/utmstack/utmstack/backend/modules/notifications"
	"github.com/utmstack/utmstack/backend/modules/soar"
	"github.com/utmstack/utmstack/backend/modules/socai"
	"github.com/utmstack/utmstack/backend/modules/storage"
	"github.com/utmstack/utmstack/backend/modules/tenant"
	tenant_domain "github.com/utmstack/utmstack/backend/modules/tenant/domain"
	tenant_dto "github.com/utmstack/utmstack/backend/modules/tenant/dto"
	"github.com/utmstack/utmstack/backend/modules/threatintel"
	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

func initHTTPServer(cfg *config) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(requestLogger())
	catcher.Info(fmt.Sprintf("CORS init: devMode=%t (origins matching http://localhost will be allowed only when devMode=true)", cfg.devMode), nil)
	// A SAML assertion arrives as a cross-site form POST that the provider itself
	// caused, and the browser sends Origin: null for it. CORS exists to police
	// what scripts may fetch, not top-level navigations, so the sign-in routes
	// are outside it — the assertion's own signature is what authorises them.
	corsGuard := cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			allowed := cfg.devMode && strings.HasPrefix(origin, "http://localhost")
			if !allowed {
				catcher.Warn(fmt.Sprintf("CORS rejected origin %q (devMode=%t) — request will be aborted with 403", origin, cfg.devMode), nil)
			}
			return allowed
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
	engine.Use(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/v1/sso/") {
			c.Next()
			return
		}
		corsGuard(c)
	})
	return engine
}

var logSkipPaths = map[string]bool{
	"/api/v1/health": true,
}

func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		if logSkipPaths[c.Request.URL.Path] {
			return
		}
		if status := c.Writer.Status(); status >= 400 {
			catcher.Error("http", fmt.Errorf("status %d", status), map[string]any{
				"method":     c.Request.Method,
				"path":       c.Request.URL.Path,
				"status":     status,
				"latency_ms": time.Since(start).Milliseconds(),
				"client_ip":  c.ClientIP(),
			})
		}
	}
}

func registerRoutes(engine *gin.Engine, m *modules, cfg *config) {
	api := engine.Group("/api/v1")

	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	for _, sub := range []string{"avatars", "branding"} {
		if err := os.MkdirAll(filepath.Join(cfg.uploadDir, sub), 0o755); err != nil {
			_ = catcher.Error("failed to create upload dir", err, nil)
		}
	}
	engine.Static("/uploads", cfg.uploadDir)

	apiKeyAuth := func(ctx context.Context, key, ip string) (*middleware.Actor, error) {
		if !m.billing.License().Current().IsEnterprise() {
			return nil, fmt.Errorf("api key authentication requires an enterprise license")
		}
		r, err := m.iam.GetAPIKeyUsecase().Authenticate(ctx, key, ip)
		if err != nil {
			return nil, err
		}
		return &middleware.Actor{
			UserID:      r.UserID,
			Email:       r.Email,
			Roles:       r.Roles,
			Permissions: r.Permissions,
			TenantID:    r.TenantID.String(),
		}, nil
	}
	supportOf := func(ctx context.Context, tenantID string) (string, error) {
		tenantUUID, err := uuid.Parse(tenantID)
		if err != nil {
			return "", err
		}
		t, err := m.tenant.GetTenantUsecase().GetByID(ctx, tenantUUID)
		if err != nil {
			return "", err
		}
		return string(t.SupportAccess), nil
	}
	userAuth := middleware.Authenticate(m.signer, apiKeyAuth, cfg.internalKey, supportOf)

	isEnterprise := func() bool { return m.billing.License().Current().IsEnterprise() }
	enterprise := middleware.RequireEnterprise(isEnterprise)
	enterpriseLicense := middleware.RequireEnterpriseLicense(isEnterprise)
	mssp := middleware.RequireMSSP(func() bool {
		return m.billing.License().Current().IsMSSP()
	})
	platform := middleware.RequirePlatform()

	api.Use(middleware.ResolveTenant(
		func() bool { return m.billing.License().Current().IsMSSP() },
		cfg.internalKey,
		func(ctx context.Context, host string) (string, string, error) {
			t, err := m.tenant.GetTenantUsecase().ResolveDomain(ctx, host)
			if err != nil {
				return "", "", err
			}
			return t.ID.String(), string(t.SupportAccess), nil
		},
	))

	tenantLister := func(ctx context.Context) ([]string, error) {
		tenants, _, err := m.tenant.GetTenantUsecase().List(ctx, tenant_dto.Filter{Size: 10000, Status: tenant_domain.StatusActive})
		if err != nil {
			return nil, err
		}
		ids := make([]string, 0, len(tenants))
		for _, t := range tenants {
			ids = append(ids, t.ID.String())
		}
		return ids, nil
	}

	iam.RegisterRoutes(api, m.iam, userAuth, enterprise, enterpriseLicense, platform)
	tenant.RegisterRoutes(api, m.tenant, userAuth, mssp, platform)
	audit.RegisterRoutes(api, m.audit, userAuth)
	appconfig.RegisterRoutes(api, m.appconfig, userAuth, enterprise, platform)
	billing.RegisterRoutes(api, m.billing, userAuth)
	alerts.RegisterRoutes(api, m.alerts, userAuth)
	soar.RegisterRoutes(api, m.soar, userAuth, apiKeyAuth, platform)
	eventprocessing.RegisterRoutes(api, m.eventProcessing, userAuth, platform, tenantLister)
	compliance.RegisterRoutes(api, m.compliance, userAuth, platform)
	dashboards.RegisterRoutes(api, m.dashboards, userAuth)
	loganalyzer.RegisterRoutes(api, m.loganalyzer, userAuth)
	integrations.RegisterRoutes(api, m.integrations, userAuth)
	incidents.RegisterRoutes(api, m.incidents, userAuth)
	notifications.RegisterRoutes(api, m.notifications, userAuth)
	socai.RegisterRoutes(api, m.socAI, userAuth)
	threatintel.RegisterRoutes(api, m.threatIntel, userAuth)
	storage.RegisterRoutes(api, m.storage, userAuth)
	datasources.RegisterRoutes(api, m.datasources, userAuth)
	adaudit.RegisterRoutes(api, m.adaudit, userAuth)
	if m.mcp != nil {
		mcpmod.RegisterRoutes(api, m.mcp, userAuth)
	}
}

// startServer starts the HTTP server and blocks until appCtx is cancelled
// (i.e. until SIGINT/SIGTERM is received in main).
// Background goroutines (e.g. the alerts scheduler) share the same appCtx and
// will stop when main cancels it — before the HTTP server shuts down.
func startServer(engine *gin.Engine, cfg *config, appCtx context.Context) {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.appPort),
		Handler: engine,
	}

	go func() {
		catcher.Info(fmt.Sprintf("Backend server starting on port %d", cfg.appPort), nil)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			_ = catcher.Error("server error", err, nil)
			panic(err)
		}
	}()

	// Block until the signal-derived context is cancelled.
	<-appCtx.Done()

	catcher.Info("shutting down server...", nil)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		_ = catcher.Error("server forced to shutdown", err, nil)
	}
	catcher.Info("server stopped", nil)
}
