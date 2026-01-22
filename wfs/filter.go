package wfs

import (
	"fmt"
	"strings"
)

// Filter represents an OGC Filter for WFS queries.
type Filter interface {
	// ToXML returns the XML representation of the filter.
	ToXML() string
	// ToCQL returns the CQL representation of the filter (if supported).
	ToCQL() string
}

// baseFilter provides common filter functionality.
type baseFilter struct{}

// PropertyIsEqualTo creates an equality filter.
func PropertyIsEqualTo(property string, value interface{}) Filter {
	return &comparisonFilter{
		operator: "PropertyIsEqualTo",
		property: property,
		value:    value,
	}
}

// PropertyEquals is an alias for PropertyIsEqualTo.
func PropertyEquals(property string, value interface{}) Filter {
	return PropertyIsEqualTo(property, value)
}

// PropertyIsNotEqualTo creates an inequality filter.
func PropertyIsNotEqualTo(property string, value interface{}) Filter {
	return &comparisonFilter{
		operator: "PropertyIsNotEqualTo",
		property: property,
		value:    value,
	}
}

// PropertyIsLessThan creates a less-than filter.
func PropertyIsLessThan(property string, value interface{}) Filter {
	return &comparisonFilter{
		operator: "PropertyIsLessThan",
		property: property,
		value:    value,
	}
}

// PropertyLessThan is an alias for PropertyIsLessThan.
func PropertyLessThan(property string, value interface{}) Filter {
	return PropertyIsLessThan(property, value)
}

// PropertyIsLessThanOrEqualTo creates a less-than-or-equal filter.
func PropertyIsLessThanOrEqualTo(property string, value interface{}) Filter {
	return &comparisonFilter{
		operator: "PropertyIsLessThanOrEqualTo",
		property: property,
		value:    value,
	}
}

// PropertyIsGreaterThan creates a greater-than filter.
func PropertyIsGreaterThan(property string, value interface{}) Filter {
	return &comparisonFilter{
		operator: "PropertyIsGreaterThan",
		property: property,
		value:    value,
	}
}

// PropertyGreaterThan is an alias for PropertyIsGreaterThan.
func PropertyGreaterThan(property string, value interface{}) Filter {
	return PropertyIsGreaterThan(property, value)
}

// PropertyIsGreaterThanOrEqualTo creates a greater-than-or-equal filter.
func PropertyIsGreaterThanOrEqualTo(property string, value interface{}) Filter {
	return &comparisonFilter{
		operator: "PropertyIsGreaterThanOrEqualTo",
		property: property,
		value:    value,
	}
}

// PropertyIsLike creates a pattern matching filter.
func PropertyIsLike(property, pattern string) Filter {
	return &likeFilter{
		property:   property,
		pattern:    pattern,
		wildCard:   "*",
		singleChar: "?",
		escapeChar: "\\",
	}
}

// PropertyLike is an alias for PropertyIsLike.
func PropertyLike(property, pattern string) Filter {
	return PropertyIsLike(property, pattern)
}

// PropertyIsNull creates a null check filter.
func PropertyIsNull(property string) Filter {
	return &nullFilter{property: property}
}

// PropertyIsBetween creates a between filter.
func PropertyIsBetween(property string, lower, upper interface{}) Filter {
	return &betweenFilter{
		property: property,
		lower:    lower,
		upper:    upper,
	}
}

// PropertyBetween is an alias for PropertyIsBetween.
func PropertyBetween(property string, lower, upper interface{}) Filter {
	return PropertyIsBetween(property, lower, upper)
}

// And creates a logical AND filter.
func And(filters ...Filter) Filter {
	return &logicalFilter{
		operator: "And",
		filters:  filters,
	}
}

// Or creates a logical OR filter.
func Or(filters ...Filter) Filter {
	return &logicalFilter{
		operator: "Or",
		filters:  filters,
	}
}

// Not creates a logical NOT filter.
func Not(filter Filter) Filter {
	return &notFilter{filter: filter}
}

// BBOX creates a bounding box spatial filter.
func BBOXFilter(geometryProperty string, minX, minY, maxX, maxY float64, srs string) Filter {
	return &bboxFilter{
		property: geometryProperty,
		minX:     minX,
		minY:     minY,
		maxX:     maxX,
		maxY:     maxY,
		srs:      srs,
	}
}

// Intersects creates a spatial intersects filter.
func Intersects(geometryProperty string, geometry interface{}) Filter {
	return &spatialFilter{
		operator: "Intersects",
		property: geometryProperty,
		geometry: geometry,
	}
}

// Contains creates a spatial contains filter.
func Contains(geometryProperty string, geometry interface{}) Filter {
	return &spatialFilter{
		operator: "Contains",
		property: geometryProperty,
		geometry: geometry,
	}
}

// Within creates a spatial within filter.
func Within(geometryProperty string, geometry interface{}) Filter {
	return &spatialFilter{
		operator: "Within",
		property: geometryProperty,
		geometry: geometry,
	}
}

// DWithin creates a distance within filter.
func DWithin(geometryProperty string, geometry interface{}, distance float64, units string) Filter {
	return &distanceFilter{
		operator: "DWithin",
		property: geometryProperty,
		geometry: geometry,
		distance: distance,
		units:    units,
	}
}

// ResourceID creates a resource ID filter.
func ResourceIDFilter(ids ...string) Filter {
	return &resourceIDFilter{ids: ids}
}

// comparisonFilter implements comparison operators.
type comparisonFilter struct {
	operator string
	property string
	value    interface{}
}

func (f *comparisonFilter) ToXML() string {
	return fmt.Sprintf(`<fes:Filter xmlns:fes="http://www.opengis.net/fes/2.0">
  <fes:%s>
    <fes:ValueReference>%s</fes:ValueReference>
    <fes:Literal>%v</fes:Literal>
  </fes:%s>
</fes:Filter>`, f.operator, escapeXML(f.property), escapeXML(fmt.Sprintf("%v", f.value)), f.operator)
}

func (f *comparisonFilter) ToCQL() string {
	val := f.value
	if s, ok := val.(string); ok {
		val = "'" + strings.ReplaceAll(s, "'", "''") + "'"
	}

	op := "="
	switch f.operator {
	case "PropertyIsNotEqualTo":
		op = "<>"
	case "PropertyIsLessThan":
		op = "<"
	case "PropertyIsLessThanOrEqualTo":
		op = "<="
	case "PropertyIsGreaterThan":
		op = ">"
	case "PropertyIsGreaterThanOrEqualTo":
		op = ">="
	}

	return fmt.Sprintf("%s %s %v", f.property, op, val)
}

// likeFilter implements pattern matching.
type likeFilter struct {
	property   string
	pattern    string
	wildCard   string
	singleChar string
	escapeChar string
}

func (f *likeFilter) ToXML() string {
	return fmt.Sprintf(`<fes:Filter xmlns:fes="http://www.opengis.net/fes/2.0">
  <fes:PropertyIsLike wildCard="%s" singleChar="%s" escapeChar="%s">
    <fes:ValueReference>%s</fes:ValueReference>
    <fes:Literal>%s</fes:Literal>
  </fes:PropertyIsLike>
</fes:Filter>`, f.wildCard, f.singleChar, f.escapeChar, escapeXML(f.property), escapeXML(f.pattern))
}

func (f *likeFilter) ToCQL() string {
	// Convert pattern to SQL-like pattern
	pattern := f.pattern
	pattern = strings.ReplaceAll(pattern, "*", "%")
	pattern = strings.ReplaceAll(pattern, "?", "_")
	return fmt.Sprintf("%s LIKE '%s'", f.property, strings.ReplaceAll(pattern, "'", "''"))
}

// nullFilter implements null check.
type nullFilter struct {
	property string
}

func (f *nullFilter) ToXML() string {
	return fmt.Sprintf(`<fes:Filter xmlns:fes="http://www.opengis.net/fes/2.0">
  <fes:PropertyIsNull>
    <fes:ValueReference>%s</fes:ValueReference>
  </fes:PropertyIsNull>
</fes:Filter>`, escapeXML(f.property))
}

func (f *nullFilter) ToCQL() string {
	return fmt.Sprintf("%s IS NULL", f.property)
}

// betweenFilter implements between check.
type betweenFilter struct {
	property string
	lower    interface{}
	upper    interface{}
}

func (f *betweenFilter) ToXML() string {
	return fmt.Sprintf(`<fes:Filter xmlns:fes="http://www.opengis.net/fes/2.0">
  <fes:PropertyIsBetween>
    <fes:ValueReference>%s</fes:ValueReference>
    <fes:LowerBoundary><fes:Literal>%v</fes:Literal></fes:LowerBoundary>
    <fes:UpperBoundary><fes:Literal>%v</fes:Literal></fes:UpperBoundary>
  </fes:PropertyIsBetween>
</fes:Filter>`, escapeXML(f.property), f.lower, f.upper)
}

func (f *betweenFilter) ToCQL() string {
	return fmt.Sprintf("%s BETWEEN %v AND %v", f.property, f.lower, f.upper)
}

// logicalFilter implements AND/OR.
type logicalFilter struct {
	operator string
	filters  []Filter
}

func (f *logicalFilter) ToXML() string {
	var inner strings.Builder
	for _, filter := range f.filters {
		// Extract the inner filter content (without the Filter wrapper)
		xml := filter.ToXML()
		// This is a simplification - in production you'd parse and extract properly
		inner.WriteString(extractFilterContent(xml))
	}

	return fmt.Sprintf(`<fes:Filter xmlns:fes="http://www.opengis.net/fes/2.0">
  <fes:%s>
%s  </fes:%s>
</fes:Filter>`, f.operator, inner.String(), f.operator)
}

func (f *logicalFilter) ToCQL() string {
	parts := make([]string, len(f.filters))
	for i, filter := range f.filters {
		parts[i] = "(" + filter.ToCQL() + ")"
	}
	return strings.Join(parts, " "+strings.ToUpper(f.operator)+" ")
}

// notFilter implements NOT.
type notFilter struct {
	filter Filter
}

func (f *notFilter) ToXML() string {
	inner := extractFilterContent(f.filter.ToXML())
	return fmt.Sprintf(`<fes:Filter xmlns:fes="http://www.opengis.net/fes/2.0">
  <fes:Not>
%s  </fes:Not>
</fes:Filter>`, inner)
}

func (f *notFilter) ToCQL() string {
	return "NOT (" + f.filter.ToCQL() + ")"
}

// bboxFilter implements bounding box spatial filter.
type bboxFilter struct {
	property string
	minX     float64
	minY     float64
	maxX     float64
	maxY     float64
	srs      string
}

func (f *bboxFilter) ToXML() string {
	srsAttr := ""
	if f.srs != "" {
		srsAttr = fmt.Sprintf(` srsName="%s"`, f.srs)
	}

	return fmt.Sprintf(`<fes:Filter xmlns:fes="http://www.opengis.net/fes/2.0" xmlns:gml="http://www.opengis.net/gml/3.2">
  <fes:BBOX>
    <fes:ValueReference>%s</fes:ValueReference>
    <gml:Envelope%s>
      <gml:lowerCorner>%f %f</gml:lowerCorner>
      <gml:upperCorner>%f %f</gml:upperCorner>
    </gml:Envelope>
  </fes:BBOX>
</fes:Filter>`, escapeXML(f.property), srsAttr, f.minX, f.minY, f.maxX, f.maxY)
}

func (f *bboxFilter) ToCQL() string {
	return fmt.Sprintf("BBOX(%s,%f,%f,%f,%f)", f.property, f.minX, f.minY, f.maxX, f.maxY)
}

// spatialFilter implements spatial operators.
type spatialFilter struct {
	operator string
	property string
	geometry interface{}
}

func (f *spatialFilter) ToXML() string {
	geomXML := geometryToGML(f.geometry)
	return fmt.Sprintf(`<fes:Filter xmlns:fes="http://www.opengis.net/fes/2.0" xmlns:gml="http://www.opengis.net/gml/3.2">
  <fes:%s>
    <fes:ValueReference>%s</fes:ValueReference>
    %s
  </fes:%s>
</fes:Filter>`, f.operator, escapeXML(f.property), geomXML, f.operator)
}

func (f *spatialFilter) ToCQL() string {
	geomWKT := geometryToWKT(f.geometry)
	return fmt.Sprintf("%s(%s, %s)", strings.ToUpper(f.operator), f.property, geomWKT)
}

// distanceFilter implements distance-based spatial operators.
type distanceFilter struct {
	operator string
	property string
	geometry interface{}
	distance float64
	units    string
}

func (f *distanceFilter) ToXML() string {
	geomXML := geometryToGML(f.geometry)
	return fmt.Sprintf(`<fes:Filter xmlns:fes="http://www.opengis.net/fes/2.0" xmlns:gml="http://www.opengis.net/gml/3.2">
  <fes:%s>
    <fes:ValueReference>%s</fes:ValueReference>
    %s
    <fes:Distance uom="%s">%f</fes:Distance>
  </fes:%s>
</fes:Filter>`, f.operator, escapeXML(f.property), geomXML, f.units, f.distance, f.operator)
}

func (f *distanceFilter) ToCQL() string {
	geomWKT := geometryToWKT(f.geometry)
	return fmt.Sprintf("DWITHIN(%s, %s, %f, %s)", f.property, geomWKT, f.distance, f.units)
}

// resourceIDFilter implements resource ID filter.
type resourceIDFilter struct {
	ids []string
}

func (f *resourceIDFilter) ToXML() string {
	var ids strings.Builder
	for _, id := range f.ids {
		ids.WriteString(fmt.Sprintf(`    <fes:ResourceId rid="%s"/>`+"\n", escapeXML(id)))
	}
	return fmt.Sprintf(`<fes:Filter xmlns:fes="http://www.opengis.net/fes/2.0">
%s</fes:Filter>`, ids.String())
}

func (f *resourceIDFilter) ToCQL() string {
	quoted := make([]string, len(f.ids))
	for i, id := range f.ids {
		quoted[i] = "'" + strings.ReplaceAll(id, "'", "''") + "'"
	}
	return "IN (" + strings.Join(quoted, ", ") + ")"
}

// CQLFilter wraps a raw CQL string as a Filter.
type CQLFilter string

// ToXML returns empty since CQL is used directly.
func (f CQLFilter) ToXML() string {
	return ""
}

// ToCQL returns the CQL string.
func (f CQLFilter) ToCQL() string {
	return string(f)
}

// escapeXML escapes special XML characters.
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}

// extractFilterContent extracts the inner content from a filter XML.
func extractFilterContent(xml string) string {
	// Find content between <fes:Filter...> and </fes:Filter>
	start := strings.Index(xml, ">")
	end := strings.LastIndex(xml, "</fes:Filter>")
	if start == -1 || end == -1 || start >= end {
		return xml
	}
	content := xml[start+1 : end]
	// Add indentation
	lines := strings.Split(content, "\n")
	var result strings.Builder
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			result.WriteString("    " + strings.TrimPrefix(line, "  ") + "\n")
		}
	}
	return result.String()
}

// geometryToGML converts a geometry to GML format.
func geometryToGML(geom interface{}) string {
	switch g := geom.(type) {
	case string:
		// Assume it's already GML or WKT
		if strings.HasPrefix(g, "<") {
			return g
		}
		// Convert WKT to GML
		return wktToGML(g)
	case Point:
		return fmt.Sprintf(`<gml:Point srsName="%s"><gml:pos>%f %f</gml:pos></gml:Point>`,
			g.SRS, g.X, g.Y)
	default:
		return fmt.Sprintf("%v", geom)
	}
}

// geometryToWKT converts a geometry to WKT format.
func geometryToWKT(geom interface{}) string {
	switch g := geom.(type) {
	case string:
		return g
	case Point:
		return fmt.Sprintf("POINT(%f %f)", g.X, g.Y)
	default:
		return fmt.Sprintf("%v", geom)
	}
}

// wktToGML converts WKT to GML (basic implementation).
func wktToGML(wkt string) string {
	wkt = strings.TrimSpace(wkt)
	upper := strings.ToUpper(wkt)

	if strings.HasPrefix(upper, "POINT") {
		coords := extractWKTCoords(wkt)
		return fmt.Sprintf(`<gml:Point><gml:pos>%s</gml:pos></gml:Point>`, coords)
	}

	// For other types, return as-is (would need full parser for production)
	return wkt
}

// extractWKTCoords extracts coordinates from WKT.
func extractWKTCoords(wkt string) string {
	start := strings.Index(wkt, "(")
	end := strings.LastIndex(wkt, ")")
	if start == -1 || end == -1 {
		return ""
	}
	coords := wkt[start+1 : end]
	// Replace commas with spaces for GML
	return strings.ReplaceAll(coords, ",", " ")
}

// Point represents a point geometry.
type Point struct {
	X   float64
	Y   float64
	SRS string
}

// NewPoint creates a new Point.
func NewPoint(x, y float64, srs string) Point {
	return Point{X: x, Y: y, SRS: srs}
}
