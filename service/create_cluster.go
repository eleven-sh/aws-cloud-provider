package service

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/eleven-sh/aws-cloud-provider/infrastructure"
	"github.com/eleven-sh/eleven/entities"
	"github.com/eleven-sh/eleven/queues"
	"github.com/eleven-sh/eleven/stepper"
)

type ClusterInfrastructure struct {
	VPC             *infrastructure.VPC             `json:"vpc"`
	InternetGateway *infrastructure.InternetGateway `json:"internet_gateway"`
	Subnet          *infrastructure.Subnet          `json:"subnet"`
	RouteTable      *infrastructure.RouteTable      `json:"route_table"`
	Route           *infrastructure.Route           `json:"route"`
}

func (a *AWS) CreateCluster(
	stepper stepper.Stepper,
	config *entities.Config,
	cluster *entities.Cluster,
) error {

	clusterInfra := &ClusterInfrastructure{}
	if len(cluster.InfrastructureJSON) > 0 {
		err := json.Unmarshal([]byte(cluster.InfrastructureJSON), clusterInfra)

		if err != nil {
			return err
		}
	}

	prefixResource := prefixClusterResource(cluster.GetNameSlug())
	ec2Client := ec2.NewFromConfig(a.sdkConfig)

	clusterInfraQueue := queues.InfrastructureQueue[*ClusterInfrastructure]{}

	createVPC := func(infra *ClusterInfrastructure) error {
		if infra.VPC != nil {
			return nil
		}

		vpc, err := infrastructure.CreateVPC(
			ec2Client,
			prefixResource("vpc"),
			"10.0.0.0/16",
		)

		if err != nil {
			return err
		}

		infra.VPC = vpc
		return nil
	}

	createInternetGateway := func(infra *ClusterInfrastructure) error {
		if infra.InternetGateway != nil {
			return nil
		}

		internetGateway, err := infrastructure.CreateInternetGateway(
			ec2Client,
			prefixResource("internet-gateway"),
		)

		if err != nil {
			return err
		}

		infra.InternetGateway = internetGateway
		return nil
	}

	clusterInfraQueue = append(
		clusterInfraQueue,
		queues.InfrastructureQueueSteps[*ClusterInfrastructure]{
			func(*ClusterInfrastructure) error {
				stepper.StartTemporaryStep("Creating a VPC and an internet gateway")
				return nil
			},
			createVPC,
			createInternetGateway,
		},
	)

	attachInternetGatewayToVPC := func(infra *ClusterInfrastructure) error {
		if infra.InternetGateway.IsAttachedToVPC {
			return nil
		}

		err := infrastructure.AttachInternetGatewayToVPC(
			ec2Client,
			infra.InternetGateway.ID,
			infra.VPC.ID,
		)

		if err != nil {
			return err
		}

		infra.InternetGateway.IsAttachedToVPC = true
		return nil
	}

	createSubnet := func(infra *ClusterInfrastructure) error {
		if infra.Subnet != nil {
			return nil
		}

		subnet, err := infrastructure.CreateSubnet(
			ec2Client,
			prefixResource("public-subnet"),
			"10.0.0.0/24",
			infra.VPC.ID,
		)

		if err != nil {
			return err
		}

		infra.Subnet = subnet
		return nil
	}

	createRouteTable := func(infra *ClusterInfrastructure) error {
		if infra.RouteTable != nil {
			return nil
		}

		routeTable, err := infrastructure.CreateRouteTable(
			ec2Client,
			prefixResource("route-table"),
			infra.VPC.ID,
		)

		if err != nil {
			return err
		}

		infra.RouteTable = routeTable
		return nil
	}

	clusterInfraQueue = append(
		clusterInfraQueue,
		queues.InfrastructureQueueSteps[*ClusterInfrastructure]{
			func(*ClusterInfrastructure) error {
				stepper.StartTemporaryStep("Creating a subnet and a route table")
				return nil
			},
			attachInternetGatewayToVPC,
			createSubnet,
			createRouteTable,
		},
	)

	createRoute := func(infra *ClusterInfrastructure) error {
		if infra.Route != nil {
			return nil
		}

		route, err := infrastructure.CreateRoute(
			ec2Client,
			infra.InternetGateway.ID,
			infra.RouteTable.ID,
		)

		if err != nil {
			return err
		}

		infra.Route = route
		return nil
	}

	associateRouteTable := func(infra *ClusterInfrastructure) error {
		if infra.RouteTable.IsAssociatedToSubnet {
			return nil
		}

		err := infrastructure.AssociateRouteTable(
			ec2Client,
			infra.Subnet.ID,
			infra.RouteTable.ID,
		)

		if err != nil {
			return err
		}

		infra.RouteTable.IsAssociatedToSubnet = true
		return nil
	}

	clusterInfraQueue = append(
		clusterInfraQueue,
		queues.InfrastructureQueueSteps[*ClusterInfrastructure]{
			func(*ClusterInfrastructure) error {
				stepper.StartTemporaryStep("Adding a route to the route table")
				return nil
			},
			createRoute,
			associateRouteTable,
		},
	)

	err := clusterInfraQueue.Run(
		clusterInfra,
	)

	// Cluster infra could be updated in the queue even
	// in case of error (partial infrastructure)
	cluster.SetInfrastructureJSON(clusterInfra)

	return err
}
