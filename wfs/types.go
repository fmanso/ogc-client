package wfs

import (
	"encoding/xml"
	"strings"
)

// Output format constants.
const (
	FormatGML3      = "application/gml+xml; version=3.2"
	FormatGML2      = "GML2"
	FormatGeoJSON   = "application/json"
	FormatShapefile = "SHAPE-ZIP"
	FormatCSV       = "csv"
	FormatKML       = "application/vnd.google-earth.kml+xml"
)

// FeatureType represents a WFS feature type.
type FeatureType struct {
	Name             string
	Title            string
	Abstract         string
	Keywords         []string
	DefaultCRS       string
	OtherCRS         []string
	OutputFormats    []string
	WGS84BoundingBox *BBox
	MetadataURL      string
}

// BBox represents a bounding box.
type BBox struct {
	LowerCorner [2]float64 // MinX, MinY
	UpperCorner [2]float64 // MaxX, MaxY
	CRS         string
}

// MinX returns the minimum X coordinate.
func (b *BBox) MinX() float64 { return b.LowerCorner[0] }

// MinY returns the minimum Y coordinate.
func (b *BBox) MinY() float64 { return b.LowerCorner[1] }

// MaxX returns the maximum X coordinate.
func (b *BBox) MaxX() float64 { return b.UpperCorner[0] }

// MaxY returns the maximum Y coordinate.
func (b *BBox) MaxY() float64 { return b.UpperCorner[1] }

// Contains checks if a point is within the bounding box.
func (b *BBox) Contains(x, y float64) bool {
	return x >= b.MinX() && x <= b.MaxX() && y >= b.MinY() && y <= b.MaxY()
}

// Capabilities represents the WFS GetCapabilities response.
type Capabilities struct {
	XMLName            xml.Name              `xml:"WFS_Capabilities"`
	Version            string                `xml:"version,attr"`
	ServiceID          ServiceIdentification `xml:"ServiceIdentification"`
	ServiceProvider    *ServiceProvider      `xml:"ServiceProvider"`
	OperationsMetadata *OperationsMetadata   `xml:"OperationsMetadata"`
	FeatureTypeList    *FeatureTypeList      `xml:"FeatureTypeList"`
	FilterCapabilities *FilterCapabilities   `xml:"Filter_Capabilities"`
}

// ServiceIdentification contains service metadata.
type ServiceIdentification struct {
	Title              string   `xml:"Title"`
	Abstract           string   `xml:"Abstract"`
	Keywords           []string `xml:"Keywords>Keyword"`
	ServiceType        string   `xml:"ServiceType"`
	ServiceTypeVersion string   `xml:"ServiceTypeVersion"`
	Fees               string   `xml:"Fees"`
	AccessConstraints  string   `xml:"AccessConstraints"`
}

// ServiceProvider contains provider information.
type ServiceProvider struct {
	ProviderName   string          `xml:"ProviderName"`
	ProviderSite   string          `xml:"ProviderSite>href,attr"`
	ServiceContact *ServiceContact `xml:"ServiceContact"`
}

// ServiceContact contains contact information.
type ServiceContact struct {
	IndividualName string `xml:"IndividualName"`
	PositionName   string `xml:"PositionName"`
	Phone          string `xml:"ContactInfo>Phone>Voice"`
	Email          string `xml:"ContactInfo>Address>ElectronicMailAddress"`
}

// OperationsMetadata describes available operations.
type OperationsMetadata struct {
	Operations  []Operation  `xml:"Operation"`
	Parameters  []Parameter  `xml:"Parameter"`
	Constraints []Constraint `xml:"Constraint"`
}

// Operation describes a single operation.
type Operation struct {
	Name       string      `xml:"name,attr"`
	DCPs       []DCP       `xml:"DCP"`
	Parameters []Parameter `xml:"Parameter"`
}

// DCP describes the distributed computing platform.
type DCP struct {
	HTTP *DCPHttp `xml:"HTTP"`
}

// DCPHttp contains HTTP endpoint URLs.
type DCPHttp struct {
	Get  string `xml:"Get>href,attr"`
	Post string `xml:"Post>href,attr"`
}

// Parameter describes an operation parameter.
type Parameter struct {
	Name          string   `xml:"name,attr"`
	AllowedValues []string `xml:"AllowedValues>Value"`
	DefaultValue  string   `xml:"DefaultValue"`
}

// Constraint describes a constraint.
type Constraint struct {
	Name          string   `xml:"name,attr"`
	DefaultValue  string   `xml:"DefaultValue"`
	AllowedValues []string `xml:"AllowedValues>Value"`
}

// FeatureTypeList contains the list of feature types.
type FeatureTypeList struct {
	FeatureTypes []XMLFeatureType `xml:"FeatureType"`
}

// XMLFeatureType is the XML representation of a feature type.
type XMLFeatureType struct {
	Name             string          `xml:"Name"`
	Title            string          `xml:"Title"`
	Abstract         string          `xml:"Abstract"`
	Keywords         []string        `xml:"Keywords>Keyword"`
	DefaultCRS       string          `xml:"DefaultCRS"`
	OtherCRS         []string        `xml:"OtherCRS"`
	OutputFormats    []string        `xml:"OutputFormats>Format"`
	WGS84BoundingBox *XMLBoundingBox `xml:"WGS84BoundingBox"`
	MetadataURL      *XMLMetadataURL `xml:"MetadataURL"`
}

// XMLBoundingBox is the XML representation of a bounding box.
type XMLBoundingBox struct {
	LowerCorner string `xml:"LowerCorner"`
	UpperCorner string `xml:"UpperCorner"`
}

// XMLMetadataURL is the XML representation of metadata URL.
type XMLMetadataURL struct {
	Type string `xml:"type,attr"`
	Href string `xml:"href,attr"`
}

// FilterCapabilities describes filter support.
type FilterCapabilities struct {
	Conformance         *Conformance         `xml:"Conformance"`
	ScalarCapabilities  *ScalarCapabilities  `xml:"Scalar_Capabilities"`
	SpatialCapabilities *SpatialCapabilities `xml:"Spatial_Capabilities"`
}

// Conformance describes conformance classes.
type Conformance struct {
	Constraints []Constraint `xml:"Constraint"`
}

// ScalarCapabilities describes scalar filter support.
type ScalarCapabilities struct {
	LogicalOperators    *LogicalOperators    `xml:"LogicalOperators"`
	ComparisonOperators *ComparisonOperators `xml:"ComparisonOperators"`
}

// LogicalOperators describes supported logical operators.
type LogicalOperators struct {
	// Presence indicates support for logical operators
}

// ComparisonOperators describes supported comparison operators.
type ComparisonOperators struct {
	Operators []string `xml:"ComparisonOperator>name,attr"`
}

// SpatialCapabilities describes spatial filter support.
type SpatialCapabilities struct {
	GeometryOperands []string `xml:"GeometryOperands>GeometryOperand>name,attr"`
	SpatialOperators []string `xml:"SpatialOperators>SpatialOperator>name,attr"`
}

// GetFeatureTypes returns all feature types.
func (c *Capabilities) GetFeatureTypes() []FeatureType {
	if c.FeatureTypeList == nil {
		return nil
	}

	types := make([]FeatureType, 0, len(c.FeatureTypeList.FeatureTypes))
	for _, xft := range c.FeatureTypeList.FeatureTypes {
		ft := FeatureType{
			Name:          xft.Name,
			Title:         xft.Title,
			Abstract:      xft.Abstract,
			Keywords:      xft.Keywords,
			DefaultCRS:    xft.DefaultCRS,
			OtherCRS:      xft.OtherCRS,
			OutputFormats: xft.OutputFormats,
		}

		if xft.WGS84BoundingBox != nil {
			ft.WGS84BoundingBox = parseBoundingBox(xft.WGS84BoundingBox, "EPSG:4326")
		}

		if xft.MetadataURL != nil {
			ft.MetadataURL = xft.MetadataURL.Href
		}

		types = append(types, ft)
	}

	return types
}

// GetFeatureType finds a feature type by name.
func (c *Capabilities) GetFeatureType(name string) *FeatureType {
	types := c.GetFeatureTypes()
	for i := range types {
		if types[i].Name == name {
			return &types[i]
		}
	}
	return nil
}

// GetFeatureTypeNames returns all feature type names.
func (c *Capabilities) GetFeatureTypeNames() []string {
	types := c.GetFeatureTypes()
	names := make([]string, len(types))
	for i, t := range types {
		names[i] = t.Name
	}
	return names
}

// SupportsOperation checks if an operation is supported.
func (c *Capabilities) SupportsOperation(name string) bool {
	if c.OperationsMetadata == nil {
		return false
	}
	for _, op := range c.OperationsMetadata.Operations {
		if strings.EqualFold(op.Name, name) {
			return true
		}
	}
	return false
}

// SupportsOutputFormat checks if a given output format is supported.
func (c *Capabilities) SupportsOutputFormat(format string) bool {
	if c.OperationsMetadata == nil {
		return false
	}
	for _, param := range c.OperationsMetadata.Parameters {
		if param.Name == "outputFormat" {
			for _, v := range param.AllowedValues {
				if strings.EqualFold(v, format) {
					return true
				}
			}
		}
	}
	return false
}

// parseBoundingBox parses an XML bounding box.
func parseBoundingBox(xbb *XMLBoundingBox, crs string) *BBox {
	if xbb == nil {
		return nil
	}

	bb := &BBox{CRS: crs}

	// Parse lower corner
	parts := strings.Fields(xbb.LowerCorner)
	if len(parts) >= 2 {
		var minX, minY float64
		_, _ = parseFloat(parts[0], &minX)
		_, _ = parseFloat(parts[1], &minY)
		bb.LowerCorner = [2]float64{minX, minY}
	}

	// Parse upper corner
	parts = strings.Fields(xbb.UpperCorner)
	if len(parts) >= 2 {
		var maxX, maxY float64
		_, _ = parseFloat(parts[0], &maxX)
		_, _ = parseFloat(parts[1], &maxY)
		bb.UpperCorner = [2]float64{maxX, maxY}
	}

	return bb
}

// parseFloat parses a float, returning success status.
func parseFloat(s string, out *float64) (bool, error) {
	var err error
	*out, err = parseFloatValue(s)
	return err == nil, err
}

// parseFloatValue parses a float value from string.
func parseFloatValue(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}

	var f float64
	_, err := strings.NewReader(s).Read(nil)
	if err != nil {
		return 0, err
	}

	// Use simple parsing
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '-' || c == '.' || (c >= '0' && c <= '9') {
			continue
		}
		if c == 'e' || c == 'E' {
			continue
		}
		if c == '+' {
			continue
		}
		s = s[:i]
		break
	}

	_, scanErr := strings.NewReader(s).Read(nil)
	if scanErr != nil {
		return 0, scanErr
	}

	// Actually parse the float
	fmt := strings.TrimSpace(s)
	if fmt == "" {
		return 0, nil
	}

	// Use strconv via reflection-free method
	isNegative := false
	if len(fmt) > 0 && fmt[0] == '-' {
		isNegative = true
		fmt = fmt[1:]
	}

	parts := strings.Split(fmt, ".")
	intPart := int64(0)
	for _, c := range parts[0] {
		if c >= '0' && c <= '9' {
			intPart = intPart*10 + int64(c-'0')
		}
	}

	f = float64(intPart)

	if len(parts) > 1 {
		fracPart := float64(0)
		divisor := float64(1)
		for _, c := range parts[1] {
			if c >= '0' && c <= '9' {
				fracPart = fracPart*10 + float64(c-'0')
				divisor *= 10
			}
		}
		f += fracPart / divisor
	}

	if isNegative {
		f = -f
	}

	return f, nil
}

// DescribeFeatureTypeResponse represents the response from DescribeFeatureType.
type DescribeFeatureTypeResponse struct {
	Schema *XSDSchema
	Raw    []byte
}

// XSDSchema represents an XML Schema.
type XSDSchema struct {
	XMLName            xml.Name         `xml:"schema"`
	TargetNS           string           `xml:"targetNamespace,attr"`
	ElementFormDefault string           `xml:"elementFormDefault,attr"`
	Elements           []XSDElement     `xml:"element"`
	ComplexTypes       []XSDComplexType `xml:"complexType"`
}

// XSDElement represents an XSD element.
type XSDElement struct {
	Name              string `xml:"name,attr"`
	Type              string `xml:"type,attr"`
	SubstitutionGroup string `xml:"substitutionGroup,attr"`
	MinOccurs         string `xml:"minOccurs,attr"`
	MaxOccurs         string `xml:"maxOccurs,attr"`
	Nillable          string `xml:"nillable,attr"`
}

// XSDComplexType represents an XSD complex type.
type XSDComplexType struct {
	Name           string             `xml:"name,attr"`
	ComplexContent *XSDComplexContent `xml:"complexContent"`
	Sequence       *XSDSequence       `xml:"sequence"`
}

// XSDComplexContent represents XSD complex content.
type XSDComplexContent struct {
	Extension *XSDExtension `xml:"extension"`
}

// XSDSequence represents an XSD sequence.
type XSDSequence struct {
	Elements []XSDElement `xml:"element"`
}

// XSDExtension represents an XSD extension.
type XSDExtension struct {
	Base     string       `xml:"base,attr"`
	Sequence *XSDSequence `xml:"sequence"`
}

// Property represents a feature property.
type Property struct {
	Name       string
	Type       string
	MinOccurs  int
	MaxOccurs  int // -1 for unbounded
	Nillable   bool
	IsGeometry bool
}

// GetProperties extracts property definitions from the schema.
func (s *XSDSchema) GetProperties(typeName string) []Property {
	if s == nil {
		return nil
	}

	var props []Property

	// Find the complex type for this feature
	for _, ct := range s.ComplexTypes {
		if strings.HasSuffix(ct.Name, "Type") && strings.Contains(ct.Name, typeName) {
			var seq *XSDSequence
			if ct.Sequence != nil {
				seq = ct.Sequence
			} else if ct.ComplexContent != nil && ct.ComplexContent.Extension != nil && ct.ComplexContent.Extension.Sequence != nil {
				seq = ct.ComplexContent.Extension.Sequence
			}

			if seq != nil {
				for _, elem := range seq.Elements {
					prop := Property{
						Name:      elem.Name,
						Type:      elem.Type,
						Nillable:  elem.Nillable == "true",
						MinOccurs: parseOccurs(elem.MinOccurs, 1),
						MaxOccurs: parseMaxOccurs(elem.MaxOccurs),
					}
					prop.IsGeometry = isGeometryType(elem.Type)
					props = append(props, prop)
				}
			}
			break
		}
	}

	return props
}

// parseOccurs parses minOccurs/maxOccurs.
func parseOccurs(s string, def int) int {
	if s == "" {
		return def
	}
	val := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			val = val*10 + int(c-'0')
		}
	}
	return val
}

// parseMaxOccurs parses maxOccurs, returning -1 for unbounded.
func parseMaxOccurs(s string) int {
	if s == "unbounded" {
		return -1
	}
	return parseOccurs(s, 1)
}

// isGeometryType checks if a type is a geometry type.
func isGeometryType(t string) bool {
	t = strings.ToLower(t)
	geomTypes := []string{
		"geometry", "point", "linestring", "polygon",
		"multipoint", "multilinestring", "multipolygon",
		"geometrycollection", "curve", "surface", "multisurface",
		"multicurve", "geometrypropertytype",
	}
	for _, gt := range geomTypes {
		if strings.Contains(t, gt) {
			return true
		}
	}
	return false
}

// TransactionResponse represents the response from a WFS-T Transaction.
type TransactionResponse struct {
	XMLName            xml.Name           `xml:"TransactionResponse"`
	Version            string             `xml:"version,attr"`
	TransactionSummary TransactionSummary `xml:"TransactionSummary"`
	InsertResults      []InsertResult     `xml:"InsertResults>Feature"`
}

// TransactionSummary summarizes a transaction.
type TransactionSummary struct {
	TotalInserted int `xml:"totalInserted"`
	TotalUpdated  int `xml:"totalUpdated"`
	TotalDeleted  int `xml:"totalDeleted"`
}

// InsertResult contains info about inserted features.
type InsertResult struct {
	Handle     string `xml:"handle,attr"`
	ResourceID string `xml:"ResourceId>rid,attr"`
}

// GetPropertyValueResponse represents the response from GetPropertyValue.
type GetPropertyValueResponse struct {
	Values []PropertyValue
}

// PropertyValue represents a single property value.
type PropertyValue struct {
	Value interface{}
}

// WFSError represents a WFS service exception.
type WFSError struct {
	Code       string
	Message    string
	Locator    string
	StatusCode int
}

// Error returns the error message.
func (e *WFSError) Error() string {
	msg := "wfs: "
	if e.Code != "" {
		msg += "[" + e.Code + "] "
	}
	msg += e.Message
	if e.Locator != "" {
		msg += " (locator: " + e.Locator + ")"
	}
	return msg
}
