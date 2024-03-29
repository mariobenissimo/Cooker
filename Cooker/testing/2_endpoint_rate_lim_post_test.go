package test

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func TestRateLimiterPost2_endpoint(t *testing.T) {
	time.Sleep(1 * time.Second)
	makeRequest := func() (*http.Response, error) {
		client := &http.Client{}
		requestPayload := map[string]interface {
		}{"Nome": "aa", "Età": 10.000000}
		requestBody, err := json.Marshal(requestPayload)
		assert.NoError(t, err)
		req, err := http.NewRequest("POST", "http://apigateway:8000/auth/user", bytes.NewBuffer(requestBody))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		token := GetTestToken()
		req.Header.Set("Authorization", "Bearer "+token)
		assert.NoError(t, err)
		return client.Do(req)
	}
	startTime := time.Now()
	i := 0
	for ; i < 10; i++ {
		resp, err := makeRequest()
		assert.NoError(t, err)
		assert.Equal(t, 201, resp.StatusCode)
	}
	endTime := time.Now()
	elapsedTime := endTime.Sub(startTime)
	if elapsedTime.Seconds() < 1 {
		resp, err := makeRequest()
		assert.NoError(t, err)
		assert.Equal(t, 429, resp.StatusCode)
	}
	time.Sleep(1 * time.Second)
	resp, err := makeRequest()
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)
}
