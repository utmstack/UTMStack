package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"

	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/plugins/aws/config"
)

const (
	defaultTenant      = "ce66672c-e36d-4761-a8c8-90058fee1a24"
	urlCheckConnection = "https://sts.amazonaws.com"
	wait               = 1 * time.Second
)

type activeGroupStream struct {
	cancel context.CancelFunc
	config AWSProcessor
}

var (
	activeStreams = make(map[int32]*activeGroupStream)
)

func main() {
	mode := plugins.GetCfg("plugin_com.utmstack.aws").Env.Mode
	if mode != "manager" {
		return
	}

	go config.StartConfigurationSystem()

	for t := 0; t < 2*runtime.NumCPU(); t++ {
		go func() {
			plugins.SendLogsFromChannel("com.utmstack.aws")
		}()
	}

	for {
		if err := connectionChecker(urlCheckConnection); err != nil {
			_ = catcher.Error("failed to connect with external service", err, map[string]any{"process": "plugin_com.utmstack.aws"})
			continue
		}
		break
	}

	watchConfigChanges()
}

func watchConfigChanges() {
	time.Sleep(3 * time.Second)

	initialConfig := config.GetConfig()
	if initialConfig != nil && initialConfig.ModuleActive {
		syncStreams(initialConfig)
	}

	for newConfig := range config.GetConfigUpdateChannel() {
		if newConfig == nil || !newConfig.ModuleActive {
			stopAllStreams()
			continue
		}

		syncStreams(newConfig)
	}
}

func syncStreams(moduleConfig *config.ConfigurationSection) {
	currentGroupIDs := make(map[int32]bool)
	for _, group := range moduleConfig.ModuleGroups {
		currentConfig := getAWSProcessor(group)
		groupID := group.Id
		currentGroupIDs[groupID] = true

		existing := activeStreams[groupID]

		if existing == nil {
			startGroupStream(groupID, group)
		} else if existing.config != currentConfig {
			catcher.Info("Configuration changed for group, restarting", map[string]any{
				"group":   group.GroupName,
				"process": "plugin_com.utmstack.aws",
			})
			existing.cancel()
			delete(activeStreams, groupID)
			startGroupStream(groupID, group)
		}
	}

	for groupID, stream := range activeStreams {
		if !currentGroupIDs[groupID] {
			catcher.Info("Group removed, stopping stream", map[string]any{
				"groupId": groupID,
				"process": "plugin_com.utmstack.aws",
			})
			stream.cancel()
			delete(activeStreams, groupID)
		}
	}
}

func startGroupStream(groupID int32, group *config.ModuleGroup) {
	ctx, cancel := context.WithCancel(context.Background())

	groupConfig := getAWSProcessor(group)

	activeStreams[groupID] = &activeGroupStream{
		cancel: cancel,
		config: groupConfig,
	}

	catcher.Info("Starting stream for group", map[string]any{
		"group":   group.GroupName,
		"process": "plugin_com.utmstack.aws",
	})

	go streamLogs(ctx, group)
}

func stopAllStreams() {
	if len(activeStreams) == 0 {
		return
	}

	catcher.Info("Stopping all active streams", map[string]any{
		"count":   len(activeStreams),
		"process": "plugin_com.utmstack.aws",
	})

	for groupID, stream := range activeStreams {
		stream.cancel()
		delete(activeStreams, groupID)
	}
}

func sleepWithCancel(ctx context.Context, d time.Duration) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(d):
		return true
	}
}

func streamLogs(ctx context.Context, group *config.ModuleGroup) {
	agent := getAWSProcessor(group)

	awsConfig, err := agent.createAWSSession()
	if err != nil {
		_ = catcher.Error("cannot create AWS session", err, map[string]any{"process": "plugin_com.utmstack.aws"})
		return
	}

	cwl := cloudwatchlogs.NewFromConfig(awsConfig)

	startTime := time.Now().UTC()

	catcher.Info("Starting streaming logs", map[string]any{
		"group":     group.GroupName,
		"logGroup":  agent.LogGroup,
		"startTime": startTime.Format(time.RFC3339),
		"process":   "plugin_com.utmstack.aws",
	})

	currentStreams := make(map[string]context.CancelFunc)
	defer func() {
		for _, cancel := range currentStreams {
			cancel()
		}
	}()

	for {
		logStreams, err := describeLogStreams(ctx, cwl, agent.LogGroup)
		if err != nil {
			_ = catcher.Error("cannot get log streams", err, map[string]any{
				"logGroup": agent.LogGroup,
				"process":  "plugin_com.utmstack.aws",
			})
			if !sleepWithCancel(ctx, 30*time.Second) {
				return
			}
			continue
		}

		for _, stream := range logStreams {
			if _, exists := currentStreams[stream]; exists {
				continue
			}

			streamCtx, streamCancel := context.WithCancel(ctx)
			currentStreams[stream] = streamCancel

			go streamLogStream(streamCtx, cwl, agent.LogGroup, stream, startTime, group.GroupName)
		}

		awsStreamsMap := make(map[string]bool)
		for _, stream := range logStreams {
			awsStreamsMap[stream] = true
		}

		for streamName, cancel := range currentStreams {
			if !awsStreamsMap[streamName] {
				catcher.Info("Log stream expired, stopping", map[string]any{
					"logGroup":  agent.LogGroup,
					"logStream": streamName,
					"process":   "plugin_com.utmstack.aws",
				})
				cancel()
				delete(currentStreams, streamName)
			}
		}

		if !sleepWithCancel(ctx, 5*time.Minute) {
			catcher.Info("Stream cancelled for group", map[string]any{
				"group":   group.GroupName,
				"process": "plugin_com.utmstack.aws",
			})
			return
		}
	}
}

type AWSProcessor struct {
	RegionName      string
	AccessKey       string
	SecretAccessKey string
	LogGroup        string
}

func getAWSProcessor(group *config.ModuleGroup) AWSProcessor {
	awsPro := AWSProcessor{}
	for _, cnf := range group.ModuleGroupConfigurations {
		switch cnf.ConfKey {
		case "aws_default_region":
			awsPro.RegionName = cnf.ConfValue
		case "aws_access_key_id":
			awsPro.AccessKey = cnf.ConfValue
		case "aws_secret_access_key":
			awsPro.SecretAccessKey = cnf.ConfValue
		case "aws_log_group_name":
			awsPro.LogGroup = cnf.ConfValue
		}
	}
	return awsPro
}

func (p *AWSProcessor) createAWSSession() (aws.Config, error) {
	if p.RegionName == "" {
		return aws.Config{}, catcher.Error("cannot create AWS session",
			errors.New("region name is empty"), map[string]any{"process": "plugin_com.utmstack.aws"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	adaptiveRetryer := retry.NewAdaptiveMode(func(ao *retry.AdaptiveModeOptions) {
		ao.StandardOptions = append(ao.StandardOptions, func(so *retry.StandardOptions) {
			so.MaxAttempts = 10              // Increment max attempts for throttling
			so.MaxBackoff = 30 * time.Second // Increase max backoff time
		})
		ao.RequestCost = 1
		ao.FailOnNoAttemptTokens = false // Allow retries even without tokens
	})

	cfg, err := awsConfig.LoadDefaultConfig(ctx,
		awsConfig.WithRegion(p.RegionName),
		awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(p.AccessKey, p.SecretAccessKey, "")),
		awsConfig.WithRetryer(func() aws.Retryer {
			return adaptiveRetryer
		}),
	)
	if err != nil {
		return aws.Config{}, catcher.Error("cannot create AWS session", err, map[string]any{"process": "plugin_com.utmstack.aws"})
	}

	return cfg, nil
}

func describeLogStreams(ctx context.Context, cwl *cloudwatchlogs.Client, logGroup string) ([]string, error) {
	var logStreams []string
	paginator := cloudwatchlogs.NewDescribeLogStreamsPaginator(cwl, &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName: aws.String(logGroup),
		OrderBy:      "LastEventTime",
		Descending:   aws.Bool(true),
	})

	for paginator.HasMorePages() {
		requestCtx, cancel := context.WithTimeout(ctx, 1*time.Minute)

		page, err := paginator.NextPage(requestCtx)
		if err != nil {
			cancel()
			return nil, catcher.Error("cannot get log streams", err, map[string]any{"process": "plugin_com.utmstack.aws"})
		}
		for _, stream := range page.LogStreams {
			logStreams = append(logStreams, *stream.LogStreamName)
		}

		cancel()
	}

	return logStreams, nil
}

func streamLogStream(ctx context.Context, cwl *cloudwatchlogs.Client, logGroup, streamName string, startTime time.Time, dataSource string) {
	var nextToken *string
	processedCount := 0

	for {
		select {
		case <-ctx.Done():
			catcher.Info("Log stream cancelled", map[string]any{
				"stream":     streamName,
				"totalCount": processedCount,
				"process":    "plugin_com.utmstack.aws",
			})
			return
		default:
		}

		input := &cloudwatchlogs.GetLogEventsInput{
			LogGroupName:  aws.String(logGroup),
			LogStreamName: aws.String(streamName),
			StartTime:     aws.Int64(startTime.Unix() * 1000),
			StartFromHead: aws.Bool(true),
			NextToken:     nextToken,
			Limit:         aws.Int32(1000),
		}

		requestCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		result, err := cwl.GetLogEvents(requestCtx, input)
		cancel()

		if err != nil {
			_ = catcher.Error("cannot get log events", err, map[string]any{
				"logGroup": logGroup,
				"stream":   streamName,
				"process":  "plugin_com.utmstack.aws",
			})
			if !sleepWithCancel(ctx, 10*time.Second) {
				return
			}
			continue
		}

		eventsInBatch := 0
		for _, event := range result.Events {
			_ = plugins.EnqueueLog(&plugins.Log{
				Id:         uuid.NewString(),
				TenantId:   defaultTenant,
				DataType:   "aws",
				DataSource: dataSource,
				Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
				Raw:        *event.Message,
			}, "com.utmstack.aws")
			processedCount++
			eventsInBatch++
		}

		if eventsInBatch > 0 {
			catcher.Info("Processed logs from stream", map[string]any{
				"stream":     streamName,
				"batchCount": eventsInBatch,
				"totalCount": processedCount,
				"dataSource": dataSource,
				"process":    "plugin_com.utmstack.aws",
			})
		}

		if result.NextForwardToken != nil {
			nextToken = result.NextForwardToken
		} else {
			if !sleepWithCancel(ctx, 5*time.Second) {
				return
			}
		}
	}
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
			_ = catcher.Error("error closing response body: %v", err, map[string]any{"process": "plugin_com.utmstack.aws"})
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
				_ = catcher.Error("An error occurred (%s), will keep retrying indefinitely...", err, map[string]any{"process": "aws-plugin"})
				xErrorWasLogged = true
			}
			time.Sleep(wait)
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
