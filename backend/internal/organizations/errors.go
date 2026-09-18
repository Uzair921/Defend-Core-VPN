package organizations

import "defendcore-vpn/internal/apperr"

// Sentinel errors for the organizations domain.
var (
ErrNotFound            = apperr.NotFound("organization_not_found", "organization not found")
ErrSlugExists          = apperr.Conflict("slug_exists", "organization slug already exists")
ErrEmailExists         = apperr.Conflict("email_exists", "email already registered")
ErrUserLimitReached    = apperr.Conflict("user_limit_reached", "maximum user limit reached")
ErrServiceLimitReached = apperr.Conflict("service_limit_reached", "maximum service limit reached")
ErrSuspended           = apperr.Forbidden("organization_suspended", "organization is suspended")
ErrAlreadyMember       = apperr.Conflict("already_member", "user is already a member")
ErrNotMember           = apperr.NotFound("not_member", "user is not a member of this organization")
)
