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
> **Note:** Edit `k8s/deployment.yaml` before deploying.

### Deploying with Kubernetes
- Apply kubectl
```
minikube start
cd k8s
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
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
	pb "backend"
	"context"
	"db"
	"log"
	"net"
	"net/http"

	"middlewares"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func initConfig() {
	viper.SetDefault("grpc.port", ":50051")
	viper.SetDefault("http.port", ":8080")
	// Load environment variables
	viper.AutomaticEnv()
}

func RegisterServers(server *grpc.Server, client *db.PrismaClient, sugar *zap.SugaredLogger) {
	pb.RegisterAuthServer(server, &pb.AuthenticatorServer{
		PrismaClient: client,
		Logger:       sugar,
	})
}

func RegisterHandlers(gwmux *runtime.ServeMux, conn *grpc.ClientConn) {
	err := pb.RegisterAuthHandler(context.Background(), gwmux, conn)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}
}

func main() {
	initConfig()
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}

	sugar := logger.Sugar()
	defer logger.Sync()
	grpcPort := viper.GetString("grpc.port")
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}
	client := db.NewClient()
	if err := client.Prisma.Connect(); err != nil {
		panic(err)
	}

	defer func() {
		if err := client.Prisma.Disconnect(); err != nil {
			panic(err)
		}
	}()
	// Initialize rate limiter (e.g., 5 requests per second, burst of 10)
	rateLimiter := middlewares.NewRateLimiter(5, 10)

	// Use custom interceptor chain
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(
			middlewares.ChainUnaryInterceptors(
				rateLimiter.RateLimiterInterceptor, // Rate limiting
				middlewares.AuthUnaryInterceptor,   // Authentication
			),
		),
	)
	RegisterServers(grpcServer, client, sugar)

	log.Println("Serving gRPC on 0.0.0.0" + grpcPort)
	go func() {
		log.Fatalln(grpcServer.Serve(lis))
	}()

	conn, err := grpc.NewClient(
		"0.0.0.0:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}

	gwmux := runtime.NewServeMux()
	RegisterHandlers(gwmux, conn)
	httpPort := viper.GetString("http.port")
	gwServer := &http.Server{
		Addr:    httpPort,
		Handler: gwmux,
	}

	log.Println("Serving gRPC-Gateway on http://0.0.0.0:8080")
	log.Fatalln(gwServer.ListenAndServe())
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
```

### Examples
- [x] https://github.com/Raezil/ProtoText

# References
- [x] https://goprisma.org/docs
- [x] https://protobuf.dev/programming-guides/proto3/
- [x] https://grpc-ecosystem.github.io/grpc-gateway/docs/tutorials/adding_annotations/
