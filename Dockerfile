# Use the official Golang image as the base for building
FROM golang:1.23.0-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN go build -o stock-quant-reports .

# Use a lightweight alpine image for the final runtime container
FROM alpine:latest  

# Add CA certificates and tzdata for secure calls and correct timezone (NY)
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/stock-quant-reports .

# Tell Docker that the container listens on the specified port
EXPOSE 8080

# Command to run the executable
CMD ["./stock-quant-reports"]
