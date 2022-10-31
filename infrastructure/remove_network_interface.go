package infrastructure

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func RemoveNetworkInterface(
	ec2Client *ec2.Client,
	networkInterfaceID string,
) error {

	_, err := ec2Client.DeleteNetworkInterface(
		context.TODO(),
		&ec2.DeleteNetworkInterfaceInput{
			NetworkInterfaceId: &networkInterfaceID,
		},
	)

	return err
}
