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

# Execute the binary
CMD ["./vdr"]
