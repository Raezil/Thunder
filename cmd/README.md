# Thunder CLI 🚀

The Thunder CLI is a dedicated command‐line tool designed to work with the Thunder backend framework. It provides developers with a streamlined way to generate code, manage configurations, and interact with various Thunder features directly from the terminal.

A custom CLI tool to automate:
- **Generating gRPC and Prisma files** (`thunder generate`)
- **Deploying Kubernetes resources** (`thunder deploy`)
- **Initializing project** (`thunder init`)
- **Docker** (`thunder build`)
- **Test**: (`thunder test`)

## Installation

### 1. Clone or Download the Repository
If you haven't already, navigate to your project directory where `generator.go` is located.

### 2. Run the Installation Script
Make sure you have **Go**, **Minikube**, and **kubectl** installed.

Run the following command:

```bash
chmod +x install.sh && ./install.sh
```

This script will:
- Compile `generator.go` into the `thunder-generate` binary.
- Move `thunder-generate` and the `thunder` CLI script to `/usr/local/bin/`.
- Make them globally accessible.

## Usage

### Generate gRPC & Prisma Files
```bash
thunder generate --proto yourfile.proto
```

### Test application
```bash
thunder test
```

### Generate project
```
thunder init projectname
```
> **Note** replace projectname with actual project name

### Deploy Kubernetes Resources
Before deploying make sure You run that command:
```
thunder build
```

Congratulations!, Now You can use deploy!
```bash
thunder deploy
```

This command will:
1. Start Minikube.
2. Apply PostgreSQL deployments and services.
3. Apply your app’s Kubernetes deployments and services.
4. Restart PgBouncer and your app deployment.
5. Forward port `8080` to access the application.


## Requirements
- **Go** (for building `thunder-generate`)
- **Minikube** (for Kubernetes)
- **kubectl** (to manage Kubernetes resources)
- **Prisma Client Go** (if using Prisma)
- **Protobuf Compiler (`protoc`)** (if using gRPC)

## Troubleshooting
- If `thunder` is not recognized, make sure `/usr/local/bin/` is in your `$PATH`:
  ```bash
  export PATH=$PATH:/usr/local/bin
  ```
