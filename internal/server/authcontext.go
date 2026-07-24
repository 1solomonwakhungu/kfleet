package server

import (
	"context"

	"github.com/1solomonwakhungu/kfleet/pkg/types"
)

type contextKey string

const userContextKey contextKey = "kfleet.authenticatedUser"

func withAuthenticatedUser(ctx context.Context, user types.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// authenticatedUser returns the user attached to the request context by
// requireAuth. It must only be called from handlers wrapped by requireAuth
// or requireRole, which guarantee the value is present.
func authenticatedUser(ctx context.Context) (types.User, bool) {
	user, ok := ctx.Value(userContextKey).(types.User)
	return user, ok
}
