package main

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestGetCorrectValue2_endpoint(t *testing.T) {
	req, err := http.NewRequest("GET", "http://apigateway:8000/user", nil)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	jsonResponse := []map[string]interface{}{}
	err = json.NewDecoder(resp.Body).Decode(&jsonResponse)
	assert.NoError(t, err)
	assert.Equal(t, len(jsonResponse), 2, "Expected at least two values in the response")
}
