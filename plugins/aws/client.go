package main

import (
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"

	"github.com/utmstack/UTMStack/plugins/shared/identity"
)

// clientCache hands out one client per (region, credentials). CloudWatch
// Logs throttling quotas are enforced per account and region and shared by
// every group using them, so client scope must match the quota, not the group.
type clientCache struct {
	mu      sync.Mutex
	clients map[string]*cloudwatchlogs.Client
}

func newClientCache() *clientCache {
	return &clientCache{clients: make(map[string]*cloudwatchlogs.Client)}
}

// Credentials are hashed because this key reaches log lines and error
// contexts, where a secret access key must never appear.
func clientCacheKey(region, accessKey, secretAccessKey string) string {
	return region + "|" + identity.Hash(accessKey, secretAccessKey)
}

// The lock is intentionally not held across construction: it blocks on I/O
// for up to a minute, so holding it would let one bad credential set stall
// every other group's lookup. The re-check below discards a racing duplicate.
func (c *clientCache) get(region, accessKey, secretAccessKey string) (*cloudwatchlogs.Client, error) {
	key := clientCacheKey(region, accessKey, secretAccessKey)

	c.mu.Lock()
	cached, ok := c.clients[key]
	c.mu.Unlock()
	if ok {
		return cached, nil
	}

	client, err := newCloudWatchLogsClient(region, accessKey, secretAccessKey)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if winner, ok := c.clients[key]; ok {
		return winner, nil
	}
	c.clients[key] = client
	return client, nil
}

var awsClients = newClientCache()

func newCloudWatchLogsClient(region, accessKey, secretAccessKey string) (*cloudwatchlogs.Client, error) {
	processor := AWSProcessor{
		RegionName:      region,
		AccessKey:       accessKey,
		SecretAccessKey: secretAccessKey,
	}

	cfg, err := processor.createAWSSession()
	if err != nil {
		return nil, err
	}

	return cloudwatchlogs.NewFromConfig(cfg), nil
}

func (p *AWSProcessor) client() (*cloudwatchlogs.Client, error) {
	return awsClients.get(p.RegionName, p.AccessKey, p.SecretAccessKey)
}
