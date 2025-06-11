# Build stage
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install git (needed for go mod download)
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with static linking
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o main .

# Runtime stage - using scratch for minimal image
FROM scratch

# Copy the statically linked binary
COPY --from=builder /app/main /main

# Copy content files (includes both .md and .css files)
COPY content/ /content/

# Expose port
EXPOSE 8080

# Set environment variables
ENV PORT=8080
ENV CONTENT_DIR=/content

# Run the application
ENTRYPOINT ["/main"]