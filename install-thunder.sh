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
    generate)
        shift
        thunder-generate "$@"
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
        # Apply hpa
        kubectl apply -f hpa.yaml

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
        echo "⚡ Usage: thunder [generate | deploy]"
        exit 1
        ;;
esac
EOF

# Make thunder executable
echo "🔧 Making thunder command executable..."
sudo chmod +x /usr/local/bin/thunder

echo "✅ Installation complete! You can now use 'thunder generate' and 'thunder deploy'."
