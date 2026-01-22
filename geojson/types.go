// Package geojson provides GeoJSON types and utilities for working with OGC services.
package geojson

import (
	"encoding/json"
	"fmt"
)

// Geometry types.
const (
	GeometryTypePoint              = "Point"
	GeometryTypeMultiPoint         = "MultiPoint"
	GeometryTypeLineString         = "LineString"
	GeometryTypeMultiLineString    = "MultiLineString"
	GeometryTypePolygon            = "Polygon"
	GeometryTypeMultiPolygon       = "MultiPolygon"
	GeometryTypeGeometryCollection = "GeometryCollection"
)

// Geometry represents a GeoJSON geometry.
type Geometry struct {
	Type        string      `json:"type"`
	Coordinates interface{} `json:"coordinates,omitempty"`
	Geometries  []Geometry  `json:"geometries,omitempty"` // For GeometryCollection
}

// NewPoint creates a new Point geometry.
func NewPoint(lon, lat float64) *Geometry {
	return &Geometry{
		Type:        GeometryTypePoint,
		Coordinates: []float64{lon, lat},
	}
}

// NewPointZ creates a new Point geometry with altitude.
func NewPointZ(lon, lat, alt float64) *Geometry {
	return &Geometry{
		Type:        GeometryTypePoint,
		Coordinates: []float64{lon, lat, alt},
	}
}

// NewLineString creates a new LineString geometry.
func NewLineString(coords [][]float64) *Geometry {
	return &Geometry{
		Type:        GeometryTypeLineString,
		Coordinates: coords,
	}
}

// NewPolygon creates a new Polygon geometry.
// coords should be an array of linear rings, where the first is the exterior ring.
func NewPolygon(coords [][][]float64) *Geometry {
	return &Geometry{
		Type:        GeometryTypePolygon,
		Coordinates: coords,
	}
}

// NewMultiPoint creates a new MultiPoint geometry.
func NewMultiPoint(coords [][]float64) *Geometry {
	return &Geometry{
		Type:        GeometryTypeMultiPoint,
		Coordinates: coords,
	}
}

// NewMultiLineString creates a new MultiLineString geometry.
func NewMultiLineString(coords [][][]float64) *Geometry {
	return &Geometry{
		Type:        GeometryTypeMultiLineString,
		Coordinates: coords,
	}
}

// NewMultiPolygon creates a new MultiPolygon geometry.
func NewMultiPolygon(coords [][][][]float64) *Geometry {
	return &Geometry{
		Type:        GeometryTypeMultiPolygon,
		Coordinates: coords,
	}
}

// NewGeometryCollection creates a new GeometryCollection.
func NewGeometryCollection(geometries ...Geometry) *Geometry {
	return &Geometry{
		Type:       GeometryTypeGeometryCollection,
		Geometries: geometries,
	}
}

// IsPoint returns true if this is a Point geometry.
func (g *Geometry) IsPoint() bool {
	return g.Type == GeometryTypePoint
}

// IsLineString returns true if this is a LineString geometry.
func (g *Geometry) IsLineString() bool {
	return g.Type == GeometryTypeLineString
}

// IsPolygon returns true if this is a Polygon geometry.
func (g *Geometry) IsPolygon() bool {
	return g.Type == GeometryTypePolygon
}

// AsPoint returns the coordinates as a point [lon, lat].
func (g *Geometry) AsPoint() ([]float64, error) {
	if !g.IsPoint() {
		return nil, fmt.Errorf("geometry is not a Point")
	}
	coords, ok := g.Coordinates.([]interface{})
	if !ok {
		// Try direct float slice
		if fc, ok := g.Coordinates.([]float64); ok {
			return fc, nil
		}
		return nil, fmt.Errorf("invalid point coordinates")
	}
	result := make([]float64, len(coords))
	for i, c := range coords {
		f, ok := c.(float64)
		if !ok {
			return nil, fmt.Errorf("invalid coordinate at index %d", i)
		}
		result[i] = f
	}
	return result, nil
}

// Bounds calculates the bounding box of the geometry.
func (g *Geometry) Bounds() (minX, minY, maxX, maxY float64, err error) {
	switch g.Type {
	case GeometryTypePoint:
		coords, err := g.AsPoint()
		if err != nil {
			return 0, 0, 0, 0, err
		}
		return coords[0], coords[1], coords[0], coords[1], nil
	default:
		// For other types, would need to traverse all coordinates
		return 0, 0, 0, 0, fmt.Errorf("bounds calculation not implemented for %s", g.Type)
	}
}

// Feature represents a GeoJSON Feature.
type Feature struct {
	Type       string                 `json:"type"`
	ID         interface{}            `json:"id,omitempty"`
	Geometry   *Geometry              `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
	BBox       []float64              `json:"bbox,omitempty"`
}

// NewFeature creates a new Feature.
func NewFeature(geometry *Geometry, properties map[string]interface{}) *Feature {
	if properties == nil {
		properties = make(map[string]interface{})
	}
	return &Feature{
		Type:       "Feature",
		Geometry:   geometry,
		Properties: properties,
	}
}

// GetProperty returns a property value.
func (f *Feature) GetProperty(name string) interface{} {
	if f.Properties == nil {
		return nil
	}
	return f.Properties[name]
}

// GetPropertyString returns a property as a string.
func (f *Feature) GetPropertyString(name string) string {
	v := f.GetProperty(name)
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

// GetPropertyFloat returns a property as a float64.
func (f *Feature) GetPropertyFloat(name string) (float64, bool) {
	v := f.GetProperty(name)
	if v == nil {
		return 0, false
	}
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}

// GetPropertyInt returns a property as an int.
func (f *Feature) GetPropertyInt(name string) (int, bool) {
	v := f.GetProperty(name)
	if v == nil {
		return 0, false
	}
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	case json.Number:
		i, err := n.Int64()
		return int(i), err == nil
	default:
		return 0, false
	}
}

// GetPropertyBool returns a property as a bool.
func (f *Feature) GetPropertyBool(name string) (bool, bool) {
	v := f.GetProperty(name)
	if v == nil {
		return false, false
	}
	if b, ok := v.(bool); ok {
		return b, true
	}
	return false, false
}

// SetProperty sets a property value.
func (f *Feature) SetProperty(name string, value interface{}) {
	if f.Properties == nil {
		f.Properties = make(map[string]interface{})
	}
	f.Properties[name] = value
}

// FeatureCollection represents a GeoJSON FeatureCollection.
type FeatureCollection struct {
	Type     string    `json:"type"`
	Features []Feature `json:"features"`
	BBox     []float64 `json:"bbox,omitempty"`
	CRS      *CRS      `json:"crs,omitempty"`
}

// NewFeatureCollection creates a new FeatureCollection.
func NewFeatureCollection(features ...Feature) *FeatureCollection {
	return &FeatureCollection{
		Type:     "FeatureCollection",
		Features: features,
	}
}

// AddFeature adds a feature to the collection.
func (fc *FeatureCollection) AddFeature(f Feature) {
	fc.Features = append(fc.Features, f)
}

// Len returns the number of features.
func (fc *FeatureCollection) Len() int {
	return len(fc.Features)
}

// Get returns the feature at the given index.
func (fc *FeatureCollection) Get(index int) *Feature {
	if index < 0 || index >= len(fc.Features) {
		return nil
	}
	return &fc.Features[index]
}

// Filter returns a new FeatureCollection with features matching the predicate.
func (fc *FeatureCollection) Filter(predicate func(*Feature) bool) *FeatureCollection {
	result := NewFeatureCollection()
	for i := range fc.Features {
		if predicate(&fc.Features[i]) {
			result.AddFeature(fc.Features[i])
		}
	}
	return result
}

// CRS represents a Coordinate Reference System.
type CRS struct {
	Type       string            `json:"type"`
	Properties map[string]string `json:"properties"`
}

// NewNamedCRS creates a CRS with a named reference.
func NewNamedCRS(name string) *CRS {
	return &CRS{
		Type: "name",
		Properties: map[string]string{
			"name": name,
		},
	}
}

// NewLinkedCRS creates a CRS with a linked reference.
func NewLinkedCRS(href, crsType string) *CRS {
	return &CRS{
		Type: "link",
		Properties: map[string]string{
			"href": href,
			"type": crsType,
		},
	}
}

// Parse parses GeoJSON from bytes.
func Parse(data []byte) (*FeatureCollection, error) {
	// First try to parse as a FeatureCollection
	var fc FeatureCollection
	if err := json.Unmarshal(data, &fc); err == nil && fc.Type == "FeatureCollection" {
		return &fc, nil
	}

	// Try to parse as a single Feature
	var f Feature
	if err := json.Unmarshal(data, &f); err == nil && f.Type == "Feature" {
		return NewFeatureCollection(f), nil
	}

	// Try to parse as a Geometry
	var g Geometry
	if err := json.Unmarshal(data, &g); err == nil && g.Type != "" {
		feature := NewFeature(&g, nil)
		return NewFeatureCollection(*feature), nil
	}

	return nil, fmt.Errorf("unable to parse GeoJSON")
}

// ParseFeature parses a single GeoJSON feature.
func ParseFeature(data []byte) (*Feature, error) {
	var f Feature
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, err
	}
	if f.Type != "Feature" {
		return nil, fmt.Errorf("expected Feature, got %s", f.Type)
	}
	return &f, nil
}

// ParseGeometry parses a GeoJSON geometry.
func ParseGeometry(data []byte) (*Geometry, error) {
	var g Geometry
	if err := json.Unmarshal(data, &g); err != nil {
		return nil, err
	}
	return &g, nil
}

// Marshal serializes to GeoJSON.
func Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// MarshalIndent serializes to formatted GeoJSON.
func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	return json.MarshalIndent(v, prefix, indent)
}
