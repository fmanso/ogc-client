package ogc

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestError(t *testing.T) {
	t.Run("basic error", func(t *testing.T) {
		err := &Error{
			Code:    ErrCodeServiceException,
			Message: "Layer not found",
		}

		str := err.Error()
		if !strings.Contains(str, "SERVICE_EXCEPTION") {
			t.Errorf("expected error to contain code, got: %s", str)
		}
		if !strings.Contains(str, "Layer not found") {
			t.Errorf("expected error to contain message, got: %s", str)
		}
	})

	t.Run("error with locator", func(t *testing.T) {
		err := &Error{
			Code:    ErrCodeInvalidParameter,
			Message: "Invalid value",
			Locator: "BBOX",
		}

		str := err.Error()
		if !strings.Contains(str, "BBOX") {
			t.Errorf("expected error to contain locator, got: %s", str)
		}
	})

	t.Run("error with cause", func(t *testing.T) {
		cause := errors.New("connection refused")
		err := &Error{
			Code:    ErrCodeRequestFailed,
			Message: "Request failed",
			Cause:   cause,
		}

		str := err.Error()
		if !strings.Contains(str, "connection refused") {
			t.Errorf("expected error to contain cause, got: %s", str)
		}
	})

	t.Run("Unwrap", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := &Error{
			Code:    ErrCodeRequestFailed,
			Message: "Request failed",
			Cause:   cause,
		}

		if err.Unwrap() != cause {
			t.Error("Unwrap() did not return cause")
		}
	})

	t.Run("Is", func(t *testing.T) {
		err1 := &Error{Code: ErrCodeServiceException}
		err2 := &Error{Code: ErrCodeServiceException}
		err3 := &Error{Code: ErrCodeInvalidParameter}

		if !err1.Is(err2) {
			t.Error("expected errors with same code to match")
		}
		if err1.Is(err3) {
			t.Error("expected errors with different code to not match")
		}
	})
}

func TestNewError(t *testing.T) {
	err := NewError(ErrCodeInvalidURL, "bad url")
	if err.Code != ErrCodeInvalidURL {
		t.Errorf("expected code %s, got %s", ErrCodeInvalidURL, err.Code)
	}
	if err.Message != "bad url" {
		t.Errorf("expected message 'bad url', got '%s'", err.Message)
	}
}

func TestWrapError(t *testing.T) {
	cause := errors.New("original error")
	err := WrapError(ErrCodeRequestFailed, "wrapped", cause)

	if err.Cause != cause {
		t.Error("expected cause to be set")
	}
	if err.Code != ErrCodeRequestFailed {
		t.Errorf("expected code %s, got %s", ErrCodeRequestFailed, err.Code)
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "layer not defined",
			err:  &Error{Code: ErrCodeLayerNotDefined},
			want: true,
		},
		{
			name: "style not defined",
			err:  &Error{Code: ErrCodeStyleNotDefined},
			want: true,
		},
		{
			name: "404 status",
			err:  &Error{StatusCode: http.StatusNotFound},
			want: true,
		},
		{
			name: "other error",
			err:  &Error{Code: ErrCodeServiceException},
			want: false,
		},
		{
			name: "non-ogc error",
			err:  errors.New("some error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotFound(tt.err); got != tt.want {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsAuthError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "unauthorized",
			err:  &Error{StatusCode: http.StatusUnauthorized},
			want: true,
		},
		{
			name: "forbidden",
			err:  &Error{StatusCode: http.StatusForbidden},
			want: true,
		},
		{
			name: "other status",
			err:  &Error{StatusCode: http.StatusBadRequest},
			want: false,
		},
		{
			name: "non-ogc error",
			err:  errors.New("auth failed"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAuthError(tt.err); got != tt.want {
				t.Errorf("IsAuthError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsServerError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "500",
			err:  &Error{StatusCode: http.StatusInternalServerError},
			want: true,
		},
		{
			name: "503",
			err:  &Error{StatusCode: http.StatusServiceUnavailable},
			want: true,
		},
		{
			name: "400",
			err:  &Error{StatusCode: http.StatusBadRequest},
			want: false,
		},
		{
			name: "non-ogc error",
			err:  errors.New("server error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsServerError(tt.err); got != tt.want {
				t.Errorf("IsServerError() = %v, want %v", got, tt.want)
			}
		})
	}
}
