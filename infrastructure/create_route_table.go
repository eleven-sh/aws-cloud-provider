package infrastructure

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type RouteTable struct {
	ID                   string `json:"id"`
	IsAssociatedToSubnet bool   `json:"is_associated_to_subnet"`
}

func CreateRouteTable(
	ec2Client *ec2.Client,
	name string,
	VPCID string,
) (returnedRouteTable *RouteTable, returnedError error) {

	createRouteTableResp, err := ec2Client.CreateRouteTable(
		context.TODO(),
		&ec2.CreateRouteTableInput{
			VpcId: &VPCID,
			TagSpecifications: []types.TagSpecification{{
				ResourceType: types.ResourceTypeRouteTable,
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

	returnedRouteTable = &RouteTable{
		ID:                   *createRouteTableResp.RouteTable.RouteTableId,
		IsAssociatedToSubnet: false,
	}
	return
}
