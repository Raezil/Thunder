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

# The database URL will come from environment (docker-compose.yml)
# but you can set a default if you wish
ENV DATABASE_URL="postgresql://postgres:postgres@postgres:5432/thunder?connection_limit=5"

COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

# Expose ports (whatever your app uses)
EXPOSE 8080 50051

# By default, run our custom entrypoint which creates the DB, migrates, and then runs the app
ENTRYPOINT ["/app/entrypoint.sh"]
