package service_test

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/eleven-sh/aws-cloud-provider/config"
	"github.com/eleven-sh/aws-cloud-provider/mocks"
	"github.com/eleven-sh/aws-cloud-provider/service"
	"github.com/eleven-sh/aws-cloud-provider/userconfig"
	"github.com/golang/mock/gomock"
)

func TestServiceBuilderBuildWithResolvedUserConfig(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	resolvedUserConfig := userconfig.NewConfig("a", "b", "c")

	userConfigResolver := mocks.NewUserConfigResolver(mockCtrl)
	userConfigResolver.EXPECT().Resolve().Return(resolvedUserConfig, nil).Times(1)

	userConfigValidator := mocks.NewUserConfigValidator(mockCtrl)
	userConfigValidator.EXPECT().Validate(resolvedUserConfig).Return(nil).Times(1)

	userConfigLoader := mocks.NewUserConfigLoader(mockCtrl)
	userConfigLoader.EXPECT().Load(resolvedUserConfig).Return(aws.Config{}, nil).Times(1)

	builder := service.NewBuilder(
		userConfigResolver,
		userConfigValidator,
		userConfigLoader,
	)
	_, err := builder.Build()

	if err != nil {
		t.Fatalf("expected no error, got '%+v'", err)
	}
}

func TestServiceBuilderBuildWithUserConfigResolverError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	resolvedUserConfig := userconfig.NewConfig("a", "b", "c")

	userConfigResolverErr := userconfig.ErrMissingAccessKeyInEnv
	userConfigResolver := mocks.NewUserConfigResolver(mockCtrl)
	userConfigResolver.EXPECT().Resolve().Return(nil, userConfigResolverErr).Times(1)

	userConfigValidator := mocks.NewUserConfigValidator(mockCtrl)
	userConfigValidator.EXPECT().Validate(resolvedUserConfig).Return(nil).Times(0)

	userConfigLoader := mocks.NewUserConfigLoader(mockCtrl)
	userConfigLoader.EXPECT().Load(resolvedUserConfig).Return(aws.Config{}, nil).Times(0)

	builder := service.NewBuilder(
		userConfigResolver,
		userConfigValidator,
		userConfigLoader,
	)
	_, err := builder.Build()

	if err == nil {
		t.Fatalf("expected error, got nothing")
	}

	if !errors.Is(err, userConfigResolverErr) {
		t.Fatalf(
			"expected error to equal '%+v', got '%+v'",
			userConfigResolverErr,
			err,
		)
	}
}

func TestServiceBuilderBuildWithUserConfigValidatorError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	resolvedUserConfig := userconfig.NewConfig("a", "b", "c")

	userConfigResolver := mocks.NewUserConfigResolver(mockCtrl)
	userConfigResolver.EXPECT().Resolve().Return(resolvedUserConfig, nil).Times(1)

	userConfigValidatorErr := config.ErrInvalidAccessKeyID{}
	userConfigValidator := mocks.NewUserConfigValidator(mockCtrl)
	userConfigValidator.EXPECT().Validate(resolvedUserConfig).Return(userConfigValidatorErr).Times(1)

	userConfigLoader := mocks.NewUserConfigLoader(mockCtrl)
	userConfigLoader.EXPECT().Load(resolvedUserConfig).Return(aws.Config{}, nil).Times(0)

	builder := service.NewBuilder(
		userConfigResolver,
		userConfigValidator,
		userConfigLoader,
	)
	_, err := builder.Build()

	if err == nil {
		t.Fatalf("expected error, got nothing")
	}

	if !errors.Is(err, userConfigValidatorErr) {
		t.Fatalf(
			"expected error to equal '%+v', got '%+v'",
			userConfigValidatorErr,
			err,
		)
	}
}

func TestBuildWithConfigLoaderError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	resolvedUserConfig := userconfig.NewConfig("a", "b", "c")

	userConfigResolver := mocks.NewUserConfigResolver(mockCtrl)
	userConfigResolver.EXPECT().Resolve().Return(resolvedUserConfig, nil).Times(1)

	userConfigValidator := mocks.NewUserConfigValidator(mockCtrl)
	userConfigValidator.EXPECT().Validate(resolvedUserConfig).Return(nil).Times(1)

	unknownError := errors.New("UnknownError")
	userConfigLoader := mocks.NewUserConfigLoader(mockCtrl)
	userConfigLoader.EXPECT().Load(resolvedUserConfig).Return(aws.Config{}, unknownError).Times(1)

	builder := service.NewBuilder(
		userConfigResolver,
		userConfigValidator,
		userConfigLoader,
	)
	_, err := builder.Build()

	if err == nil {
		t.Fatalf("expected error, got nothing")
	}

	if !errors.Is(err, unknownError) {
		t.Fatalf(
			"expected error to equal '%+v', got '%+v'",
			unknownError,
			err,
		)
	}
}
