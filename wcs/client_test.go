package wcs

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestLongitudinalProfile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("SERVICE") != "WCS" {
			t.Errorf("expected SERVICE=WCS, got %s", q.Get("SERVICE"))
		}
		if q.Get("VERSION") != "1.0.0" {
			t.Errorf("expected VERSION=1.0.0, got %s", q.Get("VERSION"))
		}
		if q.Get("REQUEST") != "GetCoverage" {
			t.Errorf("expected REQUEST=GetCoverage, got %s", q.Get("REQUEST"))
		}
		if q.Get("COVERAGE") != "dem" {
			t.Errorf("expected COVERAGE=dem, got %s", q.Get("COVERAGE"))
		}

		bboxParts := strings.Split(q.Get("BBOX"), ",")
		if len(bboxParts) != 4 {
			t.Fatalf("invalid BBOX: %s", q.Get("BBOX"))
		}
		lon, err := strconv.ParseFloat(bboxParts[0], 64)
		if err != nil {
			t.Fatalf("invalid BBOX longitude: %v", err)
		}

		fmt.Fprintf(w, "%.2f", lon*100)
	}))
	defer server.Close()

	client := NewClient(server.URL, nil, nil, "")
	profile, err := client.LongitudinalProfile(context.Background(), "dem", []Point{
		{Lon: 0, Lat: 0},
		{Lon: 1, Lat: 0},
	}, 3)
	if err != nil {
		t.Fatalf("LongitudinalProfile() error = %v", err)
	}

	if len(profile) != 3 {
		t.Fatalf("expected 3 samples, got %d", len(profile))
	}
	if profile[0].X != 0 {
		t.Errorf("expected first distance 0, got %f", profile[0].X)
	}
	if profile[0].Y != 0 {
		t.Errorf("expected first elevation 0, got %f", profile[0].Y)
	}
	if profile[1].Y != 50 {
		t.Errorf("expected middle elevation 50, got %f", profile[1].Y)
	}
	if profile[2].Y != 100 {
		t.Errorf("expected last elevation 100, got %f", profile[2].Y)
	}
	// 1 degree of longitude at the equator is ~111.2 km.
	const minExpectedMeters = 111000.0
	const maxExpectedMeters = 112500.0
	if profile[2].X <= minExpectedMeters || profile[2].X >= maxExpectedMeters {
		t.Errorf("unexpected total distance %f", profile[2].X)
	}
}

func TestLongitudinalProfileValidation(t *testing.T) {
	client := NewClient("http://localhost:8080", nil, nil, "")
	line := []Point{{Lon: 0, Lat: 0}, {Lon: 1, Lat: 1}}

	t.Run("missing coverage", func(t *testing.T) {
		_, err := client.LongitudinalProfile(context.Background(), "", line, 2)
		if err == nil {
			t.Fatal("expected error for missing coverage")
		}
	})

	t.Run("short line", func(t *testing.T) {
		_, err := client.LongitudinalProfile(context.Background(), "dem", line[:1], 2)
		if err == nil {
			t.Fatal("expected error for short line")
		}
	})

	t.Run("invalid sample count", func(t *testing.T) {
		_, err := client.LongitudinalProfile(context.Background(), "dem", line, 1)
		if err == nil {
			t.Fatal("expected error for invalid sample count")
		}
	})
}
