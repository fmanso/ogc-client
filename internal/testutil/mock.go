// Package testutil provides test utilities for OGC client tests.
package testutil

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
)

// MockServer creates a test server that responds with the given handler.
func MockServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

// XMLResponse creates a handler that responds with XML content.
func XMLResponse(statusCode int, xmlContent string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(statusCode)
		io.WriteString(w, xmlContent)
	}
}

// JSONResponse creates a handler that responds with JSON content.
func JSONResponse(statusCode int, jsonContent string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		io.WriteString(w, jsonContent)
	}
}

// ImageResponse creates a handler that responds with image content.
func ImageResponse(statusCode int, contentType string, data []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(statusCode)
		w.Write(data)
	}
}

// ServiceExceptionResponse creates a handler that responds with an OGC service exception.
func ServiceExceptionResponse(code, message string) http.HandlerFunc {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<ServiceExceptionReport version="1.3.0">
  <ServiceException code="` + code + `">` + message + `</ServiceException>
</ServiceExceptionReport>`
	return XMLResponse(http.StatusBadRequest, xml)
}

// OWSExceptionResponse creates a handler that responds with an OWS 2.0 exception.
func OWSExceptionResponse(code, message string) http.HandlerFunc {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<ows:ExceptionReport xmlns:ows="http://www.opengis.net/ows/1.1" version="2.0.0">
  <ows:Exception exceptionCode="` + code + `">
    <ows:ExceptionText>` + message + `</ows:ExceptionText>
  </ows:Exception>
</ows:ExceptionReport>`
	return XMLResponse(http.StatusBadRequest, xml)
}

// RequestRecorder records requests for verification.
type RequestRecorder struct {
	Requests []*http.Request
	Handler  http.HandlerFunc
}

// NewRequestRecorder creates a recorder that wraps another handler.
func NewRequestRecorder(handler http.HandlerFunc) *RequestRecorder {
	return &RequestRecorder{
		Handler: handler,
	}
}

// ServeHTTP implements http.Handler.
func (r *RequestRecorder) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.Requests = append(r.Requests, req)
	if r.Handler != nil {
		r.Handler(w, req)
	}
}

// LastRequest returns the last recorded request.
func (r *RequestRecorder) LastRequest() *http.Request {
	if len(r.Requests) == 0 {
		return nil
	}
	return r.Requests[len(r.Requests)-1]
}

// AssertParam checks if a query parameter has the expected value.
func AssertParam(req *http.Request, key, expected string) bool {
	actual := req.URL.Query().Get(key)
	return strings.EqualFold(actual, expected)
}
