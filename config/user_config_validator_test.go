package config_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/eleven-sh/aws-cloud-provider/config"
	"github.com/eleven-sh/aws-cloud-provider/userconfig"
)

func TestUserConfigValidator(t *testing.T) {
	testCases := []struct {
		test          string
		userconfig    *userconfig.Config
		expectedError error
	}{
		{
			test: "with valid config",
			userconfig: &userconfig.Config{
				Credentials: userconfig.Credentials{
					AccessKeyID:     strings.Repeat("B", 20),
					SecretAccessKey: strings.Repeat("b", 40),
				},
				Region: "eu-west-1",
			},
			expectedError: nil,
		},

		{
			test: "with invalid region",
			userconfig: &userconfig.Config{
				Credentials: userconfig.Credentials{
					AccessKeyID:     strings.Repeat("B", 20),
					SecretAccessKey: strings.Repeat("b", 40),
				},
				Region: "invalid_region",
			},
			expectedError: config.ErrInvalidRegion{},
		},

		{
			test: "with invalid access key ID",
			userconfig: &userconfig.Config{
				Credentials: userconfig.Credentials{
					AccessKeyID:     "invalid_access_key_id",
					SecretAccessKey: strings.Repeat("b", 40),
				},
				Region: "us-east-1",
			},
			expectedError: config.ErrInvalidAccessKeyID{},
		},

		{
			test: "with invalid secret access key",
			userconfig: &userconfig.Config{
				Credentials: userconfig.Credentials{
					AccessKeyID:     strings.Repeat("B", 20),
					SecretAccessKey: "invalid_secret_access_key",
				},
				Region: "me-central-1",
			},
			expectedError: config.ErrInvalidSecretAccessKey{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.test, func(t *testing.T) {
			configvalidator := config.NewUserConfigValidator()
			err := configvalidator.Validate(tc.userconfig)

			if tc.expectedError == nil && err != nil {
				t.Fatalf("expected no error, got '%+v'", err)
			}

			if tc.expectedError == nil {
				return
			}

			if _, ok := tc.expectedError.(config.ErrInvalidRegion); ok {
				if !errors.As(err, &config.ErrInvalidRegion{}) {
					t.Fatalf(
						"expected error to equal '%+v', got '%+v'",
						tc.expectedError,
						err,
					)
				}
			}

			if _, ok := tc.expectedError.(config.ErrInvalidAccessKeyID); ok {
				if !errors.As(err, &config.ErrInvalidAccessKeyID{}) {
					t.Fatalf(
						"expected error to equal '%+v', got '%+v'",
						tc.expectedError,
						err,
					)
				}
			}

			if _, ok := tc.expectedError.(config.ErrInvalidSecretAccessKey); ok {
				if !errors.As(err, &config.ErrInvalidSecretAccessKey{}) {
					t.Fatalf(
						"expected error to equal '%+v', got '%+v'",
						tc.expectedError,
						err,
					)
				}
			}
		})
	}
}
