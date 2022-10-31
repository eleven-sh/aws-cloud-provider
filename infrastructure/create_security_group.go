package infrastructure

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type SecurityGroup struct {
	ID string `json:"id"`
}

func CreateSecurityGroup(
	ec2Client *ec2.Client,
	name string,
	description string,
	VPCID string,
	ingressPorts []types.IpPermission,
) (returnedSecurityGroup *SecurityGroup, returnedError error) {

	createSecurityGroupResp, err := ec2Client.CreateSecurityGroup(
		context.TODO(),
		&ec2.CreateSecurityGroupInput{
			GroupName:   &name,
			Description: &description,
			VpcId:       &VPCID,
			TagSpecifications: []types.TagSpecification{{
				ResourceType: types.ResourceTypeSecurityGroup,
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

		_ = RemoveSecurityGroup(ec2Client, *createSecurityGroupResp.GroupId)
	}()

	existsWaiter := ec2.NewSecurityGroupExistsWaiter(ec2Client)
	maxWaitTime := 5 * time.Minute

	err = existsWaiter.Wait(
		context.TODO(),
		&ec2.DescribeSecurityGroupsInput{
			GroupIds: []string{
				*createSecurityGroupResp.GroupId,
			},
		},
		maxWaitTime,
	)

	if err != nil {
		returnedError = err
		return
	}

	_, err = ec2Client.AuthorizeSecurityGroupIngress(
		context.TODO(),
		&ec2.AuthorizeSecurityGroupIngressInput{
			GroupId:       createSecurityGroupResp.GroupId,
			IpPermissions: ingressPorts,
		},
	)

	if err != nil {
		returnedError = err
		return
	}

	returnedSecurityGroup = &SecurityGroup{
		ID: *createSecurityGroupResp.GroupId,
	}
	return
}
