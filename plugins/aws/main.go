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

func main() {
	mode := plugins.GetCfg().Env.Mode
	if mode != "manager" {
		return
	}

	go config.StartConfigurationSystem()

	for t := 0; t < 2*runtime.NumCPU(); t++ {
		go func() {
			plugins.SendLogsFromChannel()
		}()
	}

	for {
		if err := connectionChecker(urlCheckConnection); err != nil {
			_ = catcher.Error("External connection failure detected: %v", err, nil)
		}

		moduleConfig := config.GetConfig()
		if moduleConfig != nil && moduleConfig.ModuleActive {
			for _, grp := range moduleConfig.ModuleGroups {
				go func(group *config.ModuleGroup) {
					var invalid bool
					for _, c := range group.ModuleGroupConfigurations {
						if strings.TrimSpace(c.ConfValue) == "" {
							invalid = true
							break
						}
					}

					if !invalid {
						streamLogs(group)
					}
				}(grp)
			}
			break
		}
		time.Sleep(5 * time.Second)
	}

	select {}
}

func streamLogs(group *config.ModuleGroup) {
	agent := getAWSProcessor(group)

	awsConfig, err := agent.createAWSSession()
	if err != nil {
		_ = catcher.Error("cannot create AWS session", err, nil)
		return
	}

	cwl := cloudwatchlogs.NewFromConfig(awsConfig)

	startTime := time.Now().UTC()

	catcher.Info("Starting streaming logs", map[string]any{
		"group":     group.GroupName,
		"logGroup":  agent.LogGroup,
		"startTime": startTime.Format(time.RFC3339),
	})

	currentStreams := make(map[string]struct{})

	for {
		logStreams, err := agent.describeLogStreams(cwl, agent.LogGroup)
		if err != nil {
			_ = catcher.Error("cannot get log streams", err, map[string]any{
				"logGroup": agent.LogGroup,
			})
			time.Sleep(30 * time.Second)
			continue
		}

		for _, stream := range logStreams {
			if _, exists := currentStreams[stream]; exists {
				continue
			}
			currentStreams[stream] = struct{}{}

			catcher.Info("Starting to stream log stream", map[string]any{
				"logGroup":  agent.LogGroup,
				"logStream": stream,
			})
			go streamLogStream(cwl, agent.LogGroup, stream, startTime, group.GroupName)
		}

		time.Sleep(5 * time.Minute)
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
			errors.New("region name is empty"), nil)
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
		return aws.Config{}, catcher.Error("cannot create AWS session", err, nil)
	}

	return cfg, nil
}

func (p *AWSProcessor) describeLogStreams(cwl *cloudwatchlogs.Client, logGroup string) ([]string, error) {
	var logStreams []string
	paginator := cloudwatchlogs.NewDescribeLogStreamsPaginator(cwl, &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName: aws.String(logGroup),
		OrderBy:      "LastEventTime",
		Descending:   aws.Bool(true),
	})

	for paginator.HasMorePages() {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)

		page, err := paginator.NextPage(ctx)
		if err != nil {
			cancel()
			return nil, catcher.Error("cannot get log streams", err, nil)
		}
		for _, stream := range page.LogStreams {
			logStreams = append(logStreams, *stream.LogStreamName)
		}

		cancel()
	}

	return logStreams, nil
}

func streamLogStream(cwl *cloudwatchlogs.Client, logGroup, streamName string, startTime time.Time, dataSource string) {
	var nextToken *string
	processedCount := 0

	for {
		input := &cloudwatchlogs.GetLogEventsInput{
			LogGroupName:  aws.String(logGroup),
			LogStreamName: aws.String(streamName),
			StartTime:     aws.Int64(startTime.Unix() * 1000),
			StartFromHead: aws.Bool(true),
			NextToken:     nextToken,
			Limit:         aws.Int32(1000),
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		result, err := cwl.GetLogEvents(ctx, input)
		cancel()

		if err != nil {
			_ = catcher.Error("cannot get log events", err, map[string]any{
				"logGroup": logGroup,
				"stream":   streamName,
			})
			time.Sleep(10 * time.Second)
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
			})
			processedCount++
			eventsInBatch++
		}

		if eventsInBatch > 0 {
			catcher.Info("Processed logs from stream", map[string]any{
				"stream":     streamName,
				"batchCount": eventsInBatch,
				"totalCount": processedCount,
				"dataSource": dataSource,
			})
		} else {
			time.Sleep(5 * time.Second)
		}

		if result.NextForwardToken != nil {
			nextToken = result.NextForwardToken
		} else {
			time.Sleep(5 * time.Second)
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
			_ = catcher.Error("error closing response body: %v", err, nil)
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
				_ = catcher.Error("An error occurred (%s), will keep retrying indefinitely...", err, nil)
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
