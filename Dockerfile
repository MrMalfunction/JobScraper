FROM golang:1.25-alpine

# Install necessary packages
RUN apk add --no-cache git

# Install Air
RUN go install github.com/air-verse/air@latest

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy air configuration
COPY .air.toml ./

# Copy config.yaml for DB config
# COPY config.yaml ./
# COPY .secrets.yaml ./

# The source code will be mounted as a volume
# So we don't copy it here

EXPOSE 8080

# Run Air for live reload
CMD ["air", "-c", ".air.toml", "-d", "-build.cmd", "go build -buildvcs=false -o ./tmp/main ."]
