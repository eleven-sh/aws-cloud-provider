package infrastructure

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func getMostRecentAMI(AMIs []types.Image) (*types.Image, error) {
	var chosenAMI types.Image
	var chosenAMICreationDateTimestamp int64

	for _, AMI := range AMIs {
		AMICreationDate, err := time.Parse(time.RFC3339, *AMI.CreationDate)

		if err != nil {
			return nil, err
		}

		if chosenAMICreationDateTimestamp < AMICreationDate.Unix() {
			chosenAMI = AMI
			chosenAMICreationDateTimestamp = AMICreationDate.Unix()
		}
	}

	return &chosenAMI, nil
}
