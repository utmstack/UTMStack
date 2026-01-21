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
	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/plugins/crowdstrike/config"
)

const (
	defaultTenant      = "ce66672c-e36d-4761-a8c8-90058fee1a24"
	urlCheckConnection = "https://falcon.crowdstrike.com"
	wait               = 1 * time.Second
)

func main() {
	mode := plugins.GetCfg("plugin_com.utmstack.crowdstrike").Env.Mode
	if mode != "manager" {
		return
	}

	go config.StartConfigurationSystem()

	for t := 0; t < 2*runtime.NumCPU(); t++ {
		go func() {
			plugins.SendLogsFromChannel("com.utmstack.crowdstrike")
		}()
	}

	delay := 5 * time.Minute
	ticker := time.NewTicker(delay)
	defer ticker.Stop()

	for range ticker.C {
		if err := connectionChecker(urlCheckConnection); err != nil {
			_ = catcher.Error("External connection failure detected: %v", err, map[string]any{"process": "plugin_com.utmstack.crowdstrike"})
		}

		moduleConfig := config.GetConfig()
		if moduleConfig != nil && moduleConfig.ModuleActive {
			var wg sync.WaitGroup
			wg.Add(len(moduleConfig.ModuleGroups))
			for _, grp := range moduleConfig.ModuleGroups {
				go func(group *config.ModuleGroup) {
					defer wg.Done()
					var invalid bool
					for _, c := range group.ModuleGroupConfigurations {
						if strings.TrimSpace(c.ConfValue) == "" {
							invalid = true
							break
						}
					}

					if !invalid {
						pullCrowdStrikeEvents(group)
					}
				}(grp)
			}
			wg.Wait()
		}
	}
}

func pullCrowdStrikeEvents(group *config.ModuleGroup) {
	processor := getCrowdStrikeProcessor(group)

	events, err := processor.getEvents()
	if err != nil {
		_ = catcher.Error("cannot get CrowdStrike events", err, map[string]any{
			"group":   group.GroupName,
			"process": "plugin_com.utmstack.crowdstrike",
		})
		return
	}

	for _, event := range events {
		_ = plugins.EnqueueLog(&plugins.Log{
			Id:         uuid.NewString(),
			TenantId:   defaultTenant,
			DataType:   "crowdstrike",
			DataSource: group.GroupName,
			Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
			Raw:        event,
		}, "com.utmstack.crowdstrike")
	}
}

type CrowdStrikeProcessor struct {
	ClientID     string
	ClientSecret string
	Cloud        string
	AppName      string
}

func getCrowdStrikeProcessor(group *config.ModuleGroup) CrowdStrikeProcessor {
	processor := CrowdStrikeProcessor{}

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

func (p *CrowdStrikeProcessor) createClient() (*client.CrowdStrikeAPISpecification, error) {
	if p.ClientID == "" || p.ClientSecret == "" {
		return nil, catcher.Error("cannot create CrowdStrike client",
			errors.New("client ID or client secret is empty"), map[string]any{"process": "plugin_com.utmstack.crowdstrike"})
	}

	client, err := falcon.NewClient(&falcon.ApiConfig{
		ClientId:     p.ClientID,
		ClientSecret: p.ClientSecret,
		Cloud:        falcon.Cloud(p.Cloud),
		Context:      context.Background(),
	})
	if err != nil {
		return nil, catcher.Error("cannot create CrowdStrike client", err, map[string]any{"process": "plugin_com.utmstack.crowdstrike"})
	}

	return client, nil
}

func (p *CrowdStrikeProcessor) getEvents() ([]string, error) {
	client, err := p.createClient()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	json := "json"
	response, err := client.EventStreams.ListAvailableStreamsOAuth2(
		&event_streams.ListAvailableStreamsOAuth2Params{
			AppID:   p.AppName,
			Format:  &json,
			Context: ctx,
		},
	)
	if err != nil {
		return nil, catcher.Error("cannot list available streams", err, map[string]any{"process": "plugin_com.utmstack.crowdstrike"})
	}

	if err = falcon.AssertNoError(response.Payload.Errors); err != nil {
		return nil, catcher.Error("CrowdStrike API error", err, map[string]any{"process": "plugin_com.utmstack.crowdstrike"})
	}

	availableStreams := response.Payload.Resources
	if len(availableStreams) == 0 {
		_ = catcher.Error("no available streams found", nil, map[string]any{
			"app_id":  p.AppName,
			"process": "plugin_com.utmstack.crowdstrike",
		})
		return []string{}, nil
	}

	var events []string
	for _, availableStream := range availableStreams {
		streamEvents, err := p.getStreamEvents(ctx, client, availableStream)
		if err != nil {
			_ = catcher.Error("cannot get stream events", err, map[string]any{
				"stream":  availableStream,
				"process": "plugin_com.utmstack.crowdstrike",
			})
			continue
		}
		events = append(events, streamEvents...)
	}

	return events, nil
}

func (p *CrowdStrikeProcessor) getStreamEvents(ctx context.Context, client *client.CrowdStrikeAPISpecification, availableStream interface{}) ([]string, error) {
	stream_v2, ok := availableStream.(*models.MainAvailableStreamV2)
	if !ok {
		return nil, catcher.Error("invalid stream type", fmt.Errorf("cannot convert to MainAvailableStreamV2"), map[string]any{"process": "plugin_com.utmstack.crowdstrike"})
	}

	stream, err := falcon.NewStream(ctx, client, p.AppName, stream_v2, 0)
	if err != nil {
		return nil, catcher.Error("cannot create stream", err, map[string]any{"process": "plugin_com.utmstack.crowdstrike"})
	}
	defer stream.Close()

	var events []string
	timeout := time.NewTimer(30 * time.Second)
	defer timeout.Stop()

	for {
		select {
		case err := <-stream.Errors:
			if err.Fatal {
				return events, catcher.Error("fatal stream error", err.Err, map[string]any{"process": "plugin_com.utmstack.crowdstrike"})
			} else {
				_ = catcher.Error("stream error", err.Err, map[string]any{"process": "plugin_com.utmstack.crowdstrike"})
			}
		case event := <-stream.Events:
			eventJSON, err := json.Marshal(event)
			if err != nil {
				_ = catcher.Error("cannot marshal event", err, map[string]any{"process": "crowdstrike-plugin"})
				continue
			}
			events = append(events, string(eventJSON))

			if len(events) >= 100 {
				return events, nil
			}
		case <-timeout.C:
			return events, nil
		case <-ctx.Done():
			return events, nil
		}
	}
}
