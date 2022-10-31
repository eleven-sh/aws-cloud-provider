package infrastructure

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type InternetGateway struct {
	ID              string `json:"id"`
	IsAttachedToVPC bool   `json:"is_attached_to_vpc"`
}

func CreateInternetGateway(
	ec2Client *ec2.Client,
	name string,
) (returnedIG *InternetGateway, returnedError error) {

	createInternetGatewayResp, err := ec2Client.CreateInternetGateway(
		context.TODO(),
		&ec2.CreateInternetGatewayInput{
			TagSpecifications: []types.TagSpecification{{
				ResourceType: types.ResourceTypeInternetGateway,
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

	defer func() {
		if returnedError == nil {
			return
		}

		_ = RemoveInternetGateway(
			ec2Client,
			*createInternetGatewayResp.InternetGateway.InternetGatewayId,
		)
	}()

	existsWaiter := ec2.NewInternetGatewayExistsWaiter(ec2Client)
	maxWaitTime := 5 * time.Minute

	err = existsWaiter.Wait(
		context.TODO(),
		&ec2.DescribeInternetGatewaysInput{
			InternetGatewayIds: []string{
				*createInternetGatewayResp.InternetGateway.InternetGatewayId,
			},
		},
		maxWaitTime,
	)

	if err != nil {
		returnedError = err
		return
	}

	returnedIG = &InternetGateway{
		ID:              *createInternetGatewayResp.InternetGateway.InternetGatewayId,
		IsAttachedToVPC: false,
	}
	return
}
