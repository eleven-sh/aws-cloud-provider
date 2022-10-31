package config_test

import (
	"errors"
	"testing"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/eleven-sh/aws-cloud-provider/config"
	"github.com/eleven-sh/aws-cloud-provider/userconfig"
)

func TestProfileLoaderLoadWithExistingProfiles(t *testing.T) {
	testCases := []struct {
		test           string
		profile        string
		expectedConfig *userconfig.Config
	}{
		{
			test:    "with default profile",
			profile: userconfig.AWSConfigFileDefaultProfile,
			expectedConfig: userconfig.NewConfig(
				"default_access_key_id",
				"default_secret_access_key",
				"default_region",
			),
		},

		{
			test:    "with non-default profile",
			profile: "production",
			expectedConfig: userconfig.NewConfig(
				"production_access_key_id",
				"production_secret_access_key",
				"production_region",
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.test, func(t *testing.T) {
			profileLoader := config.NewProfileLoader()
			loadedProfile, err := profileLoader.Load(
				tc.profile,
				"./testdata/user_credentials",
				"./testdata/user_config",
			)

			if err != nil {
				t.Fatalf("expected no error, got '%+v'", err)
			}

			if loadedProfile.Credentials.AccessKeyID != tc.expectedConfig.Credentials.AccessKeyID ||
				loadedProfile.Credentials.SecretAccessKey != tc.expectedConfig.Credentials.SecretAccessKey ||
				loadedProfile.Region != tc.expectedConfig.Region {

				t.Fatalf("expected config to equal '%+v', got '%+v'", *tc.expectedConfig, loadedProfile)
			}
		})
	}
}

func TestProfileLoaderLoadWithInvalidProfile(t *testing.T) {
	profileLoader := config.NewProfileLoader()
	_, err := profileLoader.Load(
		"non_existing_profile",
		"./testdata/user_credentials",
		"./testdata/user_config",
	)

	if !errors.As(err, &awsconfig.SharedConfigProfileNotExistError{}) {
		t.Fatalf(
			"expected error to equal '%+v', got '%+v'",
			awsconfig.SharedConfigProfileNotExistError{},
			err,
		)
	}
}

func TestProfileLoaderLoadWithInvalidConfigFile(t *testing.T) {
	profileLoader := config.NewProfileLoader()
	_, err := profileLoader.Load(
		userconfig.AWSConfigFileDefaultProfile,
		"./testdata/invalid_user_config",
		"./testdata/user_credentials",
	)

	if !errors.As(err, &awsconfig.SharedConfigLoadError{}) {
		t.Fatalf(
			"expected error to equal '%+v', got '%+v'",
			awsconfig.SharedConfigProfileNotExistError{},
			err,
		)
	}
}
