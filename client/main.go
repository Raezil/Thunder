package main

import (
	. "backend"
	"context"
	"crypto/tls"
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

func main() {
	// Create a custom TLS config that skips certificate verification.
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // Only use this for testing!
	}
	tlsCreds := credentials.NewTLS(tlsConfig)

	// Dial the gRPC server using TLS credentials.
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(tlsCreds), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := NewAuthClient(conn)
	ctx := context.Background()
	registerReply, err := client.Register(ctx, &RegisterRequest{
		Email:    "kmosc111238@example.com",
		Password: "password",
		Name:     "Kamil",
		Surname:  "Mosciszko",
		Age:      27,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("Received registration response:", registerReply)

	loginReply, err := client.Login(ctx, &LoginRequest{
		Email:    "kmosc111238@example.com",
		Password: "password",
	})
	if err != nil {
		log.Fatalf("Login failed: %v", err)
	}

	token := loginReply.Token
	fmt.Println("Received JWT token:", token)
	md := metadata.Pairs("authorization", token)
	outgoingCtx := metadata.NewOutgoingContext(ctx, md)
	protectedReply, err := client.SampleProtected(outgoingCtx, &ProtectedRequest{
		Text: "Hello from client",
	})
	if err != nil {
		log.Fatalf("SampleProtected failed: %v", err)
	}
	fmt.Println("SampleProtected response:", protectedReply.Result)
}
