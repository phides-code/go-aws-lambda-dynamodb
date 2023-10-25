package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/go-playground/validator"
)

var validate *validator.Validate = validator.New()

func router(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Printf("Received req %#v", req)

	switch req.HTTPMethod {
	case "GET":
		return processGet(ctx, req)
	case "POST":
		return processPost(ctx, req)
	case "DELETE":
		return processDelete(ctx, req)
	case "PUT":
		return processPut(ctx, req)
	default:
		return clientError(http.StatusMethodNotAllowed)
	}
}

func processGet(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	id, ok := req.PathParameters["id"]
	if !ok {
		return processGetPeople(ctx)
	} else {
		return processGetPerson(ctx, id)
	}
}

func processGetPerson(ctx context.Context, id string) (events.APIGatewayProxyResponse, error) {
	log.Printf("Received GET person request with id = %s", id)

	person, err := getItem(ctx, id)
	if err != nil {
		return serverError(err)
	}

	if person == nil {
		return clientError(http.StatusNotFound)
	}

	json, err := json.Marshal(person)
	if err != nil {
		return serverError(err)
	}
	log.Printf("Successfully fetched person item %s", json)

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(json),
	}, nil
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

func processPost(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var createPerson CreatePerson
	err := json.Unmarshal([]byte(req.Body), &createPerson)
	if err != nil {
		log.Printf("Can't unmarshal body: %v", err)
		return clientError(http.StatusUnprocessableEntity)
	}

	err = validate.Struct(&createPerson)
	if err != nil {
		log.Printf("Invalid body: %v", err)
		return clientError(http.StatusBadRequest)
	}
	log.Printf("Received POST request with item: %+v", createPerson)

	res, err := insertItem(ctx, createPerson)
	if err != nil {
		return serverError(err)
	}
	log.Printf("Inserted new person: %+v", res)

	json, err := json.Marshal(res)
	if err != nil {
		return serverError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusCreated,
		Body:       string(json),
		Headers: map[string]string{
			"Location": fmt.Sprintf("/people/%s", res.Id),
		},
	}, nil
}

func processDelete(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	id, ok := req.PathParameters["id"]
	if !ok {
		return clientError(http.StatusBadRequest)
	}
	log.Printf("Received DELETE request with id = %s", id)

	person, err := deleteItem(ctx, id)
	if err != nil {
		return serverError(err)
	}

	if person == nil {
		return clientError(http.StatusNotFound)
	}

	json, err := json.Marshal(person)
	if err != nil {
		return serverError(err)
	}
	log.Printf("Successfully deleted person item %+v", person)

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(json),
	}, nil
}

func processPut(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	id, ok := req.PathParameters["id"]
	if !ok {
		return clientError(http.StatusBadRequest)
	}

	var updatePerson UpdatePerson
	err := json.Unmarshal([]byte(req.Body), &updatePerson)
	if err != nil {
		log.Printf("Can't unmarshal body: %v", err)
		return clientError(http.StatusUnprocessableEntity)
	}

	err = validate.Struct(&updatePerson)
	if err != nil {
		log.Printf("Invalid body: %v", err)
		return clientError(http.StatusBadRequest)
	}
	log.Printf("Received PUT request with item: %+v", updatePerson)

	res, err := updateItem(ctx, id, updatePerson)
	if err != nil {
		return serverError(err)
	}

	if res == nil {
		return clientError(http.StatusNotFound)
	}

	log.Printf("Updated person: %+v", res)

	json, err := json.Marshal(res)
	if err != nil {
		return serverError(err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(json),
		Headers: map[string]string{
			"Location": fmt.Sprintf("/people/%s", res.Id),
		},
	}, nil
}
