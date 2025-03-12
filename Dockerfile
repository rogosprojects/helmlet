# Use the official Golang image as the base image
FROM golang:1.23-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

COPY . .

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Build the Go app with optimization flags
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s -X main.Version=docker" -o main .


FROM alpine:latest

# Set the Current Working Directory inside the container
WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/main /usr/local/bin/helmlet


# Command to run the executable
CMD ["./main"]