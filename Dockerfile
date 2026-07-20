# Stage 1: Build
FROM golang:alpine
WORKDIR /app

# Install git and iproute2 for building/networking
RUN apk add --no-cache git iproute2

# Copy all source and modules
COPY . .

# Generate a clean go.sum and download dependencies
RUN go mod tidy

# Build the VDR binary
RUN CGO_ENABLED=0 GOOS=linux go build -o vdr ./cmd/vdr/main.go

# Stage 2: Minimal Runtime
FROM alpine:latest
WORKDIR /app

# Install necessary network tools for TUN/TAP testing (iproute2)
RUN apk add --no-cache ca-certificates tzdata iproute2 tcpdump iptables iputils

# Copy the binary and entrypoint script from builder
COPY --from=builder /app/vdr .
COPY --from=builder /app/entrypoint.sh .

# Make entrypoint executable
RUN chmod +x ./entrypoint.sh

# Execute the entrypoint script
ENTRYPOINT ["./entrypoint.sh"]
CMD ["./vdr"]
