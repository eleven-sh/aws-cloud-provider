package infrastructure

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func RemoveElasticIP(
	ec2Client *ec2.Client,
	elasticIPId string,
) error {

	_, err := ec2Client.ReleaseAddress(
		context.TODO(),
		&ec2.ReleaseAddressInput{
			AllocationId: &elasticIPId,
		},
	)

	return err
}
