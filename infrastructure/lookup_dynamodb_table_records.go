package infrastructure

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var (
	ErrElevenConfigNotFound   = errors.New("ErrElevenConfigNotFound")
	ErrElevenConfigDuplicated = errors.New("ErrElevenConfigDuplicated")
)

type DynamoDBElevenConfigTableRecord struct {
	ID         string
	ConfigJSON string
}

func LookupElevenConfigInDynamoDBTable(
	dynamoDBClient *dynamodb.Client,
) (returnedConfigJSON string, returnedError error) {

	scanResp, err := dynamoDBClient.Scan(context.TODO(), &dynamodb.ScanInput{
		TableName: aws.String(DynamoDBElevenConfigTableName),
	})

	if err != nil {
		var resourceNotFoundErr *types.ResourceNotFoundException

		if errors.As(err, &resourceNotFoundErr) { // Table not found
			returnedError = ErrElevenConfigNotFound
			return
		}

		returnedError = err
		return
	}

	if scanResp.Count == 0 { // Empty table
		returnedError = ErrElevenConfigNotFound
		return
	}

	if scanResp.Count > 1 { // Multiple rows
		returnedError = ErrElevenConfigDuplicated
		return
	}

	var records []DynamoDBElevenConfigTableRecord
	err = attributevalue.UnmarshalListOfMaps(scanResp.Items, &records)

	if err != nil {
		returnedError = err
		return
	}

	returnedConfigJSON = records[0].ConfigJSON
	return
}
