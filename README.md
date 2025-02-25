## Thunder - Backend Framework (gRPC Gateway + Prisma + Kubernetes + Golang)
Thunder is a backend framework built with Golang that leverages gRPC Gateway, Prisma, and Kubernetes to simplify the development, testing, and deployment of scalable microservices.

### 1. Developing with Protocol Buffers (proto)
## a. Create Your .proto File

Start by defining your service and messages in a .proto file (for example, example.proto):
```
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
#### b. Build and Install a Custom protoc Plugin

To generate your gRPC server implementations, you can build a custom protoc plugin. In Thunder, the plugin is built as follows:

##### Adding `protoc` Plugin
```
go build -o protoc-gen-rpc-impl ./cmd/protoc-gen-rpc-impl.go
sudo mv protoc-gen-rpc-impl /usr/local/bin
sudo chmod +x /usr/local/bin/protoc-gen-rpc-impl
```
### Code Generation
```
go run generator.go -proto=filename.proto -prisma=true
```
> **Note:** Replace `filename` with the actual name of your gRPC service.

# 2. Developing with Prisma

When the -prisma=true flag is enabled, the generator will integrate Prisma into your project. 
Although Thunder’s approach is Golang based, the principle is similar to Prisma’s usage in other ecosystems.

##### a. Example Prisma Workflow (General Approach)
Define Your Data Model: If you’re using Prisma traditionally, you’d start with a schema like this in schema.prisma:

```
datasource db {
  provider = "sqlite" // or "postgresql", "mysql", etc.
  url      = "file:dev.db"
}

generator db {
  provider = "go run github.com/steebchen/prisma-client-go"
}

model User {
  id    Int    @id @default(autoincrement())
  name  String
  email String @unique
  age   Int
}
```
##### b. Prisma Integration with Thunder

With Thunder’s generator, much of the manual work of integrating Prisma is handled automatically. The generated Prisma files ensure that your database layer is aligned with your proto definitions. This streamlines development by reducing redundancy and keeping your API and database schema in sync.

##### Mocking Tests
To mock a gRPC server:
```
cd backend
mockgen -source=yourservice_grpc.pb.go -destination=yourservice_mock.go
```
> **Note:** Replace `yourservice` with the actual name of your gRPC service.

**Examples** Look into /backend/authenticator_server_test.go to see how to develop tests or look into https://github.com/golang/mock

## Kubernetes Deployment

### Building and Pushing Docker Image
```
docker build -t app:latest .
docker login
docker push $docker_username/app:latest
```
> **Note:** Edit `k8s/app-deployment.yaml` before deploying.

### Deploying with Kubernetes
- Apply kubectl
```
minikube start
cd k8s
kubectl apply -f postgres-deployment.yaml
kubectl apply -f postgres-service.yaml
kubectl apply -f postgres-pvc.yaml
kubectl apply -f app-deployment.yaml
kubectl apply -f app-service.yaml
kubectl apply -f pgbouncer-all.yaml
```

### Rollout
```
kubectl rollout restart deployment app-deployment
```
### Port Forwarding
```
kubectl port-forward service/app-service 8080:8080 -n default
```

### Checking Pod Status
```
kubectl get pods -n default
kubectl describe pod $NAME -n default
```

## Testing API

> **Register**
```
curl --http2 -X POST https://localhost:8080/v1/auth/register \
     --cacert certs/server.crt \
     -H "Content-Type: application/json" \
     -d '{
           "email": "newuser@example.com",
           "password": "password123",
           "name": "John",
           "surname": "Doe",
           "age": 30
         }'
```

> **login**
```
curl --http2 -X POST https://localhost:8080/v1/auth/login \
     --cacert certs/server.crt \
     -H "Content-Type: application/json" \
     -d '{
           "email": "newuser@example.com",
           "password": "password123"
         }'
```
# Client and Server Examples

## TLS Certificate Generation

Before running your application, generate the TLS certificates to secure gRPC communication. Run the following commands in your project root:

```sh
mkdir certs
openssl req -x509 -newkey rsa:4096 -keyout certs/server.key -out certs/server.crt -days 365 -nodes -subj "/CN=localhost"
```

## Server Example

> Below is an example of a server implementation that registers gRPC services and a gRPC-Gateway for HTTP REST access:
```
// ./server/main.go
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

	sugar.Infof("Serving gRPC with TLS on 0.0.0.0%s", grpcPort)
	go func() {
		log.Fatalln(grpcServer.Serve(lis))
	}()

	// Setup secure connection for gRPC-Gateway.
	// Use "localhost" since the certificate is issued to "localhost".
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
	httpPort := viper.GetString("http.port")
	gwServer := &http.Server{
		Addr:    httpPort,
		Handler: gwmux,
	}

	sugar.Infof("Serving gRPC-Gateway on https://0.0.0.0%s", httpPort)
	log.Fatalln(gwServer.ListenAndServeTLS("../certs/server.crt", "../certs/server.key"))
}
```

## Client Example

> Below is an example client that connects to the gRPC server, registers a new user, logs in, and accesses a protected endpoint:

```
// ./client/main.go
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

### Examples
- [x] https://github.com/Raezil/ProtoText

# References
- [x] https://goprisma.org/docs
- [x] https://protobuf.dev/programming-guides/proto3/
- [x] https://grpc-ecosystem.github.io/grpc-gateway/docs/tutorials/adding_annotations/
