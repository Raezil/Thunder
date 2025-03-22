package main

import (
	"flag"
	"fmt"
	. "helpers"
	"log"
)

// It generates proto files and builds from Prisma schema
func main() {
	proto := flag.String("proto", "", "Path to the .proto file")
	prisma := flag.Bool("prisma", true, "Whether to run the Prisma db push command")
	generate := flag.Bool("generate", true, "Whether to generate register functions")
	flag.Parse()

	// First command: Run protoc to generate Go code from .proto file
	if *proto != "" {
		if err := RunCommand("protoc",
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
		if err := RunCommandInDir("./pkg", "go", "run", "github.com/steebchen/prisma-client-go", "db", "push"); err != nil {
			log.Fatalf("Error executing Prisma command: %v", err)
		}
		fmt.Println("Prisma database changes pushed successfully!")
	}

	// Third step: Generate gRPC registration file
	if *generate {
		services, err := LoadServicesFromJSON("services.json")
		if err != nil {
			log.Fatalf("Error loading services from JSON: %v", err)
		}
		GenerateRegisterFile(services)
	}
}
