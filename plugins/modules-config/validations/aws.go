package validations

import (
	"context"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/plugins/modules-config/config"
)

func ValidateAwsConfig(config *config.ModuleGroup) error {
	var regionName, accessKey, secretAccessKey string

	if config == nil {
		return catcher.Error("AWS_IAM_USER configuration is nil", nil, nil)
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
		return catcher.Error("Default Region is required in AWS_IAM_USER configuration", nil, nil)
	}
	if accessKey == "" {
		return catcher.Error("Access Key is required in AWS_IAM_USER configuration", nil, nil)
	}
	if secretAccessKey == "" {
		return catcher.Error("Secret Key is required in AWS_IAM_USER configuration", nil, nil)
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
		return catcher.Error("failed to load AWS configuration", err, nil)
	}

	stsClient := sts.NewFromConfig(cfg)

	_, err = stsClient.GetCallerIdentity(context.TODO(), &sts.GetCallerIdentityInput{})
	if err != nil {
		return catcher.Error("AWS credentials validation failed", err, nil)
	}

	return nil
}
