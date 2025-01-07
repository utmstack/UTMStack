package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"

	utmconf "github.com/utmstack/config-client-go"
	"github.com/utmstack/config-client-go/enum"
	"github.com/utmstack/config-client-go/types"
)

const delayCheck = 300
const defaultTenant string = "ce66672c-e36d-4761-a8c8-90058fee1a24"

var logQueue = make(chan *plugins.Log, 100*runtime.NumCPU())

func main() {
	mode := plugins.GetCfg().Env.Mode
	if mode != "manager" {
		os.Exit(0)
	}

	for t := 0; t < runtime.NumCPU(); t++ {
		go processLogs()
	}

	st := time.Now().Add(-600 * time.Second)
	for {
		utmConfig := plugins.PluginCfg("com.utmstack", false)
		internalKey := utmConfig.Get("internalKey").String()
		backendUrl := utmConfig.Get("backend").String()
		if internalKey == "" || backendUrl == "" {
			time.Sleep(5 * time.Second)
			continue
		}

		client := utmconf.NewUTMClient(internalKey, backendUrl)

		et := st.Add(299 * time.Second)
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
			st = et.Add(1)
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
						pull(st, et, group)
					}

					wg.Done()
				}(group)
			}

			wg.Wait()
		}

		time.Sleep(time.Second * delayCheck)
		st = et.Add(1)
	}
}

func etlProcess(events []*cloudwatchlogs.OutputLogEvent) []string {
	var logs = make([]string, 0)

	for _, event := range events {
		logs = append(logs, *event.Message)
	}

	return logs
}

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

func (p *AWSProcessor) createAWSSession() (*session.Session, error) {
	if p.RegionName == "" {
		return nil, catcher.Error("cannot create AWS session",
			errors.New("region name is empty"), nil)
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
		return nil, catcher.Error("cannot create AWS session", err, nil)
	}

	return sess, nil
}

func (p *AWSProcessor) DescribeLogGroups() ([]string, error) {
	awsSession, err := p.createAWSSession()
	if err != nil {
		return nil, catcher.Error("cannot create AWS session", err, nil)
	}

	cwl := cloudwatchlogs.New(awsSession)
	var logGroups []string
	err = cwl.DescribeLogGroupsPages(&cloudwatchlogs.DescribeLogGroupsInput{},
		func(page *cloudwatchlogs.DescribeLogGroupsOutput, lastPage bool) bool {
			for _, group := range page.LogGroups {
				logGroups = append(logGroups, aws.StringValue(group.LogGroupName))
			}
			return !lastPage
		})
	if err != nil {
		return nil, catcher.Error("cannot get log groups", err, nil)
	}

	return logGroups, nil
}

func (p *AWSProcessor) DescribeLogStreams(logGroup string) ([]string, error) {
	awsSession, err := p.createAWSSession()
	if err != nil {
		return nil, catcher.Error("cannot create AWS session", err, nil)
	}

	cwl := cloudwatchlogs.New(awsSession)
	var logStreams []string
	err = cwl.DescribeLogStreamsPages(&cloudwatchlogs.DescribeLogStreamsInput{
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
		return nil, catcher.Error("cannot get log streams", err, nil)
	}

	return logStreams, nil
}

func (p *AWSProcessor) GetLogs(startTime, endTime time.Time) ([]string, error) {
	awsSession, err := p.createAWSSession()
	if err != nil {
		return nil, catcher.Error("cannot create AWS session", err, nil)
	}

	cwl := cloudwatchlogs.New(awsSession)

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
			params := &cloudwatchlogs.GetLogEventsInput{
				LogGroupName:  aws.String(logGroup),
				LogStreamName: aws.String(stream),
				StartTime:     aws.Int64(startTime.Unix() * 1000),
				EndTime:       aws.Int64(endTime.Unix() * 1000),
				StartFromHead: aws.Bool(true),
			}

			err := cwl.GetLogEventsPages(params,
				func(page *cloudwatchlogs.GetLogEventsOutput, lastPage bool) bool {
					cleanLogs := etlProcess(page.Events)
					transformedLogs = append(transformedLogs, cleanLogs...)
					return !lastPage
				})
			if err != nil {
				return nil, catcher.Error("cannot get logs", err, nil)
			}
		}
	}

	return transformedLogs, nil
}

func pull(startTime time.Time, endTime time.Time, group types.ModuleGroup) {
	agent := GetAWSProcessor(group)

	logs, err := agent.GetLogs(startTime, endTime)
	if err != nil {
		_ = catcher.Error("cannot get logs", err, map[string]any{
			"startTime": startTime,
			"endTime":   endTime,
			"group":     group.GroupName,
		})
		return
	}

	for _, log := range logs {
		logQueue <- &plugins.Log{
			Id:         uuid.NewString(),
			TenantId:   defaultTenant,
			DataType:   "aws",
			DataSource: group.GroupName,
			Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
			Raw:        log,
		}
	}
}

func processLogs() {
	conn, err := grpc.NewClient(fmt.Sprintf("unix://%s",
		path.Join(plugins.GetCfg().Env.Workdir, "sockets", "engine_server.sock")),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		_ = catcher.Error("cannot connect to engine server", err, nil)
		os.Exit(1)
	}

	client := plugins.NewEngineClient(conn)

	inputClient, err := client.Input(context.Background())
	if err != nil {
		_ = catcher.Error("cannot create input client", err, nil)
		// TODO: notify engine about this error before exit
		os.Exit(1)
	}

	// TODO: implement retry system using ack and timeouts
	go func() {
		for {
			log := <-logQueue
			err := inputClient.Send(log)
			if err != nil {
				_ = catcher.Error("cannot send log", err, nil)
				// TODO: notify engine about this error before exit
				os.Exit(1)
			}
		}
	}()

	for {
		_, err = inputClient.Recv()
		if err != nil {
			_ = catcher.Error("cannot receive ack", err, nil)
			// TODO: notify engine about this error before exit
			os.Exit(1)
		}
	}
}
