<p align="center">
  <img src="https://github.com/user-attachments/assets/15766a1d-3ec2-4750-bd89-969d031ef8fe" alt="centered image">
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
![Screenshot from 2025-06-28 17-32-47](https://github.com/user-attachments/assets/97c0c576-4026-4633-b2bc-ec57d81834fd)

## **📌 Use Cases**

Thunder is designed for **scalable microservices** and **high-performance API development**, particularly suited for:
### 1. High-Performance API Development
- **gRPC-first APIs** with RESTful interfaces via gRPC-Gateway  
- **GraphQL interfaces** via grpc-graphql-gateway for strongly-typed, schema-driven queries  
- Critical performance and low-latency applications  
- Strongly-typed APIs defined in Protobuf  

### 2. Microservices Architecture
- Efficient inter-service communication with gRPC  
- Kubernetes deployments with built-in service discovery and scaling  
- Sidecar-friendly design for Istio/Linkerd environments  

### 3. Database Management with Prisma
- Type-safe queries and zero-downtime migrations  
- Multi-database support (PostgreSQL, MySQL, SQLite)  
- Auto-generated client library for Go  

### 4. Lightweight Backend Alternative
- Minimalist, modular core—no unnecessary middleware  
- Fast startup and small binary footprint  
- Pluggable extensions (auth, metrics, tracing)  

### 5. Kubernetes & Cloud-Native Applications
- Containerized with Docker, optimized for multi-container pods  
- Automatic horizontal scaling and load balancing  
- Health checks, liveness/readiness probes built in  

---

## When **Not** to Use Thunder
- If you need a traditional REST-only API (consider Gin, Fiber, or Echo)  
- If you require a feature-heavy, middleware-rich framework  
- If you’re deploying a monolithic service outside of Kubernetes  


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

### Stream Protected
### REST
```bash
wscat --no-check   -c "wss://localhost:8080/v1/auth/stream/protected?method=GET&text=hello"   -s Bearer   -s "$TOKEN"
```
### Graphql
```bash
NODE_TLS_REJECT_UNAUTHORIZED=0 wscat -c wss://localhost:8080/graphql       -H "Authorization: Bearer $TOKEN"       -s graphql-ws
```
```bash
{"id":"1","type":"start","payload":{"query":"subscription { stream(text: \"hello\") { result } }","variables":{}}}

```

### Graphql
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
