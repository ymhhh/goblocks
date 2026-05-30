package resilience

import (
	"context"
	"testing"
)

func TestRouteKey(t *testing.T) {
	got := RouteKey("post", "/ai/chat")
	want := "route:POST:/ai/chat"
	if got != want {
		t.Fatalf("RouteKey = %q, want %q", got, want)
	}
}

func TestGlobalKey(t *testing.T) {
	if GlobalKey("") != "global" {
		t.Fatalf("empty service: got %q", GlobalKey(""))
	}
	if GlobalKey("api") != "global:api" {
		t.Fatalf("named service: got %q", GlobalKey("api"))
	}
}

func TestUserKeyFromContext(t *testing.T) {
	ctx := ContextWithUserID(context.Background(), "alice")
	if got := UserKeyFromContext(ctx); got != "user:alice" {
		t.Fatalf("got %q", got)
	}
	if got := UserKeyFromContext(context.Background()); got != "user:anonymous" {
		t.Fatalf("anonymous: got %q", got)
	}
}

func TestUserIDFromContext(t *testing.T) {
	ctx := ContextWithUserID(context.Background(), "bob")
	id, ok := UserIDFromContext(ctx)
	if !ok || id != "bob" {
		t.Fatalf("got id=%q ok=%v", id, ok)
	}
	_, ok = UserIDFromContext(context.Background())
	if ok {
		t.Fatal("expected false for empty context")
	}
	_, ok = UserIDFromContext(ContextWithUserID(context.Background(), ""))
	if ok {
		t.Fatal("expected false for empty user id")
	}
}
