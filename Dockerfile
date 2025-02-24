# Use the official Go image for both building and running
FROM golang:1.22-alpine

# Install necessary packages
RUN apk add --no-cache git
ENV DATABASE_URL="postgresql://postgres:postgres@localhost:5432/Thunder?connection_limit=5"
# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire project
COPY . .

# Remove the unnecessary _gen.go file to avoid conflicts
RUN rm -f db/query-engine-debian-openssl-3.0.x_gen.go

# Install prisma-client-go
RUN go install github.com/steebchen/prisma-client-go@latest

# Add Go binaries to PATH
ENV PATH=$PATH:/go/bin

RUN PGPASSWORD=postgres psql -h localhost -U postgres -p 5432 -c "CREATE DATABASE Thunder;"
# Set environment variables if needed (e.g., DATABASE_URL)
# ARG DATABASE_URL
# ENV DATABASE_URL=${DATABASE_URL}

# Run Prisma db push
RUN prisma-client-go db push

# Expose necessary ports
EXPOSE 50051 8080

# Run the application using go run
CMD ["go", "run", "./server/main.go"]
