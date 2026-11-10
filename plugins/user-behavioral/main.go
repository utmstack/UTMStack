package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/store"
	ch "github.com/threatwinds/go-sdk/store/clickhouse"
	"github.com/tidwall/gjson"
)

const (
	pluginName = "com.utmstack.user-behavioral"

	datasetProfile    store.Dataset = "userBehaviorProfile"
	datasetLogs       store.Dataset = "logs"
	datasetHistory    store.Dataset = "userProfileHistory"
	defaultTable                    = "user_behavior_profile"
	defaultLogsTbl                  = "logs"
	defaultHistoryTbl               = "user_profile_history"
)

var (
	zScoreLimit   = 3.0
	impactMedium  = uint32(2)
	impactHigh    = uint32(3)
	maxJudgeConns = 32
	countCacheTTL = time.Minute
	judgeTimeout  = 15 * time.Second
)

var judgeSem chan struct{}

type cacheEntry struct {
	n    int64
	from time.Time
}

type countCache struct {
	sync.Mutex
	entries map[string]cacheEntry
}

var volumeCountCache = countCache{entries: map[string]cacheEntry{}}

func (c *countCache) get(key string, now time.Time) (int64, bool) {
	c.Lock()
	defer c.Unlock()
	e, ok := c.entries[key]
	if !ok || now.Sub(e.from) > countCacheTTL {
		return 0, false
	}
	return e.n, true
}

func (c *countCache) set(key string, n int64, now time.Time) {
	c.Lock()
	defer c.Unlock()
	c.entries[key] = cacheEntry{n: n, from: now}
}

var (
	profileStore store.Store
	logsStore    store.Store
	historyStore store.Store
)

func loadBehavioralConfig() {
	applyBehavioralCfg(plugins.PluginCfg("user_behavioral"))
}

func applyBehavioralCfg(b gjson.Result) {
	if v := b.Get("z_score_limit").Float(); v > 0 {
		zScoreLimit = v
	}
	if b.Get("impact_medium").Exists() {
		if v := b.Get("impact_medium").Int(); v >= 0 && v <= 3 {
			impactMedium = uint32(v)
		}
	}
	if b.Get("impact_high").Exists() {
		if v := b.Get("impact_high").Int(); v >= 0 && v <= 3 {
			impactHigh = uint32(v)
		}
	}
	if v := b.Get("max_concurrent_judgments").Int(); v > 0 {
		maxJudgeConns = int(v)
	}
	if v := b.Get("count_cache_ttl").String(); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			countCacheTTL = d
		}
	}
}

type profile struct {
	TenantId      string   `json:"tenantId"`
	User          string   `json:"user"`
	EventDayMean  float64  `json:"event_day_mean"`
	EventDayStd   float64  `json:"event_day_std"`
	EventDayP95   uint64   `json:"event_day_p95"`
	HourSin       float64  `json:"hour_sin"`
	HourCos       float64  `json:"hour_cos"`
	HourStd       float64  `json:"hour_std"`
	TopIPs        []string `json:"top_ips"`
	MinSamplesOK  bool     `json:"min_samples_ok"`
	SnapshottedAt string   `json:"snapshotted_at"`
}

func main() {
	mode := plugins.GetCfg("plugin_" + pluginName).Env.Mode
	if mode != "worker" {
		return
	}

	cfg := plugins.PluginCfg("clickhouse")
	loadBehavioralConfig()
	judgeSem = make(chan struct{}, maxJudgeConns)

	logsTable := cfg.Get("logsTable").String()
	if logsTable == "" {
		logsTable = defaultLogsTbl
	}

	var err error
	profileStore, err = ch.New(ch.Config{
		Addr:     []string{cfg.Get("host").String() + ":" + cfg.Get("port").String()},
		Database: cfg.Get("database").String(),
		Username: cfg.Get("user").String(),
		Password: cfg.Get("password").String(),
		Tables: map[store.Dataset]string{
			datasetProfile: defaultTable,
		},
		TenantColumn: "tenantId",
		TimeColumn:   "updated_at",
	})
	if err != nil {
		_ = catcher.Error("failed to build the profile store", err, map[string]any{"process": pluginName})
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}

	logsStore, err = ch.New(ch.Config{
		Addr:     []string{cfg.Get("host").String() + ":" + cfg.Get("port").String()},
		Database: cfg.Get("database").String(),
		Username: cfg.Get("user").String(),
		Password: cfg.Get("password").String(),
		Tables: map[store.Dataset]string{
			datasetLogs: logsTable,
		},
		TenantColumn:   "tenantId",
		TimeColumn:     "@timestamp",
		DataTypeColumn: "dataType",
	})
	if err != nil {
		_ = catcher.Error("failed to build the logs store", err, map[string]any{"process": pluginName})
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}

	historyStore, err = ch.New(ch.Config{
		Addr:     []string{cfg.Get("host").String() + ":" + cfg.Get("port").String()},
		Database: cfg.Get("database").String(),
		Username: cfg.Get("user").String(),
		Password: cfg.Get("password").String(),
		Tables: map[store.Dataset]string{
			datasetHistory: defaultHistoryTbl,
		},
		TenantColumn: "tenantId",
		TimeColumn:   "snapshotted_at",
	})
	if err != nil {
		_ = catcher.Error("failed to build the history store", err, map[string]any{"process": pluginName})
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}

	err = plugins.InitAnalysisPlugin(pluginName, analyze)
	if err != nil {
		_ = catcher.Error("failed to initialize analysis plugin", err, map[string]any{"process": pluginName})
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}
}

func analyze(event *plugins.Event, srv plugins.Analysis_AnalyzeServer) error {
	if !isLogin(event) {
		return nil
	}

	ctx, cancel := context.WithTimeout(srv.Context(), judgeTimeout)
	defer cancel()

	select {
	case judgeSem <- struct{}{}:
		defer func() { <-judgeSem }()
	case <-ctx.Done():
		return ctx.Err()
	}

	err := judge(ctx, srv, event)
	if err != nil {
		_ = catcher.Error("user-behavioral: failed to judge login", err, map[string]any{
			"process": pluginName,
			"tenant":  event.TenantId,
			"user":    event.Origin.User,
		})
	}
	return err
}

func isLogin(event *plugins.Event) bool {
	if event.Origin == nil || event.Origin.User == "" {
		return false
	}
	switch event.DataType {
	case "wineventlog":
		code := event.Log["eventCode"]
		return code.GetStringValue() == "4624" || code.GetNumberValue() == 4624
	case "linux":
		id := event.Log["syslogIdentifier"].GetStringValue()
		if id == "sshd" || id == "sshd-session" {
			return event.ActionResult == "success"
		}
		return id == "auditd" && event.Action == "USER_LOGIN" &&
			event.ActionResult == "success"
	case "azure":
		return strings.EqualFold(event.Log["azureKind"].GetStringValue(), "signin")
	case "o365":
		return event.Log["RecordType"].GetStringValue() == "15" &&
			event.ActionResult == "success"
	case "aws":
		return event.Log["eventSource"].GetStringValue() == "signin.amazonaws.com" &&
			event.Log["eventName"].GetStringValue() == "ConsoleLogin"
	case "google":
		return event.Log["protoPayloadServiceName"].GetStringValue() == "login.googleapis.com"
	default:
		return false
	}
}

func judge(ctx context.Context, srv plugins.Analysis_AnalyzeServer, event *plugins.Event) error {
	now := time.Now()
	atTime := eventTime(event, now)
	p, source, found, readErr := loadProfileOrHistory(ctx, event.TenantId, event.Origin.User)
	if readErr != nil {
		message := "user-behavioral: profile read failed"
		if source == "history" {
			message = "user-behavioral: history read failed"
		}
		catcher.Warn(message, map[string]any{
			"process": pluginName,
			"tenant":  event.TenantId,
			"user":    event.Origin.User,
			"error":   readErr.Error(),
		})
	}
	if !found || p == nil {
		return ctx.Err()
	}
	if source == "profile" && !p.MinSamplesOK {
		return nil
	}
	var details []string

	if anomalousHour(p, atTime) {
		details = append(details, hourAnomalyDetail(atTime, p))
	}

	if ip := event.Origin.Ip; ip != "" && len(p.TopIPs) > 0 && !knownIP(ip, p.TopIPs) {
		details = append(details, ipAnomalyDetail(ip))
	}

	if source == "profile" && p.EventDayMean > 0 {
		if n, ok, countErr := volumeCount(ctx, event); countErr != nil {
			catcher.Warn("user-behavioral: volume count failed; abnormal-volume criterion unavailable", map[string]any{
				"process": pluginName,
				"tenant":  event.TenantId,
				"user":    event.Origin.User,
				"error":   countErr.Error(),
			})
		} else if ok {
			if todayVolumeExceeded(float64(n), *p) {
				details = append(details, volumeAnomalyDetail(n, p.EventDayMean))
			}
		}
	}

	if err := ctx.Err(); err != nil {
		return err
	}
	if impact := loginSeverity(len(details)); impact > 0 {
		desc := describeAnomaly(event.Origin.User, source, p.SnapshottedAt, details)
		return sendAlert(srv, event, "User Behavior Anomaly", desc, impact)
	}
	return nil
}

func hourAnomalyDetail(at time.Time, p *profile) string {
	minutes := int(math.Round(hourMean(p)*60)) % (24 * 60)
	return fmt.Sprintf("Logged in at %s, outside the usual login time around %02d:%02d UTC.",
		at.Format("15:04 MST"), minutes/60, minutes%60)
}

func ipAnomalyDetail(ip string) string {
	return fmt.Sprintf("Logged in from %s, an unfamiliar source IP address.", ip)
}

func volumeAnomalyDetail(count int64, mean float64) string {
	return fmt.Sprintf("%d events were recorded today (UTC), compared with a usual average of %.1f per day.", count, mean)
}

func describeAnomaly(user, source, snapshot string, details []string) string {
	prefix := fmt.Sprintf("User %s: ", user)
	if source == "history" {
		if snapshot != "" {
			prefix += fmt.Sprintf("Compared with the profile snapshot from %s. ", snapshot)
		} else {
			prefix += "Compared with the user's historical profile. "
		}
	}
	return prefix + strings.Join(details, " ")
}

func loginSeverity(criteria int) uint32 {
	switch {
	case criteria >= 3:
		return impactHigh
	case criteria == 2:
		return impactMedium
	default:
		return 0
	}
}

func utcDayStart(now time.Time) time.Time {
	u := now.UTC()
	return time.Date(u.Year(), u.Month(), u.Day(), 0, 0, 0, 0, time.UTC)
}

func volumeCount(ctx context.Context, event *plugins.Event) (int64, bool, error) {
	key := event.TenantId + "\x00" + event.Origin.User
	now := time.Now()
	if n, ok := volumeCountCache.get(key, now); ok {
		return n, true, nil
	}
	n, err := logsStore.Count(ctx,
		store.Scope{Tenant: event.TenantId, Dataset: datasetLogs, From: utcDayStart(now), To: now},
		[]store.Filter{{Field: "origin.user", Op: store.OpEq, Value: event.Origin.User}})
	if err != nil {
		return 0, false, err
	}
	volumeCountCache.set(key, n, now)
	return n, true, nil
}

func todayVolumeExceeded(n float64, p profile) bool {
	if n > p.EventDayMean+zScoreLimit*p.EventDayStd {
		return true
	}
	return p.EventDayP95 > 0 && n > float64(p.EventDayP95)
}

func eventTime(event *plugins.Event, fallback time.Time) time.Time {
	if event.Timestamp == "" {
		return fallback
	}
	t, err := time.Parse(time.RFC3339Nano, event.Timestamp)
	if err != nil {
		return fallback
	}
	return t.UTC()
}

func loadProfileOrHistory(ctx context.Context, tenant, user string) (*profile, string, bool, error) {
	p, found, err := loadProfile(ctx, tenant, user)
	if err != nil || found {
		return p, "profile", found, err
	}
	h, err := loadHistory(ctx, tenant, user)
	if err != nil {
		return nil, "history", false, err
	}
	if h == nil {
		return nil, "", false, nil
	}
	return h, "history", true, nil
}

func loadHistory(ctx context.Context, tenant, user string) (*profile, error) {
	cutoff := hotCutoff(ctx)
	from := cutoff
	if !from.IsZero() {
		from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
	}
	rows, err := historyStore.FetchN(ctx, store.Scope{Tenant: tenant, Dataset: datasetHistory, From: from},
		[]store.Filter{{Field: "user", Op: store.OpEq, Value: user}}, 1)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	var h profile
	if err := json.Unmarshal(rows[0], &h); err != nil {
		return nil, err
	}
	return &h, nil
}

func hotCutoff(ctx context.Context) time.Time {
	r, err := historyStore.Retention(ctx, datasetHistory)
	return historyCutoff(r, err, time.Now())
}

func historyCutoff(ret store.Retention, err error, now time.Time) time.Time {
	if err != nil {
		return time.Time{}
	}
	if ret.ColdAfter > 0 {
		return now.Add(-ret.ColdAfter)
	}
	return time.Time{}
}

func loadProfile(ctx context.Context, tenant, user string) (*profile, bool, error) {
	rows, err := profileStore.FetchN(ctx, store.Scope{Tenant: tenant, Dataset: datasetProfile},
		[]store.Filter{{Field: "user", Op: store.OpEq, Value: user}}, 1)
	if err != nil {
		return nil, false, err
	}
	if len(rows) == 0 {
		return nil, false, nil
	}
	var p profile
	if err := json.Unmarshal(rows[0], &p); err != nil {
		return nil, false, err
	}
	return &p, true, nil
}

func anomalousHour(p *profile, at time.Time) bool {
	mean := hourMean(p)
	dist := hourDist(at.Hour(), mean)
	if p.HourStd <= 0 {
		return false
	}
	return dist > zScoreLimit*hourSpread(p)
}

func hourSpread(p *profile) float64 {
	return p.HourStd * 12 / math.Pi
}

func hourMean(p *profile) float64 {
	h := math.Atan2(p.HourSin, p.HourCos) * 12 / math.Pi
	if h < 0 {
		h += 24
	}
	return h
}

func hourDist(h int, mean float64) float64 {
	d := math.Abs(float64(h) - mean)
	return math.Min(d, 24-d)
}

func knownIP(ip string, tops []string) bool {
	if !isPublic(ip) {
		return true
	}
	for _, t := range tops {
		if t == ip || sameSubnet24(ip, t) {
			return true
		}
	}
	return false
}

func isPublic(ip string) bool {
	p := net.ParseIP(ip)
	if p == nil {
		return false
	}
	for _, cidr := range []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "169.254.0.0/16", "127.0.0.0/8", "224.0.0.0/24"} {
		if _, b, err := net.ParseCIDR(cidr); err == nil && b.Contains(p) {
			return false
		}
	}
	return true
}

func sameSubnet24(a, b string) bool {
	pa, pb := net.ParseIP(a).To4(), net.ParseIP(b).To4()
	if pa == nil || pb == nil {
		return false
	}
	return pa[0] == pb[0] && pa[1] == pb[1] && pa[2] == pb[2]
}

func sendAlert(srv plugins.Analysis_AnalyzeServer, event *plugins.Event, name, desc string, impact uint32) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	severity := "low"
	switch impact {
	case 2:
		severity = "medium"
	case 3:
		severity = "high"
	}
	alert := &plugins.Alert{
		Id:          uuid.NewString(),
		Timestamp:   now,
		LastUpdate:  now,
		Name:        name,
		TenantId:    event.TenantId,
		TenantName:  event.TenantName,
		DataSource:  event.DataSource,
		DataType:    event.DataType,
		Category:    "User Behavioral Profiling",
		Description: desc,
		Severity:    severity,
		Impact:      &plugins.Impact{Confidentiality: impact},
		ImpactScore: impact,
		Adversary:   event.Origin,
		Target:      event.Target,
		LastEvent:   event,
	}
	return srv.Send(alert)
}
