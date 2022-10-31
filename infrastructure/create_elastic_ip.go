package infrastructure

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type ElasticIP struct {
	ID                   string `json:"id"`
	Address              string `json:"address"`
	IsAttachedToInstance bool   `json:"is_attached_to_instance"`
	AssociationID        string `json:"association_id"`
}

func CreateElasticIP(
	ec2Client *ec2.Client,
	name string,
) (returnedElasticIP *ElasticIP, returnedError error) {

	createElasticIPResp, err := ec2Client.AllocateAddress(
		context.TODO(),
		&ec2.AllocateAddressInput{
			Domain: types.DomainTypeVpc,
			TagSpecifications: []types.TagSpecification{{
				ResourceType: types.ResourceTypeElasticIp,
				Tags: []types.Tag{{
					Key:   aws.String("Name"),
					Value: &name,
				}},
			}},
		},
	)

	if err != nil {
		returnedError = err
		return
	}

	returnedElasticIP = &ElasticIP{
		ID:                   *createElasticIPResp.AllocationId,
		Address:              *createElasticIPResp.PublicIp,
		IsAttachedToInstance: false,
	}
	return
}
