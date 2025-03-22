package helpers

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"text/template"
)

// Service holds the metadata needed for generating the register functions.
type Service struct {
	ServiceName     string `json:"ServiceName"`
	ServiceStruct   string `json:"ServiceStruct"`
	ServiceRegister string `json:"ServiceRegister"`
	HandlerRegister string `json:"HandlerRegister"`
}

// loadServicesFromJSON reads the service definitions from a JSON file.
func LoadServicesFromJSON(filepath string) ([]Service, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	var services []Service
	err = json.Unmarshal(data, &services)
	return services, err
}

// Template for RegisterServers and RegisterHandlers.
const templateCode = `package routes

import (
	"context"
	"log"
	. "generated"

	"google.golang.org/grpc"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"db"
	pb "services"
)

// RegisterServers registers gRPC services to the server.
func RegisterServers(server *grpc.Server, client *db.PrismaClient, sugar *zap.SugaredLogger) {
	{{range .}}
	{{.ServiceRegister}}(server, &pb.{{.ServiceStruct}}{
		PrismaClient: client,
		Logger:       sugar,
	})
	{{end}}
}

// RegisterHandlers registers gRPC-Gateway handlers.
func RegisterHandlers(gwmux *runtime.ServeMux, conn *grpc.ClientConn) {
	var err error
	{{range .}}
	err = {{.HandlerRegister}}(context.Background(), gwmux, conn)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}
	{{end}}
}
`

func RunCommand(name string, args ...string) error {
	// Create the command
	cmd := exec.Command(name, args...)

	// Set the command's output to log to the console
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	return cmd.Run()
}

// generateRegisterFile creates the register file dynamically
func GenerateRegisterFile(services []Service) {
	tmpl, err := template.New("register").Parse(templateCode)
	if err != nil {
		log.Fatalf("Error parsing template: %v", err)
	}

	file, err := os.Create("pkg/routes/generated_register.go")
	if err != nil {
		log.Fatalf("Error creating file: %v", err)
	}
	defer file.Close()

	err = tmpl.Execute(file, services)
	if err != nil {
		log.Fatalf("Error executing template: %v", err)
	}
	fmt.Println("Generated register file: pkg/internal/routes/generated_register.go")
}

func RunCommandInDir(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir // Ustawienie katalogu roboczego
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
