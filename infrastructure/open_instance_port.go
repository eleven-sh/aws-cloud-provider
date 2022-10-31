package infrastructure

import (
	"context"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func OpenInstancePort(
	ec2Client *ec2.Client,
	securityGroupID string,
	portToOpen string,
) error {

	portToOpenAsInt, _ := strconv.Atoi(portToOpen)

	_, err := ec2Client.AuthorizeSecurityGroupIngress(
		context.TODO(),
		&ec2.AuthorizeSecurityGroupIngressInput{
			CidrIp:     aws.String("0.0.0.0/0"),
			FromPort:   aws.Int32(int32(portToOpenAsInt)),
			ToPort:     aws.Int32(int32(portToOpenAsInt)),
			GroupId:    &securityGroupID,
			IpProtocol: aws.String("tcp"),
		},
	)

	return err
}
