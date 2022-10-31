package config

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
)

type ProfileLoader struct{}

func NewProfileLoader() ProfileLoader {
	return ProfileLoader{}
}

func (ProfileLoader) Load(
	profile string,
	credentialsPath string,
	configPath string,
) (config.SharedConfig, error) {

	return config.LoadSharedConfigProfile(
		context.TODO(),
		profile,
		func(l *config.LoadSharedConfigOptions) {
			l.ConfigFiles = []string{configPath}
			l.CredentialsFiles = []string{credentialsPath}
		},
	)
}
