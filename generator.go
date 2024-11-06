package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func runCommand(name string, args ...string) error {
	// Create the command
	cmd := exec.Command(name, args...)

	// Set the command's output to log to the console
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	return cmd.Run()
}

func main() {
	proto := flag.String("proto", "", "Path to the .proto file")
	prisma := flag.Bool("prisma", true, "Whether to run the Prisma db push command")
	flag.Parse()
	// First command: Run protoc to generate Go code from .proto file
	if *proto != "" {
		if err := runCommand("protoc",
			"-I", ".",
			"--go_out=./backend",
			"--go_opt=paths=source_relative",
			"--go-grpc_out=./backend",
			"--go-grpc_opt=paths=source_relative",
			"--grpc-gateway_out=./backend",
			"--grpc-gateway_opt=paths=source_relative",
			"--rpc-impl_out=.",
			*proto,
		); err != nil {
			log.Fatalf("Error executing protoc command: %v", err)
		}

		fmt.Println("Protobuf, gRPC, and gRPC Gateway files generated successfully!")

	}
	if *prisma == true {
		// Second command: Run Prisma command to push database changes
		if err := runCommand("go", "run", "github.com/steebchen/prisma-client-go", "db", "push"); err != nil {
			log.Fatalf("Error executing Prisma command: %v", err)
		}

		fmt.Println("Prisma database changes pushed successfully!")
	}
}
