package config_test

import (
	"context"
	"testing"

	"github.com/eleven-sh/aws-cloud-provider/config"
	"github.com/eleven-sh/aws-cloud-provider/userconfig"
)

func TestUserConfigLoader(t *testing.T) {
	configLoader := config.NewUserConfigLoader()

	passedUserConfig := userconfig.NewConfig("a", "b", "c")
	loadedConfig, err := configLoader.Load(passedUserConfig)

	if err != nil {
		t.Fatalf("expected no error, got '%+v'", err)
	}

	if loadedConfig.Region != passedUserConfig.Region {
		t.Errorf(
			"expected region to equal '%s', got '%s'",
			passedUserConfig.Region,
			loadedConfig.Region,
		)
	}

	credsInConfig, err := loadedConfig.Credentials.Retrieve(context.TODO())

	if err != nil {
		t.Fatalf("expected no error, got '%+v'", err)
	}

	if credsInConfig.AccessKeyID != passedUserConfig.Credentials.AccessKeyID {
		t.Errorf(
			"expected access key id to equal '%s', got '%s'",
			passedUserConfig.Credentials.AccessKeyID,
			credsInConfig.AccessKeyID,
		)
	}

	if credsInConfig.SecretAccessKey != passedUserConfig.Credentials.SecretAccessKey {
		t.Errorf(
			"expected secret access key to equal '%s', got '%s'",
			passedUserConfig.Credentials.SecretAccessKey,
			credsInConfig.SecretAccessKey,
		)
	}
}
