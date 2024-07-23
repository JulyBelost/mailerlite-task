# Use the official Golang image as the base image
FROM golang:1.22 AS builder

# Set the working directory inside the container
WORKDIR /workspace

# Copy the Go module files and download dependencies
COPY email-operator/go.mod ./
RUN go mod download

# Copy the remaining source code
COPY email-operator/ .

# Build the operator binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager main.go

# Use a minimal base image
FROM alpine:3.13

# Set the working directory
WORKDIR /

# Copy the binary from the builder stage
COPY --from=builder /workspace/manager .

# Run the binary
ENTRYPOINT ["/manager"]
