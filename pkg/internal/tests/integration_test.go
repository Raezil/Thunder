package tests

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
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

func TestContainers(t *testing.T) {
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
	defer network.Remove(ctx)

	// Launch PostgreSQL container with a network alias.
	postgresReq := testcontainers.ContainerRequest{
		Image:        "postgres:13",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		Networks: []string{networkName},
		// Set the network alias so that other containers can refer to it by name.
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
		t.Fatalf("failed to start postgres container: %v", err)
	}
	defer postgresC.Terminate(ctx)

	// Update the connection string to use the network alias "postgres".
	dbConnStr := "postgres://testuser:testpass@postgres:5432/testdb?sslmode=disable"
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
		// Retrieve and log container logs for debugging.
		if appContainer != nil {
			logs, logErr := appContainer.Logs(ctx)
			if logErr == nil {
				buf := new(strings.Builder)
				_, copyErr := io.Copy(buf, logs)
				if copyErr == nil {
					t.Logf("Application container logs:\n%s", buf.String())
				} else {
					t.Logf("Failed to read container logs: %v", copyErr)
				}
			} else {
				t.Logf("Failed to get container logs: %v", logErr)
			}
		}
		t.Fatalf("failed to start raezil/app:latest container: %v", err)
	}
	defer appContainer.Terminate(ctx)

	// Retrieve host and port for the application container.
	appHost, err := appContainer.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get app container host: %v", err)
	}
	mappedPort, err := appContainer.MappedPort(ctx, "8080/tcp")
	if err != nil {
		t.Fatalf("failed to get mapped port: %v", err)
	}
	appURL := fmt.Sprintf("https://%s:%s", appHost, mappedPort.Port())

	t.Logf("Application is running at %s", appURL)
	// Optionally wait for a few seconds to ensure the application is fully started.
	time.Sleep(15 * time.Second)

	// Configure an HTTP client. If your app doesn't use TLS, change the scheme above to "http".
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// 4. Simulate the curl commands.

	// Registration: simulate
	registerURL := appURL + "/v1/auth/register"
	registerPayload := User{
		Email:    "newuser@example.com",
		Password: "password123",
		Name:     "John",
		Surname:  "Doe",
		Age:      30,
	}
	// Expecting 201 Created or adjust as per your app behavior.
	if err := postJSON(client, registerURL, registerPayload, 200); err != nil {
		t.Fatalf("registration failed: %v", err)
	}

	// Login: simulate
	loginURL := appURL + "/v1/auth/login"
	loginPayload := User{
		Email:    "newuser@example.com",
		Password: "password123",
	}
	// Expecting 200 OK or adjust as needed.
	if err := postJSON(client, loginURL, loginPayload, 20); err != nil {
		t.Fatalf("login failed: %v", err)
	}

	t.Log("Both registration and login returned the expected status codes.")
}

// postJSON simulates a curl POST request with JSON payload,
// validates the HTTP response status code, and logs the response.
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

	// Log the response for debugging purposes.
	return nil
}
