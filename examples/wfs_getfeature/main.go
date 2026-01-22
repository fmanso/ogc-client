// Example: WFS GetFeature
//
// This example demonstrates how to use the WFS client to query features
// from a GeoServer instance, including filtering and streaming.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	ogc "github.com/fmanso/ogc-client"
	"github.com/fmanso/ogc-client/geojson"
	"github.com/fmanso/ogc-client/wfs"
)

func main() {
	// Create a new OGC client
	client, err := ogc.NewClient(
		"http://localhost:8080/geoserver",
		ogc.WithBasicAuth("admin", "geoserver"),
		ogc.WithTimeout(60*time.Second),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create client: %v\n", err)
		os.Exit(1)
	}

	// Create a WFS client for a specific workspace
	wfsClient := client.WFS("topp")

	ctx := context.Background()

	// Get capabilities to discover available feature types
	fmt.Println("Fetching WFS capabilities...")
	caps, err := wfsClient.GetCapabilities(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get capabilities: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Service: %s\n", caps.ServiceID.Title)
	fmt.Printf("Available feature types:\n")
	for _, ft := range caps.GetFeatureTypes() {
		fmt.Printf("  - %s (%s)\n", ft.Name, ft.Title)
	}

	// Describe a feature type to see its schema
	fmt.Println("\nDescribing feature type 'topp:states'...")
	schema, err := wfsClient.DescribeFeatureType(ctx, "topp:states")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to describe feature type: %v\n", err)
	} else if schema.Schema != nil {
		props := schema.Schema.GetProperties("states")
		fmt.Printf("Properties:\n")
		for _, p := range props {
			geomStr := ""
			if p.IsGeometry {
				geomStr = " [GEOMETRY]"
			}
			fmt.Printf("  - %s: %s%s\n", p.Name, p.Type, geomStr)
		}
	}

	// Basic GetFeature request
	fmt.Println("\nFetching first 5 features...")
	data, err := wfsClient.GetFeature("topp:states").
		OutputFormat(wfs.FormatGeoJSON).
		Count(5).
		PropertyNames("STATE_NAME", "PERSONS", "the_geom").
		Execute(ctx)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get features: %v\n", err)
		os.Exit(1)
	}

	// Parse the GeoJSON response
	fc, err := geojson.Parse(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse GeoJSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Received %d features\n", fc.Len())
	for i := 0; i < fc.Len(); i++ {
		f := fc.Get(i)
		name := f.GetPropertyString("STATE_NAME")
		pop, _ := f.GetPropertyInt("PERSONS")
		fmt.Printf("  %s: %d people\n", name, pop)
	}

	// GetFeature with CQL filter
	fmt.Println("\nFetching states with population > 10 million...")
	filteredData, err := wfsClient.GetFeature("topp:states").
		OutputFormat(wfs.FormatGeoJSON).
		CQLFilter("PERSONS > 10000000").
		SortBy("PERSONS D"). // Sort descending
		Execute(ctx)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get filtered features: %v\n", err)
		os.Exit(1)
	}

	filteredFC, _ := geojson.Parse(filteredData)
	fmt.Printf("Found %d states with population > 10 million:\n", filteredFC.Len())
	for i := 0; i < filteredFC.Len(); i++ {
		f := filteredFC.Get(i)
		name := f.GetPropertyString("STATE_NAME")
		pop, _ := f.GetPropertyInt("PERSONS")
		fmt.Printf("  %s: %d\n", name, pop)
	}

	// GetFeature with OGC Filter
	fmt.Println("\nFetching states using OGC Filter...")
	ogcFilteredData, err := wfsClient.GetFeature("topp:states").
		OutputFormat(wfs.FormatGeoJSON).
		Filter(wfs.And(
			wfs.PropertyGreaterThan("PERSONS", 5000000),
			wfs.PropertyLessThan("PERSONS", 15000000),
		)).
		Count(10).
		Execute(ctx)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get OGC filtered features: %v\n", err)
	} else {
		ogcFC, _ := geojson.Parse(ogcFilteredData)
		fmt.Printf("Found %d states with population between 5-15 million\n", ogcFC.Len())
	}

	// GetFeature with BBOX
	fmt.Println("\nFetching states in the western US (bbox filter)...")
	bboxData, err := wfsClient.GetFeature("topp:states").
		OutputFormat(wfs.FormatGeoJSON).
		BBox(-125, 30, -110, 50, "EPSG:4326").
		Execute(ctx)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get bbox features: %v\n", err)
	} else {
		bboxFC, _ := geojson.Parse(bboxData)
		fmt.Printf("Found %d states in western US bbox\n", bboxFC.Len())
	}

	// Streaming example
	fmt.Println("\nStreaming all features...")
	stream, err := wfsClient.GetFeature("topp:states").
		OutputFormat(wfs.FormatGeoJSON).
		Stream(ctx)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start stream: %v\n", err)
		os.Exit(1)
	}
	defer stream.Close()

	// Use the GeoJSON streaming reader
	reader := geojson.NewFeatureReader(stream)
	totalPop := 0
	count := 0

	for reader.Next() {
		f, err := reader.Feature()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading feature: %v\n", err)
			break
		}
		count++
		pop, _ := f.GetPropertyInt("PERSONS")
		totalPop += pop
	}

	if err := reader.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Stream error: %v\n", err)
	}

	fmt.Printf("Processed %d features, total population: %d\n", count, totalPop)

	// Save result to file
	fmt.Println("\nSaving features to file...")
	file, err := os.Create("states.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	encoder.Encode(fc)
	fmt.Println("Saved to states.json")
}
