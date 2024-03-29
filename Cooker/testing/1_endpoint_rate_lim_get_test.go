package test

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func TestRateLimiterGet1_endpoint(t *testing.T) {
	time.Sleep(1 * time.Second)
	makeRequest := func() (*http.Response, error) {
		client := &http.Client{}
		req, err := http.NewRequest("GET", "http://apigateway:8000/user/550e8400-e29b-41d4-a716-446655440000", nil)
		assert.NoError(t, err)
		return client.Do(req)
	}
	startTime := time.Now()
	i := 0
	for ; i < 10; i++ {
		resp, err := makeRequest()
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
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
	assert.Equal(t, 200, resp.StatusCode)
}
