package helpers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLoadServicesFromJSON_Success verifies that a proper JSON file is read correctly.
func TestLoadServicesFromJSON_Success(t *testing.T) {
	// Create a temporary file with JSON content.
	tmpFile, err := os.CreateTemp("", "services-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	servicesJSON := `[{"ServiceName": "TestService", "ServiceStruct": "TestStruct", "ServiceRegister": "RegisterTest", "HandlerRegister": "RegisterHandlerTest"}]`
	if _, err := tmpFile.WriteString(servicesJSON); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Call the function and verify the results.
	services, err := LoadServicesFromJSON(tmpFile.Name())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(services) != 1 {
		t.Fatalf("Expected 1 service, got %d", len(services))
	}
	s := services[0]
	if s.ServiceName != "TestService" || s.ServiceStruct != "TestStruct" || s.ServiceRegister != "RegisterTest" || s.HandlerRegister != "RegisterHandlerTest" {
		t.Errorf("Unexpected service values: %+v", s)
	}
}

// TestLoadServicesFromJSON_FileNotFound verifies that an error is returned when the file does not exist.
func TestLoadServicesFromJSON_FileNotFound(t *testing.T) {
	_, err := LoadServicesFromJSON("nonexistent.json")
	if err == nil {
		t.Fatalf("Expected error for nonexistent file, got nil")
	}
}

// TestGenerateRegisterFile verifies that the register file is generated correctly.
func TestGenerateRegisterFile(t *testing.T) {
	// Create a temporary directory for testing.
	tmpDir := t.TempDir()

	// Save the current working directory.
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	// Change the working directory to the temporary directory.
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	// Restore the original directory when the test is done.
	defer os.Chdir(origDir)

	// Create the necessary directory structure.
	routesDir := filepath.Join("pkg", "routes")
	if err := os.MkdirAll(routesDir, 0755); err != nil {
		t.Fatalf("Failed to create directories: %v", err)
	}

	// Prepare sample services data.
	services := []Service{
		{
			ServiceName:     "TestService",
			ServiceStruct:   "TestStruct",
			ServiceRegister: "RegisterTest",
			HandlerRegister: "RegisterHandlerTest",
		},
	}

	// Generate the register file.
	GenerateRegisterFile(services)

	// Read and verify the generated file.
	generatedFilePath := filepath.Join("pkg", "routes", "generated_register.go")
	data, err := os.ReadFile(generatedFilePath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "func RegisterServers") {
		t.Errorf("Generated file does not contain RegisterServers function")
	}
	if !strings.Contains(content, "RegisterTest") {
		t.Errorf("Generated file does not contain expected service register call")
	}
}
