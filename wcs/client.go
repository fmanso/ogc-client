// Package wcs provides a client for OGC Web Coverage Service (WCS).
package wcs

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	inthttp "github.com/fmanso/ogc-client/internal/http"
)

const (
	// ServiceName is the WCS service identifier.
	ServiceName = "WCS"
	// Version is the WCS version used for coverage requests.
	Version = "1.0.0"
)

var firstFloatPattern = regexp.MustCompile(`[-+]?(?:\d+\.?\d*|\.\d+)(?:[eE][-+]?\d+)?`)

// Client provides access to WCS operations.
type Client struct {
	http *inthttp.Client
}

// Point represents a WGS84 coordinate.
type Point struct {
	Lon float64
	Lat float64
}

// ProfilePoint represents one sample in a longitudinal profile.
// X is the distance from the beginning of the line in meters.
// Y is the terrain elevation.
type ProfilePoint struct {
	X   float64
	Y   float64
	Lon float64
	Lat float64
}

// NewClient creates a new WCS client.
func NewClient(endpoint string, httpClient *http.Client, auth func(*http.Request), userAgent string) *Client {
	return &Client{
		http: inthttp.NewClient(endpoint, httpClient, auth, userAgent),
	}
}

// LongitudinalProfile samples terrain elevation from WCS along a WGS84 line.
func (c *Client) LongitudinalProfile(ctx context.Context, coverageID string, line []Point, samples int) ([]ProfilePoint, error) {
	if strings.TrimSpace(coverageID) == "" {
		return nil, fmt.Errorf("coverageID is required")
	}
	if len(line) < 2 {
		return nil, fmt.Errorf("line must contain at least two points")
	}
	if samples < 2 {
		return nil, fmt.Errorf("samples must be at least 2")
	}

	cumulative, total := cumulativeDistances(line)
	profile := make([]ProfilePoint, 0, samples)
	for i := 0; i < samples; i++ {
		target := total * float64(i) / float64(samples-1)
		lon, lat := pointAtDistance(line, cumulative, target)
		elevation, err := c.getElevation(ctx, coverageID, lon, lat)
		if err != nil {
			return nil, err
		}
		profile = append(profile, ProfilePoint{
			X:   target,
			Y:   elevation,
			Lon: lon,
			Lat: lat,
		})
	}
	return profile, nil
}

func (c *Client) getElevation(ctx context.Context, coverageID string, lon, lat float64) (float64, error) {
	body, resp, err := c.http.Get("").
		Param("SERVICE", ServiceName).
		Param("VERSION", Version).
		Param("REQUEST", "GetCoverage").
		Param("COVERAGE", coverageID).
		Param("COVERAGEID", coverageID).
		Param("CRS", "EPSG:4326").
		Param("BBOX", fmt.Sprintf("%.8f,%.8f,%.8f,%.8f", lon, lat, lon, lat)).
		Param("WIDTH", "1").
		Param("HEIGHT", "1").
		Param("FORMAT", "text/plain").
		DoAndRead(ctx)
	if err != nil {
		return 0, fmt.Errorf("GetCoverage request failed: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("GetCoverage returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	elevation, err := parseFirstFloat(body)
	if err != nil {
		return 0, fmt.Errorf("failed to parse elevation from WCS response: %w", err)
	}
	return elevation, nil
}

func parseFirstFloat(body []byte) (float64, error) {
	match := firstFloatPattern.Find(body)
	if len(match) == 0 {
		return 0, fmt.Errorf("no numeric value found")
	}
	value, err := strconv.ParseFloat(string(match), 64)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func cumulativeDistances(line []Point) ([]float64, float64) {
	cumulative := make([]float64, len(line))
	for i := 1; i < len(line); i++ {
		cumulative[i] = cumulative[i-1] + haversineMeters(line[i-1], line[i])
	}
	return cumulative, cumulative[len(cumulative)-1]
}

func pointAtDistance(line []Point, cumulative []float64, distance float64) (lon float64, lat float64) {
	if distance <= 0 {
		return line[0].Lon, line[0].Lat
	}
	last := len(cumulative) - 1
	if distance >= cumulative[last] {
		return line[last].Lon, line[last].Lat
	}
	for i := 1; i < len(cumulative); i++ {
		if distance <= cumulative[i] {
			segmentLen := cumulative[i] - cumulative[i-1]
			if segmentLen == 0 {
				return line[i].Lon, line[i].Lat
			}
			ratio := (distance - cumulative[i-1]) / segmentLen
			return line[i-1].Lon + (line[i].Lon-line[i-1].Lon)*ratio,
				line[i-1].Lat + (line[i].Lat-line[i-1].Lat)*ratio
		}
	}
	return line[last].Lon, line[last].Lat
}

func haversineMeters(a, b Point) float64 {
	const earthRadiusMeters = 6371008.8
	lat1 := a.Lat * math.Pi / 180
	lat2 := b.Lat * math.Pi / 180
	dLat := (b.Lat - a.Lat) * math.Pi / 180
	dLon := (b.Lon - a.Lon) * math.Pi / 180
	sinDLat := math.Sin(dLat / 2)
	sinDLon := math.Sin(dLon / 2)
	h := sinDLat*sinDLat + math.Cos(lat1)*math.Cos(lat2)*sinDLon*sinDLon
	return 2 * earthRadiusMeters * math.Atan2(math.Sqrt(h), math.Sqrt(1-h))
}
