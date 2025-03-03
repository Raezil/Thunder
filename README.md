# **Thunder - A Minimalist Backend Framework in Go**
*A gRPC-Gateway-powered framework with Prisma, Kubernetes and Go for scalable microservices.*

[![Go Version](https://img.shields.io/badge/Go-1.21-blue)](https://golang.org)
[![License](https://img.shields.io/github/license/Raezil/Thunder)](LICENSE)
[![Stars](https://img.shields.io/github/stars/Raezil/Thunder)](https://github.com/Raezil/Thunder/stargazers)
[![Issues](https://img.shields.io/github/issues/Raezil/Thunder)](https://github.com/Raezil/Thunder/issues)

## **Table of Contents**
- [🚀 Features](#-features)
- [📌 Getting Started](#-getting-started)
- - [⚡ Thunder CLI](#thunder-cli)
  - [1️⃣ Install Dependencies](#1️⃣-install-dependencies)
  - [2️⃣ Define Your gRPC Service](#2️⃣-define-your-grpc-service)
- [🛠️ Prisma Integration](#️-prisma-integration)
- [🚀 Running the Server](#-running-the-server)
  - [a. Code Generation](#a-code-generation)
  - [b. Start the **gRPC + REST API** server](#b-start-the-grpc--rest-api-server)
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
✔️ **Thunder CLI** - generate, deploy, create new project by using dedicated CLI.  

---

## **📌 Getting Started**

### **Thunder CLI**
For a comprehensive guide on how to use Thunder CLI—including installation steps, available commands, and usage examples—you can refer to the official documentation here:
https://github.com/Raezil/Thunder/blob/main/thunder-cli.md

This file covers everything you need to get started with Thunder CLI and will help you integrate it into your development workflow.


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

## a. Create Your schema.prisma File
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

#### a. Code Generation
```
thunder generate -proto=filename.proto -prisma=true
```
> **Note:** Replace `filename` with the actual name of your gRPC service.
> **Note** Remember to install [ Thunder CLI](#thunder-cli)

#### b. Start the **gRPC + REST API** server:

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
thunder docker
```

### **3️⃣ Deploy to Kubernetes**
```sh
thunder deploy
```
**Note** Remember to install [ Thunder CLI](#thunder-cli)

#### Checking Pod Status
```
kubectl get pods -n default
kubectl describe pod $NAME -n default
```

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
