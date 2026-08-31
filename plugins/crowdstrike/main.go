package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/crowdstrike/gofalcon/falcon"
	"github.com/crowdstrike/gofalcon/falcon/client"
	"github.com/crowdstrike/gofalcon/falcon/client/event_streams"
	"github.com/crowdstrike/gofalcon/falcon/models"
	"github.com/crowdstrike/gofalcon/falcon/models/streaming_models"
	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/plugins/shared/coordination"
)

const (
	defaultTenant      = "ce66672c-e36d-4761-a8c8-90058fee1a24"
	urlCheckConnection = "https://falcon.crowdstrike.com"
	reconnectDelay     = 5 * time.Second
)

type activeStream struct {
	ctx       context.Context
	cancel    context.CancelFunc
	processor *CrowdStrikeProcessor
	// groupKey is ModuleGroup.Key() stored unmodified: the lease key, cursor key
	// and event identities are all derived from it.
	groupKey   string
	dataSource string
	cursor     *cursorState
	wg         sync.WaitGroup
}

var (
	activeStreams   = make(map[string]*activeStream)
	activeStreamsMu sync.RWMutex
)

// coordinationRetryDelay paces retries while coordination is unreachable.
const coordinationRetryDelay = 5 * time.Second

func main() {
	mode := plugins.GetCfg(processName).Env.Mode
	if mode != "worker" {
		return
	}

	if err := connectionChecker(urlCheckConnection); err != nil {
		_ = catcher.Error("Failed to establish connectivity, plugin will not start", err, map[string]any{
			"process": processName,
		})
		return
	}

	go StartConfigurationSystem()

	for i := 0; i < 2*runtime.NumCPU(); i++ {
		go func() {
			plugins.SendLogsFromChannel(pluginName)
		}()
	}

	ctx := context.Background()

	// Blocks until coordination succeeds and never falls back to ingesting
	// without it: an appId names one server-side subscription, so two workers on
	// the same unit would fight over it instead of splitting the feed.
	setup, err := coordination.SetupWithRetry(ctx, coordination.DialLeasePath(processName), coordinationRetryDelay, func(err error) {
		_ = catcher.Error("coordination setup failed, waiting to retry", err, map[string]any{
			"process": processName,
		})
	})
	if err != nil {
		// Only reachable if ctx is cancelled; production passes Background().
		_ = catcher.Error("coordination setup cancelled before it could succeed", err, map[string]any{
			"process": processName,
		})
		return
	}
	defer setup.Close()

	holder := uuid.NewString()
	logCoordinationReady(holder)

	go watchConfigurationChanges(ctx, setup, holder)

	select {}
}

func watchConfigurationChanges(ctx context.Context, coord coordination.LeasePath, holder string) {
	time.Sleep(3 * time.Second)

	initialConfig := GetConfig()
	if initialConfig != nil && initialConfig.ModuleActive {
		updateStreams(ctx, coord, holder, initialConfig)
	}

	for newConfig := range GetConfigUpdateChannel() {
		if newConfig == nil || !newConfig.ModuleActive {
			stopAllStreams()
			continue
		}

		updateStreams(ctx, coord, holder, newConfig)
	}
}

func updateStreams(ctx context.Context, coord coordination.LeasePath, holder string, newConfig *ConfigurationSection) {
	activeStreamsMu.Lock()
	defer activeStreamsMu.Unlock()

	newGroups := make(map[string]*ModuleGroup)
	for _, grp := range newConfig.ModuleGroups {
		newGroups[grp.Key()] = grp
	}

	for key, stream := range activeStreams {
		if _, exists := newGroups[key]; !exists {
			stream.cancel()

			go func(s *activeStream, k string) {
				s.wg.Wait()
				activeStreamsMu.Lock()
				delete(activeStreams, k)
				activeStreamsMu.Unlock()
			}(stream, key)
		}
	}

	for key, group := range newGroups {
		if !isGroupValid(group) {
			continue
		}

		existingStream, exists := activeStreams[key]

		if exists {
			newProcessor := buildProcessor(group)
			if processorChanged(existingStream.processor, newProcessor) {
				existingStream.cancel()

				go func(s *activeStream, k string, g *ModuleGroup) {
					s.wg.Wait()
					activeStreamsMu.Lock()
					delete(activeStreams, k)
					startStream(ctx, coord, holder, k, g)
					activeStreamsMu.Unlock()
				}(existingStream, key, group)
			}
		} else {
			startStream(ctx, coord, holder, key, group)
		}
	}
}

// startStream registers a unit; nothing connects until runOwnedStream holds the
// lease. ctx must be the plugin's root context so shutdown reaches every socket.
func startStream(ctx context.Context, coord coordination.LeasePath, holder, key string, group *ModuleGroup) {
	streamCtx, cancel := context.WithCancel(ctx)

	stream := &activeStream{
		ctx:        streamCtx,
		cancel:     cancel,
		processor:  buildProcessor(group),
		groupKey:   key,
		dataSource: group.GroupName,
		cursor:     newCursorState(),
	}

	activeStreams[key] = stream

	// Not tracked by stream.wg: runEventStream waits on that same WaitGroup for
	// its feed goroutines, so adding this one would deadlock the unit.
	go runOwnedStream(coord, holder, stream, maintainStreamConnection)
}

func stopAllStreams() {
	activeStreamsMu.Lock()

	if len(activeStreams) == 0 {
		activeStreamsMu.Unlock()
		return
	}

	for _, stream := range activeStreams {
		stream.cancel()
	}

	var wg sync.WaitGroup
	for _, stream := range activeStreams {
		wg.Add(1)
		go func(s *activeStream) {
			defer wg.Done()
			s.wg.Wait()
		}(stream)
	}

	activeStreamsMu.Unlock()

	wg.Wait()

	activeStreamsMu.Lock()
	for key := range activeStreams {
		delete(activeStreams, key)
	}
	activeStreamsMu.Unlock()
}

func maintainStreamConnection(stream *activeStream) {
	for {
		err := runEventStream(stream)
		if err != nil {
			select {
			case <-stream.ctx.Done():
				return
			case <-time.After(reconnectDelay):
			}
		}
	}
}

func runEventStream(stream *activeStream) error {
	apiClient, err := stream.processor.createClient()
	if err != nil {
		return catcher.Error("failed to create client", err, map[string]any{
			"process": processName,
		})
	}

	ctx, cancel := context.WithTimeout(stream.ctx, 2*time.Minute)
	defer cancel()

	jsonFormat := "json"
	response, err := apiClient.EventStreams.ListAvailableStreamsOAuth2(
		&event_streams.ListAvailableStreamsOAuth2Params{
			AppID:   stream.processor.AppID,
			Format:  &jsonFormat,
			Context: ctx,
		},
	)
	if err != nil {
		return catcher.Error("failed to list streams", err, map[string]any{
			"process": processName,
		})
	}

	if err = falcon.AssertNoError(response.Payload.Errors); err != nil {
		return catcher.Error("CrowdStrike API error", err, map[string]any{
			"process": processName,
		})
	}

	availableStreams := response.Payload.Resources

	for _, streamV2 := range availableStreams {
		if streamV2.DataFeedURL == nil {
			catcher.Error("Stream has nil DataFeedURL, skipping", nil, map[string]any{
				"process": processName,
			})
			continue
		}

		streamID := *streamV2.DataFeedURL

		stream.wg.Add(1)
		go func(streamResource *models.MainAvailableStreamV2, sid string) {
			defer stream.wg.Done()
			maintainIndividualStream(stream, apiClient, streamResource, sid)
		}(streamV2, streamID)
	}

	<-stream.ctx.Done()

	stream.wg.Wait()

	return nil
}

func maintainIndividualStream(stream *activeStream, apiClient *client.CrowdStrikeAPISpecification,
	streamResource *models.MainAvailableStreamV2, streamID string) {

	for {
		select {
		case <-stream.ctx.Done():
			return
		default:
			currentOffset := stream.cursor.offset(streamID)

			falconStream, err := falcon.NewStream(stream.ctx, apiClient, stream.processor.AppID, streamResource, currentOffset)
			if err != nil {
				catcher.Error("failed to create stream, will retry", err, map[string]any{
					"process": processName,
				})
			} else {
				logStreamOpened(stream.groupKey, streamID, currentOffset)

				err = processStreamEvents(stream, falconStream, streamID)
				falconStream.Close()

				if err != nil {
					catcher.Error("stream error, will reconnect", err, map[string]any{
						"process": processName,
					})
				}
			}

			if err != nil {
				select {
				case <-stream.ctx.Done():
					return
				case <-time.After(reconnectDelay):
					continue
				}
			}
		}
	}
}

func processStreamEvents(stream *activeStream, falconStream *falcon.StreamingHandle, streamID string) error {
	// Scoped to one open socket, so each reconnect announces its first event.
	var firstEvent firstEventGate

	for {
		select {
		case <-stream.ctx.Done():
			return nil

		case err := <-falconStream.Errors:
			if err.Fatal {
				return catcher.Error("fatal stream error", err.Err, map[string]any{
					"process": processName,
				})
			}
			catcher.Error("Non-fatal stream error", err.Err, map[string]any{
				"process": processName,
			})

		case event := <-falconStream.Events:
			if event.Metadata.EventCreationTime > stream.cursor.startsAfter() {
				processEvent(event, stream.dataSource, stream.processor.TenantId, stream.groupKey, streamID)

				stream.cursor.setOffset(streamID, event.Metadata.Offset)

				if firstEvent.take() {
					logFirstEventIngested(stream.groupKey, streamID, event.Metadata.Offset)
				}
			}
		}
	}
}

func processEvent(event *streaming_models.EventItem, dataSource string, tenantId string, groupKey string, streamID string) {
	var eventData string
	if len(event.RawMessage) > 0 {
		eventData = string(event.RawMessage)
	} else {
		eventJSON, err := json.Marshal(event)
		if err != nil {
			catcher.Error("Failed to marshal event", err, map[string]any{
				"process": processName,
			})
			return
		}
		eventData = string(eventJSON)
	}

	if tenantId == "" {
		tenantId = defaultTenant
	}

	// Identity is derived after the defaultTenant substitution above, so it is
	// scoped to the tenant the event is filed under.
	_ = plugins.EnqueueLog(&plugins.Log{
		Id:         eventIdentity(tenantId, groupKey, streamID, event.Metadata.Offset),
		TenantId:   tenantId,
		DataType:   "crowdstrike",
		DataSource: dataSource,
		Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
		Raw:        eventData,
	}, pluginName)
}

type CrowdStrikeProcessor struct {
	ClientID     string
	ClientSecret string
	Cloud        string
	// AppID comes from deriveAppID, never from configuration: two groups sharing
	// a CID and an administrator-chosen app name would share one subscription.
	AppID    string
	TenantId string
}

func isGroupValid(group *ModuleGroup) bool {
	if group == nil {
		return false
	}

	for _, cnf := range group.ModuleGroupConfigurations {
		if strings.TrimSpace(cnf.ConfValue) == "" {
			return false
		}
	}
	return true
}

func buildProcessor(group *ModuleGroup) *CrowdStrikeProcessor {
	processor := &CrowdStrikeProcessor{
		TenantId: group.TenantId,
		AppID:    deriveAppID(group.Key()),
	}

	for _, cnf := range group.ModuleGroupConfigurations {
		switch cnf.ConfKey {
		case "crowdstrike_client_id":
			processor.ClientID = cnf.ConfValue
		case "crowdstrike_client_secret":
			processor.ClientSecret = cnf.ConfValue
		case "crowdstrike_cloud_region_url":
			processor.Cloud = cnf.ConfValue
		}
	}
	return processor
}

func processorChanged(old, new *CrowdStrikeProcessor) bool {
	if old == nil || new == nil {
		return true
	}
	return old.ClientID != new.ClientID ||
		old.ClientSecret != new.ClientSecret ||
		old.Cloud != new.Cloud ||
		old.AppID != new.AppID ||
		old.TenantId != new.TenantId
}

func (p *CrowdStrikeProcessor) createClient() (*client.CrowdStrikeAPISpecification, error) {
	if p.ClientID == "" || p.ClientSecret == "" {
		return nil, catcher.Error("cannot create CrowdStrike client",
			errors.New("client ID or client secret is empty"), map[string]any{"process": processName})
	}

	cloudType, err := extractCloudFromURL(p.Cloud)
	if err != nil {
		return nil, catcher.Error("invalid cloud region configuration", err, map[string]any{
			"process":     processName,
			"cloud_value": p.Cloud,
		})
	}

	client, err := falcon.NewClient(&falcon.ApiConfig{
		ClientId:     p.ClientID,
		ClientSecret: p.ClientSecret,
		Cloud:        cloudType,
		Context:      context.Background(),
	})
	if err != nil {
		return nil, catcher.Error("cannot create CrowdStrike client", err, map[string]any{"process": processName})
	}

	return client, nil
}

func extractCloudFromURL(cloudValue string) (falcon.CloudType, error) {
	trimmed := strings.TrimSpace(cloudValue)

	urlToRegion := map[string]string{
		"api.crowdstrike.com":            "us-1",
		"api.us-2.crowdstrike.com":       "us-2",
		"api.eu-1.crowdstrike.com":       "eu-1",
		"api.laggar.gcw.crowdstrike.com": "us-gov-1",
		"api.us-gov-2.crowdstrike.mil":   "us-gov-2",
	}

	if strings.Contains(trimmed, "://") || strings.Contains(trimmed, ".crowdstrike.") {
		for host, region := range urlToRegion {
			if strings.Contains(trimmed, host) {
				return falcon.CloudValidate(region)
			}
		}
		return 0, fmt.Errorf("unrecognized CrowdStrike URL: %s", trimmed)
	}

	return falcon.CloudValidate(trimmed)
}

func connectionChecker(url string) error {
	checkConn := func() error {
		if err := checkConnection(url); err != nil {
			return fmt.Errorf("connection failed: %v", err)
		}
		return nil
	}

	if err := infiniteRetryIfXError(checkConn, "connection failed"); err != nil {
		return err
	}

	return nil
}

func checkConnection(url string) error {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			_ = catcher.Error("error closing response body: %v", err, map[string]any{"process": processName})
		}
	}()

	return nil
}

func infiniteRetryIfXError(f func() error, exception string) error {
	var xErrorWasLogged bool

	for {
		err := f()
		if err != nil && is(err, exception) {
			if !xErrorWasLogged {
				_ = catcher.Error("An error occurred (%s), will keep retrying indefinitely...", err, map[string]any{"process": processName})
				xErrorWasLogged = true
			}
			time.Sleep(reconnectDelay)
			continue
		}

		return err
	}
}

func is(e error, args ...string) bool {
	for _, arg := range args {
		if strings.Contains(e.Error(), arg) {
			return true
		}
	}
	return false
}
