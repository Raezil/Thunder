#!/bin/bash


set -e  # Exit immediately if a command exits with a non-zero status

# Build the thunder-generate binary
echo "⚙️  Building thunder-generate..."
go build -o thunder-generate generator.go

# Move thunder-generate to /usr/local/bin/
echo "🚀 Moving thunder-generate to /usr/local/bin/..."
sudo mv thunder-generate /usr/local/bin/

# Create the thunder command script
echo "🛠️ Creating thunder command script..."
cat << 'EOF' | sudo tee /usr/local/bin/thunder > /dev/null
#!/bin/bash

case "$1" in
    new)
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
    docker)
        echo "Enter your Docker Hub username:"
        read docker_username
        echo "🔨 Building Docker image..."
        docker build -t ${docker_username}/app:latest .
        echo "🔑 Logging in to Docker Hub..."
        docker login
        echo "⬆️  Pushing Docker image..."
        docker push ${docker_username}/app:latest
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
    *)
        echo "⚡ Usage: $0 [new | generate | deploy]"
        exit 1
        ;;
esac
EOF

# Make thunder executable
echo "🔧 Making thunder command executable..."
sudo chmod +x /usr/local/bin/thunder

echo "✅ Installation complete! You can now use 'thunder generate' and 'thunder deploy'."
