package infrastructure

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func RemoveSecurityGroup(
	ec2Client *ec2.Client,
	securityGroupID string,
) error {

	_, err := ec2Client.DeleteSecurityGroup(
		context.TODO(),
		&ec2.DeleteSecurityGroupInput{
			GroupId: &securityGroupID,
		},
	)

	return err
}
