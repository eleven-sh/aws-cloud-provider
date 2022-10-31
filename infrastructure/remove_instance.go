package infrastructure

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func TerminateInstance(
	ec2Client *ec2.Client,
	instanceID string,
) error {

	_, err := ec2Client.TerminateInstances(context.TODO(), &ec2.TerminateInstancesInput{
		InstanceIds: []string{instanceID},
	})

	if err != nil {
		return err
	}

	terminatedWaiter := ec2.NewInstanceTerminatedWaiter(ec2Client)
	maxWaitTime := 5 * time.Minute

	return terminatedWaiter.Wait(
		context.TODO(),
		&ec2.DescribeInstancesInput{
			InstanceIds: []string{
				instanceID,
			},
		},
		maxWaitTime,
	)
}
