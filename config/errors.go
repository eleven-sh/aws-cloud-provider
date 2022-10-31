package config

// ErrInvalidRegion represents the error
// returned when the region in user config is invalid.
type ErrInvalidRegion struct {
	Region string
}

func (ErrInvalidRegion) Error() string {
	return "ErrInvalidRegion"
}

// ErrInvalidAccessKeyID represents the error
// returned when the access key ID in user config is invalid.
type ErrInvalidAccessKeyID struct {
	AccessKeyID string
}

func (ErrInvalidAccessKeyID) Error() string {
	return "ErrInvalidAccessKeyID"
}

// ErrInvalidSecretAccessKey represents the error
// returned when the secret access key in user config is invalid.
type ErrInvalidSecretAccessKey struct {
	SecretAccessKey string
}

func (ErrInvalidSecretAccessKey) Error() string {
	return "ErrInvalidSecretAccessKey"
}
