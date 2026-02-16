package dao

import "errors"

var ErrSiteNotFound = errors.New("site not found")
var ErrRateLimitExceeded = errors.New("rate limit exceeded")
