package service

import (
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/eleven-sh/aws-cloud-provider/infrastructure"
	"github.com/eleven-sh/eleven/stepper"
)

type ErrInvalidInstanceType struct {
	InstanceType string
	Region       string
}

func (ErrInvalidInstanceType) Error() string {
	return "ErrInvalidInstanceType"
}

type ErrInvalidInstanceTypeArch struct {
	InstanceType   string
	SupportedArchs string
}

func (ErrInvalidInstanceTypeArch) Error() string {
	return "ErrInvalidInstanceTypeArch"
}

func (a *AWS) CheckInstanceTypeValidity(
	stepper stepper.Stepper,
	instanceType string,
) error {

	ec2Client := ec2.NewFromConfig(a.sdkConfig)

	_, err := infrastructure.LookupInstanceTypeInfos(
		ec2Client,
		instanceType,
	)

	if err != nil {

		if errors.Is(err, infrastructure.ErrInvalidInstanceType) {
			return ErrInvalidInstanceType{
				InstanceType: instanceType,
				Region:       a.sdkConfig.Region,
			}
		}

		if errors.Is(err, infrastructure.ErrInvalidInstanceTypeArch) {
			return ErrInvalidInstanceTypeArch{
				InstanceType:   instanceType,
				SupportedArchs: strings.Join(infrastructure.SupportedInstanceTypeArchs, ", "),
			}
		}

		return err
	}

	return nil
}
