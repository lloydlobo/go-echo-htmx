# syntax=docker/dockerfile:1

# Use alpine linux for smaller image size
FROM alpine:3.16

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum for reproducible builds
COPY go.mod go.sum ./

# Download and install Go 1.22.0
RUN apk add --no-cache --virtual build-deps \
    wget \
    gcc \
    libc-dev \
    make \
    && \
    wget -q https://dl.google.com/go/go1.22.osarch.tar.gz -O go.tar.gz \
    && \
    tar -xzf go.tar.gz -C /usr/local \
    && \
    rm go.tar.gz \
    && \
    rm -rf /var/lib/apk/cache/*

# Set environment variables for Go tools
ENV GOPATH /app
ENV GO111MODULE on
ENV GO_TAGS netgo
ENV GO_LDFLAGS '-s -w'

# Copy the rest of the project files
COPY . .

# Install dependencies
RUN go mod download

# Build the application for the current architecture
RUN go build -tags $GO_TAGS -ldflags "$GO_LDFLAGS" -o app cmd/main.go

# (Optional) Build for other architectures (modify as needed)
# RUN GOOS=linux GOARCH=amd64 go build -o app .

# Clean up build dependencies
RUN apk del build-deps

# Expose port 10000 (modify as needed)
EXPOSE 10000

# Run the application
CMD ["app"]

# 
# References:
#   - https://docs.docker.com/language/golang/build-images/
# 

# FROM golang:1.19

# # Set destination for COPY
# WORKDIR /app

# # Download Go modules
# COPY go.mod go.sum ./
# RUN go mod download

# # Copy the source code. Note the slash at the end, as explained in
# # https://docs.docker.com/engine/reference/builder/#copy
# COPY *.go ./

# # Build
# RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-gs-ping

# # Optional:
# # To bind to a TCP port, runtime parameters must be supplied to the docker command.
# # But we can document in the Dockerfile what ports
# # the application is going to listen on by default.
# # https://docs.docker.com/engine/reference/builder/#expose
# EXPOSE 8080

# # Run
# CMD ["/docker-gs-ping"]
