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
	defaultTable                    = "utmstack.user_behavior_profile"
	defaultLogsTbl                  = "utmstack.logs"
	defaultHistoryTbl               = "utmstack.user_profile_history"
)

var (
	zScoreLimit   = 3.0
	scanInterval  = 5 * time.Minute
	impactMedium  = uint32(2)
	impactHigh    = uint32(3)
	maxJudgeConns = 32
	countCacheTTL = time.Minute
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
	sendMu       sync.RWMutex
	sendServer   plugins.Analysis_AnalyzeServer
	profileStore store.Store
	logsStore    store.Store
	historyStore store.Store
	tenantNames  sync.Map
	volumeFired  sync.Map
)

func loadBehavioralConfig() {
	applyBehavioralCfg(plugins.PluginCfg("user_behavioral"))
}

func applyBehavioralCfg(b gjson.Result) {
	if v := b.Get("z_score_limit").Float(); v > 0 {
		zScoreLimit = v
	}
	if v := b.Get("scan_interval").String(); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			scanInterval = d
		}
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

	go volumeScanLoop()
}

func volumeScanLoop() {
	ticker := time.NewTicker(scanInterval)
	defer ticker.Stop()
	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
		scanVolume(ctx)
		cancel()
	}
}

func scanVolume(ctx context.Context) {
	now := time.Now()
	bucket := scanBucket(now)
	pruneVolumeFired(bucket)
	rows, err := profileStore.FetchN(ctx, store.Scope{Tenant: store.AllTenants, Dataset: datasetProfile},
		[]store.Filter{{Field: "min_samples_ok", Op: store.OpEq, Value: 1},
			{Field: "event_day_mean", Op: store.OpGt, Value: 0}}, 10000)
	if err != nil {
		catcher.Warn("user-behavioral: volume scan failed", map[string]any{"process": pluginName, "error": err.Error()})
		return
	}
	for _, raw := range rows {
		var p profile
		if err := json.Unmarshal(raw, &p); err != nil {
			continue
		}
		if p.EventDayMean == 0 {
			continue
		}
		key := p.TenantId + "\x00" + p.User
		if !scanDue(&volumeFired, key, bucket) {
			continue
		}
		n, err := logsStore.Count(ctx,
			store.Scope{Tenant: p.TenantId, Dataset: datasetLogs, From: utcDayStart(now), To: now},
			[]store.Filter{{Field: "origin.user", Op: store.OpEq, Value: p.User}})
		if err != nil {
			continue
		}
		if todayVolumeExceeded(float64(n), p) {
			if sendScanAlert(p, n) {
				volumeFired.Store(key, bucket)
			}
		}
	}
}

func analyze(event *plugins.Event, srv plugins.Analysis_AnalyzeServer) error {
	if event.TenantName != "" {
		tenantNames.Store(event.TenantId, event.TenantName)
	}
	sendMu.Lock()
	sendServer = srv
	sendMu.Unlock()
	if !isLogin(event) {
		return nil
	}

	go func() {
		judgeSem <- struct{}{}
		defer func() { <-judgeSem }()
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		judge(ctx, srv, event)
	}()

	return nil
}

func isLogin(event *plugins.Event) bool {
	if event.Origin == nil || event.Origin.User == "" {
		return false
	}
	switch event.DataType {
	case "wineventlog":
		code := event.Log["eventCode"].GetStringValue()
		return code == "4624"
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

func judge(ctx context.Context, srv plugins.Analysis_AnalyzeServer, event *plugins.Event) {
	now := time.Now()
	atTime := eventTime(event, now)
	p, source, found := loadProfileOrHistory(ctx, event.TenantId, event.Origin.User)
	if !found || p == nil {
		return
	}
	if source == "profile" && !p.MinSamplesOK {
		return
	}
	prefix := ""
	if source == "history" {
		prefix = "vs profile snapshot " + p.SnapshottedAt + ": "
	}

	var details []string

	if anomalousHour(p, atTime) {
		details = append(details, "unusual hour: logged in at "+
			prefix+atTime.Format("15:04")+
			fmt.Sprintf(", usual mean %.1fh (spread %.1fh)", hourMean(p), p.HourStd))
	}

	if ip := event.Origin.Ip; ip != "" && len(p.TopIPs) > 0 && !knownIP(ip, p.TopIPs) {
		details = append(details, "unfamiliar IP: "+prefix+fmt.Sprintf(
			"logged in from %s, not among the usual %d source IPs", ip, len(p.TopIPs)))
	}

	if source == "profile" && p.EventDayMean > 0 {
		if n, ok := volumeCount(ctx, event); ok {
			if todayVolumeExceeded(float64(n), *p) {
				details = append(details, fmt.Sprintf(
					"abnormal volume: %d events today vs usual %.1f/day (p95 %d)",
					n, p.EventDayMean, p.EventDayP95))
			}
		}
	}

	if impact := loginSeverity(len(details)); impact > 0 {
		desc := prefix + describeAnomaly(event.Origin.User, source, len(details), details)
		_ = sendAlert(srv, event, "UserBehaviorAnomaly", desc, impact)
	}
}

func criteriaTotalFor(source string) int {
	if source == "history" {
		return 2
	}
	return 3
}

func describeAnomaly(user, source string, hit int, details []string) string {
	return fmt.Sprintf("%s matches %d of %d behavioral criteria: %s.",
		user, hit, criteriaTotalFor(source), strings.Join(details, "; "))
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

func pruneVolumeFired(current string) {
	volumeFired.Range(func(k, v any) bool {
		if b, ok := v.(string); ok && b != current {
			volumeFired.Delete(k)
		}
		return true
	})
}

func scanDue(fired *sync.Map, key, bucket string) bool {
	if v, ok := fired.Load(key); ok && v == bucket {
		return false
	}
	return true
}

func utcDayStart(now time.Time) time.Time {
	u := now.UTC()
	return time.Date(u.Year(), u.Month(), u.Day(), 0, 0, 0, 0, time.UTC)
}

func scanBucket(now time.Time) string {
	u := now.UTC()
	start := time.Date(u.Year(), u.Month(), u.Day(), (u.Hour()/6)*6, 0, 0, 0, time.UTC)
	return start.Format("2006-01-02T15")
}

func volumeCount(ctx context.Context, event *plugins.Event) (int64, bool) {
	key := event.TenantId + "\x00" + event.Origin.User
	now := time.Now()
	if n, ok := volumeCountCache.get(key, now); ok {
		return n, true
	}
	n, err := logsStore.Count(ctx,
		store.Scope{Tenant: event.TenantId, Dataset: datasetLogs, From: utcDayStart(now), To: now},
		[]store.Filter{{Field: "origin.user", Op: store.OpEq, Value: event.Origin.User}})
	if err != nil {
		return 0, false
	}
	volumeCountCache.set(key, n, now)
	return n, true
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

func loadProfileOrHistory(ctx context.Context, tenant, user string) (*profile, string, bool) {
	p, found, err := loadProfile(ctx, tenant, user)
	if err != nil || found {
		return p, "profile", found
	}
	h, err := loadHistory(ctx, tenant, user)
	if err != nil {
		catcher.Warn("user-behavioral: history read failed", map[string]any{
			"process": pluginName,
			"tenant":  tenant,
			"user":    user,
		})
		return nil, "", false
	}
	if h == nil {
		return nil, "", false
	}
	return h, "history", true
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
	return dist > zScoreLimit*p.HourStd
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

func sendScanAlert(p profile, n int64) bool {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	tenantName, _ := tenantNames.Load(p.TenantId)
	alert := &plugins.Alert{
		Id:         uuid.NewString(),
		Timestamp:  now,
		LastUpdate: now,
		Name:       "UserEventVolumeAnomaly",
		TenantId:   p.TenantId,
		TenantName: fmt.Sprint(tenantName),
		DataType:   "utmstack",
		Category:   "User Behavioral Profiling",
		Description: fmt.Sprintf("%s generated %d events today, far above their 90-day daily norm (mean %.1f, p95 %d).",
			p.User, n, p.EventDayMean, p.EventDayP95),
		Impact:      &plugins.Impact{Confidentiality: impactMedium},
		ImpactScore: impactMedium,
		Adversary:   &plugins.Side{User: p.User},
	}
	sendMu.RLock()
	srv := sendServer
	sendMu.RUnlock()
	if srv == nil {
		return false
	}
	return srv.Send(alert) == nil
}

func sendAlert(srv plugins.Analysis_AnalyzeServer, event *plugins.Event, name, desc string, impact uint32) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
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
		Impact:      &plugins.Impact{Confidentiality: impact},
		ImpactScore: impact,
		Adversary:   event.Origin,
		Target:      event.Target,
		LastEvent:   event,
	}
	return srv.Send(alert)
}
