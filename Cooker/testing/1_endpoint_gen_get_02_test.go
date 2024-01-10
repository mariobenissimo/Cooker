package test

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestGetInvalidUuidValue01_endpoint(t *testing.T) {
	req, err := http.NewRequest("GET", "http://apigateway:8000/user/not-a-valid-uuid", nil)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 400, resp.StatusCode)
}
