# BikeRental-GRPC

Sample biking rental service using grpc.

## Kubernetes

### Run Docker
```
docker build -t app:latest .
docker login
docker push $docker_username/app:latest
```

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
## Sample curl requests
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
