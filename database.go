package main

import (
	"context"
	"errors"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/smithy-go"
	"github.com/google/uuid"
)

type Person struct {
	Id   string `json:"id" dynamodbav:"id"`
	Name string `json:"name" dynamodbav:"name"`
}

type CreatePerson struct {
	Name string `json:"name" validate:"required"`
}

type UpdatePerson struct {
	Name string `json:"name" validate:"required"`
}

const TableName = "People"

func getClient() (dynamodb.Client, error) {
	sdkConfig, err := config.LoadDefaultConfig(context.TODO())

	dbClient := *dynamodb.NewFromConfig(sdkConfig)

	return dbClient, err

}

func getItem(ctx context.Context, id string) (*Person, error) {
	key, err := attributevalue.Marshal(id)
	if err != nil {
		return nil, err
	}

	input := &dynamodb.GetItemInput{
		TableName: aws.String(TableName),
		Key: map[string]types.AttributeValue{
			"id": key,
		},
	}

	log.Printf("Calling DynamoDB with input: %v", input)
	result, err := db.GetItem(ctx, input)
	if err != nil {
		return nil, err
	}
	log.Printf("Executed GetItem DynamoDb successfully. Result: %#v", result)

	if result.Item == nil {
		return nil, nil
	}

	person := new(Person)
	err = attributevalue.UnmarshalMap(result.Item, person)
	if err != nil {
		return nil, err
	}

	return person, nil
}

func listItems(ctx context.Context) ([]Person, error) {
	people := make([]Person, 0)

	var token map[string]types.AttributeValue

	for {
		input := &dynamodb.ScanInput{
			TableName:         aws.String(TableName),
			ExclusiveStartKey: token,
		}

		result, err := db.Scan(ctx, input)
		if err != nil {
			return nil, err
		}

		var fetchedPeople []Person
		err = attributevalue.UnmarshalListOfMaps(result.Items, &fetchedPeople)
		if err != nil {
			return nil, err
		}

		people = append(people, fetchedPeople...)
		token = result.LastEvaluatedKey
		if token == nil {
			break
		}

	}

	return people, nil
}

func insertItem(ctx context.Context, createPerson CreatePerson) (*Person, error) {
	person := Person{
		Name: createPerson.Name,
		Id:   uuid.NewString(),
	}

	item, err := attributevalue.MarshalMap(person)
	if err != nil {
		return nil, err
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(TableName),
		Item:      item,
	}

	res, err := db.PutItem(ctx, input)
	if err != nil {
		return nil, err
	}

	err = attributevalue.UnmarshalMap(res.Attributes, &person)
	if err != nil {
		return nil, err
	}

	return &person, nil
}

func deleteItem(ctx context.Context, id string) (*Person, error) {
	key, err := attributevalue.Marshal(id)
	if err != nil {
		return nil, err
	}

	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(TableName),
		Key: map[string]types.AttributeValue{
			"id": key,
		},
		ReturnValues: types.ReturnValue(*aws.String("ALL_OLD")),
	}

	res, err := db.DeleteItem(ctx, input)
	if err != nil {
		return nil, err
	}

	if res.Attributes == nil {
		return nil, nil
	}

	person := new(Person)
	err = attributevalue.UnmarshalMap(res.Attributes, person)
	if err != nil {
		return nil, err
	}

	return person, nil
}

func updateItem(ctx context.Context, id string, updatePerson UpdatePerson) (*Person, error) {
	key, err := attributevalue.Marshal(id)
	if err != nil {
		return nil, err
	}

	expr, err := expression.NewBuilder().WithUpdate(
		expression.Set(
			expression.Name("name"),
			expression.Value(updatePerson.Name),
		),
	).WithCondition(
		expression.Equal(
			expression.Name("id"),
			expression.Value(id),
		),
	).Build()
	if err != nil {
		return nil, err
	}

	input := &dynamodb.UpdateItemInput{
		Key: map[string]types.AttributeValue{
			"id": key,
		},
		TableName:                 aws.String(TableName),
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),

		ConditionExpression: expr.Condition(),
		ReturnValues:        types.ReturnValue(*aws.String("ALL_NEW")),
	}

	res, err := db.UpdateItem(ctx, input)
	if err != nil {
		var smErr *smithy.OperationError
		if errors.As(err, &smErr) {
			var condCheckFailed *types.ConditionalCheckFailedException
			if errors.As(err, &condCheckFailed) {
				return nil, nil
			}
		}

		return nil, err
	}

	if res.Attributes == nil {
		return nil, nil
	}

	person := new(Person)
	err = attributevalue.UnmarshalMap(res.Attributes, person)
	if err != nil {
		return nil, err
	}

	return person, nil
}
