package infrastructure

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

const (
	UbuntuAMIRootUser         = "ubuntu"
	UbuntuAMINamePatternAmd64 = "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"
	UbuntuAMINamePatternArm64 = "ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-arm64-server-*"
)

type AMI struct {
	ID             string `json:"id"`
	RootUser       string `json:"root_user"`
	RootDeviceName string `json:"root_device_name"`
}

func LookupUbuntuAMIForArch(
	ec2Client *ec2.Client,
	arch InstanceTypeArch,
) (returnedAMI *AMI, returnedError error) {

	AMINamePattern := UbuntuAMINamePatternAmd64

	if arch == InstanceTypeArchArm64 {
		AMINamePattern = UbuntuAMINamePatternArm64
	}

	describeImagesResp, err := ec2Client.DescribeImages(
		context.TODO(),
		&ec2.DescribeImagesInput{
			Filters: []types.Filter{{
				Name: aws.String("name"),
				Values: []string{
					AMINamePattern,
				},
			}, {
				Name: aws.String("architecture"),
				Values: []string{
					string(arch),
				},
			}, {
				Name: aws.String("root-device-type"),
				Values: []string{
					"ebs",
				},
			}, {
				Name: aws.String("virtualization-type"),
				Values: []string{
					"hvm",
				},
			}},
			Owners: []string{
				"099720109477",
			},
		},
	)

	if err != nil {
		returnedError = err
		return
	}

	AMIs := describeImagesResp.Images

	if len(AMIs) == 0 {
		returnedError = errors.New("no AMIs found for arch and region")
		return
	}

	mostRecentAMI, err := getMostRecentAMI(AMIs)

	if err != nil {
		returnedError = err
		return
	}

	returnedAMI = &AMI{
		ID:             *mostRecentAMI.ImageId,
		RootUser:       UbuntuAMIRootUser,
		RootDeviceName: *mostRecentAMI.RootDeviceName,
	}
	return
}
