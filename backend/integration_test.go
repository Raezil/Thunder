package backend

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

const baseURL = "http://localhost:8080/api/v1"

func TestRegisterUserREST(t *testing.T) {
	reqBody := `{"username": "testuser", "password": "password123"}`
	resp, err := http.Post(baseURL+"/auth/register", "application/json", bytes.NewBuffer([]byte(reqBody)))

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestLoginUserREST(t *testing.T) {
	reqBody := `{"username": "testuser", "password": "password123"}`
	resp, err := http.Post(baseURL+"/auth/login", "application/json", bytes.NewBuffer([]byte(reqBody)))

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
