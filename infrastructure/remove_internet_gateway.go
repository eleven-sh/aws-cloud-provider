package infrastructure

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func RemoveInternetGateway(
	ec2Client *ec2.Client,
	internetGatewayId string,
) error {

	_, err := ec2Client.DeleteInternetGateway(
		context.TODO(),
		&ec2.DeleteInternetGatewayInput{
			InternetGatewayId: &internetGatewayId,
		},
	)

	return err
}
