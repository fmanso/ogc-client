# OGC Client for Go

A production-ready Go HTTP client library for OGC WMS 1.3.0 and WFS 2.0.0 services, specifically tailored for GeoServer.

## Features

- **WMS 1.3.0 Support**
  - GetCapabilities
  - GetMap (with builder pattern)
  - DescribeLayer
  - GetLegendGraphic

- **WFS 2.0.0 Support**
  - GetCapabilities
  - DescribeFeatureType
  - GetFeature (with streaming support)
  - GetPropertyValue
  - Transactions (Insert, Update, Delete)

- **WCS Support**
  - Longitudinal terrain profile sampling along WGS84 lines

- **Additional Features**
  - OGC Filter Encoding / CQL support
  - GeoJSON types and GML conversion
  - Streaming reader/writer for large datasets
  - Basic authentication
  - Structured OGC exception parsing
  - Context-based cancellation
  - Zero external dependencies (standard library only)

## Installation

```bash
go get github.com/fmanso/ogc-client
```

## Quick Start

### WMS Client

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    ogc "github.com/fmanso/ogc-client"
    "github.com/fmanso/ogc-client/wms"
)

func main() {
    // Create client
    client := ogc.NewClient("http://localhost:8080/geoserver/wms")

    // Create WMS client
    wmsClient := wms.NewClient(client)

    ctx := context.Background()

    // Get capabilities
    caps, err := wmsClient.GetCapabilities(ctx)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Service: %s\n", caps.Service.Title)
    fmt.Printf("Layers: %v\n", caps.GetLayerNames())

    // Get a map image
    img, err := wmsClient.GetMap(ctx, &wms.GetMapRequest{
        Layers: []string{"topp:states"},
        CRS:    "EPSG:4326",
        BBox:   &wms.BBox{MinX: -125, MinY: 24, MaxX: -66, MaxY: 50},
        Width:  800,
        Height: 600,
        Format: "image/png",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer img.Close()

    // Save to file
    f, _ := os.Create("map.png")
    defer f.Close()
    io.Copy(f, img)
}
```

### WMS GetMap Builder Pattern

```go
// Using the builder pattern for more complex requests
img, err := wmsClient.GetMapBuilder().
    Layers("topp:states", "topp:cities").
    CRS("EPSG:4326").
    BBox(-125, 24, -66, 50).
    Size(800, 600).
    Format("image/png").
    Transparent(true).
    Style("population").
    Execute(ctx)
```

### WFS Client

```go
package main

import (
    "context"
    "fmt"
    "log"

    ogc "github.com/fmanso/ogc-client"
    "github.com/fmanso/ogc-client/wfs"
)

func main() {
    // Create client with authentication
    client := ogc.NewClient(
        "http://localhost:8080/geoserver/wfs",
        ogc.WithBasicAuth("admin", "geoserver"),
    )

    // Create WFS client
    wfsClient := wfs.NewClient(client)

    ctx := context.Background()

    // Get capabilities
    caps, err := wfsClient.GetCapabilities(ctx)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Service: %s\n", caps.ServiceIdentification.Title)
    for _, ft := range caps.FeatureTypeList.FeatureTypes {
        fmt.Printf("  - %s\n", ft.Name)
    }

    // Get features as GeoJSON
    fc, err := wfsClient.GetFeature(ctx, &wfs.GetFeatureRequest{
        TypeNames: []string{"topp:states"},
        Count:     10,
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Got %d features\n", len(fc.Features))
    for _, f := range fc.Features {
        fmt.Printf("  ID: %s, Properties: %v\n", f.ID, f.Properties)
    }
}
```

### WFS Filtering

The library supports OGC Filter Encoding for querying features:

```go
import "github.com/fmanso/ogc-client/wfs"

// Property equals
filter := wfs.PropertyEquals("STATE_NAME", "California")

// Property like (wildcard matching)
filter := wfs.PropertyLike("STATE_NAME", "New*")

// Property between
filter := wfs.PropertyBetween("POPULATION", "1000000", "5000000")

// Spatial filter - BBOX
filter := wfs.BBOX("the_geom", -125, 24, -66, 50, "EPSG:4326")

// Spatial filter - Intersects
filter := wfs.Intersects("the_geom", geojson.Polygon{...})

// Spatial filter - DWithin (distance)
filter := wfs.DWithin("the_geom", geojson.Point{-122.4, 37.8}, 1000, "meters")

// Combine filters with AND/OR
filter := wfs.And(
    wfs.PropertyEquals("STATE_NAME", "California"),
    wfs.PropertyBetween("POPULATION", "100000", "500000"),
)

filter := wfs.Or(
    wfs.PropertyEquals("STATE_NAME", "California"),
    wfs.PropertyEquals("STATE_NAME", "Oregon"),
)

// Negate a filter
filter := wfs.Not(wfs.PropertyEquals("STATE_NAME", "Texas"))

// Filter by resource ID
filter := wfs.ResourceID("states.1", "states.2", "states.3")

// Use filter in GetFeature request
fc, err := wfsClient.GetFeature(ctx, &wfs.GetFeatureRequest{
    TypeNames: []string{"topp:states"},
    Filter:    filter,
})
```

### WFS Transactions

Insert, update, and delete features using WFS-T:

```go
import (
    "github.com/fmanso/ogc-client/wfs"
    "github.com/fmanso/ogc-client/geojson"
)

// Create a transaction
tx := wfs.NewTransaction()

// Insert a new feature
tx.Insert("topp:states", &geojson.Feature{
    Geometry: &geojson.Polygon{
        Coordinates: [][][]float64{
            {{-120, 35}, {-120, 40}, {-115, 40}, {-115, 35}, {-120, 35}},
        },
    },
    Properties: map[string]any{
        "STATE_NAME": "New State",
        "POPULATION": 1000000,
    },
})

// Update existing features
tx.Update("topp:states",
    map[string]any{"POPULATION": 2000000},
    wfs.PropertyEquals("STATE_NAME", "California"),
)

// Delete features
tx.Delete("topp:states",
    wfs.PropertyEquals("STATE_NAME", "Old State"),
)

// Execute the transaction
result, err := wfsClient.Transaction(ctx, tx)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Inserted: %d, Updated: %d, Deleted: %d\n",
    result.TotalInserted, result.TotalUpdated, result.TotalDeleted)
```

### Streaming Large Datasets

For large feature collections, use streaming to avoid loading everything into memory:

```go
// Stream features one at a time
reader, err := wfsClient.GetFeatureStream(ctx, &wfs.GetFeatureRequest{
    TypeNames: []string{"topp:states"},
})
if err != nil {
    log.Fatal(err)
}
defer reader.Close()

for {
    feature, err := reader.Next()
    if err == io.EOF {
        break
    }
    if err != nil {
        log.Fatal(err)
    }
    
    // Process feature
    fmt.Printf("Feature: %s\n", feature.ID)
}
```

### GeoJSON Types

The library includes GeoJSON types for working with spatial data:

```go
import "github.com/fmanso/ogc-client/geojson"

// Point
point := &geojson.Point{Coordinates: []float64{-122.4, 37.8}}

// LineString
line := &geojson.LineString{
    Coordinates: [][]float64{{-122.4, 37.8}, {-122.5, 37.9}},
}

// Polygon
polygon := &geojson.Polygon{
    Coordinates: [][][]float64{
        {{-120, 35}, {-120, 40}, {-115, 40}, {-115, 35}, {-120, 35}},
    },
}

// Feature
feature := &geojson.Feature{
    ID:       "feature.1",
    Geometry: point,
    Properties: map[string]any{
        "name": "San Francisco",
    },
}

// FeatureCollection
fc := &geojson.FeatureCollection{
    Features: []*geojson.Feature{feature},
}

// Serialize to JSON
data, err := json.Marshal(fc)
```

## Client Options

Configure the client with various options:

```go
import ogc "github.com/fmanso/ogc-client"

client := ogc.NewClient(
    "http://localhost:8080/geoserver/wms",
    
    // Basic authentication
    ogc.WithBasicAuth("username", "password"),
    
    // Custom HTTP client
    ogc.WithHTTPClient(&http.Client{
        Timeout: 30 * time.Second,
    }),
    
    // Custom headers
    ogc.WithHeader("X-Custom-Header", "value"),
    
    // Custom user agent
    ogc.WithUserAgent("MyApp/1.0"),
)
```

## Error Handling

The library parses OGC exception reports and returns structured errors:

```go
fc, err := wfsClient.GetFeature(ctx, &wfs.GetFeatureRequest{
    TypeNames: []string{"nonexistent:layer"},
})
if err != nil {
    // Check if it's an OGC exception
    if ogcErr, ok := ogc.AsOGCError(err); ok {
        fmt.Printf("OGC Exception Code: %s\n", ogcErr.Code)
        fmt.Printf("OGC Exception Text: %s\n", ogcErr.Text)
        fmt.Printf("Locator: %s\n", ogcErr.Locator)
    } else {
        // Regular error
        log.Fatal(err)
    }
}
```

## Context Support

All operations support context for cancellation and timeouts:

```go
// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

caps, err := wmsClient.GetCapabilities(ctx)

// With cancellation
ctx, cancel := context.WithCancel(context.Background())
go func() {
    time.Sleep(5 * time.Second)
    cancel() // Cancel after 5 seconds
}()

fc, err := wfsClient.GetFeature(ctx, req)
if errors.Is(err, context.Canceled) {
    fmt.Println("Request was cancelled")
}
```

## Supported Services

This library is designed to work with:

- **GeoServer** (primary target)
- Any OGC-compliant WMS 1.3.0 server
- Any OGC-compliant WFS 2.0.0 server

## Project Structure

```
ogc-client/
├── client.go           # Main client with options pattern
├── auth.go             # Authentication support
├── errors.go           # OGC exception parsing
├── wms/
│   ├── client.go       # WMS client implementation
│   └── types.go        # WMS types and capabilities
├── wfs/
│   ├── client.go       # WFS client implementation
│   ├── types.go        # WFS types and capabilities
│   ├── filter.go       # OGC Filter builder
│   └── transaction.go  # WFS-T support
├── geojson/
│   ├── types.go        # GeoJSON types
│   ├── parser.go       # GML to GeoJSON conversion
│   └── stream.go       # Streaming reader/writer
└── internal/
    ├── http/           # HTTP client wrapper
    └── xml/            # XML utilities
```

## Testing

Run all tests:

```bash
go test ./...
```

Run with verbose output:

```bash
go test -v ./...
```

Run integration tests (requires a running GeoServer):

```bash
go test -tags=integration ./...
```

## License

MIT License - see LICENSE file for details.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.
