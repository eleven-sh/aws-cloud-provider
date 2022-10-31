package service

import (
	"encoding/json"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/eleven-sh/aws-cloud-provider/infrastructure"
	"github.com/eleven-sh/eleven/entities"
	"github.com/eleven-sh/eleven/stepper"
)

func (a *AWS) CreateElevenConfigStorage(
	stepper stepper.Stepper,
) error {

	dynamoDBClient := dynamodb.NewFromConfig(a.sdkConfig)

	stepper.StartTemporaryStep("Creating a DynamoDB table to store the Eleven configuration")

	err := infrastructure.CreateDynamoDBTableForElevenConfig(
		dynamoDBClient,
	)

	if err != nil && errors.Is(err, infrastructure.ErrElevenConfigTableAlreadyExists) {
		return nil
	}

	return err
}

func (a *AWS) LookupElevenConfig(
	stepper stepper.Stepper,
) (*entities.Config, error) {

	dynamoDBClient := dynamodb.NewFromConfig(a.sdkConfig)

	configJSON, err := infrastructure.LookupElevenConfigInDynamoDBTable(
		dynamoDBClient,
	)

	if err != nil {

		if errors.Is(err, infrastructure.ErrElevenConfigNotFound) {
			// No config table or no records.
			return nil, entities.ErrElevenNotInstalled
		}

		return nil, err
	}

	var elevenConfig *entities.Config
	err = json.Unmarshal([]byte(configJSON), &elevenConfig)

	if err != nil {
		return nil, err
	}

	return elevenConfig, nil
}

func (a *AWS) SaveElevenConfig(
	stepper stepper.Stepper,
	config *entities.Config,
) error {

	configJSON, err := json.Marshal(config)

	if err != nil {
		return err
	}

	dynamoDBClient := dynamodb.NewFromConfig(a.sdkConfig)

	return infrastructure.UpdateElevenConfigInDynamoDBTable(
		dynamoDBClient,
		config.ID,
		string(configJSON),
	)
}

func (a *AWS) RemoveElevenConfigStorage(
	stepper stepper.Stepper,
) error {

	dynamoDBClient := dynamodb.NewFromConfig(a.sdkConfig)

	stepper.StartTemporaryStep("Removing the DynamoDB table used to store the Eleven configuration")

	return infrastructure.RemoveDynamoDBTableForElevenConfig(
		dynamoDBClient,
	)
}
