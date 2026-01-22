// Example: WMS GetMap
//
// This example demonstrates how to use the WMS client to retrieve map images
// from a GeoServer instance.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	ogc "github.com/fmanso/ogc-client"
)

func main() {
	// Create a new OGC client
	client, err := ogc.NewClient(
		"http://localhost:8080/geoserver",
		ogc.WithBasicAuth("admin", "geoserver"),
		ogc.WithTimeout(30*time.Second),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create client: %v\n", err)
		os.Exit(1)
	}

	// Create a WMS client for a specific workspace
	wms := client.WMS("topp")

	ctx := context.Background()

	// Get capabilities to discover available layers
	fmt.Println("Fetching WMS capabilities...")
	caps, err := wms.GetCapabilities(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get capabilities: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Service: %s\n", caps.Service.Title)
	fmt.Printf("Available layers:\n")
	for _, layer := range caps.GetLayers() {
		fmt.Printf("  - %s (%s)\n", layer.Name, layer.Title)
	}

	// Request a map image
	fmt.Println("\nRequesting map image...")
	resp, err := wms.GetMap().
		Layers("topp:states").
		CRS("EPSG:4326").
		BBox(-125, 24, -66, 50). // Continental US
		Size(800, 600).
		Format("image/png").
		Transparent(true).
		Execute(ctx)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get map: %v\n", err)
		os.Exit(1)
	}

	// Save the image to a file
	filename := "map.png"
	err = os.WriteFile(filename, resp.Data, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to save image: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Map saved to %s (%d bytes)\n", filename, len(resp.Data))

	// Get legend graphic
	fmt.Println("\nRequesting legend graphic...")
	legend, err := wms.GetLegendGraphic().
		Layer("topp:states").
		Format("image/png").
		Size(20, 20).
		Execute(ctx)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get legend: %v\n", err)
		os.Exit(1)
	}

	legendFile := "legend.png"
	err = os.WriteFile(legendFile, legend.Data, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to save legend: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Legend saved to %s (%d bytes)\n", legendFile, len(legend.Data))
}
