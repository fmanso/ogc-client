package ogc

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Error codes for OGC client errors.
const (
	ErrCodeInvalidURL          = "INVALID_URL"
	ErrCodeRequestFailed       = "REQUEST_FAILED"
	ErrCodeInvalidResponse     = "INVALID_RESPONSE"
	ErrCodeServiceException    = "SERVICE_EXCEPTION"
	ErrCodeMissingParameter    = "MISSING_PARAMETER"
	ErrCodeInvalidParameter    = "INVALID_PARAMETER"
	ErrCodeOperationNotAllowed = "OPERATION_NOT_ALLOWED"
	ErrCodeLayerNotDefined     = "LAYER_NOT_DEFINED"
	ErrCodeStyleNotDefined     = "STYLE_NOT_DEFINED"
	ErrCodeInvalidCRS          = "INVALID_CRS"
	ErrCodeInvalidBBox         = "INVALID_BBOX"
	ErrCodeInvalidFormat       = "INVALID_FORMAT"
	ErrCodeNoApplicableCode    = "NO_APPLICABLE_CODE"
)

// Error represents an error from the OGC client.
type Error struct {
	Code       string // OGC exception code or internal error code
	Message    string // Human-readable error message
	Locator    string // Parameter or location that caused the error
	StatusCode int    // HTTP status code (0 if not applicable)
	Cause      error  // Underlying error, if any
}

// Error returns the error message.
func (e *Error) Error() string {
	var sb strings.Builder
	sb.WriteString("ogc: ")

	if e.Code != "" {
		sb.WriteString("[")
		sb.WriteString(e.Code)
		sb.WriteString("] ")
	}

	sb.WriteString(e.Message)

	if e.Locator != "" {
		sb.WriteString(" (locator: ")
		sb.WriteString(e.Locator)
		sb.WriteString(")")
	}

	if e.Cause != nil {
		sb.WriteString(": ")
		sb.WriteString(e.Cause.Error())
	}

	return sb.String()
}

// Unwrap returns the underlying error.
func (e *Error) Unwrap() error {
	return e.Cause
}

// Is reports whether target matches this error.
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// NewError creates a new OGC error.
func NewError(code, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// WrapError wraps an error with OGC error context.
func WrapError(code, message string, cause error) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// OGC Exception XML structures for parsing service exceptions.

// ServiceExceptionReport represents an OGC 1.1 service exception report.
type ServiceExceptionReport struct {
	XMLName           xml.Name           `xml:"ServiceExceptionReport"`
	Version           string             `xml:"version,attr"`
	ServiceExceptions []ServiceException `xml:"ServiceException"`
}

// ServiceException represents a single service exception.
type ServiceException struct {
	Code    string `xml:"code,attr"`
	Locator string `xml:"locator,attr"`
	Text    string `xml:",chardata"`
}

// ExceptionReport represents an OGC 2.0 (OWS) exception report.
type ExceptionReport struct {
	XMLName    xml.Name    `xml:"ExceptionReport"`
	Version    string      `xml:"version,attr"`
	Exceptions []Exception `xml:"Exception"`
}

// Exception represents a single OWS exception.
type Exception struct {
	ExceptionCode string   `xml:"exceptionCode,attr"`
	Locator       string   `xml:"locator,attr"`
	ExceptionText []string `xml:"ExceptionText"`
}

// ParseServiceException attempts to parse an OGC service exception from the response.
// It handles both OGC 1.x ServiceExceptionReport and OGC 2.0 ExceptionReport formats.
func ParseServiceException(resp *http.Response) error {
	if resp == nil {
		return nil
	}

	// Check content type for XML
	contentType := resp.Header.Get("Content-Type")
	isXML := strings.Contains(contentType, "xml") ||
		strings.Contains(contentType, "application/vnd.ogc.se_xml")

	// If not XML and status is OK, no exception
	if !isXML && resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	// Read the body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &Error{
			Code:       ErrCodeRequestFailed,
			Message:    "failed to read response body",
			StatusCode: resp.StatusCode,
			Cause:      err,
		}
	}

	// Try to parse as OGC 1.x ServiceExceptionReport
	var seReport ServiceExceptionReport
	if err := xml.Unmarshal(body, &seReport); err == nil && len(seReport.ServiceExceptions) > 0 {
		se := seReport.ServiceExceptions[0]
		code := se.Code
		if code == "" {
			code = ErrCodeServiceException
		}
		return &Error{
			Code:       code,
			Message:    strings.TrimSpace(se.Text),
			Locator:    se.Locator,
			StatusCode: resp.StatusCode,
		}
	}

	// Try to parse as OGC 2.0 ExceptionReport
	var exReport ExceptionReport
	if err := xml.Unmarshal(body, &exReport); err == nil && len(exReport.Exceptions) > 0 {
		ex := exReport.Exceptions[0]
		code := ex.ExceptionCode
		if code == "" {
			code = ErrCodeServiceException
		}
		message := ""
		if len(ex.ExceptionText) > 0 {
			message = strings.Join(ex.ExceptionText, "; ")
		}
		return &Error{
			Code:       code,
			Message:    strings.TrimSpace(message),
			Locator:    ex.Locator,
			StatusCode: resp.StatusCode,
		}
	}

	// If we couldn't parse an exception but status is not OK, return generic error
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &Error{
			Code:       ErrCodeRequestFailed,
			Message:    fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)),
			StatusCode: resp.StatusCode,
		}
	}

	return nil
}

// IsNotFound returns true if the error indicates a resource was not found.
func IsNotFound(err error) bool {
	var ogcErr *Error
	if errors.As(err, &ogcErr) {
		return ogcErr.Code == ErrCodeLayerNotDefined ||
			ogcErr.Code == ErrCodeStyleNotDefined ||
			ogcErr.StatusCode == http.StatusNotFound
	}
	return false
}

// IsAuthError returns true if the error indicates an authentication failure.
func IsAuthError(err error) bool {
	var ogcErr *Error
	if errors.As(err, &ogcErr) {
		return ogcErr.StatusCode == http.StatusUnauthorized ||
			ogcErr.StatusCode == http.StatusForbidden
	}
	return false
}

// IsServerError returns true if the error indicates a server-side failure.
func IsServerError(err error) bool {
	var ogcErr *Error
	if errors.As(err, &ogcErr) {
		return ogcErr.StatusCode >= 500
	}
	return false
}
