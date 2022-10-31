package userconfig_test

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/eleven-sh/aws-cloud-provider/mocks"
	"github.com/eleven-sh/aws-cloud-provider/userconfig"
	"github.com/golang/mock/gomock"
)

func TestFilesResolving(t *testing.T) {
	unknownError := errors.New("UnknownError")

	testCases := []struct {
		test                    string
		configInFiles           *userconfig.Config
		regionOpts              string
		regionEnvVar            string
		profileOpts             string
		credentialsFilePathOpts string
		configFilePathOpts      string
		errorReturnedByLoader   error
		expectedError           error
		expectedConfig          *userconfig.Config
	}{
		{
			test:           "valid",
			configInFiles:  userconfig.NewConfig("a", "b", "c"),
			expectedConfig: userconfig.NewConfig("a", "b", "c"),
			expectedError:  nil,
		},

		{
			test:           "valid with region option",
			configInFiles:  userconfig.NewConfig("a", "b", "c"),
			regionOpts:     "d",
			expectedConfig: userconfig.NewConfig("a", "b", "d"),
			expectedError:  nil,
		},

		{
			test:           "valid with region sets as env var",
			configInFiles:  userconfig.NewConfig("a", "b", "c"),
			regionEnvVar:   "d",
			expectedConfig: userconfig.NewConfig("a", "b", "d"),
			expectedError:  nil,
		},

		{
			test:           "valid with region sets as env var and option",
			configInFiles:  userconfig.NewConfig("a", "b", "c"),
			regionOpts:     "d",
			regionEnvVar:   "e",
			expectedConfig: userconfig.NewConfig("a", "b", "d"),
			expectedError:  nil,
		},

		{
			test:           "missing region",
			configInFiles:  userconfig.NewConfig("a", "b", ""),
			expectedError:  userconfig.ErrMissingRegionInFiles,
			expectedConfig: nil,
		},

		{
			test:           "missing region with region option",
			configInFiles:  userconfig.NewConfig("a", "b", ""),
			regionOpts:     "c",
			expectedConfig: userconfig.NewConfig("a", "b", "c"),
			expectedError:  nil,
		},

		{
			test:           "missing credentials",
			configInFiles:  userconfig.NewConfig("", "", "c"),
			expectedError:  userconfig.ErrMissingConfig,
			expectedConfig: nil,
		},

		{
			test:                  "missing config files without profile option",
			errorReturnedByLoader: config.SharedConfigProfileNotExistError{},
			expectedError:         userconfig.ErrMissingConfig,
			expectedConfig:        nil,
		},

		{
			test:                  "missing config files with profile option",
			profileOpts:           "profile",
			errorReturnedByLoader: config.SharedConfigProfileNotExistError{},
			expectedError:         userconfig.ErrProfileNotFound{},
			expectedConfig:        nil,
		},

		{
			test:                  "unknown error during config loading",
			errorReturnedByLoader: unknownError,
			expectedError:         unknownError,
			expectedConfig:        nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.test, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			profileToLoad := tc.profileOpts
			if len(profileToLoad) == 0 {
				profileToLoad = userconfig.AWSConfigFileDefaultProfile
			}

			configAsReturnedByProfileLoader := config.SharedConfig{}
			if tc.configInFiles != nil {
				configAsReturnedByProfileLoader.Credentials.AccessKeyID = tc.configInFiles.Credentials.AccessKeyID
				configAsReturnedByProfileLoader.Credentials.SecretAccessKey = tc.configInFiles.Credentials.SecretAccessKey
				configAsReturnedByProfileLoader.Region = tc.configInFiles.Region
			}

			profileLoaderMock := mocks.NewUserConfigProfileLoader(mockCtrl)
			profileLoaderMock.
				EXPECT().
				Load(profileToLoad, tc.credentialsFilePathOpts, tc.configFilePathOpts).
				Return(configAsReturnedByProfileLoader, tc.errorReturnedByLoader).
				AnyTimes()

			envVarsGetterMock := mocks.NewUserConfigEnvVarsGetter(mockCtrl)
			envVarsGetterMock.
				EXPECT().
				Get(userconfig.AWSRegionEnvVar).
				Return(tc.regionEnvVar).
				AnyTimes()

			resolver := userconfig.NewFilesResolver(
				profileLoaderMock,
				userconfig.FilesResolverOpts{
					Region:              tc.regionOpts,
					Profile:             tc.profileOpts,
					CredentialsFilePath: tc.credentialsFilePathOpts,
					ConfigFilePath:      tc.configFilePathOpts,
				},
				envVarsGetterMock,
			)

			resolvedConfig, err := resolver.Resolve()

			if tc.expectedError == nil && err != nil {
				t.Fatalf("expected no error, got '%+v'", err)
			}

			if tc.expectedError != nil && !errors.Is(err, tc.expectedError) && !errors.As(err, &tc.expectedError) {
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
