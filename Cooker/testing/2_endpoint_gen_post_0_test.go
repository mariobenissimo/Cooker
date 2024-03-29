package test

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestPostUnAuthorized2_endpoint(t *testing.T) {
	requestPayload := map[string]interface {
	}{"Nome": "aa", "Età": 10.000000}
	requestBody, err := json.Marshal(requestPayload)
	assert.NoError(t, err)
	req, err := http.NewRequest("POST", "http://apigateway:8000/auth/user", bytes.NewBuffer(requestBody))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 401, resp.StatusCode)
}
