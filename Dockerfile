
# Stage 1: Build a static Go binary
FROM golang:1.23 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 👇 Fully static binary (no glibc dependencies)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	go build -ldflags="-w -s" -o server ./cmd/api

# Stage 2: Minimal secure image (no glibc, no shell)
FROM gcr.io/distroless/static:nonroot

WORKDIR /

# Copy the statically compiled binary
COPY --from=builder /app/server /

# Copy runtime assets
COPY --from=builder /app/templates /templates
COPY --from=builder /app/static /static

USER nonroot:nonroot
ENTRYPOINT ["/server"]

