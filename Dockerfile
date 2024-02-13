# syntax=docker/dockerfile:1

FROM golang:1.22

# Set working directory
WORKDIR /usr/src/app

# Copy go.mod and go.sum for reproducible builds
COPY go.mod go.sum ./

# Set environment variables for Go tools
ENV GO111MODULE on

# Copy the rest of the project files
COPY . .

# Install dependencies
RUN go mod download && go mod verify

# Build the application for the current architecture
RUN CGO_ENABLED=0 GOOS=linux go build -tags netgo -ldflags '-s -w' -o /usr/local/bin/app cmd/main.go

# Expose port 10000 (modify as needed)
EXPOSE 10000

# Run the application
CMD ["/usr/local/bin/app"]

# References:
# 
#   - https://docs.docker.com/language/golang/build-images/
#   - https://hub.docker.com/_/golang