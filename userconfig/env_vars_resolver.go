package userconfig

import (
	"errors"
)

var (
	// ErrMissingAccessKeyInEnv represents the error
	// returned when a secret is set but an access key cannot be found.
	ErrMissingAccessKeyInEnv = errors.New("ErrMissingAccessKeyInEnv")

	// ErrMissingSecretInEnv represents the error
	// returned when access key is set but a secret cannot be found.
	ErrMissingSecretInEnv = errors.New("ErrMissingSecretInEnv")

	// ErrMissingRegionInEnv represents the error
	// returned when access key and secret are set but a region was not
	// passed as an option nor set as a environment variable.
	ErrMissingRegionInEnv = errors.New("ErrMissingRegionInEnv")
)

const (
	// AWSAccessKeyIDEnvVar represents the environment variable name
	// that the resolver will look for when resolving the AWS access key id.
	AWSAccessKeyIDEnvVar = "AWS_ACCESS_KEY_ID"

	// AWSSecretAccessKeyEnvVar represents the environment variable name
	// that the resolver will look for when resolving the AWS secret access key.
	AWSSecretAccessKeyEnvVar = "AWS_SECRET_ACCESS_KEY"

	// AWSRegionEnvVar represents the environment variable name
	// that the resolver will look for when resolving the AWS region.
	AWSRegionEnvVar = "AWS_REGION"
)

//go:generate go run github.com/golang/mock/mockgen -destination ../mocks/user_config_env_vars_getter.go -package mocks -mock_names EnvVarsGetter=UserConfigEnvVarsGetter  github.com/eleven-sh/aws-cloud-provider/userconfig EnvVarsGetter
// EnvVarsGetter represents the interface
// used to access environment variables.
type EnvVarsGetter interface {
	Get(string) string
}

// EnvVarsResolverOpts represents the options
// used to configure the EnvVarsResolver.
type EnvVarsResolverOpts struct {
	// Region specifies which region will be used in the resulting config.
	// Default to the one found in sandbox if not set.
	Region string
}

// EnvVarsResolver retrieves the AWS account
// configuration from environment variables.
type EnvVarsResolver struct {
	opts    EnvVarsResolverOpts
	envVars EnvVarsGetter
}

// NewFilesResolver constructs the EnvVarsResolver struct.
func NewEnvVarsResolver(
	envVars EnvVarsGetter,
	opts EnvVarsResolverOpts,
) EnvVarsResolver {

	return EnvVarsResolver{
		opts:    opts,
		envVars: envVars,
	}
}

// Resolve retrieves the AWS account configuration
// from environment variables.
//
// The Region option takes precedence over the one
// found in sandbox.
//
// Partial configurations return an adequate errror.
//
// Env vars are retrieved via the EnvVarsGetter interface
// passed as constructor argument.
func (e EnvVarsResolver) Resolve() (*Config, error) {
	resolvedConfig := NewConfig(
		e.envVars.Get(AWSAccessKeyIDEnvVar),
		e.envVars.Get(AWSSecretAccessKeyEnvVar),
		e.resolveRegion(e.envVars.Get(AWSRegionEnvVar)),
	)

	if resolvedConfig.Credentials.HasKeys() &&
		len(resolvedConfig.Region) > 0 {

		return resolvedConfig, nil
	}

	if resolvedConfig.Credentials.HasKeys() &&
		len(resolvedConfig.Region) == 0 {

		return nil, ErrMissingRegionInEnv
	}

	if len(resolvedConfig.Credentials.AccessKeyID) == 0 &&
		len(resolvedConfig.Credentials.SecretAccessKey) > 0 {

		return nil, ErrMissingAccessKeyInEnv
	}

	if len(resolvedConfig.Credentials.AccessKeyID) > 0 &&
		len(resolvedConfig.Credentials.SecretAccessKey) == 0 {

		return nil, ErrMissingSecretInEnv
	}

	return nil, ErrMissingConfig
}

func (e EnvVarsResolver) resolveRegion(regionInEnvVars string) string {
	if len(e.opts.Region) > 0 {
		return e.opts.Region
	}

	return regionInEnvVars
}
