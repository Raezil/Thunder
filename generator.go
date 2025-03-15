package main

import (
	"encoding/json"
	"flag"
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
func loadServicesFromJSON(filepath string) ([]Service, error) {
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

func runCommand(name string, args ...string) error {
	// Create the command
	cmd := exec.Command(name, args...)

	// Set the command's output to log to the console
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	return cmd.Run()
}

// generateRegisterFile creates the register file dynamically
func generateRegisterFile(services []Service) {
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

func runCommandInDir(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir // Ustawienie katalogu roboczego
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// It generates proto files and builds from Prisma schema
func main() {
	proto := flag.String("proto", "", "Path to the .proto file")
	prisma := flag.Bool("prisma", true, "Whether to run the Prisma db push command")
	generate := flag.Bool("generate", true, "Whether to generate register functions")
	flag.Parse()

	// First command: Run protoc to generate Go code from .proto file
	if *proto != "" {
		if err := runCommand("protoc",
			"-I", ".",
			"--go_out=./pkg/services/generated",
			"--go_opt=paths=source_relative",
			"--go-grpc_out=./pkg/services/generated",
			"--go-grpc_opt=paths=source_relative",
			"--grpc-gateway_out=./pkg/services/generated",
			"--grpc-gateway_opt=paths=source_relative",
			"--rpc-impl_out=./pkg/services",
			"--openapiv2_out=./pkg/services",
			"--openapiv2_opt=logtostderr=true",
			*proto,
		); err != nil {
			log.Fatalf("Error executing protoc command: %v", err)
		}

		fmt.Println("Protobuf, gRPC, and gRPC Gateway files generated successfully!")
	}

	// Second command: Run Prisma command to push database changes
	if *prisma {
		if err := runCommandInDir("./pkg", "go", "run", "github.com/steebchen/prisma-client-go", "db", "push"); err != nil {
			log.Fatalf("Error executing Prisma command: %v", err)
		}
		fmt.Println("Prisma database changes pushed successfully!")
	}

	// Third step: Generate gRPC registration file
	if *generate {
		services, err := loadServicesFromJSON("services.json")
		if err != nil {
			log.Fatalf("Error loading services from JSON: %v", err)
		}
		generateRegisterFile(services)
	}
}
