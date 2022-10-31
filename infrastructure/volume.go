package infrastructure

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type CreateVolumeFromSnapshotResp struct {
	Err      error
	VolumeID string
}

func CreateVolumeFromSnapshot(
	ec2Client *ec2.Client,
	name string,
	availabilityZone string,
	snapshotID string,
) (resp CreateVolumeFromSnapshotResp) {

	createVolumeResp, err := ec2Client.CreateVolume(
		context.TODO(),
		&ec2.CreateVolumeInput{
			AvailabilityZone: &availabilityZone,
			SnapshotId:       &snapshotID,
			VolumeType:       types.VolumeTypeGp2,
			TagSpecifications: []types.TagSpecification{{
				ResourceType: types.ResourceTypeVolume,
				Tags: []types.Tag{{
					Key:   aws.String("Name"),
					Value: &name,
				}},
			}},
		},
	)

	if err != nil {
		resp.Err = err
		return
	}

	availableWaiter := ec2.NewVolumeAvailableWaiter(ec2Client)
	maxWaitTime := 5 * time.Minute

	err = availableWaiter.Wait(context.TODO(), &ec2.DescribeVolumesInput{
		VolumeIds: []string{
			*createVolumeResp.VolumeId,
		},
	}, maxWaitTime)

	if err != nil {
		resp.Err = err
		return
	}

	resp.VolumeID = *createVolumeResp.VolumeId
	return
}

type RemoveVolumeResp struct {
	Err error
}

func RemoveVolume(
	ec2Client *ec2.Client,
	volumeID string,
) (resp RemoveVolumeResp) {

	_, err := ec2Client.DeleteVolume(
		context.TODO(),
		&ec2.DeleteVolumeInput{
			VolumeId: &volumeID,
		},
	)

	if err != nil {
		resp.Err = err
		return
	}

	deletedWaiter := ec2.NewVolumeDeletedWaiter(ec2Client)
	maxWaitTime := 5 * time.Minute

	resp.Err = deletedWaiter.Wait(
		context.TODO(),
		&ec2.DescribeVolumesInput{
			VolumeIds: []string{
				volumeID,
			},
		},
		maxWaitTime,
	)

	return
}

type DetachVolumeResp struct {
	Err error
}

func DetachVolume(
	ec2Client *ec2.Client,
	instanceID string,
	volumeID string,
	deviceName string,
) (resp DetachVolumeResp) {

	_, err := ec2Client.DetachVolume(
		context.TODO(),
		&ec2.DetachVolumeInput{
			InstanceId: &instanceID,
			VolumeId:   &volumeID,
			Device:     &deviceName,
		},
	)

	if err != nil {
		resp.Err = err
		return
	}

	availableWaiter := ec2.NewVolumeAvailableWaiter(ec2Client)
	maxWaitTime := 5 * time.Minute

	resp.Err = availableWaiter.Wait(
		context.TODO(),
		&ec2.DescribeVolumesInput{
			VolumeIds: []string{
				volumeID,
			},
		},
		maxWaitTime,
	)

	return
}

type AttachVolumeResp struct {
	Err error
}

func AttachVolume(
	ec2Client *ec2.Client,
	instanceID string,
	volumeID string,
	deviceName string,
) (resp AttachVolumeResp) {

	_, err := ec2Client.AttachVolume(
		context.TODO(),
		&ec2.AttachVolumeInput{
			InstanceId: &instanceID,
			VolumeId:   &volumeID,
			Device:     &deviceName,
		},
	)

	if err != nil {
		resp.Err = err
		return
	}

	inUseWaiter := ec2.NewVolumeInUseWaiter(ec2Client)
	maxWaitTime := 5 * time.Minute

	resp.Err = inUseWaiter.Wait(
		context.TODO(),
		&ec2.DescribeVolumesInput{
			VolumeIds: []string{
				volumeID,
			},
		},
		maxWaitTime,
	)

	return
}

type CreateSnapshotForVolumeResp struct {
	Err        error
	SnapshotID string
}

func CreateSnapshotForVolume(
	ec2Client *ec2.Client,
	name string,
	volumeID string,
) (resp CreateSnapshotForVolumeResp) {

	createSnapshotResp, err := ec2Client.CreateSnapshot(
		context.TODO(),
		&ec2.CreateSnapshotInput{
			VolumeId: &volumeID,
			TagSpecifications: []types.TagSpecification{{
				ResourceType: types.ResourceTypeSnapshot,
				Tags: []types.Tag{{
					Key:   aws.String("Name"),
					Value: &name,
				}},
			}},
		},
	)

	if err != nil {
		resp.Err = err
		return
	}

	snapshotID := *createSnapshotResp.SnapshotId

	completedWaiter := ec2.NewSnapshotCompletedWaiter(ec2Client)
	maxWaitTime := 24 * time.Hour

	err = completedWaiter.Wait(
		context.TODO(),
		&ec2.DescribeSnapshotsInput{
			SnapshotIds: []string{
				snapshotID,
			},
		},
		maxWaitTime,
	)

	if err != nil {
		resp.Err = err
		return
	}

	resp.SnapshotID = snapshotID
	return
}

type RemoveVolumeSnapshotResp struct {
	Err error
}

func RemoveVolumeSnapshot(
	ec2Client *ec2.Client,
	snapshotID string,
) (resp RemoveVolumeSnapshotResp) {

	_, err := ec2Client.DeleteSnapshot(
		context.TODO(),
		&ec2.DeleteSnapshotInput{
			SnapshotId: &snapshotID,
		},
	)

	resp.Err = err
	return
}
