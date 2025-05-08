package utils

import (
	"context"
	"crypto/rand"
	"encoding/base64"
)

type contextKey string

const nonceKey contextKey = "nonce"

// GenerateNonce returns a base64 nonce string
func GenerateNonce() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

// WithNonce adds a nonce to the request context
func WithNonce(ctx context.Context, nonce string) context.Context {
	return context.WithValue(ctx, nonceKey, nonce)
}

// GetNonce retrieves the nonce from context
func GetNonce(ctx context.Context) (string, bool) {
	nonce, ok := ctx.Value(nonceKey).(string)
	return nonce, ok
}
