package infrastructure

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type KeyPair struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	PEMContent string `json:"pem_content"`
}

func CreateKeyPair(
	ec2Client *ec2.Client,
	keyPairName string,
) (returnedKeyPair *KeyPair, returnedError error) {

	createKeyPairResp, err := ec2Client.CreateKeyPair(
		context.TODO(),
		&ec2.CreateKeyPairInput{
			KeyName: &keyPairName,
			KeyType: types.KeyTypeEd25519,
		},
	)

	if err != nil {
		returnedError = err
		return
	}

	defer func() {
		if returnedError == nil {
			return
		}

		_ = RemoveKeyPair(ec2Client, *createKeyPairResp.KeyPairId)
	}()

	existsWaiter := ec2.NewKeyPairExistsWaiter(ec2Client)
	maxWaitTime := 5 * time.Minute

	err = existsWaiter.Wait(context.TODO(), &ec2.DescribeKeyPairsInput{
		KeyPairIds: []string{
			*createKeyPairResp.KeyPairId,
		},
	}, maxWaitTime)

	if err != nil {
		returnedError = err
		return
	}

	returnedKeyPair = &KeyPair{
		ID:         *createKeyPairResp.KeyPairId,
		Name:       *createKeyPairResp.KeyName,
		PEMContent: *createKeyPairResp.KeyMaterial,
	}
	return
}
