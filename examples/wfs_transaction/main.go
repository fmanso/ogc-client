// Example: WFS Transaction (WFS-T)
//
// This example demonstrates how to use the WFS client to insert, update,
// and delete features in a GeoServer instance using WFS-T.
//
// CAUTION: This example modifies data on the server!
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	ogc "github.com/fmanso/ogc-client"
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

	// Create a WFS client
	wfsClient := client.WFS("cite")

	ctx := context.Background()

	// Example 1: Insert a new feature
	fmt.Println("=== Insert Example ===")
	insertResp, err := wfsClient.Transaction().
		Insert("cite:buildings").
		Feature(map[string]interface{}{
			"name":    "New Building",
			"address": "123 Main St",
			"floors":  5,
			// For geometry, you would use GML format
			// "the_geom": "<gml:Point srsName='EPSG:4326'><gml:pos>-74.0060 40.7128</gml:pos></gml:Point>",
		}).
		Handle("insert-building-1").
		SRSName("EPSG:4326").
		Add().
		Execute(ctx)

	if err != nil {
		fmt.Printf("Insert failed (expected if layer doesn't exist): %v\n", err)
	} else {
		fmt.Printf("Insert successful!\n")
		fmt.Printf("  Total inserted: %d\n", insertResp.TransactionSummary.TotalInserted)
		for _, r := range insertResp.InsertResults {
			fmt.Printf("  New feature ID: %s\n", r.ResourceID)
		}
	}

	// Example 2: Update features
	fmt.Println("\n=== Update Example ===")
	updateResp, err := wfsClient.Transaction().
		Update("cite:buildings").
		Set("floors", 10).
		Set("updated_at", time.Now().Format(time.RFC3339)).
		Filter(wfs.PropertyEquals("name", "New Building")).
		Handle("update-building-1").
		Add().
		Execute(ctx)

	if err != nil {
		fmt.Printf("Update failed (expected if layer doesn't exist): %v\n", err)
	} else {
		fmt.Printf("Update successful!\n")
		fmt.Printf("  Total updated: %d\n", updateResp.TransactionSummary.TotalUpdated)
	}

	// Example 3: Delete features
	fmt.Println("\n=== Delete Example ===")
	deleteResp, err := wfsClient.Transaction().
		Delete("cite:buildings").
		Filter(wfs.PropertyEquals("name", "New Building")).
		Handle("delete-building-1").
		Add().
		Execute(ctx)

	if err != nil {
		fmt.Printf("Delete failed (expected if layer doesn't exist): %v\n", err)
	} else {
		fmt.Printf("Delete successful!\n")
		fmt.Printf("  Total deleted: %d\n", deleteResp.TransactionSummary.TotalDeleted)
	}

	// Example 4: Multiple operations in a single transaction
	fmt.Println("\n=== Multi-Operation Transaction Example ===")
	multiResp, err := wfsClient.Transaction().
		// Insert two new features
		Insert("cite:buildings").
		Feature(map[string]interface{}{
			"name":    "Building A",
			"address": "100 First Ave",
			"floors":  3,
		}).
		Feature(map[string]interface{}{
			"name":    "Building B",
			"address": "200 Second Ave",
			"floors":  4,
		}).
		Add().
		// Update existing features
		Update("cite:buildings").
		Set("status", "verified").
		Filter(wfs.PropertyEquals("verified", false)).
		Add().
		// Delete old features
		Delete("cite:buildings").
		Filter(wfs.PropertyLessThan("year_built", 1900)).
		Add().
		Execute(ctx)

	if err != nil {
		fmt.Printf("Multi-operation transaction failed (expected if layer doesn't exist): %v\n", err)
	} else {
		fmt.Printf("Multi-operation transaction successful!\n")
		fmt.Printf("  Total inserted: %d\n", multiResp.TransactionSummary.TotalInserted)
		fmt.Printf("  Total updated: %d\n", multiResp.TransactionSummary.TotalUpdated)
		fmt.Printf("  Total deleted: %d\n", multiResp.TransactionSummary.TotalDeleted)
	}

	// Example 5: Update with spatial filter
	fmt.Println("\n=== Spatial Update Example ===")
	spatialUpdateResp, err := wfsClient.Transaction().
		Update("cite:buildings").
		Set("zone", "downtown").
		Filter(wfs.BBOXFilter("the_geom", -74.01, 40.70, -73.99, 40.72, "EPSG:4326")).
		Add().
		Execute(ctx)

	if err != nil {
		fmt.Printf("Spatial update failed (expected if layer doesn't exist): %v\n", err)
	} else {
		fmt.Printf("Spatial update successful!\n")
		fmt.Printf("  Total updated: %d\n", spatialUpdateResp.TransactionSummary.TotalUpdated)
	}

	// Example 6: Delete with compound filter
	fmt.Println("\n=== Compound Filter Delete Example ===")
	compoundDeleteResp, err := wfsClient.Transaction().
		Delete("cite:buildings").
		Filter(wfs.And(
			wfs.PropertyEquals("status", "demolished"),
			wfs.PropertyLessThan("year_demolished", 2020),
		)).
		Add().
		Execute(ctx)

	if err != nil {
		fmt.Printf("Compound filter delete failed (expected if layer doesn't exist): %v\n", err)
	} else {
		fmt.Printf("Compound filter delete successful!\n")
		fmt.Printf("  Total deleted: %d\n", compoundDeleteResp.TransactionSummary.TotalDeleted)
	}

	fmt.Println("\nTransaction examples completed.")
}
