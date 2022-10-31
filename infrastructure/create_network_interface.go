package infrastructure

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type NetworkInterface struct {
	ID string `json:"id"`
}

func CreateNetworkInterface(
	ec2Client *ec2.Client,
	name string,
	description string,
	subnetID string,
	securityGroupIDs []string,
) (returnedNetworkInterface *NetworkInterface, returnedError error) {

	createNetworkInterfaceResp, err := ec2Client.CreateNetworkInterface(
		context.TODO(),
		&ec2.CreateNetworkInterfaceInput{
			SubnetId:    &subnetID,
			Groups:      securityGroupIDs,
			Description: &description,
			TagSpecifications: []types.TagSpecification{{
				ResourceType: types.ResourceTypeNetworkInterface,
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

		_ = RemoveNetworkInterface(
			ec2Client,
			*createNetworkInterfaceResp.NetworkInterface.NetworkInterfaceId,
		)
	}()

	availableWaiter := ec2.NewNetworkInterfaceAvailableWaiter(ec2Client)
	maxWaitTime := 5 * time.Minute

	err = availableWaiter.Wait(
		context.TODO(),
		&ec2.DescribeNetworkInterfacesInput{
			NetworkInterfaceIds: []string{
				*createNetworkInterfaceResp.NetworkInterface.NetworkInterfaceId,
			},
		},
		maxWaitTime,
	)

	if err != nil {
		returnedError = err
		return
	}

	returnedNetworkInterface = &NetworkInterface{
		ID: *createNetworkInterfaceResp.NetworkInterface.NetworkInterfaceId,
	}
	return
}
