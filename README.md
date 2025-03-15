# **Thunder - A Minimalist Backend Framework in Go**
*A gRPC-Gateway-powered framework with Prisma, Kubernetes, and Go for scalable microservices.*

[![libs.tech recommends](https://libs.tech/project/882664523/badge.svg)](https://libs.tech/project/882664523/thunder)
[![Go Version](https://img.shields.io/badge/Go-1.23-blue)](https://golang.org)
[![License](https://img.shields.io/github/license/Raezil/Thunder)](LICENSE)
[![Stars](https://img.shields.io/github/stars/Raezil/Thunder)](https://github.com/Raezil/Thunder/stargazers)
[![Issues](https://img.shields.io/github/issues/Raezil/Thunder)](https://github.com/Raezil/Thunder/issues)

## **🚀 Features**
- ✔️ **gRPC + REST (gRPC-Gateway)** – Automatically expose RESTful APIs from gRPC services.
- ✔️ **Prisma Integration** – Efficient database management and migrations.
- ✔️ **Kubernetes Ready** – Easily deploy and scale with Kubernetes.
- ✔️ **TLS Security** – Secure gRPC communications with TLS.
- ✔️ **Structured Logging** – Built-in `zap` logging.
- ✔️ **Rate Limiting & Authentication** – Pre-configured middleware.
- ✔️ **Modular & Extensible** – Easily extend Thunder for custom use cases.
- ✔️ **Thunder CLI** - Generate, deploy, and create new projects effortlessly.

## **🏗️ Architecture Overview**
![421386849-54a1cead-6886-400a-a41a-f5eb4f375dc7(1)](https://github.com/user-attachments/assets/5074e533-b023-415d-9092-e8f5270ec88f)

## **📌 Use Cases**

Thunder is designed for **scalable microservices** and **high-performance API development**, particularly suited for:

### **1. High-Performance API Development**
- gRPC-first APIs with RESTful interfaces via gRPC-Gateway.
- Critical performance and low latency applications.
- Strongly-typed APIs with protobufs.

### **2. Microservices Architecture**
- Efficient inter-service communication.
- Kubernetes deployments with built-in service discovery and scaling.

### **3. Database Management with Prisma**
- Type-safe queries and easy database migrations.
- Support for multiple databases (PostgreSQL, MySQL, SQLite).

### **4. Lightweight Backend Alternative**
- A minimalist and powerful alternative to traditional frameworks like Gin or Echo.
- Fast, simple, and modular backend without unnecessary overhead.

### **5. Kubernetes & Cloud-Native Applications**
- Containerized environments using Docker.
- Automatic service scaling and load balancing.

### **When Not to Use Thunder**
- If you need a traditional REST-only API (use Gin, Fiber, or Echo instead).
- If you require a feature-heavy web framework with extensive middleware.
- If you're not deploying on Kubernetes or prefer a monolithic backend.

## **📌 Getting Started**

### **Installation**
```bash
git clone https://github.com/Raezil/Thunder.git
cd Thunder
chmod +x install-thunder.sh
./install-thunder.sh
```

### **Setup**
Create a new Thunder application:
```bash
thunder new myapp
cd myapp
```

### **Install Dependencies**
```bash
go mod tidy
```

### **Define Your gRPC Service**
Create a `.proto` file (e.g., `example.proto`):

```proto
syntax = "proto3";

package example;

import "google/api/annotations.proto";

service Example {
	rpc SayHello(HelloRequest) returns (HelloResponse) {
		option (google.api.http) = {
			post: "/v1/example/sayhello"
			body: "*"
		};
	};
}

message HelloRequest {
	string name = 1;
}

message HelloResponse {
	string message = 1;
}
```

Add your service entry in `services.json`:
```go
[
    {
      "ServiceName": "Example",
      "ServiceStruct": "ExampleServiceServer",
      "ServiceRegister": "RegisterExampleServer",
      "HandlerRegister": "RegisterExampleHandler"
    }
]
```

## **🛠️ Prisma Integration**
Define your schema in `schema.prisma`:

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

Generate the service implementation:
```bash
thunder generate --proto=example.proto
```

## **🚀 Running the Server**

Start the server:
```bash
go run ./pkg/app/server/main.go
```

Server accessible via HTTP at `localhost:8080` and gRPC at `localhost:50051`.

## **🚀 Running the Tests**

### Mocking Tests
```bash
cd pkg/services
mockgen -source=yourservice_grpc.pb.go -destination=../yourservice_mock.go
```

### Run Tests
```bash
go test ./pkg/db ./pkg/middlewares/ ./pkg/services/ ./pkg/services/generated
```

## **🔧 Kubernetes Deployment**

### Generate TLS Certificates
```bash
cd pkg
mkdir certs
openssl req -x509 -newkey rsa:4096 -keyout certs/server.key -out certs/server.crt -days 365 -nodes \
  -subj "/CN=localhost" \
  -addext "subjectAltName=DNS:localhost,IP:127.0.0.1"
```

### Generate Kubernetes Secrets
```bash
kubectl create secret generic app-secret   --from-literal=DATABASE_URL="postgres://postgres:postgres@pgbouncer-service:6432/thunder?sslmode=disable"   --from-literal=JWT_SECRET="secret"

kubectl create secret generic postgres-secret   --from-literal=POSTGRES_USER=postgres   --from-literal=POSTGRES_PASSWORD=postgres   --from-literal=POSTGRES_DB=thunder
```

### Build & Deploy Docker Image
```bash
thunder docker
thunder deploy
```

Check pod status:
```bash
kubectl get pods -n default
kubectl describe pod $NAME -n default
```

## **📡 API Testing**

### Register User
```bash
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

### User Login
```bash
curl -k --http2 -X POST https://localhost:8080/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{
           "email": "newuser@example.com",
           "password": "password123"
         }'
```

## **📜 Contributing**

1. Fork the repository.
2. Create a feature branch: `git checkout -b feature-new`
3. Commit changes: `git commit -m "Added feature"`
4. Push to your branch: `git push origin feature-new`
5. Submit a pull request.

## **🔗 References**
- [Go Documentation](https://golang.org/doc/)
- [gRPC-Gateway](https://grpc-ecosystem.github.io/grpc-gateway/)
- [Prisma ORM](https://www.prisma.io/docs/)
- [Kubernetes Docs](https://kubernetes.io/docs/)

## **📣 Stay Connected**
⭐ Star the repository if you find it useful!  
📧 For support, use [GitHub Issues](https://github.com/Raezil/Thunder/issues).

## **License**
Thunder is released under the MIT License.
