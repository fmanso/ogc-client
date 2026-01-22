// Package integration provides integration tests for the OGC client.
// These tests require a running GeoServer instance.
//
// To run integration tests:
//   GEOSERVER_URL=http://localhost:8080/geoserver go test -tags=integration ./...
//
// Required environment variables:
//   GEOSERVER_URL - Base URL of the GeoServer instance
//   GEOSERVER_USER - (optional) Username for basic auth
//   GEOSERVER_PASS - (optional) Password for basic auth
//   GEOSERVER_WORKSPACE - (optional) Workspace to use for tests
//   GEOSERVER_LAYER - (optional) Layer name for WMS tests
//   GEOSERVER_FEATURE_TYPE - (optional) Feature type for WFS tests
//
//go:build integration

package integration

import (
	"context"
	"os"
	"testing"
	"time"

	ogc "github.com/fmanso/ogc-client"
)

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func skipIfNoGeoServer(t *testing.T) {
	if os.Getenv("GEOSERVER_URL") == "" {
		t.Skip("GEOSERVER_URL not set, skipping integration test")
	}
}

func newClient(t *testing.T) *ogc.Client {
	skipIfNoGeoServer(t)

	url := os.Getenv("GEOSERVER_URL")
	opts := []ogc.Option{
		ogc.WithTimeout(30 * time.Second),
	}

	user := os.Getenv("GEOSERVER_USER")
	pass := os.Getenv("GEOSERVER_PASS")
	if user != "" && pass != "" {
		opts = append(opts, ogc.WithBasicAuth(user, pass))
	}

	client, err := ogc.NewClient(url, opts...)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	return client
}

func TestWMSGetCapabilities(t *testing.T) {
	client := newClient(t)
	workspace := getEnvOrDefault("GEOSERVER_WORKSPACE", "")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	wms := client.WMS(workspace)
	caps, err := wms.GetCapabilities(ctx)
	if err != nil {
		t.Fatalf("GetCapabilities() error = %v", err)
	}

	t.Logf("WMS Version: %s", caps.Version)
	t.Logf("Service Title: %s", caps.Service.Title)
	t.Logf("Number of Layers: %d", len(caps.GetLayers()))

	for _, layer := range caps.GetLayers()[:min(5, len(caps.GetLayers()))] {
		t.Logf("  Layer: %s (%s)", layer.Name, layer.Title)
	}
}

func TestWMSGetMap(t *testing.T) {
	client := newClient(t)
	workspace := getEnvOrDefault("GEOSERVER_WORKSPACE", "")
	layerName := getEnvOrDefault("GEOSERVER_LAYER", "")

	if layerName == "" {
		t.Skip("GEOSERVER_LAYER not set, skipping GetMap test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	wms := client.WMS(workspace)

	// Get capabilities to find layer bounds
	caps, err := wms.GetCapabilities(ctx)
	if err != nil {
		t.Fatalf("GetCapabilities() error = %v", err)
	}

	layer := caps.GetLayer(layerName)
	if layer == nil {
		t.Fatalf("Layer %s not found", layerName)
	}

	bbox := layer.BoundingBox["EPSG:4326"]
	if bbox == nil {
		t.Skip("Layer does not have EPSG:4326 bounding box")
	}

	resp, err := wms.GetMap().
		Layers(layerName).
		CRS("EPSG:4326").
		BBox(bbox.MinX, bbox.MinY, bbox.MaxX, bbox.MaxY).
		Size(256, 256).
		Format("image/png").
		Transparent(true).
		Execute(ctx)

	if err != nil {
		t.Fatalf("GetMap() error = %v", err)
	}

	t.Logf("GetMap Response: %d bytes, content-type: %s", len(resp.Data), resp.ContentType)

	// Check PNG signature
	if len(resp.Data) < 8 || resp.Data[0] != 0x89 || resp.Data[1] != 0x50 {
		t.Error("Response does not appear to be a valid PNG")
	}
}

func TestWFSGetCapabilities(t *testing.T) {
	client := newClient(t)
	workspace := getEnvOrDefault("GEOSERVER_WORKSPACE", "")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	wfs := client.WFS(workspace)
	caps, err := wfs.GetCapabilities(ctx)
	if err != nil {
		t.Fatalf("GetCapabilities() error = %v", err)
	}

	t.Logf("WFS Version: %s", caps.Version)
	t.Logf("Service Title: %s", caps.ServiceID.Title)
	t.Logf("Number of Feature Types: %d", len(caps.GetFeatureTypes()))

	for _, ft := range caps.GetFeatureTypes()[:min(5, len(caps.GetFeatureTypes()))] {
		t.Logf("  Feature Type: %s (%s)", ft.Name, ft.Title)
	}
}

func TestWFSGetFeature(t *testing.T) {
	client := newClient(t)
	workspace := getEnvOrDefault("GEOSERVER_WORKSPACE", "")
	featureType := getEnvOrDefault("GEOSERVER_FEATURE_TYPE", "")

	if featureType == "" {
		t.Skip("GEOSERVER_FEATURE_TYPE not set, skipping GetFeature test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	wfs := client.WFS(workspace)

	data, err := wfs.GetFeature(featureType).
		Count(10).
		Execute(ctx)

	if err != nil {
		t.Fatalf("GetFeature() error = %v", err)
	}

	t.Logf("GetFeature Response: %d bytes", len(data))
	t.Logf("First 500 chars: %s", string(data[:min(500, len(data))]))
}

func TestWFSDescribeFeatureType(t *testing.T) {
	client := newClient(t)
	workspace := getEnvOrDefault("GEOSERVER_WORKSPACE", "")
	featureType := getEnvOrDefault("GEOSERVER_FEATURE_TYPE", "")

	if featureType == "" {
		t.Skip("GEOSERVER_FEATURE_TYPE not set, skipping DescribeFeatureType test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	wfs := client.WFS(workspace)

	resp, err := wfs.DescribeFeatureType(ctx, featureType)
	if err != nil {
		t.Fatalf("DescribeFeatureType() error = %v", err)
	}

	if resp.Schema != nil {
		t.Logf("Schema target namespace: %s", resp.Schema.TargetNS)
		props := resp.Schema.GetProperties(featureType)
		t.Logf("Properties: %d", len(props))
		for _, p := range props {
			t.Logf("  %s: %s (geometry=%v)", p.Name, p.Type, p.IsGeometry)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
