FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .

# Build the server binary
RUN CGO_ENABLED=0 go build -trimpath \
    -tags 'netgo osusergo static_build' \
    -ldflags "-s -w" \
    -o /app/build/conf-sync-server \
    cmd/server/main.go

FROM alpine:3.19

# Install basic tools
RUN apk add --no-cache ca-certificates tzdata

# Copy the binary
COPY --from=builder /app/build/conf-sync-server /usr/local/bin/conf-sync-server

# Set permissions
RUN chmod +x /usr/local/bin/conf-sync-server

# Create non-root user
RUN adduser -D -H -h /app confuser
USER confuser

ENTRYPOINT ["conf-sync-server"]
