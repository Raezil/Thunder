## Thunder Framework

Thunder is a minimalistic backend framework built with Go, gRPC-Gateway, and Prisma, designed for simplicity, scalability, and efficiency.

### Installation

Clone the repository:
```bash
git clone https://github.com/Raezil/Thunder.git
cd Thunder
```

Install CLI by running the installation script:
```bash
chmod +x install-thunder.sh
./install-thunder.sh
```

### Setup

Create a new Thunder application using the CLI:
```bash
thunder new myapp
cd myapp
```

### Examples

#### Basic Server Initialization

Start your Thunder server:

```bash
go run main.go
```

#### Defining a gRPC Service

Generate a new gRPC service with automatic implementation:

```bash
thunder generate --proto=example.proto
```

Example proto file (`example.proto`):
```proto
syntax = "proto3";

package example;

import "google/api/annotations.proto";

service ExampleService {
	rpc SayHello(HelloRequest) returns (HelloResponse) {
		option (google.api.http) = {
			post: "/v1/example/sayhello"
			body: "*"
		};
	};
}

message HelloRequest {
	string name = 1;
}

message HelloResponse {
	string message = 1;
}
```

This command generates the Go implementation automatically, add logic to yours backend/example_server.go. You only need to add your service entry to `routes/route.go` before you generate:

```go
package routes

var Services = []Service{
	{
		ServiceName:     "Auth",
		ServiceStruct:   "AuthServiceServer",
		ServiceRegister: "RegisterAuthServer",
		HandlerRegister: "RegisterAuthHandler",
	},
	{
		ServiceName:     "Example",
		ServiceStruct:   "ExampleServiceServer",
		ServiceRegister: "RegisterExampleServer",
		HandlerRegister: "RegisterExampleHandler",
	},
}
```

### Running Your Server

```bash
go run ./server/main.go
```

Your Thunder application is accessible via HTTP at `localhost:8080` and gRPC at `localhost:50051`.

### License

Thunder is released under the MIT License.
