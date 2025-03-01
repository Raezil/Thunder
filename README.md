# **Thunder - A Minimalist Backend Framework in Go**
*A gRPC-Gateway-powered framework with Prisma, Kubernetes and Go for scalable microservices.*

[![Go Version](https://img.shields.io/badge/Go-1.21-blue)](https://golang.org)
[![License](https://img.shields.io/github/license/Raezil/Thunder)](LICENSE)
[![Stars](https://img.shields.io/github/stars/Raezil/Thunder)](https://github.com/Raezil/Thunder/stargazers)
[![Issues](https://img.shields.io/github/issues/Raezil/Thunder)](https://github.com/Raezil/Thunder/issues)

## **Table of Contents**
- [🚀 Features](#-features)
- [📌 Getting Started](#-getting-started)
  - [1️⃣ Install Dependencies](#1️⃣-install-dependencies)
  - [2️⃣ Define Your gRPC Service](#2️⃣-define-your-grpc-service)
- [🛠️ Prisma Integration](#️-prisma-integration)
- [🚀 Running the Server](#-running-the-server)
  - [a. Build and Install a Custom protoc Plugin](#a-build-and-install-a-custom-protoc-plugin)
  - [b. Code Generation](#b-code-generation)
  - [c. Start the **gRPC + REST API** server](#c-start-the-grpc--rest-api-server)
- [🚀 Running the Tests](#-running-the-tests)
  - [a. Mocking Tests](#a-mocking-tests)
  - [b. Running the Tests](#b-running-the-tests)
- [🔧 Kubernetes Deployment](#-kubernetes-deployment)
  - [1️⃣ Generate TLS Certificates](#1️⃣-generate-tls-certificates)
  - [2️⃣ Build & Push Docker Image](#2️⃣-build--push-docker-image)
  - [3️⃣ Deploy to Kubernetes](#3️⃣-deploy-to-kubernetes)
- [📡 API Testing](#-api-testing)
  - [Register a User](#register-a-user)
  - [Login](#login)
- [💡 Example Implementations](#-example-implementations)
  - [🔹 Server](#-server)
  - [🔹 Client](#-client)
- [📜 Contributing](#-contributing)
- [🔗 References](#-references)
- [📣 Stay Connected](#-stay-connected)

---

## **🚀 Features**
✔️ **gRPC + REST (gRPC-Gateway)** – Automatically expose RESTful APIs from gRPC services.  
✔️ **Prisma Integration** – Use Prisma for efficient database access in Go.  
✔️ **Kubernetes Ready** – Easily deploy and scale with Kubernetes.  
✔️ **TLS Security** – Secure gRPC communications with TLS.  
✔️ **Structured Logging** – Built-in `zap` logging.  
✔️ **Rate Limiting & Authentication** – Pre-configured middleware.  
✔️ **Modular & Extensible** – Easily extend Thunder for custom use cases.

---

## **📌 Getting Started**
### **1️⃣ Install Dependencies**
Ensure you have Go, `protoc`, and Prisma installed.  

```sh
go mod tidy
```

### **2️⃣ Define Your gRPC Service**
Create a `.proto` file, e.g., `user.proto`:

```proto
syntax = "proto3";

package example;

option go_package = "backend/";

import "google/api/annotations.proto";

// A simple service definition.
service UserService {
  rpc GetUser (UserRequest) returns (UserResponse) {
    option (google.api.http) = {
      get: "/v1/users/{id}"
    };
  }
}

// Request and response messages.
message UserRequest {
  int32 id = 1;
}

message UserResponse {
  int32 id = 1;
  string name = 2;
  int32 age = 3;
}
```
---

## **🛠️ Prisma Integration**
Thunder automatically integrates Prisma for database management. Define your schema:

## a. Create Your .proto File
```prisma
datasource db {
  provider = "postgresql"
  url      = env("DATABASE_URL")
}

model User {
  id    String @default(cuid()) @id
  name  String
  email String @unique
}
```

## **🚀 Running the Server**

#### a. Build and Install a Custom protoc Plugin

To generate your gRPC server implementations, you can build a custom protoc plugin. In Thunder, the plugin is built as follows:

##### Adding `protoc` Plugin
```
go build -o protoc-gen-rpc-impl ./cmd/protoc-gen-rpc-impl.go
sudo mv protoc-gen-rpc-impl /usr/local/bin
sudo chmod +x /usr/local/bin/protoc-gen-rpc-impl
```
#### b. Code Generation
```
go run generator.go -proto=filename.proto -prisma=true
```
> **Note:** Replace `filename` with the actual name of your gRPC service.

#### c. Start the **gRPC + REST API** server:

```sh
go run ./server/main.go
```
> **Note:** Generate TLS certificates prior running the server.

## **🚀 Running the Tests**
#### a. Mocking Tests
To mock a gRPC server:
```
cd backend
mockgen -source=yourservice_grpc.pb.go -destination=yourservice_mock.go
```
> **Note:** Replace `yourservice` with the actual name of your gRPC service. Look into /backend/authenticator_server_test.go to see how to develop tests or look into https://github.com/golang/mock

#### b. Running the Tests
```
go test ./backend/... ./db/...
```

---

## **🔧 Kubernetes Deployment**
### **1️⃣ Generate TLS Certificates**
```sh
mkdir certs
openssl req -x509 -newkey rsa:4096 -keyout certs/server.key -out certs/server.crt -days 365 -nodes \
  -subj "/CN=localhost" \
  -addext "subjectAltName=DNS:localhost,IP:127.0.0.1"
```

### **2️⃣ Build & Push Docker Image**
```
docker build -t app:latest .
docker login
docker push $docker_username/app:latest
```

> **Note:** Note $docker_username is your username, change it in k8s/app-deployment as well

### **3️⃣ Deploy to Kubernetes**
```sh
minikube start
cd k8s
kubectl apply -f postgres-deployment.yaml
kubectl apply -f postgres-service.yaml
kubectl apply -f postgres-pvc.yaml
kubectl apply -f app-deployment.yaml
kubectl apply -f app-service.yaml
kubectl apply -f app-loadbalancer.yaml
kubectl apply -f pgbouncer-all.yaml
kubectl apply -f hpa.yaml
kubectl rollout restart deployment pgbouncer
kubectl rollout restart deployment app-deployment
kubectl port-forward service/app-service 8080:8080
```

#### Checking Pod Status
```
kubectl get pods -n default
kubectl describe pod $NAME -n default
```

### Thunder CLI
For a comprehensive guide on how to use Thunder CLI—including installation steps, available commands, and usage examples—you can refer to the official documentation here:
https://github.com/Raezil/Thunder/blob/main/thunder-cli.md

This file covers everything you need to get started with Thunder CLI and will help you integrate it into your development workflow.

---

## **📡 API Testing**
### **Register a User**
```sh
curl -k --http2 -X POST https://localhost:8080/v1/auth/register \
     -H "Content-Type: application/json" \
     -d '{
           "email": "newuser@example.com",
           "password": "password123",
           "name": "John",
           "surname": "Doe",
           "age": 30
         }'
```

### **Login**
```sh
curl -k --http2 -X POST https://localhost:8080/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{
           "email": "newuser@example.com",
           "password": "password123"
         }'
```

---

## **💡 Example Implementations**
### **🔹 Server**
Example of a gRPC server running with authentication and rate limiting:

```go
package main

import (
	"context"
	"db"
	"io"
	"log"
	"net"
	"net/http"

	pb "backend"
	"middlewares"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/opentracing/opentracing-go"
	"github.com/spf13/viper"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// initConfig sets default values and loads environment variables.
func initConfig() {
	viper.SetDefault("grpc.port", ":50051")
	viper.SetDefault("http.port", ":8080")
	viper.AutomaticEnv()
}

// initJaeger initializes a Jaeger tracer based on environment configuration.
func initJaeger(service string) (opentracing.Tracer, io.Closer) {
	cfg, err := config.FromEnv()
	if err != nil {
		log.Fatalf("Failed to read Jaeger env vars: %v", err)
	}
	cfg.ServiceName = service
	tracer, closer, err := cfg.NewTracer(config.Logger(jaeger.StdLogger))
	if err != nil {
		log.Fatalf("Could not initialize Jaeger tracer: %v", err)
	}
	opentracing.SetGlobalTracer(tracer)
	return tracer, closer
}

// RegisterServers registers gRPC servers.
func RegisterServers(server *grpc.Server, client *db.PrismaClient, sugar *zap.SugaredLogger) {
	pb.RegisterAuthServer(server, &pb.AuthenticatorServer{
		PrismaClient: client,
		Logger:       sugar,
	})
}

// RegisterHandlers registers gRPC-Gateway handlers.
func RegisterHandlers(gwmux *runtime.ServeMux, conn *grpc.ClientConn) {
	err := pb.RegisterAuthHandler(context.Background(), gwmux, conn)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}
}

func main() {
	// Initialize configuration.
	initConfig()

	// Setup structured logging.
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	sugar := logger.Sugar()
	defer logger.Sync()

	// Initialize Jaeger tracer.
	_, closer := initJaeger("thunder-grpc")
	defer closer.Close()

	// Load TLS credentials for the gRPC server.
	certFile := "../certs/server.crt"
	keyFile := "../certs/server.key"
	creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
	if err != nil {
		sugar.Fatalf("Failed to load TLS credentials: %v", err)
	}

	// Listen on the configured gRPC port.
	grpcPort := viper.GetString("grpc.port")
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}

	// Connect to the database.
	client := db.NewClient()
	if err := client.Prisma.Connect(); err != nil {
		panic(err)
	}
	defer func() {
		if err := client.Prisma.Disconnect(); err != nil {
			panic(err)
		}
	}()

	// Initialize rate limiter (e.g., 5 requests per second, burst of 10).
	rateLimiter := middlewares.NewRateLimiter(5, 10)

	// Create the gRPC server with TLS and custom interceptors.
	grpcServer := grpc.NewServer(
		grpc.Creds(creds),
		grpc.UnaryInterceptor(
			middlewares.ChainUnaryInterceptors(
				rateLimiter.RateLimiterInterceptor, // Rate limiting
				middlewares.AuthUnaryInterceptor,   // Authentication
			),
		),
	)
	RegisterServers(grpcServer, client, sugar)

	// Register gRPC Health service.
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	sugar.Infof("Serving gRPC with TLS on 0.0.0.0%s", grpcPort)
	go func() {
		log.Fatalln(grpcServer.Serve(lis))
	}()

	// Setup secure connection for gRPC-Gateway.
	clientCreds, err := credentials.NewClientTLSFromFile(certFile, "localhost")
	if err != nil {
		sugar.Fatalf("Failed to load client TLS credentials: %v", err)
	}
	conn, err := grpc.Dial(
		"localhost"+grpcPort,
		grpc.WithTransportCredentials(clientCreds),
	)
	if err != nil {
		log.Fatalln("Failed to dial gRPC server:", err)
	}

	// Register gRPC-Gateway handlers.
	gwmux := runtime.NewServeMux()
	RegisterHandlers(gwmux, conn)

	// Create a new HTTP mux and add health and readiness endpoints.
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// You can add more logic here if needed.
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		// You might include checks (e.g., database connectivity) before reporting readiness.
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Ready"))
	})
	mux.Handle("/", gwmux)

	httpPort := viper.GetString("http.port")
	gwServer := &http.Server{
		Addr:    httpPort,
		Handler: mux,
	}

	sugar.Infof("Serving gRPC-Gateway on https://0.0.0.0%s", httpPort)
	log.Fatalln(gwServer.ListenAndServeTLS("../certs/server.crt", "../certs/server.key"))
}
```

### **🔹 Client**
A simple gRPC client:

```go
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
		Email:    "kmosc1238@example.com",
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
		Email:    "kmosc1238@example.com",
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
```

---

## **📜 Contributing**
Want to improve Thunder? 🚀  
1. Fork the repo  
2. Create a feature branch (`git checkout -b feature-new`)  
3. Commit your changes (`git commit -m "Added feature"`)  
4. Push to your branch (`git push origin feature-new`)  
5. Submit a PR!  

---

## **🔗 References**
- 📜 [Go Documentation](https://golang.org/doc/)  
- 📘 [gRPC-Gateway](https://grpc-ecosystem.github.io/grpc-gateway/)  
- 🛠️ [Prisma ORM](https://www.prisma.io/docs/)  
- ☁️ [Kubernetes Docs](https://kubernetes.io/docs/)  

---

## **📣 Stay Connected**
⭐ Star the repo if you find it useful!  
📧 For questions, reach out via [GitHub Issues](https://github.com/Raezil/Thunder/issues).  
