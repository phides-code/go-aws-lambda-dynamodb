package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

func clientError(status int) (events.APIGatewayProxyResponse, error) {

	errorString := http.StatusText(status)

	response := ResponseStructure{
		Data:         nil,
		ErrorMessage: &errorString,
	}

	responseJson, _ := json.Marshal(response)

	return events.APIGatewayProxyResponse{
		Body:       string(responseJson),
		StatusCode: status,
	}, nil
}

func serverError(err error) (events.APIGatewayProxyResponse, error) {
	log.Println(err.Error())

	errorString := http.StatusText(http.StatusInternalServerError)

	response := ResponseStructure{
		Data:         nil,
		ErrorMessage: &errorString,
	}

	responseJson, _ := json.Marshal(response)

	return events.APIGatewayProxyResponse{
		Body:       string(responseJson),
		StatusCode: http.StatusInternalServerError,
	}, nil
}
