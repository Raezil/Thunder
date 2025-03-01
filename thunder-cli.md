# Thunder CLI 🚀

The Thunder CLI is a dedicated command‐line tool designed to work with the Thunder backend framework. It provides developers with a streamlined way to generate code, manage configurations, and interact with various Thunder features directly from the terminal.

A custom CLI tool to automate:
- **Generating gRPC and Prisma files** (`thunder generate`)
- **Deploying Kubernetes resources** (`thunder deploy`)

## Installation

### 1. Clone or Download the Repository
If you haven't already, navigate to your project directory where `generator.go` is located.

### 2. Run the Installation Script
Make sure you have **Go**, **Minikube**, and **kubectl** installed.

Run the following command:

```bash
chmod +x install-thunder.sh && ./install-thunder.sh
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
or
```bash
thunder generate --prisma
```

### Generate project
```
thunder new projectname
```
> **Note** replace projectname with actual project name

### Deploy Kubernetes Resources
Before deploying make sure You run those commands:
```
docker build -t app:latest .
docker login
docker push $docker_username/app:latest
```
> **Note** $docker_username is your username, change it in k8s/app-deployment as well

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

## Uninstallation
To remove the CLI:

```bash
sudo rm /usr/local/bin/thunder /usr/local/bin/thunder-generate
```

## Troubleshooting
- If `thunder` is not recognized, make sure `/usr/local/bin/` is in your `$PATH`:
  ```bash
  export PATH=$PATH:/usr/local/bin
  ```
