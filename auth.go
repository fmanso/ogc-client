package ogc

import "net/http"

// Authenticator defines the interface for authentication strategies.
type Authenticator interface {
	// Apply adds authentication to the HTTP request.
	Apply(req *http.Request)
}

// BasicAuth implements HTTP Basic authentication.
type BasicAuth struct {
	Username string
	Password string
}

// Apply adds Basic authentication headers to the request.
func (a *BasicAuth) Apply(req *http.Request) {
	req.SetBasicAuth(a.Username, a.Password)
}

// Ensure BasicAuth implements Authenticator.
var _ Authenticator = (*BasicAuth)(nil)
