package infrastructure

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func RemoveVPC(
	ec2Client *ec2.Client,
	VPCID string,
) error {

	_, err := ec2Client.DeleteVpc(
		context.TODO(),
		&ec2.DeleteVpcInput{
			VpcId: &VPCID,
		},
	)

	return err
}
