# Use the official Golang image to create a build artifact.
# This is based on Debian and sets the GOPATH to /go.
# https://hub.docker.com/_/golang
FROM golang:1.18.6 as builder
ENV GO111MODULE=on
# Create and change to the app directory.
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the binary.
RUN CGO_ENABLED=0 go build -v -o cli .

FROM alpine:3.13.5 as deploy
# Copy the binary to the production image from the builder stage.
RUN apk update && apk add ca-certificates
COPY --from=builder /app/cli .
# Run the web service on container startup.
CMD ["./cli"]
