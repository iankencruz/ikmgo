
# -------- Build stage --------
FROM --platform=linux/amd64 golang:1.23-alpine AS builder

WORKDIR /app

# Install git for module fetching
RUN apk add --no-cache git

# Copy go.mod and download deps
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Copy source
COPY . .

# ✅ Build using correct arch (no GOARCH/GOOS needed here)
RUN go build -o app ./cmd/api

# Ensure binary is executable
RUN chmod +x /app/app

# -------- Runtime stage --------
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/app .

EXPOSE 8080

CMD ["./app"]


