package infrastructure

import (
	"context"
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type InstanceTypeArch string

const (
	InstanceTypeArchArm64 = "arm64"
	InstanceTypeArchX8664 = "x86_64"
)

var (
	ErrInvalidInstanceType     = errors.New("ErrInvalidInstanceType")
	ErrInvalidInstanceTypeArch = errors.New("ErrInvalidInstanceTypeArch")

	SupportedInstanceTypeArchs = []string{
		string(InstanceTypeArchArm64),
		string(InstanceTypeArchX8664),
	}
)

type InstanceTypeInfos struct {
	Type string           `json:"type"`
	Arch InstanceTypeArch `json:"arch"`
}

func LookupInstanceTypeInfos(
	ec2Client *ec2.Client,
	instanceType string,
) (returnedInstanceTypeInfos *InstanceTypeInfos, returnedError error) {

	describeInstanceTypesResp, err := ec2Client.DescribeInstanceTypes(
		context.TODO(),
		&ec2.DescribeInstanceTypesInput{
			InstanceTypes: []types.InstanceType{
				types.InstanceType(instanceType),
			},
			Filters: []types.Filter{{
				Name:   aws.String("processor-info.supported-architecture"),
				Values: SupportedInstanceTypeArchs,
			}, {
				Name:   aws.String("supported-root-device-type"),
				Values: []string{"ebs"},
			}, {
				Name:   aws.String("supported-usage-class"),
				Values: []string{"on-demand"},
			}},
		},
	)

	if err != nil {
		if strings.Contains(err.Error(), "InvalidInstanceType") {
			returnedError = ErrInvalidInstanceType
			return
		}

		returnedError = err
		return
	}

	instanceTypes := describeInstanceTypesResp.InstanceTypes

	if len(instanceTypes) == 0 {
		returnedError = ErrInvalidInstanceTypeArch
		return
	}

	if len(instanceTypes) > 1 {
		returnedError = errors.New("multiple instance types match")
		return
	}

	supportedArchs := instanceTypes[0].ProcessorInfo.SupportedArchitectures

	returnedInstanceTypeInfos = &InstanceTypeInfos{
		Type: instanceType,
	}

	for _, supportedArch := range supportedArchs {
		if supportedArch == types.ArchitectureTypeArm64 {
			returnedInstanceTypeInfos.Arch = InstanceTypeArchArm64
			return
		}
	}

	returnedInstanceTypeInfos.Arch = InstanceTypeArchX8664
	return
}
