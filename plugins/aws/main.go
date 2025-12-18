package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
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

	delay := 5 * time.Minute
	ticker := time.NewTicker(delay)
	defer ticker.Stop()

	startTime := time.Now().UTC().Add(-delay)

	for range ticker.C {
		endTime := time.Now().UTC()

		if err := connectionChecker(urlCheckConnection); err != nil {
			_ = catcher.Error("External connection failure detected: %v", err, nil)
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
						pull(startTime, endTime, group)
					}
				}(grp)
			}
			wg.Wait()
		}

		startTime = endTime.Add(1 * time.Nanosecond)
	}
}

func pull(startTime time.Time, endTime time.Time, group *config.ModuleGroup) {
	agent := getAWSProcessor(group)

	logs, err := agent.getLogs(startTime, endTime)
	if err != nil {
		_ = catcher.Error("cannot get logs", err, map[string]any{
			"startTime": startTime,
			"endTime":   endTime,
			"group":     group.GroupName,
		})
		return
	}

	for _, log := range logs {
		_ = plugins.EnqueueLog(&plugins.Log{
			Id:         uuid.NewString(),
			TenantId:   defaultTenant,
			DataType:   "aws",
			DataSource: group.GroupName,
			Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
			Raw:        log,
		})
	}
}

type AWSProcessor struct {
	RegionName      string
	AccessKey       string
	SecretAccessKey string
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

	cfg, err := awsConfig.LoadDefaultConfig(ctx,
		awsConfig.WithRegion(p.RegionName),
		awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(p.AccessKey, p.SecretAccessKey, "")),
	)
	if err != nil {
		return aws.Config{}, catcher.Error("cannot create AWS session", err, nil)
	}

	return cfg, nil
}

func (p *AWSProcessor) describeLogGroups() ([]string, error) {
	awsConfig, err := p.createAWSSession()
	if err != nil {
		return nil, catcher.Error("cannot create AWS session", err, nil)
	}

	cwl := cloudwatchlogs.NewFromConfig(awsConfig)
	var logGroups []string
	paginator := cloudwatchlogs.NewDescribeLogGroupsPaginator(cwl, &cloudwatchlogs.DescribeLogGroupsInput{})

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, catcher.Error("cannot get log groups", err, nil)
		}
		for _, group := range page.LogGroups {
			logGroups = append(logGroups, *group.LogGroupName)
		}
	}

	return logGroups, nil
}

func (p *AWSProcessor) describeLogStreams(logGroup string) ([]string, error) {
	awsConfig, err := p.createAWSSession()
	if err != nil {
		return nil, catcher.Error("cannot create AWS session", err, nil)
	}

	cwl := cloudwatchlogs.NewFromConfig(awsConfig)
	var logStreams []string
	paginator := cloudwatchlogs.NewDescribeLogStreamsPaginator(cwl, &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName: aws.String(logGroup),
		OrderBy:      "LastEventTime",
		Descending:   aws.Bool(true),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, catcher.Error("cannot get log streams", err, nil)
		}
		for _, stream := range page.LogStreams {
			logStreams = append(logStreams, *stream.LogStreamName)
		}
	}

	return logStreams, nil
}

func (p *AWSProcessor) getLogs(startTime, endTime time.Time) ([]string, error) {
	// Retry logic for AWS session creation
	maxRetries := 3
	retryDelay := 2 * time.Second
	var awsConfig aws.Config
	var err error

	for retry := 0; retry < maxRetries; retry++ {
		awsConfig, err = p.createAWSSession()
		if err == nil {
			break
		}

		_ = catcher.Error("cannot create AWS session, retrying", err, map[string]any{
			"retry":      retry + 1,
			"maxRetries": maxRetries,
		})

		if retry < maxRetries-1 {
			time.Sleep(retryDelay)
			// Increase delay for next retry
			retryDelay *= 2
		}
	}

	if err != nil {
		return nil, catcher.Error("all retries failed when creating AWS session", err, nil)
	}

	cwl := cloudwatchlogs.NewFromConfig(awsConfig)

	// Retry logic for describing log groups
	retryDelay = 2 * time.Second
	var logGroups []string

	for retry := 0; retry < maxRetries; retry++ {
		logGroups, err = p.describeLogGroups()
		if err == nil {
			break
		}

		_ = catcher.Error("cannot get log groups, retrying", err, map[string]any{
			"retry":      retry + 1,
			"maxRetries": maxRetries,
		})

		if retry < maxRetries-1 {
			time.Sleep(retryDelay)
			// Increase delay for next retry
			retryDelay *= 2
		}
	}

	if err != nil {
		return nil, catcher.Error("all retries failed when getting log groups", err, nil)
	}

	transformedLogs := make([]string, 0, 10)
	for _, logGroup := range logGroups {
		// Retry logic for describing log streams
		retryDelay = 2 * time.Second
		var logStreams []string

		for retry := 0; retry < maxRetries; retry++ {
			logStreams, err = p.describeLogStreams(logGroup)
			if err == nil {
				break
			}

			_ = catcher.Error("cannot get log streams, retrying", err, map[string]any{
				"retry":      retry + 1,
				"maxRetries": maxRetries,
				"logGroup":   logGroup,
			})

			if retry < maxRetries-1 {
				time.Sleep(retryDelay)
				// Increase delay for next retry
				retryDelay *= 2
			}
		}

		if err != nil {
			_ = catcher.Error("all retries failed when getting log streams", err, map[string]any{
				"logGroup": logGroup,
			})
			continue // Skip this log group and try the next one
		}

		for _, stream := range logStreams {
			paginator := cloudwatchlogs.NewGetLogEventsPaginator(cwl, &cloudwatchlogs.GetLogEventsInput{
				LogGroupName:  aws.String(logGroup),
				LogStreamName: aws.String(stream),
				StartTime:     aws.Int64(startTime.Unix() * 1000),
				EndTime:       aws.Int64(endTime.Unix() * 1000),
				StartFromHead: aws.Bool(true),
			}, func(options *cloudwatchlogs.GetLogEventsPaginatorOptions) {
				options.StopOnDuplicateToken = true
				options.Limit = 10000
			})

			for paginator.HasMorePages() {
				// Retry logic for getting log events
				retryDelay = 2 * time.Second
				var page *cloudwatchlogs.GetLogEventsOutput

				for retry := 0; retry < maxRetries; retry++ {
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute) 
					
					page, err = paginator.NextPage(ctx)
					if err == nil {
						cancel()
						break
					}

					_ = catcher.Error("cannot get logs, retrying", err, map[string]any{
						"retry":      retry + 1,
						"maxRetries": maxRetries,
						"logGroup":   logGroup,
						"stream":     stream,
					})

					if retry < maxRetries-1 {
						time.Sleep(retryDelay)
						// Increase delay for next retry
						retryDelay *= 2
					}
					cancel()
				}

				if err != nil {
					_ = catcher.Error("all retries failed when getting logs", err, map[string]any{
						"logGroup": logGroup,
						"stream":   stream,
					})
					continue // Skip this page and try the next one
				}

				if page == nil {
					continue
				}

				for _, event := range page.Events {
					transformedLogs = append(transformedLogs, *event.Message)
				}
			}
		}
	}

	return transformedLogs, nil
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
