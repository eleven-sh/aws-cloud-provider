package service

import (
	"github.com/aws/aws-sdk-go-v2/aws"
)

type AWS struct {
	sdkConfig aws.Config
}

func NewAWS(SDKConfig aws.Config) *AWS {
	return &AWS{
		sdkConfig: SDKConfig,
	}
}
