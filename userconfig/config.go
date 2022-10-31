package userconfig

// Credentials represents the AWS credentials resolved from user config.
type Credentials struct {
	// AccessKeyID represents the access key.
	AccessKeyID string

	// SecretAccessKey represents the secret associated with the access key.
	SecretAccessKey string

	// SessionToken represents an AWS session token that
	// could replace the access key + secret set.
	// Currently not supported.
	SessionToken string
}

// HasKeys is an helper method used to check
// that the credentials in the Credentials struct are not empty.
func (u Credentials) HasKeys() bool {
	return len(u.AccessKeyID) > 0 && len(u.SecretAccessKey) > 0
}

// Config represents the resolved user config.
type Config struct {
	// Credentials represents the resolved credentials (access key + secret).
	Credentials Credentials

	// Region represents the resolved region.
	Region string
}

// NewConfig constructs a new resolved user config.
func NewConfig(
	accessKeyID string,
	secretAccessKey string,
	region string,
) *Config {

	return &Config{
		Credentials: Credentials{
			AccessKeyID:     accessKeyID,
			SecretAccessKey: secretAccessKey,
		},
		Region: region,
	}
}
