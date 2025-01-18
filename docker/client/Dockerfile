FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .

# Build the client binary
RUN CGO_ENABLED=0 go build -trimpath \
    -tags 'netgo osusergo static_build' \
    -ldflags "-s -w" \
    -o /app/build/conf-sync-client \
    cmd/client/main.go

FROM alpine:3.19

# Install basic tools
RUN apk add --no-cache ca-certificates tzdata

# Create config directory
RUN mkdir -p /etc/conf-sync

# Copy the binary
COPY --from=builder /app/build/conf-sync-client /usr/local/bin/conf-sync-client

# Set permissions
RUN chmod +x /usr/local/bin/conf-sync-client

# Create non-root user
RUN adduser -D -H -h /app confuser
USER confuser

ENTRYPOINT ["conf-sync-client"]
CMD ["watch", "--config", "/etc/conf-sync/client.yaml"]
