package main

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	utmconf "github.com/utmstack/config-client-go"
	"github.com/utmstack/config-client-go/enum"
	"github.com/utmstack/config-client-go/types"
)

const (
	delayCheck    = 300
	defaultTenant = "ce66672c-e36d-4761-a8c8-90058fee1a24"
)

func main() {
	mode := plugins.GetCfg().Env.Mode
	if mode != "manager" {
		os.Exit(0)
	}

	for t := 0; t < 2*runtime.NumCPU(); t++ {
		go plugins.SendLogsFromChannel()
	}

	startTime := time.Now().UTC().Add(-1 * delayCheck * time.Second)
	endTime := time.Now().UTC()
	for {
		utmConfig := plugins.PluginCfg("com.utmstack", false)
		internalKey := utmConfig.Get("internalKey").String()
		backendUrl := utmConfig.Get("backend").String()
		if internalKey == "" || backendUrl == "" {
			time.Sleep(5 * time.Second)
			continue
		}

		client := utmconf.NewUTMClient(internalKey, backendUrl)

		moduleConfig, err := client.GetUTMConfig(enum.AWS_IAM_USER)
		if err != nil {
			if strings.Contains(err.Error(), "invalid character '<'") {
				time.Sleep(time.Second * delayCheck)
				continue
			}
			if (err.Error() != "") && (err.Error() != " ") {
				_ = catcher.Error("cannot obtain module configuration", err, nil)
			}

			time.Sleep(time.Second * delayCheck)
			startTime = time.Now().UTC().Add(-1 * delayCheck * time.Second)
			endTime = time.Now().UTC()
			continue
		}

		if moduleConfig.ModuleActive {
			var wg sync.WaitGroup
			wg.Add(len(moduleConfig.ConfigurationGroups))

			for _, group := range moduleConfig.ConfigurationGroups {
				go func(group types.ModuleGroup) {
					var skip bool

					for _, cnf := range group.Configurations {
						if cnf.ConfValue == "" || cnf.ConfValue == " " {
							skip = true
							break
						}
					}

					if !skip {
						pull(startTime, endTime, group)
					}

					wg.Done()
				}(group)
			}

			wg.Wait()
		}

		time.Sleep(time.Second * delayCheck)

		startTime = endTime.Add(1)
		endTime = time.Now().UTC()
	}
}

func pull(startTime time.Time, endTime time.Time, group types.ModuleGroup) {
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
		plugins.EnqueueLog(&plugins.Log{
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

func getAWSProcessor(group types.ModuleGroup) AWSProcessor {
	awsPro := AWSProcessor{}
	for _, cnf := range group.Configurations {
		switch cnf.ConfName {
		case "Default Region":
			awsPro.RegionName = cnf.ConfValue
		case "Access Key":
			awsPro.AccessKey = cnf.ConfValue
		case "Secret Key":
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

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(p.RegionName),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(p.AccessKey, p.SecretAccessKey, "")),
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
	awsConfig, err := p.createAWSSession()
	if err != nil {
		return nil, catcher.Error("cannot create AWS session", err, nil)
	}

	cwl := cloudwatchlogs.NewFromConfig(awsConfig)

	logGroups, err := p.describeLogGroups()
	if err != nil {
		return nil, catcher.Error("cannot get log groups", err, nil)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	transformedLogs := make([]string, 0, 10)
	for _, logGroup := range logGroups {
		logStreams, err := p.describeLogStreams(logGroup)
		if err != nil {
			return nil, catcher.Error("cannot get log streams", err, nil)
		}

		for _, stream := range logStreams {
			paginator := cloudwatchlogs.NewGetLogEventsPaginator(cwl, &cloudwatchlogs.GetLogEventsInput{
				LogGroupName:  aws.String(logGroup),
				LogStreamName: aws.String(stream),
				StartTime:     aws.Int64(startTime.Unix() * 1000),
				EndTime:       aws.Int64(endTime.Unix() * 1000),
				StartFromHead: aws.Bool(true),
			})

			for paginator.HasMorePages() {
				page, err := paginator.NextPage(ctx)
				if err != nil {
					return nil, catcher.Error("cannot get logs", err, nil)
				}
				for _, event := range page.Events {
					transformedLogs = append(transformedLogs, *event.Message)
				}
			}
		}
	}

	return transformedLogs, nil
}
