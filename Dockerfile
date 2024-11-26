# Base image
FROM golang:1.23.2-alpine

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -o main .

# Expose port (adjust as needed for your application)
EXPOSE 8080

# Command to run the application
CMD ["./main"]