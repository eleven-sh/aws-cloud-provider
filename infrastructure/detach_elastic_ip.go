package infrastructure

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func DetachElasticIPFromInstance(
	ec2Client *ec2.Client,
	elasticIPAssociationId string,
) error {

	_, err := ec2Client.DisassociateAddress(
		context.TODO(),
		&ec2.DisassociateAddressInput{
			AssociationId: &elasticIPAssociationId,
		},
	)

	return err
}
