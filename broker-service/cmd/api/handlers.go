package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "Hit the broker",
	}

	_ = app.writeJSON(w, http.StatusOK, payload)
}

func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload
	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	switch requestPayload.Action {
	case "authenticate":
		app.authenticate(w, requestPayload.Auth)
	default:
		app.errorJSON(w, errors.New("unknown action"), http.StatusBadRequest)
	}
}

func (app *Config) authenticate(w http.ResponseWriter, a AuthPayload) {
	// create some json we'll send to the auth microservice
	jsonData, err := json.MarshalIndent(a, "", "\t")
	if err != nil {
		log.Printf("Could not marshal JSON: %v", err)
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	// call the service
	resp, err := http.Post("http://auth-service/authenticate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// make sure we get back the correct status code
	if resp.StatusCode != http.StatusOK {
		log.Printf("Unexpected status code while requesting authentication-service: %d", resp.StatusCode)
		app.errorJSON(w, errors.New("invalid credentials"), http.StatusUnauthorized)
		return
	}

	// read the response
	var jsonFromService jsonResponse
	err = json.NewDecoder(resp.Body).Decode(&jsonFromService)
	if err != nil {
		log.Printf("Could not decode JSON response from authentication-service: %v", err)
		app.errorJSON(w, err)
		return
	}

	if jsonFromService.Error {
		log.Printf("Error from authentication-service: %s", jsonFromService.Message)
		app.errorJSON(w, errors.New(jsonFromService.Message), http.StatusUnauthorized)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = jsonFromService.Message
	payload.Data = jsonFromService.Data

	app.writeJSON(w, http.StatusOK, payload)
}
