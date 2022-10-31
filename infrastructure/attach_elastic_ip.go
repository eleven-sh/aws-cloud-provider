package infrastructure

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func AttachElasticIPToInstance(
	ec2Client *ec2.Client,
	elasticIPId string,
	instanceId string,
) (string, error) {

	attachElasticIPResp, err := ec2Client.AssociateAddress(
		context.TODO(),
		&ec2.AssociateAddressInput{
			AllocationId: &elasticIPId,
			InstanceId:   &instanceId,
		},
	)

	if err != nil {
		return "", err
	}

	return *attachElasticIPResp.AssociationId, nil
}
