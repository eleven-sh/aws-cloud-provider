package infrastructure

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func AttachInternetGatewayToVPC(
	ec2Client *ec2.Client,
	internetGatewayId string,
	VPCID string,
) error {

	_, err := ec2Client.AttachInternetGateway(
		context.TODO(),
		&ec2.AttachInternetGatewayInput{
			InternetGatewayId: &internetGatewayId,
			VpcId:             &VPCID,
		},
	)

	return err
}
