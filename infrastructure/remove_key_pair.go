package infrastructure

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func RemoveKeyPair(
	ec2Client *ec2.Client,
	keyPairID string,
) error {

	_, err := ec2Client.DeleteKeyPair(
		context.TODO(),
		&ec2.DeleteKeyPairInput{
			KeyPairId: &keyPairID,
		},
	)

	return err
}
