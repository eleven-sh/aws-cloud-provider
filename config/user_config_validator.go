package config

import (
	"regexp"

	"github.com/eleven-sh/aws-cloud-provider/userconfig"
)

const (
	awsAccessKeyIDPattern     = "^[A-Z0-9]{20}$"
	awsSecretAccessKeyPattern = "^[A-Za-z0-9/+=]{40}$"
)

var validAWSRegions = map[string]bool{
	"us-east-2":      true,
	"us-east-1":      true,
	"us-west-1":      true,
	"us-west-2":      true,
	"af-south-1":     true,
	"ap-east-1":      true,
	"ap-southeast-3": true,
	"ap-south-1":     true,
	"ap-northeast-3": true,
	"ap-southeast-1": true,
	"ap-southeast-2": true,
	"ap-northeast-1": true,
	"ca-central-1":   true,
	"eu-central-1":   true,
	"eu-west-1":      true,
	"eu-west-2":      true,
	"eu-south-1":     true,
	"eu-west-3":      true,
	"eu-north-1":     true,
	"me-south-1":     true,
	"me-central-1":   true,
	"sa-east-1":      true,
}

type UserConfigValidator struct{}

func NewUserConfigValidator() UserConfigValidator {
	return UserConfigValidator{}
}

func (u UserConfigValidator) Validate(userConfig *userconfig.Config) error {
	region := userConfig.Region

	if err := u.validateRegion(region); err != nil {
		return err
	}

	creds := userConfig.Credentials
	accessKeyID := creds.AccessKeyID
	secretAccessKey := creds.SecretAccessKey

	if err := u.validateAccessKeyID(accessKeyID); err != nil {
		return err
	}

	if err := u.validateSecretAccessKey(secretAccessKey); err != nil {
		return err
	}

	return nil
}

func (UserConfigValidator) validateRegion(region string) error {
	if _, ok := validAWSRegions[region]; !ok {
		return ErrInvalidRegion{
			Region: region,
		}
	}

	return nil
}

func (UserConfigValidator) validateAccessKeyID(accessKeyID string) error {
	match, err := regexp.MatchString(awsAccessKeyIDPattern, accessKeyID)

	if err != nil {
		return err
	}

	if !match {
		return ErrInvalidAccessKeyID{
			AccessKeyID: accessKeyID,
		}
	}

	return nil
}

func (UserConfigValidator) validateSecretAccessKey(secretAccessKey string) error {
	match, err := regexp.MatchString(awsSecretAccessKeyPattern, secretAccessKey)

	if err != nil {
		return err
	}

	if !match {
		return ErrInvalidSecretAccessKey{
			SecretAccessKey: secretAccessKey,
		}
	}

	return nil
}
