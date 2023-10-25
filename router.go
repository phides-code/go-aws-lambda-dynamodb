package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

func router(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Printf("Received req %#v", req)

	switch req.HTTPMethod {
	case "GET":
		return processGet(ctx, req)
	default:
		return clientError(http.StatusMethodNotAllowed)
	}
}

func processGet(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// id, ok := req.PathParameters["id"]
	// if !ok {
	// 	return processGetPeople(ctx)
	// } else {
	// 	return processGetPerson(ctx, id)
	// }

	return processGetPeople(ctx)
}

func processGetPeople(ctx context.Context) (events.APIGatewayProxyResponse, error) {
	log.Print("Received GET people request")

	people, err := listItems(ctx)
	if err != nil {
		return serverError(err)
	}

	json, err := json.Marshal(people)
	if err != nil {
		return serverError(err)
	}
	log.Printf("Successfully fetched people: %s", json)

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(json),
	}, nil
}
