
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