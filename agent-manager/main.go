package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/agent-manager/agent"
	"github.com/utmstack/UTMStack/agent-manager/authcache"
	"github.com/utmstack/UTMStack/agent-manager/database"
	"github.com/utmstack/UTMStack/agent-manager/recovery"
	"github.com/utmstack/UTMStack/agent-manager/updates"
)

type recoveryProvider struct {
	server       *agent.AgentService
	lastSeenServ *agent.LastSeenService
}

func (p *recoveryProvider) GetStream(agentID uint) (recovery.AgentStream, bool) {
	p.server.AgentStreamMutex.Lock()
	defer p.server.AgentStreamMutex.Unlock()
	s, ok := p.server.AgentStreamMap[agentID]
	if !ok {
		return nil, false
	}
	return s, true
}

func (p *recoveryProvider) IsOnline(agentID uint) bool {
	if p.lastSeenServ == nil {
		return false
	}
	st, _, err := p.lastSeenServ.GetLastSeenStatus(agentID, "agent")
	if err != nil {
		return false
	}
	return st == agent.Status_ONLINE
}

func main() {
	catcher.Info("Starting Agent Manager", map[string]any{"process": "agent-manager"})

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	if err := database.MigrateDatabase(); err != nil {
		_ = catcher.Error("failed to migrate database", err, map[string]any{"process": "agent-manager"})
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}

	if err := agent.InitAgentService(); err != nil {
		_ = catcher.Error("failed to init agent service", err, map[string]any{"process": "agent-manager"})
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}

	go agent.InitCollectorService()
	agent.InitLastSeenService()

	// Publishing keys where the log input reads them is what makes a deletion
	// take effect at once instead of when the input's own copy expires.
	agent.AuthCache = authcache.New(
		os.Getenv("REDIS_ADDR"),
		os.Getenv("REDIS_PASSWORD"),
		0,
	)
	defer func() { _ = agent.AuthCache.Close() }()
	go agent.AuthCache.Run(ctx, agent.AuthSnapshot)

	recovery.SetStreamProvider(&recoveryProvider{
		server:       agent.AgentServ,
		lastSeenServ: agent.LastSeenServ,
	})
	agent.RegisterRecoveryHooks(
		recovery.OnAgentConnect,
		recovery.OnAgentRegister,
		recovery.OnAgentUpdate,
		recovery.OnCommandResult,
		func(agentID uint) sync.Locker { return recovery.StreamMutex.For(agentID) },
	)
	if err := recovery.Init(ctx, database.GetDB().GormDB()); err != nil {
		_ = catcher.Error("failed to init recovery", err, map[string]any{"process": "agent-manager"})
	}

	go updates.InitUpdatesManager()
	go agent.StartGrpcServer()

	<-ctx.Done()
	catcher.Info("Shutdown signal received; draining recovery dispatches", map[string]any{"process": "agent-manager"})
	if err := recovery.Shutdown(15 * time.Second); err != nil {
		_ = catcher.Error("recovery shutdown error", err, map[string]any{"process": "agent-manager"})
	}
	catcher.Info("Agent Manager shut down cleanly", map[string]any{"process": "agent-manager"})
}
