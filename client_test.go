package ogc

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		wantErr bool
	}{
		{
			name:    "valid URL",
			baseURL: "https://geoserver.example.com/geoserver",
			wantErr: false,
		},
		{
			name:    "valid URL with trailing slash",
			baseURL: "https://geoserver.example.com/geoserver/",
			wantErr: false,
		},
		{
			name:    "localhost URL",
			baseURL: "http://localhost:8080/geoserver",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.baseURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && client == nil {
				t.Error("NewClient() returned nil client without error")
			}
		})
	}
}

func TestClientOptions(t *testing.T) {
	t.Run("WithTimeout", func(t *testing.T) {
		client, err := NewClient("http://localhost:8080/geoserver",
			WithTimeout(60*time.Second),
		)
		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}
		if client.httpClient.Timeout != 60*time.Second {
			t.Errorf("expected timeout 60s, got %v", client.httpClient.Timeout)
		}
	})

	t.Run("WithBasicAuth", func(t *testing.T) {
		client, err := NewClient("http://localhost:8080/geoserver",
			WithBasicAuth("admin", "geoserver"),
		)
		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}
		if client.auth == nil {
			t.Error("expected auth to be set")
		}
	})

	t.Run("WithUserAgent", func(t *testing.T) {
		client, err := NewClient("http://localhost:8080/geoserver",
			WithUserAgent("custom-agent/1.0"),
		)
		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}
		if client.userAgent != "custom-agent/1.0" {
			t.Errorf("expected user agent 'custom-agent/1.0', got '%s'", client.userAgent)
		}
	})

	t.Run("WithHTTPClient", func(t *testing.T) {
		customHTTP := &http.Client{Timeout: 120 * time.Second}
		client, err := NewClient("http://localhost:8080/geoserver",
			WithHTTPClient(customHTTP),
		)
		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}
		if client.httpClient != customHTTP {
			t.Error("expected custom HTTP client to be set")
		}
	})
}

func TestClientWMS(t *testing.T) {
	client, err := NewClient("http://localhost:8080/geoserver")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	t.Run("global endpoint", func(t *testing.T) {
		wms := client.WMS("")
		if wms == nil {
			t.Error("WMS() returned nil")
		}
	})

	t.Run("workspace endpoint", func(t *testing.T) {
		wms := client.WMS("myworkspace")
		if wms == nil {
			t.Error("WMS() returned nil")
		}
	})
}

func TestClientWFS(t *testing.T) {
	client, err := NewClient("http://localhost:8080/geoserver")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	t.Run("global endpoint", func(t *testing.T) {
		wfs := client.WFS("")
		if wfs == nil {
			t.Error("WFS() returned nil")
		}
	})

	t.Run("workspace endpoint", func(t *testing.T) {
		wfs := client.WFS("myworkspace")
		if wfs == nil {
			t.Error("WFS() returned nil")
		}
	})
}

func TestClientWCS(t *testing.T) {
	client, err := NewClient("http://localhost:8080/geoserver")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	t.Run("global endpoint", func(t *testing.T) {
		wcs := client.WCS("")
		if wcs == nil {
			t.Error("WCS() returned nil")
		}
	})

	t.Run("workspace endpoint", func(t *testing.T) {
		wcs := client.WCS("myworkspace")
		if wcs == nil {
			t.Error("WCS() returned nil")
		}
	})
}

func TestClientBaseURL(t *testing.T) {
	client, err := NewClient("http://localhost:8080/geoserver/")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	expected := "http://localhost:8080/geoserver"
	if client.BaseURL() != expected {
		t.Errorf("BaseURL() = %s, want %s", client.BaseURL(), expected)
	}
}

func TestBasicAuth(t *testing.T) {
	auth := &BasicAuth{
		Username: "admin",
		Password: "secret",
	}

	req := httptest.NewRequest("GET", "http://example.com", nil)
	auth.Apply(req)

	username, password, ok := req.BasicAuth()
	if !ok {
		t.Error("BasicAuth not applied to request")
	}
	if username != "admin" {
		t.Errorf("expected username 'admin', got '%s'", username)
	}
	if password != "secret" {
		t.Errorf("expected password 'secret', got '%s'", password)
	}
}
