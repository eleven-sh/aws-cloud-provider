package infrastructure

import (
	"context"
	_ "embed"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/eleven-sh/agent/config"
)

const (
	InstanceRootDeviceSizeGb = 16
)

var (
	//go:embed init_instance.sh
	instanceInitScript string
)

type InstanceVolume struct {
	ID           string `json:"id"`
	DeviceName   string `json:"device_name"`
	SnapshotID   string `json:"snapshot_id"`
	IsRootVolume bool   `json:"is_root_volume"`
}

type Instance struct {
	ID                 string                     `json:"id"`
	Type               string                     `json:"type"`
	TmpPublicIPAddress string                     `json:"tmp_public_ip_address"`
	Volumes            []InstanceVolume           `json:"volumes"`
	InitScriptResults  *InitInstanceScriptResults `json:"init_script_results"`
}

func CreateInstance(
	ec2Client *ec2.Client,
	name string,
	AMIID string,
	rootDeviceName string,
	instanceType string,
	networkInterfaceID string,
	keyName string,
) (returnedInstance *Instance, returnedError error) {

	updatedInstanceInitScript := strings.ReplaceAll(
		instanceInitScript,
		"${ELEVEN_CONFIG_DIR}",
		config.ElevenConfigDirPath,
	)

	updatedInstanceInitScript = strings.ReplaceAll(
		updatedInstanceInitScript,
		"${ELEVEN_AGENT_CONFIG_DIR}",
		config.ElevenAgentConfigDirPath,
	)

	instanceInitScriptAsB64 := base64.StdEncoding.EncodeToString(
		[]byte(updatedInstanceInitScript),
	)

	runInstancesResp, err := ec2Client.RunInstances(context.TODO(), &ec2.RunInstancesInput{
		ImageId:      &AMIID,
		InstanceType: types.InstanceType(instanceType),
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
		NetworkInterfaces: []types.InstanceNetworkInterfaceSpecification{
			{
				DeviceIndex:        aws.Int32(0),
				NetworkInterfaceId: aws.String(networkInterfaceID),
			},
		},
		KeyName:  &keyName,
		UserData: &instanceInitScriptAsB64,
		BlockDeviceMappings: []types.BlockDeviceMapping{
			{
				DeviceName: &rootDeviceName,
				Ebs: &types.EbsBlockDevice{
					VolumeSize: aws.Int32(InstanceRootDeviceSizeGb),
				},
			},
		},
		TagSpecifications: []types.TagSpecification{{
			ResourceType: types.ResourceTypeInstance,
			Tags: []types.Tag{{
				Key:   aws.String("Name"),
				Value: &name,
			}},
		}},
	})

	if err != nil {
		returnedError = err
		return
	}

	instanceID := *runInstancesResp.Instances[0].InstanceId

	defer func() {
		if returnedError == nil {
			return
		}

		_ = TerminateInstance(ec2Client, instanceID)
	}()

	runningWaiter := ec2.NewInstanceRunningWaiter(ec2Client)
	maxWaitTime := 5 * time.Minute

	err = runningWaiter.Wait(context.TODO(), &ec2.DescribeInstancesInput{
		InstanceIds: []string{
			instanceID,
		},
	}, maxWaitTime)

	if err != nil {
		returnedError = err
		return
	}

	/* Public IP / DNS are only available
	when instance is running */

	createdInstance, err := lookupInstance(ec2Client, instanceID)

	if err != nil {
		returnedError = err
		return
	}

	returnedInstance = &Instance{
		ID:                 *createdInstance.InstanceId,
		TmpPublicIPAddress: *createdInstance.PublicIpAddress,
		Type:               string(createdInstance.InstanceType),
	}

	var volumes []InstanceVolume
	for _, blockDevice := range createdInstance.BlockDeviceMappings {
		volumes = append(volumes, InstanceVolume{
			ID:           *blockDevice.Ebs.VolumeId,
			DeviceName:   *blockDevice.DeviceName,
			IsRootVolume: true,
		})
	}

	// Sanity check
	if len(volumes) != 1 {
		returnedError = fmt.Errorf(
			"expected only one root volume, got %d",
			len(volumes),
		)
		return
	}

	returnedInstance.Volumes = volumes
	return
}
