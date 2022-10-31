package infrastructure

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type Route struct{}

func CreateRoute(
	ec2Client *ec2.Client,
	internetGatewayID string,
	routeTableID string,
) (returnedRoute *Route, returnedError error) {

	_, err := ec2Client.CreateRoute(context.TODO(), &ec2.CreateRouteInput{
		RouteTableId:         &routeTableID,
		DestinationCidrBlock: aws.String("0.0.0.0/0"),
		GatewayId:            &internetGatewayID,
	})

	if err != nil {
		returnedError = err
		return
	}

	returnedRoute = &Route{}

	return
}
