package userconfig

import (
	"errors"

	"github.com/aws/aws-sdk-go-v2/config"
)

var (
	// ErrMissingRegionInFiles represents the error
	// returned when no region is found in config files and
	// the Region option is not set.
	ErrMissingRegionInFiles = errors.New("ErrMissingRegionInFiles")
)

// ErrProfileNotFound represents the error
// returned when the Profile passed in option was not found.
type ErrProfileNotFound struct {
	Profile             string
	CredentialsFilePath string
	ConfigFilePath      string
}

func (ErrProfileNotFound) Error() string {
	return "ErrProfileNotFound"
}

const (
	// AWSConfigFileDefaultProfile represents the configuration profile
	// that will be loaded by default if the Profile option is not set.
	AWSConfigFileDefaultProfile = "default"
)

//go:generate go run github.com/golang/mock/mockgen -destination ../mocks/user_config_profile_loader.go -package mocks -mock_names ProfileLoader=UserConfigProfileLoader github.com/eleven-sh/aws-cloud-provider/userconfig ProfileLoader
// ProfileLoader represents the interface
// used to load configuration profile from files
type ProfileLoader interface {
	Load(
		profile string,
		credentialsFilePath string,
		configFilePath string,
	) (config.SharedConfig, error)
}

// FilesResolverOpts represents the options
// used to configure the FilesResolver.
type FilesResolverOpts struct {
	// Profile specifies which configuration profile will be loaded.
	// Default to AWSConfigFileDefaultProfile if not set.
	Profile string

	// Region specifies which region will be used in the resulting config.
	// Default to the one found in config files if not set.
	Region string

	// CredentialsFilePath specifies the file path of the credentials file.
	CredentialsFilePath string

	// ConfigFilePath specifies the file path of the config file.
	ConfigFilePath string
}

// FilesResolver retrieves the AWS account
// configuration from config files.
type FilesResolver struct {
	opts          FilesResolverOpts
	profileLoader ProfileLoader
	envVars       EnvVarsGetter
}

// NewFilesResolver constructs the FilesResolver struct.
func NewFilesResolver(
	profileLoader ProfileLoader,
	opts FilesResolverOpts,
	envVars EnvVarsGetter,
) FilesResolver {

	return FilesResolver{
		profileLoader: profileLoader,
		opts:          opts,
		envVars:       envVars,
	}
}

// Resolve retrieves the AWS account configuration from config files.
//
// The CredentialsFilePath and ConfigFilePath options
// are used to locate the credentials and config files.
//
// The Profile option specifies which configuration profile
// will be loaded (the function fallback to AWSConfigFileDefaultProfile
// if not set).
//
// The Region option takes precedence over the region found in config files.
//
// Config files are loaded via the ProfileLoader interface
// passed as constructor argument.
func (f FilesResolver) Resolve() (*Config, error) {
	loadedProfile, err := f.profileLoader.Load(
		f.resolveProfile(),
		f.opts.CredentialsFilePath,
		f.opts.ConfigFilePath,
	)

	if err != nil {
		if errors.As(err, &config.SharedConfigProfileNotExistError{}) {
			if len(f.opts.Profile) > 0 {
				return nil, ErrProfileNotFound{
					Profile:             f.opts.Profile,
					CredentialsFilePath: f.opts.CredentialsFilePath,
					ConfigFilePath:      f.opts.ConfigFilePath,
				}
			}

			return nil, ErrMissingConfig
		}

		return nil, err
	}

	if !loadedProfile.Credentials.HasKeys() {
		// the config file is set but
		// the credentials one is missing
		return nil, ErrMissingConfig
	}

	resolvedRegion := f.resolveRegion(loadedProfile.Region)

	if len(resolvedRegion) == 0 {
		return nil, ErrMissingRegionInFiles
	}

	resolvedConfig := NewConfig(
		loadedProfile.Credentials.AccessKeyID,
		loadedProfile.Credentials.SecretAccessKey,
		resolvedRegion,
	)

	return resolvedConfig, nil
}

func (f FilesResolver) resolveProfile() string {
	if len(f.opts.Profile) > 0 {
		return f.opts.Profile
	}

	return AWSConfigFileDefaultProfile
}

func (f FilesResolver) resolveRegion(regionInFile string) string {
	if len(f.opts.Region) > 0 {
		return f.opts.Region
	}

	if len(f.envVars.Get(AWSRegionEnvVar)) > 0 {
		return f.envVars.Get(AWSRegionEnvVar)
	}

	return regionInFile
}
