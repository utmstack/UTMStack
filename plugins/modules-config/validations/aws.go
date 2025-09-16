package validations

import (
	"context"
	"fmt"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/utmstack/UTMStack/plugins/modules-config/config"
)

func ValidateAwsConfig(config *config.ModuleGroup) error {
	var regionName, accessKey, secretAccessKey string

	if config == nil {
		return fmt.Errorf("AWS_IAM_USER configuration is nil")
	}

	for _, cnf := range config.ModuleGroupConfigurations {
		switch cnf.ConfKey {
		case "aws_default_region":
			regionName = cnf.ConfValue
		case "aws_access_key_id":
			accessKey = cnf.ConfValue
		case "aws_secret_access_key":
			secretAccessKey = cnf.ConfValue
		}
	}

	if regionName == "" {
		return fmt.Errorf("Default Region is required in AWS_IAM_USER configuration")
	}
	if accessKey == "" {
		return fmt.Errorf("Access Key is required in AWS_IAM_USER configuration")
	}
	if secretAccessKey == "" {
		return fmt.Errorf("Secret Key is required in AWS_IAM_USER configuration")
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
		return fmt.Errorf("failed to load AWS configuration: %w", err)
	}

	stsClient := sts.NewFromConfig(cfg)

	_, err = stsClient.GetCallerIdentity(context.TODO(), &sts.GetCallerIdentityInput{})
	if err != nil {
		return fmt.Errorf("AWS credentials validation failed: %w", err)
	}

	return nil
}
