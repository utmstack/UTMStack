package validations

import (
	"context"
	"fmt"
	"strings"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/utmstack/UTMStack/plugins/modules-config/config"
)

func ValidateAwsConfig(config *config.ModuleGroup) error {
	var regionName, accessKey, secretAccessKey, logGroup string

	if config == nil {
		return fmt.Errorf("AWS configuration is not provided")
	}

	for _, cnf := range config.ModuleGroupConfigurations {
		switch cnf.ConfKey {
		case "aws_default_region":
			regionName = cnf.ConfValue
		case "aws_access_key_id":
			accessKey = cnf.ConfValue
		case "aws_secret_access_key":
			secretAccessKey = cnf.ConfValue
		case "aws_log_group_name":
			logGroup = cnf.ConfValue
		}
	}

	if regionName == "" {
		return fmt.Errorf("Default Region is required in AWS configuration")
	}
	if accessKey == "" {
		return fmt.Errorf("Access Key is required in AWS configuration")
	}
	if secretAccessKey == "" {
		return fmt.Errorf("Secret Key is required in AWS configuration")
	}
	if logGroup == "" {
		return fmt.Errorf("Log Group is required in AWS configuration")
	}

	cfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion(regionName),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKey,
			secretAccessKey,
			"",
		)),
	)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "region") {
			return fmt.Errorf("Invalid Default Region '%s'. Please verify the AWS region is correct (e.g., us-east-1).", regionName)
		}
		return fmt.Errorf("Invalid AWS configuration. Please check your Access Key, Secret Key, and Default Region.")
	}

	stsClient := sts.NewFromConfig(cfg)

	_, err = stsClient.GetCallerIdentity(context.TODO(), &sts.GetCallerIdentityInput{})
	if err != nil {
		errMsg := err.Error()
		errLower := strings.ToLower(errMsg)
		if strings.Contains(errMsg, "InvalidClientTokenId") {
			return fmt.Errorf("Invalid Access Key ID. Please verify your AWS Access Key is correct.")
		}
		if strings.Contains(errMsg, "SignatureDoesNotMatch") {
			return fmt.Errorf("Invalid Secret Access Key. The Access Key ID appears correct but the Secret Key does not match.")
		}
		if strings.Contains(errMsg, "ExpiredToken") {
			return fmt.Errorf("AWS credentials have expired. Please generate new Access Key and Secret Key.")
		}
		if strings.Contains(errLower, "region") {
			return fmt.Errorf("Invalid Default Region '%s'. Please verify the AWS region is correct (e.g., us-east-1).", regionName)
		}
		return fmt.Errorf("AWS credentials are invalid. Please verify your Access Key and Secret Key are correct.")
	}

	cwlClient := cloudwatchlogs.NewFromConfig(cfg)
	logGroupInput := &cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePrefix: &logGroup,
	}

	logGroupOutput, err := cwlClient.DescribeLogGroups(context.TODO(), logGroupInput)
	if err != nil {
		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, "accessdenied") || strings.Contains(errMsg, "not authorized") {
			return fmt.Errorf("AWS credentials do not have permission to access CloudWatch in region '%s'. Please verify your IAM permissions.", regionName)
		}
		return fmt.Errorf("Cannot access CloudWatch Log Groups in region '%s'. Please verify the region and your IAM permissions.", regionName)
	}

	logGroupExists := false
	for _, lg := range logGroupOutput.LogGroups {
		if lg.LogGroupName != nil && *lg.LogGroupName == logGroup {
			logGroupExists = true
			break
		}
	}

	if !logGroupExists {
		return fmt.Errorf("The CloudWatch Log Group '%s' was not found in region '%s'. Please verify the Log Group name and region.", logGroup, regionName)
	}

	return nil
}
