package main

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestPostUpperIntValue10_endpoint(t *testing.T) {
	requestPayload := map[string]interface {
	}{"Età": "unsupported", "Nome": "aa"}
	requestBody, err := json.Marshal(requestPayload)
	assert.NoError(t, err)
	req, err := http.NewRequest("POST", "http://apigateway:8000/auth/user", bytes.NewBuffer(requestBody))
	assert.NoError(t, err)
	token := GetTestToken()
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 400, resp.StatusCode)
}
