package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"routes"
	"text/template"
)

// Service holds the metadata needed for generating the register functions.
type Service struct {
	ServiceName     string
	ServiceStruct   string
	ServiceRegister string
	HandlerRegister string
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
func generateRegisterFile() {

	tmpl, err := template.New("register").Parse(templateCode)
	if err != nil {
		log.Fatalf("Error parsing template: %v", err)
	}

	file, err := os.Create("app/internal/routes/generated_register.go")
	if err != nil {
		log.Fatalf("Error creating file: %v", err)
	}
	defer file.Close()

	err = tmpl.Execute(file, routes.Services)
	if err != nil {
		log.Fatalf("Error executing template: %v", err)
	}
	fmt.Println("Generated register file: backend/generated_register.go")
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
			"--go_out=./app/internal/services/generated",
			"--go_opt=paths=source_relative",
			"--go-grpc_out=./app/internal/services/generated",
			"--go-grpc_opt=paths=source_relative",
			"--grpc-gateway_out=./app/internal/services/generated",
			"--grpc-gateway_opt=paths=source_relative",
			"--rpc-impl_out=./app/internal/services",
			"--openapiv2_out=./app/internal/services",
			"--openapiv2_opt=logtostderr=true",
			*proto,
		); err != nil {
			log.Fatalf("Error executing protoc command: %v", err)
		}

		fmt.Println("Protobuf, gRPC, and gRPC Gateway files generated successfully!")
	}

	// Second command: Run Prisma command to push database changes
	if *prisma {
		if err := runCommand("go", "run", "github.com/steebchen/prisma-client-go", "db", "push"); err != nil {
			log.Fatalf("Error executing Prisma command: %v", err)
		}

		fmt.Println("Prisma database changes pushed successfully!")
	}

	// Third step: Generate gRPC registration file
	if *generate {
		generateRegisterFile()
	}
}
