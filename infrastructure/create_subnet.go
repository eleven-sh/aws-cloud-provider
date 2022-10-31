package infrastructure

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type Subnet struct {
	ID               string `json:"id"`
	AvailabilityZone string `json:"availability_zone"`
}

func CreateSubnet(
	ec2Client *ec2.Client,
	name string,
	cidrBlock string,
	VPCID string,
) (returnedSubnet *Subnet, returnedError error) {

	createSubnetResp, err := ec2Client.CreateSubnet(
		context.TODO(),
		&ec2.CreateSubnetInput{
			CidrBlock: &cidrBlock,
			VpcId:     &VPCID,
			TagSpecifications: []types.TagSpecification{{
				ResourceType: types.ResourceTypeSubnet,
				Tags: []types.Tag{{
					Key:   aws.String("Name"),
					Value: &name,
				}},
			},
			},
		})

	if err != nil {
		returnedError = err
		return
	}

	defer func() {
		if returnedError == nil {
			return
		}

		_ = RemoveSubnet(ec2Client, *createSubnetResp.Subnet.SubnetId)
	}()

	availableWaiter := ec2.NewSubnetAvailableWaiter(ec2Client)
	maxWaitTime := 5 * time.Minute

	err = availableWaiter.Wait(context.TODO(), &ec2.DescribeSubnetsInput{
		SubnetIds: []string{
			*createSubnetResp.Subnet.SubnetId,
		},
	}, maxWaitTime)

	if err != nil {
		returnedError = err
		return
	}

	_, err = ec2Client.ModifySubnetAttribute(
		context.TODO(),
		&ec2.ModifySubnetAttributeInput{
			SubnetId: createSubnetResp.Subnet.SubnetId,
			MapPublicIpOnLaunch: &types.AttributeBooleanValue{
				Value: aws.Bool(true),
			},
		},
	)

	if err != nil {
		returnedError = err
		return
	}

	returnedSubnet = &Subnet{
		AvailabilityZone: *createSubnetResp.Subnet.AvailabilityZone,
		ID:               *createSubnetResp.Subnet.SubnetId,
	}
	return
}
