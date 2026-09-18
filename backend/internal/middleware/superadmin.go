package middleware

import (
"context"
"net/http"

"defendcore-vpn/internal/httputil"
)

// SuperAdminChecker is implemented by the users repository to determine
// whether a given user has super-admin privileges.
type SuperAdminChecker interface {
IsSuperAdmin(ctx context.Context, userID any) (bool, error)
}

// SuperAdmin gates a route so that only authenticated super-admins pass.
// It must be used after the Auth middleware, which places the user ID in
// the request context.
func SuperAdmin(checker SuperAdminChecker) func(http.Handler) http.Handler {
return func(next http.Handler) http.Handler {
return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
userID, ok := UserIDFromContext(r.Context())
if !ok {
httputil.Error(w, http.StatusUnauthorized, "unauthenticated", "Not authenticated")
return
}

isAdmin, err := checker.IsSuperAdmin(r.Context(), userID)
if err != nil {
httputil.Error(w, http.StatusInternalServerError, "internal_error", "Failed to verify privileges")
return
}
if !isAdmin {
httputil.Error(w, http.StatusForbidden, "forbidden", "Super-admin access required")
return
}

next.ServeHTTP(w, r)
})
}
}
