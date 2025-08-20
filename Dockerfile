

FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the binary from cmd/server
RUN go build -o url-shortener ./cmd/server

FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/url-shortener .

EXPOSE 8080
CMD ["./url-shortener"]
