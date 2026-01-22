// Package http provides HTTP client utilities for OGC services.
package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// AuthFunc is a function that applies authentication to a request.
type AuthFunc func(*http.Request)

// Client wraps http.Client with OGC-specific functionality.
type Client struct {
	httpClient *http.Client
	auth       AuthFunc
	userAgent  string
	baseURL    string
}

// NewClient creates a new HTTP client wrapper.
func NewClient(baseURL string, httpClient *http.Client, auth AuthFunc, userAgent string) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	if userAgent == "" {
		userAgent = "ogc-client/1.0"
	}
	return &Client{
		httpClient: httpClient,
		auth:       auth,
		userAgent:  userAgent,
		baseURL:    strings.TrimSuffix(baseURL, "/"),
	}
}

// Request represents an HTTP request being built.
type Request struct {
	client      *Client
	method      string
	path        string
	params      url.Values
	body        io.Reader
	contentType string
	headers     http.Header
}

// Get creates a new GET request.
func (c *Client) Get(path string) *Request {
	return &Request{
		client:  c,
		method:  http.MethodGet,
		path:    path,
		params:  make(url.Values),
		headers: make(http.Header),
	}
}

// Post creates a new POST request.
func (c *Client) Post(path string) *Request {
	return &Request{
		client:  c,
		method:  http.MethodPost,
		path:    path,
		params:  make(url.Values),
		headers: make(http.Header),
	}
}

// Param adds a query parameter.
func (r *Request) Param(key, value string) *Request {
	r.params.Set(key, value)
	return r
}

// Params adds multiple query parameters.
func (r *Request) Params(params map[string]string) *Request {
	for k, v := range params {
		r.params.Set(k, v)
	}
	return r
}

// Body sets the request body.
func (r *Request) Body(body io.Reader, contentType string) *Request {
	r.body = body
	r.contentType = contentType
	return r
}

// BodyBytes sets the request body from bytes.
func (r *Request) BodyBytes(body []byte, contentType string) *Request {
	return r.Body(bytes.NewReader(body), contentType)
}

// BodyXML sets the request body as XML.
func (r *Request) BodyXML(body []byte) *Request {
	return r.BodyBytes(body, "application/xml")
}

// Header adds a header.
func (r *Request) Header(key, value string) *Request {
	r.headers.Set(key, value)
	return r
}

// buildURL builds the full request URL.
func (r *Request) buildURL() string {
	u := r.client.baseURL
	if r.path != "" {
		u = u + "/" + strings.TrimPrefix(r.path, "/")
	}
	if len(r.params) > 0 {
		u = u + "?" + r.params.Encode()
	}
	return u
}

// Do executes the request and returns the response.
func (r *Request) Do(ctx context.Context) (*http.Response, error) {
	reqURL := r.buildURL()

	req, err := http.NewRequestWithContext(ctx, r.method, reqURL, r.body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("User-Agent", r.client.userAgent)
	if r.contentType != "" {
		req.Header.Set("Content-Type", r.contentType)
	}
	for key, values := range r.headers {
		for _, v := range values {
			req.Header.Add(key, v)
		}
	}

	// Apply authentication
	if r.client.auth != nil {
		r.client.auth(req)
	}

	return r.client.httpClient.Do(req)
}

// DoAndRead executes the request and reads the response body.
func (r *Request) DoAndRead(ctx context.Context) ([]byte, *http.Response, error) {
	resp, err := r.Do(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, resp, nil
}

// BaseURL returns the base URL.
func (c *Client) BaseURL() string {
	return c.baseURL
}
