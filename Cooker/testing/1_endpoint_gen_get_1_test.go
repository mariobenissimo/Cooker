package test

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestGetCorrectValue1_endpoint(t *testing.T) {
	req, err := http.NewRequest("GET", "http://apigateway:8000/user/550e8400-e29b-41d4-a716-446655440000", nil)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	jsonResponse := []map[string]interface{}{}
	err = json.NewDecoder(resp.Body).Decode(&jsonResponse)
	assert.NoError(t, err)
	assert.Equal(t, len(jsonResponse), 1, "Expected TODO in the response")
}
