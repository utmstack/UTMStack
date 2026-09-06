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
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"

	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/plugins/shared/coordination"
)

const (
	defaultTenant      = "ce66672c-e36d-4761-a8c8-90058fee1a24"
	urlCheckConnection = "https://sts.amazonaws.com"
	wait               = 1 * time.Second
)

// One goroutine polls per log stream, so idle request volume scales with
// stream count, not log volume. idlePollMax stays well under the 5-minute
// discovery interval in streamLogs so it never dominates ingestion lag.
const (
	idlePollBase = 5 * time.Second
	idlePollMax  = 60 * time.Second
)

type activeGroupStream struct {
	cancel  context.CancelFunc
	config  AWSProcessor
	cursors *cursorMap
}

var (
	activeStreams = make(map[string]*activeGroupStream)
)

// coordinationRetryDelay paces retries while coordination is unreachable.
const coordinationRetryDelay = 5 * time.Second

func main() {
	mode := plugins.GetCfg(processName).Env.Mode
	if mode != "worker" {
		return
	}

	go StartConfigurationSystem()

	for t := 0; t < 2*runtime.NumCPU(); t++ {
		go func() {
			plugins.SendLogsFromChannel(pluginName)
		}()
	}

	ctx := context.Background()

	waitForConnectivity(ctx, connectionChecker, urlCheckConnection, connectivityRetryDelay)

	// Blocks until coordination succeeds; there is no uncoordinated fallback,
	// because without a lease two workers can claim the same group.
	setup, err := coordination.SetupWithRetry(ctx, coordination.DialLeasePath(processName), coordinationRetryDelay, func(err error) {
		_ = catcher.Error("coordination setup failed, waiting to retry", err, map[string]any{
			"process": processName,
		})
	})
	if err != nil {
		// Only reachable if ctx is cancelled.
		_ = catcher.Error("coordination setup cancelled before it could succeed", err, map[string]any{"process": processName})
		return
	}
	defer setup.Close()

	holder := uuid.NewString()

	watchConfigChanges(ctx, setup, holder)
}

func watchConfigChanges(ctx context.Context, coord coordination.LeasePath, holder string) {
	time.Sleep(3 * time.Second)

	initialConfig := GetConfig()
	if initialConfig != nil && initialConfig.ModuleActive {
		syncStreams(ctx, coord, holder, initialConfig)
	}

	for newConfig := range GetConfigUpdateChannel() {
		if newConfig == nil || !newConfig.ModuleActive {
			stopAllStreams()
			continue
		}

		syncStreams(ctx, coord, holder, newConfig)
	}
}

func syncStreams(ctx context.Context, coord coordination.LeasePath, holder string, moduleConfig *ConfigurationSection) {
	currentKeys := make(map[string]bool)
	for _, group := range moduleConfig.ModuleGroups {
		currentConfig := getAWSProcessor(group)
		groupKey := group.Key()
		currentKeys[groupKey] = true

		existing := activeStreams[groupKey]

		if existing == nil {
			startGroupStream(ctx, coord, holder, groupKey, group)
		} else if existing.config != currentConfig {
			catcher.Info("Configuration changed for group, restarting", map[string]any{
				"group":   group.GroupName,
				"process": processName,
			})
			existing.cancel()
			delete(activeStreams, groupKey)
			startGroupStream(ctx, coord, holder, groupKey, group)
		}
	}

	for groupKey, stream := range activeStreams {
		if !currentKeys[groupKey] {
			catcher.Info("Group removed, stopping stream", map[string]any{
				"group":   groupKey,
				"process": processName,
			})
			stream.cancel()
			delete(activeStreams, groupKey)
		}
	}
}

func startGroupStream(ctx context.Context, coord coordination.LeasePath, holder, groupKey string, group *ModuleGroup) {
	groupCtx, cancel := context.WithCancel(ctx)

	groupConfig := getAWSProcessor(group)
	cursors := newCursorMap()

	activeStreams[groupKey] = &activeGroupStream{
		cancel:  cancel,
		config:  groupConfig,
		cursors: cursors,
	}

	catcher.Info("Starting stream for group", map[string]any{
		"group":   group.GroupName,
		"process": processName,
	})

	go acquireLeaseAndStream(groupCtx, coord, holder, groupKey, cursors, cancel, func() {
		streamLogs(groupCtx, group, groupKey, cursors)
	})
}

// Must be called in the group's own goroutine: this blocks until the lease
// is owned, so a contended lease would stall syncStreams for every group.
func acquireLeaseAndStream(ctx context.Context, coord coordination.LeasePath, holder, groupKey string, cursors *cursorMap, cancel context.CancelFunc, run func()) {
	leaseKey := "aws." + groupKey

	lease, err := coordination.AcquireWithRetry(
		ctx, coord.Leases, leaseKey, holder,
		coordination.LeaseTTL, coordination.AcquireRetryInterval,
		func(err error) {
			_ = catcher.Error("failed to acquire group lease", err, map[string]any{
				"process": processName,
				"group":   groupKey,
			})
		},
	)
	if err != nil {
		return
	}

	cursorRev := loadGroupCursor(ctx, coord.Cursors, leaseKey, groupKey, cursors)

	go groupHeartbeat(coord, leaseKey, cursors, cancel).Run(ctx, lease, cursorRev)

	run()
}

func loadGroupCursor(ctx context.Context, cursors coordination.CursorStore, leaseKey, groupKey string, dst *cursorMap) uint64 {
	var entries map[string]*string

	rev, found, err := coordination.LoadCursorInto(ctx, cursors, leaseKey, &entries)
	if err != nil {
		_ = catcher.Error("failed to load persisted cursor snapshot", err, map[string]any{
			"process": processName,
			"group":   groupKey,
		})
	}
	if found {
		dst.replace(entries)
	}

	// Returned even when the decode failed: the next Save needs this
	// revision to be a valid CAS update rather than a doomed create.
	return rev
}

func stopAllStreams() {
	if len(activeStreams) == 0 {
		return
	}

	catcher.Info("Stopping all active streams", map[string]any{
		"count":   len(activeStreams),
		"process": processName,
	})

	for groupKey, stream := range activeStreams {
		stream.cancel()
		delete(activeStreams, groupKey)
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

func streamLogs(ctx context.Context, group *ModuleGroup, groupKey string, cursors *cursorMap) {
	agent := getAWSProcessor(group)

	cwl, err := agent.client()
	if err != nil {
		_ = catcher.Error("cannot create AWS session", err, map[string]any{"process": processName})
		return
	}

	startTime := time.Now().UTC()

	catcher.Info("Starting streaming logs", map[string]any{
		"group":     group.GroupName,
		"logGroup":  agent.LogGroup,
		"startTime": startTime.Format(time.RFC3339),
		"process":   processName,
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
				"process":  processName,
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

			go streamLogStream(streamCtx, cwl, agent.LogGroup, stream, startTime, group.GroupName, agent.TenantId, groupKey, cursors)
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
					"process":   processName,
				})
				cancel()
				// Prune only after cancel, so no in-flight cursors.set can
				// land after the delete and resurrect the entry.
				delete(currentStreams, streamName)
				cursors.delete(streamName)
			}
		}

		if !sleepWithCancel(ctx, 5*time.Minute) {
			catcher.Info("Stream cancelled for group", map[string]any{
				"group":   group.GroupName,
				"process": processName,
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
	TenantId        string
}

func getAWSProcessor(group *ModuleGroup) AWSProcessor {
	awsPro := AWSProcessor{TenantId: group.TenantId}
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
			errors.New("region name is empty"), map[string]any{"process": processName})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	cfg, err := awsConfig.LoadDefaultConfig(ctx,
		awsConfig.WithRegion(p.RegionName),
		awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(p.AccessKey, p.SecretAccessKey, "")),
	)
	if err != nil {
		return aws.Config{}, catcher.Error("cannot create AWS session", err, map[string]any{"process": processName})
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
			return nil, catcher.Error("cannot get log streams", err, map[string]any{"process": processName})
		}
		for _, stream := range page.LogStreams {
			logStreams = append(logStreams, *stream.LogStreamName)
		}

		cancel()
	}

	return logStreams, nil
}

type logEventsAPI interface {
	GetLogEvents(ctx context.Context, params *cloudwatchlogs.GetLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.GetLogEventsOutput, error)
}

func streamLogStream(ctx context.Context, cwl logEventsAPI, logGroup, streamName string, startTime time.Time, dataSource, tenantId, groupKey string, cursors *cursorMap) {
	if tenantId == "" {
		tenantId = defaultTenant
	}
	nextToken := seedNextToken(cursors, streamName)
	processedCount := 0
	idleWait := idlePollBase

	for {
		select {
		case <-ctx.Done():
			catcher.Info("Log stream cancelled", map[string]any{
				"stream":     streamName,
				"totalCount": processedCount,
				"process":    processName,
			})
			return
		default:
		}

		// At end of stream CloudWatch returns the same token it was sent.
		// That equality is the only end-of-stream signal it gives.
		sentToken := nextToken

		input := &cloudwatchlogs.GetLogEventsInput{
			LogGroupName:  aws.String(logGroup),
			LogStreamName: aws.String(streamName),
			StartTime:     aws.Int64(startTime.Unix() * 1000),
			StartFromHead: aws.Bool(true),
			NextToken:     sentToken,
			// Left unset so the API default applies: 1 MB per response, up
			// to 10,000 events. An event count would not bound bytes.
		}

		requestCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		result, err := cwl.GetLogEvents(requestCtx, input)
		cancel()

		if err != nil {
			// Deleted or aged out of retention. Stop instead of retrying:
			// the discovery loop restarts this stream if it reappears.
			var missing *types.ResourceNotFoundException
			if errors.As(err, &missing) {
				catcher.Info("Log stream or group no longer exists, stopping", map[string]any{
					"logGroup":   logGroup,
					"stream":     streamName,
					"totalCount": processedCount,
					"process":    processName,
				})
				cursors.delete(streamName)
				return
			}

			// NextForwardToken expires after 24 hours, and CloudWatch then
			// rejects it as InvalidParameterException. Dropping it falls
			// back to StartTime, re-reading an overlap rather than stranding
			// the stream. The sentToken guard bounds this to one immediate
			// retry; discarding on every rejection would hot-loop the API.
			var badParam *types.InvalidParameterException
			if sentToken != nil && errors.As(err, &badParam) {
				_ = catcher.Error("log stream cursor rejected, falling back to the time floor", err, map[string]any{
					"logGroup":  logGroup,
					"stream":    streamName,
					"startTime": startTime.Format(time.RFC3339),
					"process":   processName,
				})
				nextToken = nil
				if ctx.Err() == nil {
					cursors.set(streamName, nil)
				}
				continue
			}

			_ = catcher.Error("cannot get log events", err, map[string]any{
				"logGroup": logGroup,
				"stream":   streamName,
				"process":  processName,
			})
			if !sleepWithCancel(ctx, 10*time.Second) {
				return
			}
			continue
		}

		eventsInBatch := 0
		for _, event := range result.Events {
			_ = plugins.EnqueueLog(&plugins.Log{
				Id:         eventIdentity(groupKey, logGroup, streamName, *event.Timestamp, *event.Message),
				TenantId:   tenantId,
				DataType:   "aws",
				DataSource: dataSource,
				Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
				Raw:        *event.Message,
			}, pluginName)
			processedCount++
			eventsInBatch++
		}

		if eventsInBatch > 0 {
			catcher.Info("Processed logs from stream", map[string]any{
				"stream":     streamName,
				"batchCount": eventsInBatch,
				"totalCount": processedCount,
				"dataSource": dataSource,
				"process":    processName,
			})
		}

		// A nil sentToken is the first call on an unseeded stream and can
		// never match, so it must not count as end of stream.
		atEnd := false
		if result.NextForwardToken != nil {
			atEnd = sentToken != nil && *sentToken == *result.NextForwardToken
			nextToken = result.NextForwardToken
			// Skip if already cancelled, so a pruned entry is not resurrected.
			if ctx.Err() == nil {
				cursors.set(streamName, nextToken)
			}
		} else {
			// No token to advance to, so there is nothing to poll forward.
			atEnd = true
		}

		if eventsInBatch > 0 {
			idleWait = idlePollBase
		}

		if !atEnd {
			// The token moved, so poll again immediately even on an empty
			// page: a sparse region returns those while still catching up.
			continue
		}

		if !sleepWithCancel(ctx, idleWait) {
			return
		}
		idleWait *= 2
		if idleWait > idlePollMax {
			idleWait = idlePollMax
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
			_ = catcher.Error("error closing response body: %v", err, map[string]any{"process": processName})
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
				_ = catcher.Error("An error occurred (%s), will keep retrying indefinitely...", err, map[string]any{"process": processName})
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
