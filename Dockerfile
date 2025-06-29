# Stage 1: Build
FROM golang:alpine AS builder

# Set the working directory
WORKDIR /app

# Copy project files to the container
COPY . .

# Build the executable
RUN go build -o /app/raidark .

# Stage 2: Runtime
FROM alpine:latest

# Create the directory where the executable will reside
WORKDIR /app

# Copy the executable from the build stage
COPY --from=builder /app/raidark /app/raidark

EXPOSE 8080
ENV LOG_LEVEL=INFO
ENV API_PORT=8080

# Set the entry command
ENTRYPOINT ["/app/raidark", "api"]
