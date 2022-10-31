package infrastructure

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func RemoveSubnet(
	ec2Client *ec2.Client,
	subnetID string,
) error {

	_, err := ec2Client.DeleteSubnet(
		context.TODO(),
		&ec2.DeleteSubnetInput{
			SubnetId: &subnetID,
		},
	)

	return err
}
