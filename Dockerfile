FROM golang:alpine

WORKDIR /app

# Install iproute2 for debugging networking (like ip a)
RUN apk add --no-cache iproute2

# Copy all source and modules
COPY . .

# Generate a clean go.sum and download dependencies
RUN go mod tidy

# Build the VDR binary
RUN go build -o vdr ./cmd/vdr/main.go

# Make entrypoint executable
RUN chmod +x ./entrypoint.sh

# Execute the entrypoint script
ENTRYPOINT ["./entrypoint.sh"]
CMD ["./vdr"]
