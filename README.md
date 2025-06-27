<p align="center">
  <img src="https://github.com/user-attachments/assets/c98e19fd-ebf6-4dca-8a9c-ae3d82e3ee54" alt="centered image">
</p>

# **Thunder- A Minimalist Backend Framework in Go**


*A scalable microservices framework powered by Go, gRPC-Gateway, Prisma, and Kubernetes. It exposes REST, gRPC and Graphql*

[![libs.tech recommends](https://libs.tech/project/882664523/badge.svg)](https://libs.tech/project/882664523/thunder)
[![Go Version](https://img.shields.io/badge/Go-1.23-blue)](https://golang.org)
[![License](https://img.shields.io/github/license/Raezil/Thunder)](LICENSE)
[![Stars](https://img.shields.io/github/stars/Raezil/Thunder)](https://github.com/Raezil/Thunder/stargazers)

## **🚀 Features**
- ✔️ **gRPC + REST (gRPC-Gateway)** – Automatically expose RESTful APIs from gRPC services.
- ✔️ **Prisma Integration** – Efficient database management and migrations.
- ✔️ **Kubernetes Ready** – Easily deploy and scale with Kubernetes.
- ✔️ **TLS Security** – Secure gRPC communications with TLS.
- ✔️ **Structured Logging** – Built-in `zap` logging.
- ✔️ **Rate Limiting & Authentication** – Pre-configured middleware.
- ✔️ **Modular & Extensible** – Easily extend Thunder for custom use cases.
- ✔️ **Thunder CLI** - Generate, deploy, and create new projects effortlessly.
- ✔️ **Graphql support** - Transform grpc services into graphql queries

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
chmod +x install.sh
./install.sh
```
> Remember to install prerequisites, there is tutorial for this https://github.com/Raezil/Thunder/issues/99

### **Setup**
Create a new Thunder application:
```bash
thunder init myapp
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
import "graphql.proto";

service Example {
	rpc SayHello(HelloRequest) returns (HelloResponse) {
		option (google.api.http) = {
			get: "/v1/example/sayhello"
		};
    		option (graphql.schema) = {
      			type: QUERY   // declare as Query
      			name: "sayhello" // query name
    		};
	};
}

message HelloRequest {
	string name = 1 [(graphql.field) = {required: true}];
}

message HelloResponse {
	string message = 1;
}
```

### 🔨 Generate a Service Scaffold

Use the new `scaffold` command to spin up a full CRUD `.proto` file—complete with gRPC, REST (gRPC-Gateway) and GraphQL annotations. Pass your fields as a comma-separated list of `name:type` pairs:

```bash
thunder scaffold   -service UserService   -entity User   -fields "id:string,name:string,email:string,age:int32"
```

Add your service entry in `services.json`:
```go
[
    {
      "ServiceName": "Example",
      "ServiceStruct": "ExampleServiceServer",
      "ServiceRegister": "RegisterExampleServer",
      "HandlerRegister": "RegisterExampleHandler"
      "GraphqlHandlerRegister": "RegisterExampleGraphqlHandler"

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
thunder generate --proto=example.proto --graphql=true
```

## **🚀 Running the Server**

Start the server:
```bash
go run ./cmd/app/server/main.go
```

Server accessible via HTTP at `localhost:8080` and gRPC at `localhost:50051`.

## **🚀 Running the Tests**

### Mocking Tests
```bash
cd pkg/services/generated
mockgen -source=yourservice_grpc.pb.go -destination=./yourservice_mock.go
```

### Run Tests
```bash
go test ./pkg/db ./pkg/middlewares/ ./pkg/services/ ./pkg/services/generated
```

## **🔧 Kubernetes Deployment**
### PgBouncer Configuration

This setup configures PgBouncer to connect to a PostgreSQL database using Kubernetes resources.

### Updating the `userlist.txt` Secret

To regenerate and update the `userlist.txt` secret, use the following command to encode the credentials:

```bash
echo '"postgres" "postgres"' | base64
```

Now, update `pgbouncer-all.yaml` under the `Secret` section with the new base64-encoded value:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: pgbouncer-secret
type: Opaque
data:
  userlist.txt: <BASE64_ENCODED_VALUE>  # "postgres" "postgres" in base64
```

### Generate TLS Certificates
```bash
cd cmd
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
thunder build
thunder deploy
```

Check pod status:
```bash
kubectl get pods -n default
kubectl describe pod $NAME -n default
```

## **📡 API Testing**

### Register User
#### REST
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
#### Graphql
```bash
curl -k -X POST https://localhost:8080/graphql \
     -H "Content-Type: application/json" \
     -d '{"query":"mutation{register(email:\"newuser1211@example.com\",password:\"password123\",name:\"John\",surname:\"Doe\",age:30){reply}}"}'
```

### User Login
#### REST
```bash
curl -k --http2 -X POST https://localhost:8080/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{
           "email": "newuser@example.com",
           "password": "password123"
         }'
```
#### Graphql
```
curl -k -X POST https://localhost:8080/graphql \
     -H "Content-Type: application/json" \
     -d '{"query":"query Login($email:String!,$password:String!){login(email:$email,password:$password){token}}","variables":{"email":"newuser@example.com","password":"password123"}}'
```

### Sample protected
#### REST
```bash
curl -k -X GET "https://localhost:8080/v1/auth/protected?text=hello" \
  -H "Authorization: Bearer $token"
```
> $token is returned by login
#### Graphql
```bash
curl -k -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $token" \
  -d '{
    "query": "query { protected(text: \"Hello World\") { result } }"
  }' \
  https://localhost:8080/graphql
```

## **📜 Contributing**

1. Fork the repository.
2. Create a feature branch: `git checkout -b feature-new`
3. Commit changes: `git commit -m "Added feature"`
4. Push to your branch: `git push origin feature-new`
5. Submit a pull request.

## **🔗 References**
- [Go Documentation](https://golang.org/doc/)
- [gRPC](https://grpc.io/docs/languages/go/quickstart/)
- [gRPC-Gateway](https://grpc-ecosystem.github.io/grpc-gateway/)
- [Prisma ORM](https://www.prisma.io/docs/)
- [Kubernetes Docs](https://kubernetes.io/docs/)
- [Tutorial](https://gist.github.com/Raezil/f649ae8c5201f60d479ed796299af679)

## **📣 Stay Connected**
⭐ Star the repository if you find it useful!  
📧 For support, use [GitHub Issues](https://github.com/Raezil/Thunder/issues).

## **License**
Thunder is released under the MIT License.
