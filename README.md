## Thunder - backend Framework gRPC Gateway + Prisma + Kubernetes + Golang

## Generator
### Add protoc plugin
```
go build -o protoc-gen-rpc-impl ./cmd/protoc-gen-rpc-impl.go
sudo mv protoc-gen-rpc-impl /usr/local/bin
sudo chmod +x /usr/local/bin/protoc-gen-rpc-impl
```
```
go run generator.go yourfilename.proto
```
## Kubernetes
### Run Docker
```
docker build -t app:latest .
docker login
docker push $docker_username/app:latest
```
- [x] edit k8s/deployment.yaml
- Apply kubectl
```
minikube start
cd k8s
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
```
- Port Foward
```
kubectl port-forward service/app-service 8080:8080 -n default
```
- Check pods
```
kubectl get pods -n default
kubectl describe pod $NAME -n default
```
## Testing API
Register:
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
Log in:
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
