package middleware

import (
	"context"
	"net/http"
	"strings"

	"vibe/api"
)

type ctxKey string

const (
	ContextCredentials ctxKey = "credentials"
)

// Auth 中间件：从 Authorization header 解析凭证并注入 context
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			authHeader = strings.TrimPrefix(authHeader, "Bearer ")
		}

		creds := api.ParseCredentials(authHeader)
		ctx := context.WithValue(r.Context(), ContextCredentials, creds)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetCredentials 从 context 获取凭证
func GetCredentials(ctx context.Context) *api.Credentials {
	if creds, ok := ctx.Value(ContextCredentials).(*api.Credentials); ok {
		return creds
	}
	return &api.Credentials{}
}