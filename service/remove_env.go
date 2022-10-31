package service

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/eleven-sh/aws-cloud-provider/infrastructure"
	"github.com/eleven-sh/eleven/entities"
	"github.com/eleven-sh/eleven/queues"
	"github.com/eleven-sh/eleven/stepper"
)

func (a *AWS) RemoveEnv(
	stepper stepper.Stepper,
	config *entities.Config,
	cluster *entities.Cluster,
	env *entities.Env,
) error {

	var envInfra *EnvInfrastructure
	err := json.Unmarshal([]byte(env.InfrastructureJSON), &envInfra)

	if err != nil {
		return err
	}

	ec2Client := ec2.NewFromConfig(a.sdkConfig)
	envInfraQueue := queues.InfrastructureQueue[*EnvInfrastructure]{}

	terminateInstance := func(infra *EnvInfrastructure) error {
		if infra.Instance == nil {
			return nil
		}

		err := infrastructure.TerminateInstance(
			ec2Client,
			infra.Instance.ID,
		)

		if err != nil {
			return err
		}

		infra.Instance = nil
		return nil
	}

	detachElasticIP := func(infra *EnvInfrastructure) error {
		if infra.ElasticIP == nil || !infra.ElasticIP.IsAttachedToInstance {
			return nil
		}

		err := infrastructure.DetachElasticIPFromInstance(
			ec2Client,
			infra.ElasticIP.AssociationID,
		)

		if err != nil {
			return err
		}

		infra.ElasticIP.AssociationID = ""
		infra.ElasticIP.IsAttachedToInstance = false
		return nil
	}

	envInfraQueue = append(
		envInfraQueue,
		queues.InfrastructureQueueSteps[*EnvInfrastructure]{
			func(*EnvInfrastructure) error {
				stepper.StartTemporaryStep("Waiting for the EC2 instance to terminate")
				return nil
			},
			terminateInstance,
			detachElasticIP,
		},
	)

	removeKeyPair := func(infra *EnvInfrastructure) error {
		if infra.KeyPair == nil {
			return nil
		}

		err := infrastructure.RemoveKeyPair(
			ec2Client,
			infra.KeyPair.ID,
		)

		if err != nil {
			return err
		}

		infra.KeyPair = nil
		return nil
	}

	removeElasticIP := func(infra *EnvInfrastructure) error {
		if infra.ElasticIP == nil {
			return nil
		}

		err := infrastructure.RemoveElasticIP(
			ec2Client,
			infra.ElasticIP.ID,
		)

		if err != nil {
			return err
		}

		infra.ElasticIP = nil
		return nil
	}

	envInfraQueue = append(
		envInfraQueue,
		queues.InfrastructureQueueSteps[*EnvInfrastructure]{
			func(*EnvInfrastructure) error {
				stepper.StartTemporaryStep("Removing the key pair and the elastic IP")
				return nil
			},
			removeKeyPair,
			removeElasticIP,
		},
	)

	removeNetworkInterface := func(infra *EnvInfrastructure) error {
		if infra.NetworkInterface == nil {
			return nil
		}

		err := infrastructure.RemoveNetworkInterface(
			ec2Client,
			infra.NetworkInterface.ID,
		)

		if err != nil {
			return err
		}

		infra.NetworkInterface = nil
		return nil
	}

	envInfraQueue = append(
		envInfraQueue,
		queues.InfrastructureQueueSteps[*EnvInfrastructure]{
			func(*EnvInfrastructure) error {
				stepper.StartTemporaryStep("Removing the network interface")
				return nil
			},
			removeNetworkInterface,
		},
	)

	removeSecurityGroup := func(infra *EnvInfrastructure) error {
		if infra.SecurityGroup == nil {
			return nil
		}

		err := infrastructure.RemoveSecurityGroup(
			ec2Client,
			infra.SecurityGroup.ID,
		)

		if err != nil {
			return err
		}

		infra.SecurityGroup = nil
		return nil
	}

	envInfraQueue = append(
		envInfraQueue,
		queues.InfrastructureQueueSteps[*EnvInfrastructure]{
			func(*EnvInfrastructure) error {
				stepper.StartTemporaryStep("Removing the security group")
				return nil
			},
			removeSecurityGroup,
		},
	)

	err = envInfraQueue.Run(
		envInfra,
	)

	// Env infra could be updated in the queue even
	// in case of error (partial infrastructure)
	env.SetInfrastructureJSON(envInfra)

	return err
}
