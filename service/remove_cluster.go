package service

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/eleven-sh/aws-cloud-provider/infrastructure"
	"github.com/eleven-sh/eleven/entities"
	"github.com/eleven-sh/eleven/queues"
	"github.com/eleven-sh/eleven/stepper"
)

func (a *AWS) RemoveCluster(
	stepper stepper.Stepper,
	config *entities.Config,
	cluster *entities.Cluster,
) error {

	var clusterInfra *ClusterInfrastructure
	err := json.Unmarshal([]byte(cluster.InfrastructureJSON), &clusterInfra)

	if err != nil {
		return err
	}

	ec2Client := ec2.NewFromConfig(a.sdkConfig)
	clusterInfraQueue := queues.InfrastructureQueue[*ClusterInfrastructure]{}

	removeSubnet := func(infra *ClusterInfrastructure) error {
		if infra.Subnet == nil {
			return nil
		}

		err := infrastructure.RemoveSubnet(
			ec2Client,
			infra.Subnet.ID,
		)

		if err != nil {
			return err
		}

		infra.Subnet = nil
		return nil
	}

	clusterInfraQueue = append(
		clusterInfraQueue,
		queues.InfrastructureQueueSteps[*ClusterInfrastructure]{
			func(*ClusterInfrastructure) error {
				stepper.StartTemporaryStep("Removing the subnet")
				return nil
			},
			removeSubnet,
		},
	)

	removeRouteTable := func(infra *ClusterInfrastructure) error {
		if infra.RouteTable == nil {
			return nil
		}

		err := infrastructure.RemoveRouteTable(
			ec2Client,
			infra.RouteTable.ID,
		)

		if err != nil {
			return err
		}

		infra.RouteTable = nil
		return nil
	}

	clusterInfraQueue = append(
		clusterInfraQueue,
		queues.InfrastructureQueueSteps[*ClusterInfrastructure]{
			func(*ClusterInfrastructure) error {
				stepper.StartTemporaryStep("Removing the route table")
				return nil
			},
			removeRouteTable,
		},
	)

	detachIGFromVPC := func(infra *ClusterInfrastructure) error {
		if infra.InternetGateway == nil || !infra.InternetGateway.IsAttachedToVPC {
			return nil
		}

		err := infrastructure.DetachInternetGatewayFromVPC(
			ec2Client,
			infra.InternetGateway.ID,
			infra.VPC.ID,
		)

		if err != nil {
			return err
		}

		infra.InternetGateway.IsAttachedToVPC = false
		return nil
	}

	clusterInfraQueue = append(
		clusterInfraQueue,
		queues.InfrastructureQueueSteps[*ClusterInfrastructure]{
			func(*ClusterInfrastructure) error {
				stepper.StartTemporaryStep("Detaching the internet gateway from the VPC")
				return nil
			},
			detachIGFromVPC,
		},
	)

	removeInternetGateway := func(infra *ClusterInfrastructure) error {
		if infra.InternetGateway == nil {
			return nil
		}

		err := infrastructure.RemoveInternetGateway(
			ec2Client,
			infra.InternetGateway.ID,
		)

		if err != nil {
			return err
		}

		infra.InternetGateway = nil
		return nil
	}

	clusterInfraQueue = append(
		clusterInfraQueue,
		queues.InfrastructureQueueSteps[*ClusterInfrastructure]{
			func(*ClusterInfrastructure) error {
				stepper.StartTemporaryStep("Removing the internet gateway")
				return nil
			},
			removeInternetGateway,
		},
	)

	removeVPC := func(infra *ClusterInfrastructure) error {
		if infra.VPC == nil {
			return nil
		}

		err := infrastructure.RemoveVPC(
			ec2Client,
			infra.VPC.ID,
		)

		if err != nil {
			return err
		}

		infra.VPC = nil
		return nil
	}

	clusterInfraQueue = append(
		clusterInfraQueue,
		queues.InfrastructureQueueSteps[*ClusterInfrastructure]{
			func(*ClusterInfrastructure) error {
				stepper.StartTemporaryStep("Removing the VPC")
				return nil
			},
			removeVPC,
		},
	)

	err = clusterInfraQueue.Run(
		clusterInfra,
	)

	// Cluster infra could be updated in the queue even
	// in case of error (partial infrastructure)
	cluster.SetInfrastructureJSON(clusterInfra)

	return err
}
