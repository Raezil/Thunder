package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	. "generated"
	"io/ioutil"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

func main() {

	serverCert, err := ioutil.ReadFile("../../certs/server.crt")
	if err != nil {
		log.Fatalf("could not read server certificate: %v", err)
	}

	// Create a certificate pool from the server certificate.
	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(serverCert); !ok {
		log.Fatal("failed to append server certificate")
	}

	// Set up TLS configuration with the certificate pool and enforce TLS 1.2 or higher.
	tlsConfig := &tls.Config{
		RootCAs:    certPool,
		MinVersion: tls.VersionTLS12,
	}
	tlsCreds := credentials.NewTLS(tlsConfig)

	// Dial the gRPC server using TLS credentials.
	conn, err := grpc.Dial("localhost:50051",
		grpc.WithTransportCredentials(tlsCreds),
		grpc.WithBlock(),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1024*5)),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := NewAuthClient(conn)
	ctx := context.Background()
	registerReply, err := client.Register(ctx, &RegisterRequest{
		Email:    "kmosc1211@example.com",
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
		Email:    "kmosc1211@example.com",
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
