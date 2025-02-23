package main

import (
	. "backend"
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := NewAuthClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	registerReply, err := client.Register(ctx, &RegisterRequest{
		Email:    "kmosc1231@example.com", // Use a new email address here
		Password: "password",
		Name:     "Kamil",
		Surname:  "Mosciszko",
		Age:      27,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("Received JWT token:", registerReply)

	loginReply, err := client.Login(ctx, &LoginRequest{
		Email:    "kmosc1231@example.com",
		Password: "password",
	})
	if err != nil {
		log.Fatalf("Login failed: %v", err)
	}

	token := loginReply.Token
	fmt.Println("Received JWT token:", token)
	md := metadata.Pairs("authorization", token)
	context := metadata.NewOutgoingContext(ctx, md)
	protectedReply, err := client.SampleProtected(context, &ProtectedRequest{
		Text: "Hello from client",
	})
	if err != nil {
		log.Fatalf("SampleProtected failed: %v", err)
	}
	fmt.Println("SampleProtected response:", protectedReply.Result)
}
