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
     curl --http2 -X POST http://localhost:8080/v1/auth/register \
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
     curl --http2 -X POST http://localhost:8080/v1/auth/login \
          -H "Content-Type: application/json" \
          -d '{
                "email": "user@example.com",
                "password": "password123"
              }'

```

### Examples
- [x] https://github.com/Raezil/ProtoText

# References
- [x] https://goprisma.org/docs
- [x] https://protobuf.dev/programming-guides/proto3/
- [x] https://grpc-ecosystem.github.io/grpc-gateway/docs/tutorials/adding_annotations/
