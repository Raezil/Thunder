package tests

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// User represents the payload structure for both registration and login.
type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name,omitempty"`
	Surname  string `json:"surname,omitempty"`
	Age      int    `json:"age,omitempty"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

// setupTestEnvironment sets up the network, PostgreSQL, and application containers,
// and returns the application URL along with a cleanup function.
func setupTestEnvironment(t *testing.T) (appURL string, cleanup func()) {
	ctx := context.Background()
	networkName := fmt.Sprintf("test-network-%d", time.Now().UnixNano())
	networkReq := testcontainers.NetworkRequest{
		Name:           networkName,
		CheckDuplicate: true,
	}
	network, err := testcontainers.GenericNetwork(ctx, testcontainers.GenericNetworkRequest{
		NetworkRequest: networkReq,
	})
	if err != nil {
		t.Fatalf("failed to create network: %v", err)
	}

	cleanupNetwork := func() {
		network.Remove(ctx)
	}

	// Launch PostgreSQL container with a network alias.
	postgresReq := testcontainers.ContainerRequest{
		Image:        "postgres:13",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_DB":       "testdb",
		},
		Networks: []string{networkName},
		NetworkAliases: map[string][]string{
			networkName: {"postgres"},
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}
	postgresC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: postgresReq,
		Started:          true,
	})
	if err != nil {
		cleanupNetwork()
		t.Fatalf("failed to start postgres container: %v", err)
	}
	cleanupPostgres := func() {
		postgresC.Terminate(ctx)
	}

	// Update the connection string to use the network alias "postgres".
	dbConnStr := "postgres://postgres:postgres@postgres:5432/testdb?sslmode=disable"
	t.Logf("Postgres connection string: %s", dbConnStr)

	// Launch the application container on the same network.
	appReq := testcontainers.ContainerRequest{
		Image:        "raezil/app:latest",
		ExposedPorts: []string{"8080/tcp"},
		Env: map[string]string{
			"DATABASE_URL": dbConnStr,
			"JWT_SECRET":   "supersecret",
		},
		Networks:   []string{networkName},
		WaitingFor: wait.ForListeningPort("8080/tcp"),
	}
	appContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: appReq,
		Started:          true,
	})
	if err != nil {
		cleanupPostgres()
		cleanupNetwork()
		t.Fatalf("failed to start raezil/app:latest container: %v", err)
	}
	cleanupApp := func() {
		appContainer.Terminate(ctx)
	}

	// Retrieve host and port for the application container.
	appHost, err := appContainer.Host(ctx)
	if err != nil {
		cleanupApp()
		cleanupPostgres()
		cleanupNetwork()
		t.Fatalf("failed to get app container host: %v", err)
	}
	mappedPort, err := appContainer.MappedPort(ctx, "8080/tcp")
	if err != nil {
		cleanupApp()
		cleanupPostgres()
		cleanupNetwork()
		t.Fatalf("failed to get mapped port: %v", err)
	}
	url := fmt.Sprintf("https://%s:%s", appHost, mappedPort.Port())
	t.Logf("Application is running at %s", url)

	// Allow the application time to fully initialize.
	time.Sleep(45 * time.Second)

	cleanupFunc := func() {
		cleanupApp()
		cleanupPostgres()
		cleanupNetwork()
	}
	return url, cleanupFunc
}

// postJSONWithResponse sends a POST request with JSON payload, verifies the status code,
// and decodes the response into the provided response interface.
func postJSONWithResponse(client *http.Client, url string, data interface{}, expectedStatus int, response interface{}) error {
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	req, err := http.NewRequest("POST", url, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != expectedStatus {
		return fmt.Errorf("unexpected status code: got %d, expected %d", resp.StatusCode, expectedStatus)
	}
	return json.NewDecoder(resp.Body).Decode(response)
}

// postJSON sends a POST request with a JSON payload and validates the HTTP response status.
func postJSON(client *http.Client, url string, data interface{}, expectedStatus int) error {
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	req, err := http.NewRequest("POST", url, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	log.Println("Response:", string(body))
	if resp.StatusCode != expectedStatus {
		return fmt.Errorf("unexpected status code: got %d, expected %d. Response: %s", resp.StatusCode, expectedStatus, string(body))
	}
	return nil
}

// postJSONWithAuth sends a POST request with a JSON payload and an Authorization header.
func postJSONWithAuth(client *http.Client, url string, data interface{}, expectedStatus int, token string) error {
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	req, err := http.NewRequest("POST", url, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	log.Println("Response:", string(body))
	if resp.StatusCode != expectedStatus {
		return fmt.Errorf("unexpected status code: got %d, expected %d", resp.StatusCode, expectedStatus)
	}
	return nil
}

// Existing integration test for successful registration, login, and protected endpoint.
func TestContainers(t *testing.T) {
	appURL, cleanup := setupTestEnvironment(t)
	defer cleanup()

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// Registration simulation.
	registerURL := appURL + "/v1/auth/register"
	registerPayload := User{
		Email:    "newuser@example.com",
		Password: "password123",
		Name:     "John",
		Surname:  "Doe",
		Age:      30,
	}
	if err := postJSON(client, registerURL, registerPayload, 200); err != nil {
		t.Fatalf("registration failed: %v", err)
	}

	// Login simulation.
	loginURL := appURL + "/v1/auth/login"
	loginPayload := User{
		Email:    "newuser@example.com",
		Password: "password123",
	}
	if err := postJSON(client, loginURL, loginPayload, 200); err != nil {
		t.Fatalf("login failed: %v", err)
	}

	var tokenResp TokenResponse
	if err := postJSONWithResponse(client, loginURL, loginPayload, 200, &tokenResp); err != nil {
		t.Fatalf("login failed: %v", err)
	}

	// Protected endpoint simulation.
	protectedURL := appURL + "/v1/auth/protected"
	protectedPayload := map[string]string{"text": "sample request"}
	if err := postJSONWithAuth(client, protectedURL, protectedPayload, 200, tokenResp.Token); err != nil {
		t.Fatalf("protected request failed: %v", err)
	}
	t.Log("Both registration and login returned the expected status codes.")
}

// TestDuplicateRegistrationIntegration verifies that duplicate user registration fails.
func TestDuplicateRegistrationIntegration(t *testing.T) {
	appURL, cleanup := setupTestEnvironment(t)
	defer cleanup()

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	registerURL := appURL + "/v1/auth/register"
	userPayload := User{
		Email:    "dupuser@example.com",
		Password: "password123",
		Name:     "Dup",
		Surname:  "User",
		Age:      28,
	}
	// First registration should succeed.
	if err := postJSON(client, registerURL, userPayload, 200); err != nil {
		t.Fatalf("first registration failed: %v", err)
	}
	// Duplicate registration should fail (expecting status 400).
	if err := postJSON(client, registerURL, userPayload, 400); err == nil {
		t.Fatalf("duplicate registration succeeded, but expected failure")
	}
}

// TestLoginWithWrongPasswordIntegration verifies that login fails when using an incorrect password.
func TestLoginWithWrongPasswordIntegration(t *testing.T) {
	appURL, cleanup := setupTestEnvironment(t)
	defer cleanup()

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	registerURL := appURL + "/v1/auth/register"
	loginURL := appURL + "/v1/auth/login"
	userPayload := User{
		Email:    "wrongpass@example.com",
		Password: "correctpass",
		Name:     "Wrong",
		Surname:  "Password",
		Age:      30,
	}
	// Register the user.
	if err := postJSON(client, registerURL, userPayload, 200); err != nil {
		t.Fatalf("registration failed: %v", err)
	}
	// Attempt login with an incorrect password (expecting status 401).
	wrongLoginPayload := User{
		Email:    "wrongpass@example.com",
		Password: "incorrectpass",
	}
	if err := postJSON(client, loginURL, wrongLoginPayload, 401); err == nil {
		t.Fatalf("login with wrong password succeeded, expected failure")
	}
}

// TestProtectedWithoutTokenIntegration verifies that accessing the protected endpoint without a token fails.
func TestProtectedWithoutTokenIntegration(t *testing.T) {
	appURL, cleanup := setupTestEnvironment(t)
	defer cleanup()

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	protectedURL := appURL + "/v1/auth/protected"
	payload := map[string]string{"text": "sample request"}
	// Call protected endpoint without Authorization header (expecting status 401).
	if err := postJSON(client, protectedURL, payload, 401); err == nil {
		t.Fatalf("accessing protected endpoint without token succeeded, expected failure")
	}
}

// TestProtectedWithInvalidTokenIntegration verifies that using an invalid token to access the protected endpoint fails.
func TestProtectedWithInvalidTokenIntegration(t *testing.T) {
	appURL, cleanup := setupTestEnvironment(t)
	defer cleanup()

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	protectedURL := appURL + "/v1/auth/protected"
	payload := map[string]string{"text": "sample request"}
	invalidToken := "invalid.token.here"
	// Call protected endpoint with an invalid token (expecting status 401).
	if err := postJSONWithAuth(client, protectedURL, payload, 401, invalidToken); err == nil {
		t.Fatalf("accessing protected endpoint with invalid token succeeded, expected failure")
	}
}
