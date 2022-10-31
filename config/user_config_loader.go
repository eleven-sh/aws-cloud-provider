package config

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/eleven-sh/aws-cloud-provider/userconfig"
)

type UserConfigLoader struct{}

func NewUserConfigLoader() UserConfigLoader {
	return UserConfigLoader{}
}

func (UserConfigLoader) Load(userConfig *userconfig.Config) (aws.Config, error) {
	return config.LoadDefaultConfig(
		context.TODO(),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				userConfig.Credentials.AccessKeyID,
				userConfig.Credentials.SecretAccessKey,
				userConfig.Credentials.SessionToken,
			),
		),
		config.WithRegion(userConfig.Region),
	)
}
