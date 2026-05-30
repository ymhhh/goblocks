package resilience

import (
	"context"
	"fmt"
	"strings"
)

const (
	globalKeyPrefix = "global"
	userKeyPrefix   = "user:"
	routeKeyPrefix  = "route:"
)

// UserIDContextKey is the context key for user id in rate-limit helpers.
type userIDContextKey struct{}

// ContextWithUserID stores user id for rate limiting and handlers.
func ContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDContextKey{}, userID)
}

// UserIDFromContext reads user id from context.
func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDContextKey{}).(string)
	return id, ok && id != ""
}

// UserKeyFromContext builds an L2 key from context user id or "anonymous".
func UserKeyFromContext(ctx context.Context) string {
	if id, ok := UserIDFromContext(ctx); ok {
		return userKeyPrefix + id
	}
	return userKeyPrefix + "anonymous"
}

// RouteKey builds an L3 key from HTTP method and path.
func RouteKey(method, path string) string {
	return fmt.Sprintf("%s%s:%s", routeKeyPrefix, strings.ToUpper(method), path)
}

// GlobalKey returns the L1 service-wide key.
func GlobalKey(serviceName string) string {
	if serviceName == "" {
		return globalKeyPrefix
	}
	return globalKeyPrefix + ":" + serviceName
}
