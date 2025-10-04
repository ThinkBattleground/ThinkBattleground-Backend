# Use official Golang image as the base
FROM golang:1.24-alpine

# Install CA certificates for TLS verification
RUN apk --no-cache add ca-certificates && update-ca-certificates

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go app
RUN go build -o main .

# Expose port
EXPOSE 8080

# Start the app
CMD ["./main"]
