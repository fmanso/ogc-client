package geojson

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewPoint(t *testing.T) {
	p := NewPoint(-122.4, 37.8)

	if p.Type != GeometryTypePoint {
		t.Errorf("expected type Point, got %s", p.Type)
	}

	coords, ok := p.Coordinates.([]float64)
	if !ok {
		t.Fatal("expected coordinates to be []float64")
	}
	if len(coords) != 2 {
		t.Errorf("expected 2 coordinates, got %d", len(coords))
	}
	if coords[0] != -122.4 {
		t.Errorf("expected lon -122.4, got %f", coords[0])
	}
	if coords[1] != 37.8 {
		t.Errorf("expected lat 37.8, got %f", coords[1])
	}
}

func TestNewLineString(t *testing.T) {
	coords := [][]float64{{0, 0}, {1, 1}, {2, 2}}
	ls := NewLineString(coords)

	if ls.Type != GeometryTypeLineString {
		t.Errorf("expected type LineString, got %s", ls.Type)
	}
}

func TestNewPolygon(t *testing.T) {
	coords := [][][]float64{{{0, 0}, {1, 0}, {1, 1}, {0, 1}, {0, 0}}}
	poly := NewPolygon(coords)

	if poly.Type != GeometryTypePolygon {
		t.Errorf("expected type Polygon, got %s", poly.Type)
	}
}

func TestNewFeature(t *testing.T) {
	geom := NewPoint(0, 0)
	props := map[string]interface{}{
		"name":  "Test",
		"value": 42,
	}

	f := NewFeature(geom, props)

	if f.Type != "Feature" {
		t.Errorf("expected type Feature, got %s", f.Type)
	}
	if f.Geometry != geom {
		t.Error("expected geometry to be set")
	}
	if f.Properties["name"] != "Test" {
		t.Errorf("expected name='Test', got %v", f.Properties["name"])
	}
}

func TestFeatureGetProperty(t *testing.T) {
	f := NewFeature(nil, map[string]interface{}{
		"name":   "Test",
		"count":  42,
		"active": true,
		"rate":   3.14,
	})

	t.Run("GetProperty", func(t *testing.T) {
		if f.GetProperty("name") != "Test" {
			t.Error("GetProperty failed for string")
		}
		if f.GetProperty("missing") != nil {
			t.Error("GetProperty should return nil for missing key")
		}
	})

	t.Run("GetPropertyString", func(t *testing.T) {
		if f.GetPropertyString("name") != "Test" {
			t.Error("GetPropertyString failed")
		}
		if f.GetPropertyString("missing") != "" {
			t.Error("GetPropertyString should return empty for missing key")
		}
	})

	t.Run("GetPropertyFloat", func(t *testing.T) {
		val, ok := f.GetPropertyFloat("rate")
		if !ok || val != 3.14 {
			t.Error("GetPropertyFloat failed")
		}
		_, ok = f.GetPropertyFloat("missing")
		if ok {
			t.Error("GetPropertyFloat should return false for missing key")
		}
	})

	t.Run("GetPropertyInt", func(t *testing.T) {
		val, ok := f.GetPropertyInt("count")
		if !ok || val != 42 {
			t.Error("GetPropertyInt failed")
		}
	})

	t.Run("GetPropertyBool", func(t *testing.T) {
		val, ok := f.GetPropertyBool("active")
		if !ok || val != true {
			t.Error("GetPropertyBool failed")
		}
	})
}

func TestFeatureCollection(t *testing.T) {
	fc := NewFeatureCollection()

	if fc.Type != "FeatureCollection" {
		t.Errorf("expected type FeatureCollection, got %s", fc.Type)
	}

	f1 := *NewFeature(NewPoint(0, 0), map[string]interface{}{"id": 1})
	f2 := *NewFeature(NewPoint(1, 1), map[string]interface{}{"id": 2})

	fc.AddFeature(f1)
	fc.AddFeature(f2)

	if fc.Len() != 2 {
		t.Errorf("expected 2 features, got %d", fc.Len())
	}

	if fc.Get(0) == nil {
		t.Error("Get(0) returned nil")
	}
	if fc.Get(10) != nil {
		t.Error("Get(10) should return nil for out of bounds")
	}
}

func TestFeatureCollectionFilter(t *testing.T) {
	fc := NewFeatureCollection(
		*NewFeature(nil, map[string]interface{}{"active": true}),
		*NewFeature(nil, map[string]interface{}{"active": false}),
		*NewFeature(nil, map[string]interface{}{"active": true}),
	)

	filtered := fc.Filter(func(f *Feature) bool {
		active, _ := f.GetPropertyBool("active")
		return active
	})

	if filtered.Len() != 2 {
		t.Errorf("expected 2 filtered features, got %d", filtered.Len())
	}
}

func TestParse(t *testing.T) {
	t.Run("FeatureCollection", func(t *testing.T) {
		json := `{"type":"FeatureCollection","features":[{"type":"Feature","geometry":{"type":"Point","coordinates":[0,0]},"properties":{"name":"test"}}]}`
		fc, err := Parse([]byte(json))
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}
		if fc.Len() != 1 {
			t.Errorf("expected 1 feature, got %d", fc.Len())
		}
	})

	t.Run("Single Feature", func(t *testing.T) {
		json := `{"type":"Feature","geometry":{"type":"Point","coordinates":[0,0]},"properties":{"name":"test"}}`
		fc, err := Parse([]byte(json))
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}
		if fc.Len() != 1 {
			t.Errorf("expected 1 feature, got %d", fc.Len())
		}
	})

	t.Run("Geometry", func(t *testing.T) {
		json := `{"type":"Point","coordinates":[0,0]}`
		fc, err := Parse([]byte(json))
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}
		if fc.Len() != 1 {
			t.Errorf("expected 1 feature, got %d", fc.Len())
		}
	})
}

func TestMarshal(t *testing.T) {
	fc := NewFeatureCollection(
		*NewFeature(NewPoint(0, 0), map[string]interface{}{"name": "test"}),
	)

	data, err := Marshal(fc)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	if !strings.Contains(string(data), "FeatureCollection") {
		t.Error("expected output to contain FeatureCollection")
	}
}

func TestCRS(t *testing.T) {
	t.Run("Named CRS", func(t *testing.T) {
		crs := NewNamedCRS("urn:ogc:def:crs:EPSG::4326")
		if crs.Type != "name" {
			t.Errorf("expected type 'name', got '%s'", crs.Type)
		}
		if crs.Properties["name"] != "urn:ogc:def:crs:EPSG::4326" {
			t.Error("expected name property to be set")
		}
	})

	t.Run("Linked CRS", func(t *testing.T) {
		crs := NewLinkedCRS("http://example.com/crs", "proj4")
		if crs.Type != "link" {
			t.Errorf("expected type 'link', got '%s'", crs.Type)
		}
		if crs.Properties["href"] != "http://example.com/crs" {
			t.Error("expected href property to be set")
		}
	})
}

func TestFeatureReader(t *testing.T) {
	json := `{"type":"FeatureCollection","features":[
		{"type":"Feature","geometry":{"type":"Point","coordinates":[0,0]},"properties":{"id":1}},
		{"type":"Feature","geometry":{"type":"Point","coordinates":[1,1]},"properties":{"id":2}},
		{"type":"Feature","geometry":{"type":"Point","coordinates":[2,2]},"properties":{"id":3}}
	]}`

	reader := NewFeatureReader(strings.NewReader(json))
	count := 0

	for reader.Next() {
		f, err := reader.Feature()
		if err != nil {
			t.Fatalf("Feature() error = %v", err)
		}
		count++
		id, ok := f.GetPropertyInt("id")
		if !ok {
			t.Error("expected id property")
		}
		if id != count {
			t.Errorf("expected id=%d, got %d", count, id)
		}
	}

	if err := reader.Err(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if count != 3 {
		t.Errorf("expected 3 features, got %d", count)
	}
}

func TestFeatureWriter(t *testing.T) {
	var buf bytes.Buffer
	writer := NewFeatureWriter(&buf)

	writer.WriteFeature(NewFeature(NewPoint(0, 0), map[string]interface{}{"id": 1}))
	writer.WriteFeature(NewFeature(NewPoint(1, 1), map[string]interface{}{"id": 2}))
	writer.Close()

	output := buf.String()

	if !strings.Contains(output, "FeatureCollection") {
		t.Error("expected output to contain FeatureCollection")
	}
	if !strings.Contains(output, `"id":1`) {
		t.Error("expected output to contain first feature")
	}
	if !strings.Contains(output, `"id":2`) {
		t.Error("expected output to contain second feature")
	}

	// Verify it's valid JSON
	_, err := Parse([]byte(output))
	if err != nil {
		t.Errorf("output is not valid GeoJSON: %v", err)
	}

	if writer.Count() != 2 {
		t.Errorf("expected count=2, got %d", writer.Count())
	}
}

func TestReadAllWriteAll(t *testing.T) {
	original := `{"type":"FeatureCollection","features":[
		{"type":"Feature","geometry":{"type":"Point","coordinates":[0,0]},"properties":{"name":"A"}},
		{"type":"Feature","geometry":{"type":"Point","coordinates":[1,1]},"properties":{"name":"B"}}
	]}`

	// Read all
	fc, err := ReadAll(strings.NewReader(original))
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if fc.Len() != 2 {
		t.Errorf("expected 2 features, got %d", fc.Len())
	}

	// Write all
	var buf bytes.Buffer
	err = WriteAll(&buf, fc)
	if err != nil {
		t.Fatalf("WriteAll() error = %v", err)
	}

	// Re-read to verify
	fc2, err := Parse(buf.Bytes())
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if fc2.Len() != 2 {
		t.Errorf("expected 2 features after round-trip, got %d", fc2.Len())
	}
}
