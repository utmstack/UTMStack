package verifier

import (
	"context"
	"fmt"
	"strings"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// verifyAWSIAMUser — ported from modulekinds/aws_iam_user/validate.go.
func verifyAWSIAMUser(c map[string]string) error {
	if err := required(c, "AWS",
		"aws_default_region",
		"aws_access_key_id",
		"aws_secret_access_key",
		"aws_log_group_name",
	); err != nil {
		return err
	}

	regionName := c["aws_default_region"]
	accessKey := c["aws_access_key_id"]
	secretAccessKey := c["aws_secret_access_key"]
	logGroup := c["aws_log_group_name"]

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(regionName),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretAccessKey, "")),
	)
	if err != nil {
		if strings.Contains(err.Error(), "region") {
			return fmt.Errorf("Invalid Default Region '%s'. Please verify the AWS region is correct (e.g., us-east-1).", regionName)
		}
		return fmt.Errorf("Invalid AWS configuration. Please check your Access Key, Secret Key, and Default Region.")
	}

	stsClient := sts.NewFromConfig(cfg)
	if _, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{}); err != nil {
		errMsg := err.Error()
		errLower := strings.ToLower(errMsg)
		switch {
		case strings.Contains(errMsg, "InvalidClientTokenId"):
			return fmt.Errorf("Invalid Access Key ID. Please verify your AWS Access Key is correct.")
		case strings.Contains(errMsg, "SignatureDoesNotMatch"):
			return fmt.Errorf("Invalid Secret Access Key. The Access Key ID appears correct but the Secret Key does not match.")
		case strings.Contains(errMsg, "ExpiredToken"):
			return fmt.Errorf("AWS credentials have expired. Please generate new Access Key and Secret Key.")
		case strings.Contains(errLower, "region"):
			return fmt.Errorf("Invalid Default Region '%s'. Please verify the AWS region is correct (e.g., us-east-1).", regionName)
		}
		return fmt.Errorf("AWS credentials are invalid. Please verify your Access Key and Secret Key are correct.")
	}

	cwlClient := cloudwatchlogs.NewFromConfig(cfg)
	out, err := cwlClient.DescribeLogGroups(ctx, &cloudwatchlogs.DescribeLogGroupsInput{LogGroupNamePrefix: &logGroup})
	if err != nil {
		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, "accessdenied") || strings.Contains(errMsg, "not authorized") {
			return fmt.Errorf("AWS credentials do not have permission to access CloudWatch in region '%s'. Please verify your IAM permissions.", regionName)
		}
		return fmt.Errorf("Cannot access CloudWatch Log Groups in region '%s'. Please verify the region and your IAM permissions.", regionName)
	}

	for _, lg := range out.LogGroups {
		if lg.LogGroupName != nil && *lg.LogGroupName == logGroup {
			return nil
		}
	}
	return fmt.Errorf("The CloudWatch Log Group '%s' was not found in region '%s'. Please verify the Log Group name and region.", logGroup, regionName)
}
