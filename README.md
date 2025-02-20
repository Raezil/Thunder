## Thunder - Backend Framework (gRPC Gateway + Prisma + Kubernetes + Golang)

### Mocking Tests
To mock a gRPC server:
```
cd backend
mockgen -source=yourservice_grpc.pb.go -destination=yourservice_mock.go
```
> **Note:** Replace `yourservice` with the actual name of your gRPC service.

### Adding `protoc` Plugin
```
go build -o protoc-gen-rpc-impl ./cmd/protoc-gen-rpc-impl.go
sudo mv protoc-gen-rpc-impl /usr/local/bin
sudo chmod +x /usr/local/bin/protoc-gen-rpc-impl
```
### Code Generation
```
go run generator.go -proto=filename.proto -prisma=true
```

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

### Register
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

### Login
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
