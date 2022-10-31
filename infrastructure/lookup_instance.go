package infrastructure

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

var (
	ErrInstanceNotFound = errors.New("ErrInstanceNotFound")
)

func lookupInstance(
	ec2Client *ec2.Client,
	instanceID string,
) (*types.Instance, error) {

	describeInstancesResp, err := ec2Client.DescribeInstances(
		context.TODO(),
		&ec2.DescribeInstancesInput{
			InstanceIds: []string{instanceID},
		},
	)

	if err != nil {
		return nil, err
	}

	if len(describeInstancesResp.Reservations) == 0 ||
		len(describeInstancesResp.Reservations[0].Instances) == 0 {

		return nil, ErrInstanceNotFound
	}

	instance := describeInstancesResp.Reservations[0].Instances[0]
	return &instance, nil
}
