package infrastructure

import (
	"context"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func CloseInstancePort(
	ec2Client *ec2.Client,
	securityGroupID string,
	portToClose string,
) error {

	portToCloseAsInt, _ := strconv.Atoi(portToClose)

	_, err := ec2Client.RevokeSecurityGroupIngress(
		context.TODO(),
		&ec2.RevokeSecurityGroupIngressInput{
			CidrIp:     aws.String("0.0.0.0/0"),
			FromPort:   aws.Int32(int32(portToCloseAsInt)),
			ToPort:     aws.Int32(int32(portToCloseAsInt)),
			GroupId:    &securityGroupID,
			IpProtocol: aws.String("tcp"),
		},
	)

	return err
}
