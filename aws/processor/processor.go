package processor

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/threatwinds/logger"
	"github.com/utmstack/UTMStack/aws/utils"
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

func (p *AWSProcessor) createAWSSession() (*session.Session, *logger.Error) {
	if p.RegionName == "" {
		return nil, utils.Logger.ErrorF("Region is not configured")
	}

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(p.RegionName),
		Credentials: credentials.NewStaticCredentialsFromCreds(
			credentials.Value{
				AccessKeyID:     p.AccessKey,
				SecretAccessKey: p.SecretAccessKey,
			},
		),
	})
	if err != nil {
		return nil, utils.Logger.ErrorF("error creating aws session: %v", err)
	}

	return sess, nil
}

func (p *AWSProcessor) DescribeLogGroups() ([]string, *logger.Error) {
	sess, sessionErr := p.createAWSSession()
	if sessionErr != nil {
		return nil, sessionErr
	}

	cwl := cloudwatchlogs.New(sess)
	var logGroups []string
	err := cwl.DescribeLogGroupsPages(&cloudwatchlogs.DescribeLogGroupsInput{},
		func(page *cloudwatchlogs.DescribeLogGroupsOutput, lastPage bool) bool {
			for _, group := range page.LogGroups {
				logGroups = append(logGroups, aws.StringValue(group.LogGroupName))
			}
			return !lastPage
		})
	if err != nil {
		return nil, utils.Logger.ErrorF("error getting log groups: %v", err)
	}

	return logGroups, nil
}

func (p *AWSProcessor) DescribeLogStreams(logGroup string) ([]string, *logger.Error) {
	sess, sessionErr := p.createAWSSession()
	if sessionErr != nil {
		return nil, sessionErr
	}

	cwl := cloudwatchlogs.New(sess)
	var logStreams []string
	err := cwl.DescribeLogStreamsPages(&cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName: aws.String(logGroup),
		OrderBy:      aws.String("LastEventTime"),
		Descending:   aws.Bool(true),
	},
		func(page *cloudwatchlogs.DescribeLogStreamsOutput, lastPage bool) bool {
			for _, stream := range page.LogStreams {
				logStreams = append(logStreams, aws.StringValue(stream.LogStreamName))
			}
			return !lastPage
		})
	if err != nil {
		return nil, utils.Logger.ErrorF("error getting log streams: %v", err)
	}

	return logStreams, nil
}

func (p *AWSProcessor) GetLogs(startTime, endTime time.Time, group types.ModuleGroup) ([]TransformedLog, *logger.Error) {
	transformedLogs := []TransformedLog{}

	sess, sessionErr := p.createAWSSession()
	if sessionErr != nil {
		return nil, sessionErr
	}

	cwl := cloudwatchlogs.New(sess)

	logGroups, err := p.DescribeLogGroups()
	if err != nil {
		return nil, err
	}

	for _, logGroup := range logGroups {
		logStreams, err := p.DescribeLogStreams(logGroup)
		if err != nil {
			return nil, err
		}

		for _, stream := range logStreams {
			utils.Logger.Info("Processing stream %s from group %s", stream, logGroup)
			params := &cloudwatchlogs.GetLogEventsInput{
				LogGroupName:  aws.String(logGroup),
				LogStreamName: aws.String(stream),
				StartTime:     aws.Int64(startTime.Unix() * 1000),
				EndTime:       aws.Int64(endTime.Unix() * 1000),
				StartFromHead: aws.Bool(true),
			}

			err := cwl.GetLogEventsPages(params,
				func(page *cloudwatchlogs.GetLogEventsOutput, lastPage bool) bool {
					cleanLogs := ETLProcess(page.Events, group, logGroup, stream)
					transformedLogs = append(transformedLogs, cleanLogs...)
					return !lastPage
				})
			if err != nil {
				return nil, utils.Logger.ErrorF("error getting log pages: %v", err)
			}
		}
	}

	return transformedLogs, nil
}
