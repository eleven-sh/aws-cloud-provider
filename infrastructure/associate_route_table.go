package infrastructure

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func AssociateRouteTable(
	ec2Client *ec2.Client,
	subnetID string,
	routeTableID string,
) error {

	_, err := ec2Client.AssociateRouteTable(
		context.TODO(),
		&ec2.AssociateRouteTableInput{
			RouteTableId: &routeTableID,
			SubnetId:     &subnetID,
		},
	)

	return err
}
