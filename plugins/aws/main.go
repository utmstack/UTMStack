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

func (p *AWSProcessor) describeLogGroups(cwl *cloudwatchlogs.Client) ([]string, error) {
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

func (p *AWSProcessor) describeLogStreams(cwl *cloudwatchlogs.Client, logGroup string) ([]string, error) {
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
	awsConfig, err := p.createAWSSession()
	if err != nil {
		return nil, catcher.Error("cannot create AWS session", err, nil)
	}

	cwl := cloudwatchlogs.NewFromConfig(awsConfig)

	logGroups, err := p.describeLogGroups(cwl)
	if err != nil {
		return nil, catcher.Error("cannot get log groups", err, nil)
	}

	transformedLogs := make([]string, 0, 10)
	for _, logGroup := range logGroups {
		time.Sleep(500 * time.Millisecond)

		logStreams, err := p.describeLogStreams(cwl, logGroup)
		if err != nil {
			_ = catcher.Error("cannot get log streams, skipping log group", err, map[string]any{
				"logGroup": logGroup,
			})
			continue
		}

		for i, stream := range logStreams {
			if i > 0 && i%5 == 0 {
				time.Sleep(2 * time.Second)
			} else if i > 0 {
				time.Sleep(300 * time.Millisecond)
			}

			paginator := cloudwatchlogs.NewGetLogEventsPaginator(cwl, &cloudwatchlogs.GetLogEventsInput{
				LogGroupName:  aws.String(logGroup),
				LogStreamName: aws.String(stream),
				StartTime:     aws.Int64(startTime.Unix() * 1000),
				EndTime:       aws.Int64(endTime.Unix() * 1000),
				StartFromHead: aws.Bool(true),
			}, func(options *cloudwatchlogs.GetLogEventsPaginatorOptions) {
				options.StopOnDuplicateToken = true
				options.Limit = 1000
			})

			pageCount := 0
			for paginator.HasMorePages() {
				if pageCount > 0 {
					time.Sleep(200 * time.Millisecond)
				}
				pageCount++

				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
				page, err := paginator.NextPage(ctx)
				cancel()

				if err != nil {
					_ = catcher.Error("cannot get log events, skipping stream", err, map[string]any{
						"logGroup": logGroup,
						"stream":   stream,
					})
					break
				}

				if page == nil {
					break
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
