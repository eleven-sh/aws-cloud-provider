package userconfig_test

import (
	"errors"
	"testing"

	"github.com/eleven-sh/aws-cloud-provider/mocks"
	"github.com/eleven-sh/aws-cloud-provider/userconfig"
	"github.com/golang/mock/gomock"
)

func TestEnvVarsResolving(t *testing.T) {
	testCases := []struct {
		test                  string
		accessKeyIDEnvVar     string
		secretAccessKeyEnvVar string
		regionEnvVar          string
		regionOpts            string
		expectedError         error
		expectedConfig        *userconfig.Config
	}{
		{
			test:                  "valid",
			accessKeyIDEnvVar:     "a",
			secretAccessKeyEnvVar: "b",
			regionEnvVar:          "c",
			expectedConfig:        userconfig.NewConfig("a", "b", "c"),
			expectedError:         nil,
		},

		{
			test:                  "valid with region opts",
			accessKeyIDEnvVar:     "a",
			secretAccessKeyEnvVar: "b",
			regionEnvVar:          "c",
			regionOpts:            "d",
			expectedConfig:        userconfig.NewConfig("a", "b", "d"),
			expectedError:         nil,
		},

		{
			test:                  "missing region with region opts",
			accessKeyIDEnvVar:     "a",
			secretAccessKeyEnvVar: "b",
			regionOpts:            "c",
			expectedConfig:        userconfig.NewConfig("a", "b", "c"),
			expectedError:         nil,
		},

		{
			test:                  "missing region",
			accessKeyIDEnvVar:     "a",
			secretAccessKeyEnvVar: "b",
			expectedError:         userconfig.ErrMissingRegionInEnv,
			expectedConfig:        nil,
		},

		{
			test:                  "missing access key",
			secretAccessKeyEnvVar: "b",
			regionEnvVar:          "c",
			expectedError:         userconfig.ErrMissingAccessKeyInEnv,
			expectedConfig:        nil,
		},

		{
			test:                  "missing access key and region",
			secretAccessKeyEnvVar: "b",
			expectedError:         userconfig.ErrMissingAccessKeyInEnv,
			expectedConfig:        nil,
		},

		{
			test:              "missing secret",
			accessKeyIDEnvVar: "b",
			regionEnvVar:      "c",
			expectedError:     userconfig.ErrMissingSecretInEnv,
			expectedConfig:    nil,
		},

		{
			test:              "missing secret and region",
			accessKeyIDEnvVar: "b",
			expectedError:     userconfig.ErrMissingSecretInEnv,
			expectedConfig:    nil,
		},

		{
			test:           "missing access key and secret",
			regionEnvVar:   "a",
			expectedError:  userconfig.ErrMissingConfig,
			expectedConfig: nil,
		},

		{
			test:           "no env vars",
			expectedError:  userconfig.ErrMissingConfig,
			expectedConfig: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.test, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			envVarsGetterMock := mocks.NewUserConfigEnvVarsGetter(mockCtrl)
			envVarsGetterMock.EXPECT().Get(userconfig.AWSAccessKeyIDEnvVar).Return(tc.accessKeyIDEnvVar).AnyTimes()
			envVarsGetterMock.EXPECT().Get(userconfig.AWSSecretAccessKeyEnvVar).Return(tc.secretAccessKeyEnvVar).AnyTimes()
			envVarsGetterMock.EXPECT().Get(userconfig.AWSRegionEnvVar).Return(tc.regionEnvVar).AnyTimes()

			resolver := userconfig.NewEnvVarsResolver(
				envVarsGetterMock,
				userconfig.EnvVarsResolverOpts{
					Region: tc.regionOpts,
				},
			)

			resolvedConfig, err := resolver.Resolve()

			if tc.expectedError == nil && err != nil {
				t.Fatalf("expected no error, got '%+v'", err)
			}

			if tc.expectedError != nil && !errors.Is(err, tc.expectedError) {
				t.Fatalf("expected error to equal '%+v', got '%+v'", tc.expectedError, err)
			}

			if tc.expectedConfig != nil && *resolvedConfig != *tc.expectedConfig {
				t.Fatalf("expected config to equal '%+v', got '%+v'", *tc.expectedConfig, *resolvedConfig)
			}

			if tc.expectedConfig == nil && resolvedConfig != nil {
				t.Fatalf("expected no config, got '%+v'", *resolvedConfig)
			}
		})
	}
}
