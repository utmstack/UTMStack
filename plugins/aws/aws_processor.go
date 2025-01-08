package main

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/config-client-go/types"
)

type AWSProcessor struct {
	RegionName      string
	AccessKey       string
	SecretAccessKey string
}

func GetAWSProcessor(group types.ModuleGroup) AWSProcessor {
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

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(p.RegionName),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(p.AccessKey, p.SecretAccessKey, "")),
	)
	if err != nil {
		return aws.Config{}, catcher.Error("cannot create AWS session", err, nil)
	}

	return cfg, nil
}

func (p *AWSProcessor) DescribeLogGroups() ([]string, error) {
	awsConfig, err := p.createAWSSession()
	if err != nil {
		return nil, catcher.Error("cannot create AWS session", err, nil)
	}

	cwl := cloudwatchlogs.NewFromConfig(awsConfig)
	var logGroups []string
	paginator := cloudwatchlogs.NewDescribeLogGroupsPaginator(cwl, &cloudwatchlogs.DescribeLogGroupsInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, catcher.Error("cannot get log groups", err, nil)
		}
		for _, group := range page.LogGroups {
			logGroups = append(logGroups, *group.LogGroupName)
		}
	}

	return logGroups, nil
}

func (p *AWSProcessor) DescribeLogStreams(logGroup string) ([]string, error) {
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
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, catcher.Error("cannot get log streams", err, nil)
		}
		for _, stream := range page.LogStreams {
			logStreams = append(logStreams, *stream.LogStreamName)
		}
	}

	return logStreams, nil
}

func (p *AWSProcessor) GetLogs(startTime, endTime time.Time) ([]string, error) {
	awsConfig, err := p.createAWSSession()
	if err != nil {
		return nil, catcher.Error("cannot create AWS session", err, nil)
	}

	cwl := cloudwatchlogs.NewFromConfig(awsConfig)

	logGroups, err := p.DescribeLogGroups()
	if err != nil {
		return nil, catcher.Error("cannot get log groups", err, nil)
	}

	transformedLogs := make([]string, 0)
	for _, logGroup := range logGroups {
		logStreams, err := p.DescribeLogStreams(logGroup)
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
				page, err := paginator.NextPage(context.TODO())
				if err != nil {
					return nil, catcher.Error("cannot get logs", err, nil)
				}
				cleanLogs := etlProcess(page.Events)
				transformedLogs = append(transformedLogs, cleanLogs...)
			}
		}
	}

	return transformedLogs, nil
}
