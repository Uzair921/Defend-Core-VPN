package middleware

import (
"context"
"net/http"

"github.com/google/uuid"

"defendcore-vpn/internal/httputil"
)

// OrgMembershipResolver resolves a user's primary organization membership.
// It is implemented by the organizations repository or a dedicated resolver.
type OrgMembershipResolver interface {
PrimaryOrgForUser(ctx context.Context, userID uuid.UUID) (uuid.UUID, error)
}

// Tenant sets the organization context for the current request. It requires
// the user ID to already be present in the context (populated by Auth).
//
// If the user does not belong to any organization, the request is rejected
// with 403 unless the route explicitly allows org-less users.
func Tenant(resolver OrgMembershipResolver) func(http.Handler) http.Handler {
return func(next http.Handler) http.Handler {
return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
userID, ok := UserIDFromContext(r.Context())
if !ok {
httputil.Error(w, http.StatusUnauthorized, "unauthenticated", "Not authenticated")
return
}

orgID, err := resolver.PrimaryOrgForUser(r.Context(), userID)
if err != nil {
httputil.Error(w, http.StatusForbidden, "no_organization", "User is not a member of any organization")
return
}

ctx := context.WithValue(r.Context(), OrgIDKey, orgID)
next.ServeHTTP(w, r.WithContext(ctx))
})
}
}
