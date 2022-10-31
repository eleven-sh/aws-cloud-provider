package infrastructure

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func DetachInternetGatewayFromVPC(
	ec2Client *ec2.Client,
	internetGatewayId string,
	VPCID string,
) error {

	_, err := ec2Client.DetachInternetGateway(
		context.TODO(),
		&ec2.DetachInternetGatewayInput{
			InternetGatewayId: &internetGatewayId,
			VpcId:             &VPCID,
		},
	)

	return err
}
