package geojson

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
)

// GMLParser parses GML to GeoJSON.
type GMLParser struct {
	// DefaultSRS is the default SRS to use if not specified in the GML.
	DefaultSRS string
}

// NewGMLParser creates a new GML parser.
func NewGMLParser() *GMLParser {
	return &GMLParser{
		DefaultSRS: "EPSG:4326",
	}
}

// ParseGML parses a GML feature collection to GeoJSON.
func (p *GMLParser) ParseGML(data []byte) (*FeatureCollection, error) {
	// Try to parse as a WFS FeatureCollection
	var wfsFC wfsFeatureCollection
	if err := xml.Unmarshal(data, &wfsFC); err == nil && len(wfsFC.Members) > 0 {
		return p.convertWFSFeatureCollection(&wfsFC)
	}

	// Try alternative member element
	var wfsFC2 wfsFeatureCollection2
	if err := xml.Unmarshal(data, &wfsFC2); err == nil && len(wfsFC2.Members) > 0 {
		return p.convertWFSFeatureCollection2(&wfsFC2)
	}

	return nil, fmt.Errorf("unable to parse GML")
}

// wfsFeatureCollection represents a WFS 2.0 FeatureCollection.
type wfsFeatureCollection struct {
	XMLName        xml.Name    `xml:"FeatureCollection"`
	NumberMatched  string      `xml:"numberMatched,attr"`
	NumberReturned string      `xml:"numberReturned,attr"`
	Members        []gmlMember `xml:"member"`
}

type wfsFeatureCollection2 struct {
	XMLName xml.Name    `xml:"FeatureCollection"`
	Members []gmlMember `xml:"featureMember"`
}

type gmlMember struct {
	Inner []byte `xml:",innerxml"`
}

func (p *GMLParser) convertWFSFeatureCollection(wfs *wfsFeatureCollection) (*FeatureCollection, error) {
	fc := NewFeatureCollection()

	for _, member := range wfs.Members {
		feature, err := p.parseGMLFeature(member.Inner)
		if err != nil {
			continue // Skip invalid features
		}
		fc.AddFeature(*feature)
	}

	return fc, nil
}

func (p *GMLParser) convertWFSFeatureCollection2(wfs *wfsFeatureCollection2) (*FeatureCollection, error) {
	fc := NewFeatureCollection()

	for _, member := range wfs.Members {
		feature, err := p.parseGMLFeature(member.Inner)
		if err != nil {
			continue
		}
		fc.AddFeature(*feature)
	}

	return fc, nil
}

func (p *GMLParser) parseGMLFeature(data []byte) (*Feature, error) {
	// Parse the raw XML to extract properties and geometry
	decoder := xml.NewDecoder(strings.NewReader(string(data)))

	feature := NewFeature(nil, nil)
	var currentElement string
	var featureID string

	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}

		switch t := token.(type) {
		case xml.StartElement:
			currentElement = t.Name.Local

			// Check for feature ID
			for _, attr := range t.Attr {
				if attr.Name.Local == "id" || attr.Name.Local == "fid" {
					featureID = attr.Value
				}
			}

			// Check if this is a geometry element
			if isGMLGeometryElement(currentElement) {
				geom, err := p.parseGMLGeometry(decoder, t)
				if err == nil {
					feature.Geometry = geom
				}
			}

		case xml.CharData:
			text := strings.TrimSpace(string(t))
			if text != "" && currentElement != "" && !isGMLGeometryElement(currentElement) {
				// This is a property value
				feature.SetProperty(currentElement, parseValue(text))
			}

		case xml.EndElement:
			currentElement = ""
		}
	}

	if featureID != "" {
		feature.ID = featureID
	}

	return feature, nil
}

func isGMLGeometryElement(name string) bool {
	name = strings.ToLower(name)
	geometryElements := []string{
		"point", "linestring", "polygon", "multipoint",
		"multilinestring", "multipolygon", "multicurve",
		"multisurface", "geometrycollection", "curve",
		"surface", "geometry", "the_geom", "geom", "shape",
	}
	for _, g := range geometryElements {
		if strings.Contains(name, g) {
			return true
		}
	}
	return false
}

func (p *GMLParser) parseGMLGeometry(decoder *xml.Decoder, start xml.StartElement) (*Geometry, error) {
	name := strings.ToLower(start.Name.Local)

	switch {
	case strings.Contains(name, "point"):
		return p.parseGMLPoint(decoder)
	case strings.Contains(name, "linestring"):
		return p.parseGMLLineString(decoder)
	case strings.Contains(name, "polygon"):
		return p.parseGMLPolygon(decoder)
	case strings.Contains(name, "multipoint"):
		return p.parseGMLMultiPoint(decoder)
	case strings.Contains(name, "multilinestring"), strings.Contains(name, "multicurve"):
		return p.parseGMLMultiLineString(decoder)
	case strings.Contains(name, "multipolygon"), strings.Contains(name, "multisurface"):
		return p.parseGMLMultiPolygon(decoder)
	default:
		// Try to find nested geometry
		return p.parseNestedGeometry(decoder)
	}
}

func (p *GMLParser) parseGMLPoint(decoder *xml.Decoder) (*Geometry, error) {
	var coords []float64

	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}

		switch t := token.(type) {
		case xml.StartElement:
			if t.Name.Local == "pos" || t.Name.Local == "coordinates" {
				coordsStr, _ := getElementText(decoder)
				coords = parseCoordinates(coordsStr)
			}
		case xml.EndElement:
			if strings.Contains(strings.ToLower(t.Name.Local), "point") {
				if len(coords) >= 2 {
					return NewPoint(coords[0], coords[1]), nil
				}
				return nil, fmt.Errorf("invalid point coordinates")
			}
		}
	}

	return nil, fmt.Errorf("failed to parse point")
}

func (p *GMLParser) parseGMLLineString(decoder *xml.Decoder) (*Geometry, error) {
	var coords [][]float64

	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}

		switch t := token.(type) {
		case xml.StartElement:
			if t.Name.Local == "posList" || t.Name.Local == "coordinates" {
				coordsStr, _ := getElementText(decoder)
				coords = parseCoordinateList(coordsStr)
			}
		case xml.EndElement:
			if strings.Contains(strings.ToLower(t.Name.Local), "linestring") {
				return NewLineString(coords), nil
			}
		}
	}

	return nil, fmt.Errorf("failed to parse linestring")
}

func (p *GMLParser) parseGMLPolygon(decoder *xml.Decoder) (*Geometry, error) {
	var rings [][][]float64

	depth := 1
	for depth > 0 {
		token, err := decoder.Token()
		if err != nil {
			break
		}

		switch t := token.(type) {
		case xml.StartElement:
			depth++
			if t.Name.Local == "posList" || t.Name.Local == "coordinates" {
				coordsStr, _ := getElementText(decoder)
				ring := parseCoordinateList(coordsStr)
				if len(ring) > 0 {
					rings = append(rings, ring)
				}
			}
		case xml.EndElement:
			depth--
			if depth == 0 {
				return NewPolygon(rings), nil
			}
		}
	}

	return nil, fmt.Errorf("failed to parse polygon")
}

func (p *GMLParser) parseGMLMultiPoint(decoder *xml.Decoder) (*Geometry, error) {
	var coords [][]float64

	depth := 1
	for depth > 0 {
		token, err := decoder.Token()
		if err != nil {
			break
		}

		switch t := token.(type) {
		case xml.StartElement:
			depth++
			if t.Name.Local == "pos" {
				coordsStr, _ := getElementText(decoder)
				point := parseCoordinates(coordsStr)
				if len(point) >= 2 {
					coords = append(coords, point[:2])
				}
			}
		case xml.EndElement:
			depth--
			if depth == 0 {
				return NewMultiPoint(coords), nil
			}
		}
	}

	return nil, fmt.Errorf("failed to parse multipoint")
}

func (p *GMLParser) parseGMLMultiLineString(decoder *xml.Decoder) (*Geometry, error) {
	var lines [][][]float64

	depth := 1
	for depth > 0 {
		token, err := decoder.Token()
		if err != nil {
			break
		}

		switch t := token.(type) {
		case xml.StartElement:
			depth++
			if t.Name.Local == "posList" || t.Name.Local == "coordinates" {
				coordsStr, _ := getElementText(decoder)
				line := parseCoordinateList(coordsStr)
				if len(line) > 0 {
					lines = append(lines, line)
				}
			}
		case xml.EndElement:
			depth--
			if depth == 0 {
				return NewMultiLineString(lines), nil
			}
		}
	}

	return nil, fmt.Errorf("failed to parse multilinestring")
}

func (p *GMLParser) parseGMLMultiPolygon(decoder *xml.Decoder) (*Geometry, error) {
	var polygons [][][][]float64
	var currentPolygon [][][]float64

	depth := 1
	for depth > 0 {
		token, err := decoder.Token()
		if err != nil {
			break
		}

		switch t := token.(type) {
		case xml.StartElement:
			depth++
			if strings.Contains(strings.ToLower(t.Name.Local), "polygon") {
				currentPolygon = nil
			}
			if t.Name.Local == "posList" || t.Name.Local == "coordinates" {
				coordsStr, _ := getElementText(decoder)
				ring := parseCoordinateList(coordsStr)
				if len(ring) > 0 {
					currentPolygon = append(currentPolygon, ring)
				}
			}
		case xml.EndElement:
			depth--
			if strings.Contains(strings.ToLower(t.Name.Local), "polygon") && len(currentPolygon) > 0 {
				polygons = append(polygons, currentPolygon)
				currentPolygon = nil
			}
			if depth == 0 {
				return NewMultiPolygon(polygons), nil
			}
		}
	}

	return nil, fmt.Errorf("failed to parse multipolygon")
}

func (p *GMLParser) parseNestedGeometry(decoder *xml.Decoder) (*Geometry, error) {
	depth := 1
	for depth > 0 {
		token, err := decoder.Token()
		if err != nil {
			break
		}

		switch t := token.(type) {
		case xml.StartElement:
			depth++
			if isGMLGeometryElement(t.Name.Local) {
				return p.parseGMLGeometry(decoder, t)
			}
		case xml.EndElement:
			depth--
		}
	}

	return nil, fmt.Errorf("no geometry found")
}

func getElementText(decoder *xml.Decoder) (string, error) {
	var text strings.Builder

	for {
		token, err := decoder.Token()
		if err != nil {
			return text.String(), err
		}

		switch t := token.(type) {
		case xml.CharData:
			text.Write(t)
		case xml.EndElement:
			return text.String(), nil
		}
	}
}

func parseCoordinates(s string) []float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	// Handle comma-separated or space-separated
	s = strings.ReplaceAll(s, ",", " ")
	parts := strings.Fields(s)

	coords := make([]float64, 0, len(parts))
	for _, p := range parts {
		f, err := strconv.ParseFloat(p, 64)
		if err == nil {
			coords = append(coords, f)
		}
	}

	return coords
}

func parseCoordinateList(s string) [][]float64 {
	coords := parseCoordinates(s)
	if len(coords) < 2 {
		return nil
	}

	// Group into pairs (or triples if 3D)
	dimension := 2
	if len(coords)%3 == 0 && len(coords)%2 != 0 {
		dimension = 3
	}

	result := make([][]float64, 0, len(coords)/dimension)
	for i := 0; i+dimension <= len(coords); i += dimension {
		point := make([]float64, dimension)
		copy(point, coords[i:i+dimension])
		result = append(result, point)
	}

	return result
}

func parseValue(s string) interface{} {
	// Try to parse as number
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	// Try to parse as bool
	if strings.EqualFold(s, "true") {
		return true
	}
	if strings.EqualFold(s, "false") {
		return false
	}
	// Return as string
	return s
}
