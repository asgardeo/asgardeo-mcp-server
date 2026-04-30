# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o asgardeo-mcp-server .

# Runtime stage
FROM alpine:3.21

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/asgardeo-mcp-server .

# Default to SSE mode for containerized deployments.
# SERVER_BASE_URL must be set at runtime to the server's public URL
# (e.g. https://your-host.com) — required for OAuth callback registration.
ENV TRANSPORT_MODE=sse
ENV SERVER_PORT=8080

EXPOSE 8080

ENTRYPOINT ["./asgardeo-mcp-server"]
