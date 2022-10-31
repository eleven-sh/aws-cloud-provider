package infrastructure

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func RemoveRouteTable(
	ec2Client *ec2.Client,
	routeTableID string,
) error {

	_, err := ec2Client.DeleteRouteTable(
		context.TODO(),
		&ec2.DeleteRouteTableInput{
			RouteTableId: &routeTableID,
		},
	)

	return err
}
