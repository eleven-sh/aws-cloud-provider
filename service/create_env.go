package service

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	agentConfig "github.com/eleven-sh/agent/config"
	"github.com/eleven-sh/aws-cloud-provider/infrastructure"
	"github.com/eleven-sh/eleven/entities"
	"github.com/eleven-sh/eleven/queues"
	"github.com/eleven-sh/eleven/stepper"
)

type EnvInfrastructure struct {
	SecurityGroup     *infrastructure.SecurityGroup     `json:"security_group"`
	KeyPair           *infrastructure.KeyPair           `json:"key_pair"`
	NetworkInterface  *infrastructure.NetworkInterface  `json:"network_interface"`
	InstanceTypeInfos *infrastructure.InstanceTypeInfos `json:"instance_type_infos"`
	InstanceAMI       *infrastructure.AMI               `json:"instance_ami"`
	Instance          *infrastructure.Instance          `json:"instance"`
	ElasticIP         *infrastructure.ElasticIP         `json:"elastic_ip"`
}

func (a *AWS) CreateEnv(
	stepper stepper.Stepper,
	config *entities.Config,
	cluster *entities.Cluster,
	env *entities.Env,
) error {

	var clusterInfra *ClusterInfrastructure
	err := json.Unmarshal([]byte(cluster.InfrastructureJSON), &clusterInfra)

	if err != nil {
		return err
	}

	envInfra := &EnvInfrastructure{}
	if len(env.InfrastructureJSON) > 0 {
		err := json.Unmarshal([]byte(env.InfrastructureJSON), envInfra)

		if err != nil {
			return err
		}
	}

	prefixResource := prefixEnvResource(cluster.GetNameSlug(), env.GetNameSlug())
	ec2Client := ec2.NewFromConfig(a.sdkConfig)

	envInfraQueue := queues.InfrastructureQueue[*EnvInfrastructure]{}

	createSecurityGroup := func(infra *EnvInfrastructure) error {
		if infra.SecurityGroup != nil {
			return nil
		}

		elevenSSHServerListenPort, _ := strconv.ParseInt(
			agentConfig.SSHServerListenPort,
			10,
			64,
		)

		elevenHTTPServerListenPort, _ := strconv.ParseInt(
			agentConfig.HTTPServerListenPort,
			10,
			64,
		)

		elevenHTTPSServerListenPort, _ := strconv.ParseInt(
			agentConfig.HTTPSServerListenPort,
			10,
			64,
		)

		securityGroup, err := infrastructure.CreateSecurityGroup(
			ec2Client,
			prefixResource("security-group"),
			"The security group attached to your sandbox",
			clusterInfra.VPC.ID,
			[]types.IpPermission{
				{
					IpProtocol: aws.String("tcp"),
					FromPort:   aws.Int32(infrastructure.InstanceSSHPort),
					ToPort:     aws.Int32(infrastructure.InstanceSSHPort),
					IpRanges: []types.IpRange{
						{
							CidrIp: aws.String("0.0.0.0/0"),
						},
					},
				},

				{
					IpProtocol: aws.String("tcp"),
					FromPort:   aws.Int32(int32(elevenSSHServerListenPort)),
					ToPort:     aws.Int32(int32(elevenSSHServerListenPort)),
					IpRanges: []types.IpRange{
						{
							CidrIp: aws.String("0.0.0.0/0"),
						},
					},
				},

				{
					IpProtocol: aws.String("tcp"),
					FromPort:   aws.Int32(int32(elevenHTTPServerListenPort)),
					ToPort:     aws.Int32(int32(elevenHTTPServerListenPort)),
					IpRanges: []types.IpRange{
						{
							CidrIp: aws.String("0.0.0.0/0"),
						},
					},
				},

				{
					IpProtocol: aws.String("tcp"),
					FromPort:   aws.Int32(int32(elevenHTTPSServerListenPort)),
					ToPort:     aws.Int32(int32(elevenHTTPSServerListenPort)),
					IpRanges: []types.IpRange{
						{
							CidrIp: aws.String("0.0.0.0/0"),
						},
					},
				},
			},
		)

		if err != nil {
			return err
		}

		infra.SecurityGroup = securityGroup
		return nil
	}

	createKeyPair := func(infra *EnvInfrastructure) error {
		if infra.KeyPair != nil {
			return nil
		}

		keyPair, err := infrastructure.CreateKeyPair(
			ec2Client,
			prefixResource("key-pair"),
		)

		if err != nil {
			return err
		}

		infra.KeyPair = keyPair
		return nil
	}

	createElasticIP := func(infra *EnvInfrastructure) error {
		if infra.ElasticIP != nil {
			return nil
		}

		elasticIP, err := infrastructure.CreateElasticIP(
			ec2Client,
			prefixResource("elastic-ip"),
		)

		if err != nil {
			return err
		}

		infra.ElasticIP = elasticIP
		return nil
	}

	envInfraQueue = append(
		envInfraQueue,
		queues.InfrastructureQueueSteps[*EnvInfrastructure]{
			func(*EnvInfrastructure) error {
				stepper.StartTemporaryStep("Creating a security group, a key pair and an elastic IP")
				return nil
			},
			createSecurityGroup,
			createKeyPair,
			createElasticIP,
		},
	)

	createNetworkInterface := func(infra *EnvInfrastructure) error {
		if infra.NetworkInterface != nil {
			return nil
		}

		networkInterface, err := infrastructure.CreateNetworkInterface(
			ec2Client,
			prefixResource("network-interface"),
			"The network interface attached to your sandbox",
			clusterInfra.Subnet.ID,
			[]string{infra.SecurityGroup.ID},
		)

		if err != nil {
			return err
		}

		infra.NetworkInterface = networkInterface
		return nil
	}

	envInfraQueue = append(
		envInfraQueue,
		queues.InfrastructureQueueSteps[*EnvInfrastructure]{
			func(*EnvInfrastructure) error {
				stepper.StartTemporaryStep("Creating a network interface")
				return nil
			},
			createNetworkInterface,
		},
	)

	lookupInstanceTypeInfos := func(infra *EnvInfrastructure) error {
		if infra.InstanceTypeInfos != nil {
			return nil
		}

		instanceTypeInfos, err := infrastructure.LookupInstanceTypeInfos(
			ec2Client,
			env.InstanceType,
		)

		if err != nil {
			return err
		}

		infra.InstanceTypeInfos = instanceTypeInfos
		return nil
	}

	envInfraQueue = append(
		envInfraQueue,
		queues.InfrastructureQueueSteps[*EnvInfrastructure]{
			func(*EnvInfrastructure) error {
				stepper.StartTemporaryStep("Looking up instance type infos")
				return nil
			},
			lookupInstanceTypeInfos,
		},
	)

	lookupUbuntuAMIForArchAndRegion := func(infra *EnvInfrastructure) error {
		if infra.InstanceAMI != nil {
			return nil
		}

		instanceAMI, err := infrastructure.LookupUbuntuAMIForArch(
			ec2Client,
			infra.InstanceTypeInfos.Arch,
		)

		if err != nil {
			return err
		}

		infra.InstanceAMI = instanceAMI
		return nil
	}

	envInfraQueue = append(
		envInfraQueue,
		queues.InfrastructureQueueSteps[*EnvInfrastructure]{
			func(*EnvInfrastructure) error {
				stepper.StartTemporaryStep("Looking up the AMI details")
				return nil
			},
			lookupUbuntuAMIForArchAndRegion,
		},
	)

	createInstance := func(infra *EnvInfrastructure) error {
		if infra.Instance != nil {
			return nil
		}

		instance, err := infrastructure.CreateInstance(
			ec2Client,
			prefixResource("instance"),
			infra.InstanceAMI.ID,
			infra.InstanceAMI.RootDeviceName,
			infra.InstanceTypeInfos.Type,
			infra.NetworkInterface.ID,
			infra.KeyPair.Name,
		)

		if err != nil {
			return err
		}

		infra.Instance = instance
		return nil
	}

	envInfraQueue = append(
		envInfraQueue,
		queues.InfrastructureQueueSteps[*EnvInfrastructure]{
			func(*EnvInfrastructure) error {
				stepper.StartTemporaryStep("Creating an EC2 instance")
				return nil
			},
			createInstance,
		},
	)

	lookupInstanceInitScriptResults := func(infra *EnvInfrastructure) error {
		if infra.Instance.InitScriptResults != nil {
			return nil
		}

		initScriptResults, err := infrastructure.LookupInitInstanceScriptResults(
			ec2Client,
			infra.Instance.TmpPublicIPAddress,
			fmt.Sprintf("%d", infrastructure.InstanceSSHPort),
			infrastructure.InstanceRootUser,
			infra.KeyPair.PEMContent,
		)

		if err != nil {
			return err
		}

		infra.Instance.InitScriptResults = initScriptResults
		return nil
	}

	envInfraQueue = append(
		envInfraQueue,
		queues.InfrastructureQueueSteps[*EnvInfrastructure]{
			func(*EnvInfrastructure) error {
				stepper.StartTemporaryStep("Waiting for the EC2 instance to be ready")
				return nil
			},
			lookupInstanceInitScriptResults,
		},
	)

	// The Elastic IP is attached now
	// to avoid network errors during
	// instance initialization (due to IP
	// switching).
	attachElasticIP := func(infra *EnvInfrastructure) error {
		if infra.ElasticIP.IsAttachedToInstance {
			return nil
		}

		associationID, err := infrastructure.AttachElasticIPToInstance(
			ec2Client,
			infra.ElasticIP.ID,
			infra.Instance.ID,
		)

		if err != nil {
			return err
		}

		infra.Instance.TmpPublicIPAddress = ""
		infra.ElasticIP.AssociationID = associationID
		infra.ElasticIP.IsAttachedToInstance = true
		return nil
	}

	envInfraQueue = append(
		envInfraQueue,
		queues.InfrastructureQueueSteps[*EnvInfrastructure]{
			func(*EnvInfrastructure) error {
				stepper.StartTemporaryStep("Attaching a public IP to the instance")
				return nil
			},
			attachElasticIP,
		},
	)

	waitForEIPToBeReachable := func(infra *EnvInfrastructure) error {
		return infrastructure.WaitForSSHAvailableInInstance(
			ec2Client,
			infra.ElasticIP.Address,
			agentConfig.SSHServerListenPort,
		)
	}

	envInfraQueue = append(
		envInfraQueue,
		queues.InfrastructureQueueSteps[*EnvInfrastructure]{
			func(*EnvInfrastructure) error {
				stepper.StartTemporaryStep("Waiting for the public IP to be reachable")
				return nil
			},
			waitForEIPToBeReachable,
		},
	)

	err = envInfraQueue.Run(envInfra)

	// Env infra could be updated in the queue even
	// in case of error (partial infrastructure)
	env.SetInfrastructureJSON(envInfra)

	if err != nil {
		return err
	}

	env.InstancePublicIPAddress = envInfra.ElasticIP.Address

	env.SSHHostKeys = envInfra.Instance.InitScriptResults.SSHHostKeys
	env.SSHKeyPairPEMContent = envInfra.KeyPair.PEMContent

	return nil
}
