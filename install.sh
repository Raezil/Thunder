#!/bin/bash

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check for required dependencies
dependencies=( "go" "git" "docker" "minikube" "kubectl" "sudo" )
missing=()

for dep in "${dependencies[@]}"; do
    if ! command_exists "$dep"; then
        missing+=("$dep")
    fi
done

if [ ${#missing[@]} -ne 0 ]; then
    echo "❌ The following dependencies are missing: ${missing[*]}"
    echo "Please install them before running this script."
    exit 1
fi

# Build the thunder-generate binary
echo "⚙️  Building thunder-generate..."
go build -o ./cmd/protoc-gen-rpc-impl ./cmd/protoc-gen-rpc-impl.go
sudo mv ./cmd/protoc-gen-rpc-impl /usr/local/bin
sudo chmod +x /usr/local/bin/protoc-gen-rpc-impl
go build -o thunder-generate generator.go

# Move thunder-generate to /usr/local/bin/
echo "🚀 Moving thunder-generate to /usr/local/bin/..."
sudo mv thunder-generate /usr/local/bin/

# Create the thunder command script
echo "🛠️ Creating thunder command script..."
cat << 'EOF' | sudo tee /usr/local/bin/thunder > /dev/null
#!/bin/bash

case "$1" in
    init)
        shift
        # Use an optional directory name; default to "Thunder"
        TARGET_DIR="${1:-Thunder}"
        echo "Cloning Thunder repository into '${TARGET_DIR}'..."
        git clone https://github.com/Raezil/Thunder "$TARGET_DIR" || { echo "❌ Error: Cloning failed."; exit 1; }
        echo "Removing .git folder from '${TARGET_DIR}'..."
        rm -rf "$TARGET_DIR/.git" || { echo "❌ Error: Could not remove .git folder."; exit 1; }
        echo "Repository cloned to '${TARGET_DIR}' with git history removed."
        ;;
    generate)
        shift
        thunder-generate "$@"
        ;;
    build)
        DEPLOYMENT_FILE="./k8s/app-deployment.yaml"

        # Verify the deployment file exists.
        if [ ! -f "$DEPLOYMENT_FILE" ]; then
            echo "Error: $DEPLOYMENT_FILE not found."
            exit 1
        fi
        echo "Enter your Docker Hub username:"
        read docker_username
        echo "Enter your Docker project name:"
        read docker_project
        echo "🔨 Building Docker image..."
        NEW_IMAGE="${docker_username}/${docker_project}:latest"
        sed -i'' -E '/busybox/! s#^([[:space:]]*image:[[:space:]])[^[:space:]]+#\1'"${NEW_IMAGE}"'#' k8s/app-deployment.yaml
        docker build -t ${docker_username}/${docker_project}:latest .
        echo "🔑 Logging in to Docker Hub..."
        docker login
        echo "⬆️  Pushing Docker image..."
        docker push ${docker_username}/${docker_project}:latest
        ;;
    deploy)
        echo "🚀 Starting Minikube..."
        minikube start

        # Change to Kubernetes manifests directory
        cd k8s || { echo "❌ Directory k8s not found!"; exit 1; }

        # Apply PostgreSQL resources
        echo "📦 Deploying PostgreSQL..."
        kubectl apply -f postgres-deployment.yaml
        kubectl apply -f postgres-service.yaml
        kubectl apply -f postgres-pvc.yaml

        # Wait for PostgreSQL to be ready
        echo "⏳ Waiting for PostgreSQL to be ready..."
        kubectl wait --for=condition=ready pod -l app=postgres --timeout=60s

        # Apply application deployments and services
        echo "⚙️ Deploying Thunder API..."
        kubectl apply -f app-deployment.yaml
        kubectl apply -f app-service.yaml
        kubectl apply -f app-loadbalancer.yaml
        # Apply HPA configuration
        kubectl apply -f hpa.yaml
        kubectl apply -f network-policy.yaml
        # Apply PgBouncer for database connection pooling
        echo "🔄 Deploying PgBouncer..."
        kubectl apply -f pgbouncer-all.yaml

        # Restart necessary deployments
        echo "🔄 Restarting PgBouncer and Thunder API deployments..."
        kubectl rollout restart deployment pgbouncer
        kubectl rollout restart deployment app-deployment

        # Port forward the app service
        echo "🔗 Forwarding port 8080 to app-service..."
        kubectl port-forward service/app-service 8080:8080 &
        ;;
    test)
        echo "Running tests..."
        go test -v ./pkg/db ./pkg/middlewares/ ./pkg/services/ ./pkg/services/generated
        exit 0
        ;;
    *)
        echo "⚡ Usage: $0 [init | docker | generate | deploy | test]"
        exit 1
        ;;
esac
EOF

# Make thunder executable
echo "🔧 Making thunder command executable..."
sudo chmod +x /usr/local/bin/thunder

echo "✅ Installation complete! You can now use 'thunder init', 'thunder build', 'thunder generate' and 'thunder deploy'."
