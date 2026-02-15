// Package ogc provides a production-ready HTTP client library for OGC WMS/WFS services,
// specifically tailored for GeoServer.
package ogc

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/fmanso/ogc-client/wfs"
	"github.com/fmanso/ogc-client/wcs"
	"github.com/fmanso/ogc-client/wms"
)

// Client is the main entry point for interacting with OGC services.
// It provides access to WMS and WFS clients configured with shared settings.
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
	auth       Authenticator
	userAgent  string
}

// Option configures a Client.
type Option func(*Client)

// NewClient creates a new OGC client for the given GeoServer base URL.
// The baseURL should be the root GeoServer URL (e.g., "https://geoserver.example.com/geoserver").
func NewClient(baseURL string, opts ...Option) (*Client, error) {
	// Normalize the base URL
	baseURL = strings.TrimSuffix(baseURL, "/")
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, &Error{
			Code:    ErrCodeInvalidURL,
			Message: "invalid base URL",
			Cause:   err,
		}
	}

	c := &Client{
		baseURL: parsed,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		userAgent: "ogc-client/1.0",
	}

	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		c.httpClient = client
	}
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// WithBasicAuth configures Basic authentication.
func WithBasicAuth(username, password string) Option {
	return func(c *Client) {
		c.auth = &BasicAuth{
			Username: username,
			Password: password,
		}
	}
}

// WithAuth sets a custom authenticator.
func WithAuth(auth Authenticator) Option {
	return func(c *Client) {
		c.auth = auth
	}
}

// WithUserAgent sets the User-Agent header for all requests.
func WithUserAgent(userAgent string) Option {
	return func(c *Client) {
		c.userAgent = userAgent
	}
}

// WMS returns a WMS client for the specified workspace.
// If workspace is empty, the global WMS endpoint is used.
func (c *Client) WMS(workspace string) *wms.Client {
	endpoint := c.buildServiceURL(workspace, "wms")
	return wms.NewClient(endpoint, c.httpClient, c.wrapAuth(), c.userAgent)
}

// WFS returns a WFS client for the specified workspace.
// If workspace is empty, the global WFS endpoint is used.
func (c *Client) WFS(workspace string) *wfs.Client {
	endpoint := c.buildServiceURL(workspace, "wfs")
	return wfs.NewClient(endpoint, c.httpClient, c.wrapAuth(), c.userAgent)
}

// WCS returns a WCS client for the specified workspace.
// If workspace is empty, the global WCS endpoint is used.
func (c *Client) WCS(workspace string) *wcs.Client {
	endpoint := c.buildServiceURL(workspace, "wcs")
	return wcs.NewClient(endpoint, c.httpClient, c.wrapAuth(), c.userAgent)
}

// buildServiceURL constructs the service endpoint URL.
func (c *Client) buildServiceURL(workspace, service string) string {
	var path string
	if workspace != "" {
		path = "/" + workspace + "/" + service
	} else {
		path = "/" + service
	}
	return c.baseURL.String() + path
}

// wrapAuth wraps the authenticator for use by sub-clients.
func (c *Client) wrapAuth() func(*http.Request) {
	if c.auth == nil {
		return nil
	}
	return func(req *http.Request) {
		c.auth.Apply(req)
	}
}

// BaseURL returns the configured base URL.
func (c *Client) BaseURL() string {
	return c.baseURL.String()
}
