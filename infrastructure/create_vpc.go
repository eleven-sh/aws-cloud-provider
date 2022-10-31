package infrastructure

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type VPC struct {
	ID string `json:"id"`
}

func CreateVPC(
	ec2Client *ec2.Client,
	VPCName string,
	CIDRBlock string,
) (returnedVPC *VPC, returnedError error) {

	createVPCResp, err := ec2Client.CreateVpc(
		context.TODO(),
		&ec2.CreateVpcInput{
			CidrBlock: &CIDRBlock,
			TagSpecifications: []types.TagSpecification{{
				ResourceType: types.ResourceTypeVpc,
				Tags: []types.Tag{{
					Key:   aws.String("Name"),
					Value: &VPCName,
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

		_ = RemoveVPC(ec2Client, *createVPCResp.Vpc.VpcId)
	}()

	availableWaiter := ec2.NewVpcAvailableWaiter(ec2Client)
	maxWaitTime := 5 * time.Minute

	err = availableWaiter.Wait(context.TODO(), &ec2.DescribeVpcsInput{
		VpcIds: []string{
			*createVPCResp.Vpc.VpcId,
		},
	}, maxWaitTime)

	if err != nil {
		returnedError = err
		return
	}

	/* From AWS docs:
	   You cannot modify the DNS support
	   and DNS hostnames attributes in the same request.
	   Use separate requests for each attribute. */

	enableDNSSupportChan := make(chan error)
	enableDNSHostnamesChan := make(chan error)

	go func() {
		_, err := ec2Client.ModifyVpcAttribute(
			context.TODO(),
			&ec2.ModifyVpcAttributeInput{
				EnableDnsSupport: &types.AttributeBooleanValue{
					Value: aws.Bool(true),
				},
				VpcId: createVPCResp.Vpc.VpcId,
			},
		)

		enableDNSSupportChan <- err
	}()

	go func() {
		_, err := ec2Client.ModifyVpcAttribute(
			context.TODO(),
			&ec2.ModifyVpcAttributeInput{
				EnableDnsHostnames: &types.AttributeBooleanValue{
					Value: aws.Bool(true),
				},
				VpcId: createVPCResp.Vpc.VpcId,
			},
		)

		enableDNSHostnamesChan <- err
	}()

	enableDNSSupportErr := <-enableDNSSupportChan
	enableDNSHostnamesErr := <-enableDNSHostnamesChan

	if enableDNSSupportErr != nil {
		returnedError = enableDNSSupportErr
		return
	}

	if enableDNSHostnamesErr != nil {
		returnedError = enableDNSHostnamesErr
		return
	}

	returnedVPC = &VPC{
		ID: *createVPCResp.Vpc.VpcId,
	}
	return
}
