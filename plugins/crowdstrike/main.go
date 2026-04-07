package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	"github.com/utmstack/UTMStack/plugins/crowdstrike/config"
)

const (
	defaultTenant      = "ce66672c-e36d-4761-a8c8-90058fee1a24"
	urlCheckConnection = "https://falcon.crowdstrike.com"
	reconnectDelay     = 5 * time.Second
)

type streamKey struct {
	groupID   int32
	groupName string
}

type activeStream struct {
	ctx             context.Context
	cancel          context.CancelFunc
	processor       *CrowdStrikeProcessor
	dataSource      string
	offsets         map[string]uint64
	streamStartTime uint64
	wg              sync.WaitGroup
	mu              sync.Mutex
}

var (
	activeStreams   = make(map[streamKey]*activeStream)
	activeStreamsMu sync.RWMutex
)

func main() {
	mode := plugins.GetCfg("plugin_com.utmstack.crowdstrike").Env.Mode
	if mode != "manager" {
		return
	}

	if err := connectionChecker(urlCheckConnection); err != nil {
		_ = catcher.Error("Failed to establish connectivity, plugin will not start", err, map[string]any{
			"process": "plugin_com.utmstack.crowdstrike",
		})
		return
	}

	go config.StartConfigurationSystem()

	for i := 0; i < 2*runtime.NumCPU(); i++ {
		go func() {
			plugins.SendLogsFromChannel("com.utmstack.crowdstrike")
		}()
	}

	go watchConfigurationChanges()

	select {}
}

func watchConfigurationChanges() {
	time.Sleep(3 * time.Second)

	initialConfig := config.GetConfig()
	if initialConfig != nil && initialConfig.ModuleActive {
		updateStreams(initialConfig)
	}

	for newConfig := range config.GetConfigUpdateChannel() {
		if newConfig == nil || !newConfig.ModuleActive {
			stopAllStreams()
			continue
		}

		updateStreams(newConfig)
	}
}

func updateStreams(newConfig *config.ConfigurationSection) {
	activeStreamsMu.Lock()
	defer activeStreamsMu.Unlock()

	newGroups := make(map[streamKey]*config.ModuleGroup)
	for _, grp := range newConfig.ModuleGroups {
		key := streamKey{groupID: grp.Id, groupName: grp.GroupName}
		newGroups[key] = grp
	}

	for key, stream := range activeStreams {
		if _, exists := newGroups[key]; !exists {
			stream.cancel()

			go func(s *activeStream, k streamKey) {
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

				go func(s *activeStream, k streamKey, g *config.ModuleGroup) {
					s.wg.Wait()
					activeStreamsMu.Lock()
					delete(activeStreams, k)
					startStream(k, g)
					activeStreamsMu.Unlock()
				}(existingStream, key, group)
			}
		} else {
			startStream(key, group)
		}
	}
}

func startStream(key streamKey, group *config.ModuleGroup) {
	ctx, cancel := context.WithCancel(context.Background())

	processor := buildProcessor(group)

	stream := &activeStream{
		ctx:             ctx,
		cancel:          cancel,
		processor:       processor,
		dataSource:      group.GroupName,
		offsets:         make(map[string]uint64),
		streamStartTime: uint64(time.Now().UnixMilli()),
	}

	activeStreams[key] = stream

	go maintainStreamConnection(stream)
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
			"process": "plugin_com.utmstack.crowdstrike",
		})
	}

	ctx, cancel := context.WithTimeout(stream.ctx, 2*time.Minute)
	defer cancel()

	jsonFormat := "json"
	response, err := apiClient.EventStreams.ListAvailableStreamsOAuth2(
		&event_streams.ListAvailableStreamsOAuth2Params{
			AppID:   stream.processor.AppName,
			Format:  &jsonFormat,
			Context: ctx,
		},
	)
	if err != nil {
		return catcher.Error("failed to list streams", err, map[string]any{
			"process": "plugin_com.utmstack.crowdstrike",
		})
	}

	if err = falcon.AssertNoError(response.Payload.Errors); err != nil {
		return catcher.Error("CrowdStrike API error", err, map[string]any{
			"process": "plugin_com.utmstack.crowdstrike",
		})
	}

	availableStreams := response.Payload.Resources

	for _, streamV2 := range availableStreams {
		if streamV2.DataFeedURL == nil {
			catcher.Error("Stream has nil DataFeedURL, skipping", nil, map[string]any{
				"process": "plugin_com.utmstack.crowdstrike",
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
			stream.mu.Lock()
			currentOffset := stream.offsets[streamID]
			stream.mu.Unlock()

			falconStream, err := falcon.NewStream(stream.ctx, apiClient, stream.processor.AppName, streamResource, currentOffset)
			if err != nil {
				catcher.Error("failed to create stream, will retry", err, map[string]any{
					"process": "plugin_com.utmstack.crowdstrike",
				})
			} else {
				err = processStreamEvents(stream, falconStream, streamID)
				falconStream.Close()

				if err != nil {
					catcher.Error("stream error, will reconnect", err, map[string]any{
						"process": "plugin_com.utmstack.crowdstrike",
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
	for {
		select {
		case <-stream.ctx.Done():
			return nil

		case err := <-falconStream.Errors:
			if err.Fatal {
				return catcher.Error("fatal stream error", err.Err, map[string]any{
					"process": "plugin_com.utmstack.crowdstrike",
				})
			}
			catcher.Error("Non-fatal stream error", err.Err, map[string]any{
				"process": "plugin_com.utmstack.crowdstrike",
			})

		case event := <-falconStream.Events:
			if event.Metadata.EventCreationTime > stream.streamStartTime {
				processEvent(event, stream.dataSource)

				stream.mu.Lock()
				stream.offsets[streamID] = event.Metadata.Offset
				stream.mu.Unlock()
			}
		}
	}
}

func processEvent(event *streaming_models.EventItem, dataSource string) {
	var eventData string
	if len(event.RawMessage) > 0 {
		eventData = string(event.RawMessage)
	} else {
		eventJSON, err := json.Marshal(event)
		if err != nil {
			catcher.Error("Failed to marshal event", err, map[string]any{
				"process": "plugin_com.utmstack.crowdstrike",
			})
			return
		}
		eventData = string(eventJSON)
	}

	_ = plugins.EnqueueLog(&plugins.Log{
		Id:         uuid.NewString(),
		TenantId:   defaultTenant,
		DataType:   "crowdstrike",
		DataSource: dataSource,
		Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
		Raw:        eventData,
	}, "com.utmstack.crowdstrike")
}

type CrowdStrikeProcessor struct {
	ClientID     string
	ClientSecret string
	Cloud        string
	AppName      string
}

func isGroupValid(group *config.ModuleGroup) bool {
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

func buildProcessor(group *config.ModuleGroup) *CrowdStrikeProcessor {
	processor := &CrowdStrikeProcessor{}

	for _, cnf := range group.ModuleGroupConfigurations {
		switch cnf.ConfKey {
		case "crowdstrike_client_id":
			processor.ClientID = cnf.ConfValue
		case "crowdstrike_client_secret":
			processor.ClientSecret = cnf.ConfValue
		case "crowdstrike_cloud_region_url":
			processor.Cloud = cnf.ConfValue
		case "crowdstrike_app_name":
			processor.AppName = cnf.ConfValue
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
		old.AppName != new.AppName
}

func (p *CrowdStrikeProcessor) createClient() (*client.CrowdStrikeAPISpecification, error) {
	if p.ClientID == "" || p.ClientSecret == "" {
		return nil, catcher.Error("cannot create CrowdStrike client",
			errors.New("client ID or client secret is empty"), map[string]any{"process": "plugin_com.utmstack.crowdstrike"})
	}

	cloudType, err := extractCloudFromURL(p.Cloud)
	if err != nil {
		return nil, catcher.Error("invalid cloud region configuration", err, map[string]any{
			"process":     "plugin_com.utmstack.crowdstrike",
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
		return nil, catcher.Error("cannot create CrowdStrike client", err, map[string]any{"process": "plugin_com.utmstack.crowdstrike"})
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
