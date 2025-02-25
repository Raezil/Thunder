# Use the official Go image
FROM golang:1.22-alpine

# Install git (needed for 'go get' in some cases)
RUN apk add --no-cache git

# Create an app directory
WORKDIR /app

# Copy module files first
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of your code
COPY . .

# (Optional) Remove the unnecessary file if it exists
RUN rm -f db/query-engine-debian-openssl-3.0.x_gen.go

# Install prisma-client-go
RUN go install github.com/steebchen/prisma-client-go@latest

# Add Go binaries to PATH
ENV PATH=$PATH:/go/bin

# Set the database URL from the environment or default
ENV DATABASE_URL="postgres://postgres:postgres@pgbouncer-service:6432/thunder?sslmode=disable"
ENV JWT_SECRET="secret"

# Expose necessary ports
EXPOSE 50051 8080

# Copy the entrypoint script
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

# Copy the certificates directory
COPY certs ../certs

# Set the entrypoint and default command
ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["go", "run", "./server/main.go"]
