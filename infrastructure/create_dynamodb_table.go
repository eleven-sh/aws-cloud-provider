package infrastructure

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const DynamoDBElevenConfigTableName = "eleven-configuration-dynamodb-table"

var (
	ErrElevenConfigTableAlreadyExists = errors.New("ErrElevenConfigTableAlreadyExists")
)

func CreateDynamoDBTableForElevenConfig(
	dynamoDBClient *dynamodb.Client,
) error {

	_, err := dynamoDBClient.CreateTable(
		context.TODO(),

		&dynamodb.CreateTableInput{
			AttributeDefinitions: []types.AttributeDefinition{
				{
					AttributeName: aws.String("ID"),
					AttributeType: types.ScalarAttributeTypeS,
				},
			},

			KeySchema: []types.KeySchemaElement{
				{
					AttributeName: aws.String("ID"),
					KeyType:       types.KeyTypeHash,
				},
			},

			BillingMode: types.BillingModePayPerRequest,

			TableName: aws.String(DynamoDBElevenConfigTableName),
		},
	)

	if err != nil {
		var tableExistsError *types.ResourceInUseException

		if errors.As(err, &tableExistsError) {
			return ErrElevenConfigTableAlreadyExists
		}

		return err
	}

	existsWaiter := dynamodb.NewTableExistsWaiter(dynamoDBClient)
	maxWaitTime := 5 * time.Minute

	return existsWaiter.Wait(context.TODO(), &dynamodb.DescribeTableInput{
		TableName: aws.String(DynamoDBElevenConfigTableName),
	}, maxWaitTime)
}
