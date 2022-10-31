package infrastructure

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func TestGetMostRecentAMI(t *testing.T) {
	AMIs := []types.Image{{
		ImageId:      aws.String("1"),
		CreationDate: aws.String("2022-01-19T02:44:42.000Z"),
	}, {
		ImageId:      aws.String("2"),
		CreationDate: aws.String("2021-10-14T18:00:13.000Z"),
	}, {
		ImageId:      aws.String("3"),
		CreationDate: aws.String("2022-01-11T12:58:48.000Z"),
	}, {
		ImageId:      aws.String("4"),
		CreationDate: aws.String("2021-12-11T11:25:52.000Z"),
	}, {
		ImageId:      aws.String("5"),
		CreationDate: aws.String("2022-02-01T12:05:27.000Z"),
	}, {
		ImageId:      aws.String("6"),
		CreationDate: aws.String("2022-03-09T18:33:14.000Z"),
	}}

	expectedAMIID := "6"

	AMI, err := getMostRecentAMI(AMIs)

	if err != nil {
		t.Fatalf("expected no error, got %s", err)
	}

	AMIID := *AMI.ImageId

	if AMIID != expectedAMIID {
		t.Fatalf(
			"expected AMI ID to equal '%s', got '%s'",
			expectedAMIID,
			AMIID,
		)
	}
}
