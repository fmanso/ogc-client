package wms

import (
	"encoding/xml"
	"fmt"
	"strings"
)

// BBox represents a bounding box.
type BBox struct {
	MinX float64
	MinY float64
	MaxX float64
	MaxY float64
}

// String returns the BBOX string representation.
func (b *BBox) String() string {
	return fmt.Sprintf("%f,%f,%f,%f", b.MinX, b.MinY, b.MaxX, b.MaxY)
}

// Contains checks if a point is within the bounding box.
func (b *BBox) Contains(x, y float64) bool {
	return x >= b.MinX && x <= b.MaxX && y >= b.MinY && y <= b.MaxY
}

// Intersects checks if two bounding boxes intersect.
func (b *BBox) Intersects(other *BBox) bool {
	return b.MinX <= other.MaxX && b.MaxX >= other.MinX &&
		b.MinY <= other.MaxY && b.MaxY >= other.MinY
}

// Layer represents a WMS layer.
type Layer struct {
	Name        string
	Title       string
	Abstract    string
	Queryable   bool
	Opaque      bool
	CRS         []string
	BoundingBox map[string]*BBox
	Styles      []Style
	Layers      []Layer // Nested layers
	MinScale    float64
	MaxScale    float64
	Dimensions  []Dimension
	Attribution *Attribution
	MetadataURL []MetadataURL
}

// Style represents a layer style.
type Style struct {
	Name      string
	Title     string
	Abstract  string
	LegendURL *LegendURL
}

// LegendURL contains legend graphic information.
type LegendURL struct {
	Format string
	Width  int
	Height int
	URL    string
}

// Dimension represents a layer dimension (e.g., TIME, ELEVATION).
type Dimension struct {
	Name           string
	Units          string
	Default        string
	Values         string
	MultipleValues bool
	NearestValue   bool
}

// Attribution represents layer attribution.
type Attribution struct {
	Title    string
	URL      string
	LogoURL  string
	LogoType string
}

// MetadataURL represents a metadata URL.
type MetadataURL struct {
	Type   string
	Format string
	URL    string
}

// Capabilities represents the WMS GetCapabilities response.
type Capabilities struct {
	XMLName    xml.Name   `xml:"WMS_Capabilities"`
	Version    string     `xml:"version,attr"`
	Service    Service    `xml:"Service"`
	Capability Capability `xml:"Capability"`
}

// Service contains service metadata.
type Service struct {
	Name              string          `xml:"Name"`
	Title             string          `xml:"Title"`
	Abstract          string          `xml:"Abstract"`
	Keywords          []string        `xml:"KeywordList>Keyword"`
	OnlineResourceEl  *OnlineResource `xml:"OnlineResource"`
	ContactInfo       *ContactInfo    `xml:"ContactInformation"`
	Fees              string          `xml:"Fees"`
	AccessConstraints string          `xml:"AccessConstraints"`
	MaxWidth          int             `xml:"MaxWidth"`
	MaxHeight         int             `xml:"MaxHeight"`
}

// OnlineResource returns the online resource URL.
func (s *Service) OnlineResource() string {
	if s.OnlineResourceEl != nil {
		return s.OnlineResourceEl.Href
	}
	return ""
}

// OnlineResource represents an xlink:href reference.
type OnlineResource struct {
	Href string `xml:"href,attr"`
}

// ContactInfo contains contact information.
type ContactInfo struct {
	PersonPrimary *PersonPrimary `xml:"ContactPersonPrimary"`
	Position      string         `xml:"ContactPosition"`
	Address       *Address       `xml:"ContactAddress"`
	Phone         string         `xml:"ContactVoiceTelephone"`
	Fax           string         `xml:"ContactFacsimileTelephone"`
	Email         string         `xml:"ContactElectronicMailAddress"`
}

// PersonPrimary contains primary contact person details.
type PersonPrimary struct {
	Person       string `xml:"ContactPerson"`
	Organization string `xml:"ContactOrganization"`
}

// Address contains address information.
type Address struct {
	Type     string `xml:"AddressType"`
	Address  string `xml:"Address"`
	City     string `xml:"City"`
	State    string `xml:"StateOrProvince"`
	PostCode string `xml:"PostCode"`
	Country  string `xml:"Country"`
}

// Capability contains capability information.
type Capability struct {
	Request   RequestCapability `xml:"Request"`
	Exception Exception         `xml:"Exception"`
	Layer     *XMLLayer         `xml:"Layer"`
}

// RequestCapability lists supported operations.
type RequestCapability struct {
	GetCapabilities  *OperationType `xml:"GetCapabilities"`
	GetMap           *OperationType `xml:"GetMap"`
	GetFeatureInfo   *OperationType `xml:"GetFeatureInfo"`
	DescribeLayer    *OperationType `xml:"DescribeLayer"`
	GetLegendGraphic *OperationType `xml:"GetLegendGraphic"`
}

// OperationType describes an operation.
type OperationType struct {
	Formats []string  `xml:"Format"`
	DCPType []DCPType `xml:"DCPType"`
}

// DCPType describes HTTP methods.
type DCPType struct {
	HTTP HTTPD `xml:"HTTP"`
}

// HTTPD contains HTTP endpoint information.
type HTTPD struct {
	Get  *HTTPMethod `xml:"Get"`
	Post *HTTPMethod `xml:"Post"`
}

// HTTPMethod contains method-specific URL.
type HTTPMethod struct {
	OnlineResourceEl *OnlineResource `xml:"OnlineResource"`
}

// URL returns the online resource URL.
func (m *HTTPMethod) URL() string {
	if m.OnlineResourceEl != nil {
		return m.OnlineResourceEl.Href
	}
	return ""
}

// Exception lists supported exception formats.
type Exception struct {
	Formats []string `xml:"Format"`
}

// XMLLayer is the XML representation of a layer for unmarshaling.
type XMLLayer struct {
	Queryable     string            `xml:"queryable,attr"`
	Opaque        string            `xml:"opaque,attr"`
	Name          string            `xml:"Name"`
	Title         string            `xml:"Title"`
	Abstract      string            `xml:"Abstract"`
	CRS           []string          `xml:"CRS"`
	EXBoundingBox *XMLEXBoundingBox `xml:"EX_GeographicBoundingBox"`
	BoundingBoxes []XMLBoundingBox  `xml:"BoundingBox"`
	Styles        []XMLStyle        `xml:"Style"`
	Dimensions    []XMLDimension    `xml:"Dimension"`
	Attribution   *XMLAttribution   `xml:"Attribution"`
	MetadataURLs  []XMLMetadataURL  `xml:"MetadataURL"`
	MinScale      float64           `xml:"MinScaleDenominator"`
	MaxScale      float64           `xml:"MaxScaleDenominator"`
	Layers        []XMLLayer        `xml:"Layer"`
}

// XMLEXBoundingBox is the geographic bounding box.
type XMLEXBoundingBox struct {
	WestLon  float64 `xml:"westBoundLongitude"`
	EastLon  float64 `xml:"eastBoundLongitude"`
	SouthLat float64 `xml:"southBoundLatitude"`
	NorthLat float64 `xml:"northBoundLatitude"`
}

// XMLBoundingBox is a CRS-specific bounding box.
type XMLBoundingBox struct {
	CRS  string  `xml:"CRS,attr"`
	MinX float64 `xml:"minx,attr"`
	MinY float64 `xml:"miny,attr"`
	MaxX float64 `xml:"maxx,attr"`
	MaxY float64 `xml:"maxy,attr"`
}

// XMLStyle is the XML representation of a style.
type XMLStyle struct {
	Name      string        `xml:"Name"`
	Title     string        `xml:"Title"`
	Abstract  string        `xml:"Abstract"`
	LegendURL *XMLLegendURL `xml:"LegendURL"`
}

// XMLLegendURL is the XML representation of legend URL.
type XMLLegendURL struct {
	Width            int             `xml:"width,attr"`
	Height           int             `xml:"height,attr"`
	Format           string          `xml:"Format"`
	OnlineResourceEl *OnlineResource `xml:"OnlineResource"`
}

// URL returns the legend URL.
func (l *XMLLegendURL) URL() string {
	if l.OnlineResourceEl != nil {
		return l.OnlineResourceEl.Href
	}
	return ""
}

// XMLDimension is the XML representation of a dimension.
type XMLDimension struct {
	Name           string `xml:"name,attr"`
	Units          string `xml:"units,attr"`
	Default        string `xml:"default,attr"`
	MultipleValues string `xml:"multipleValues,attr"`
	NearestValue   string `xml:"nearestValue,attr"`
	Values         string `xml:",chardata"`
}

// XMLAttribution is the XML representation of attribution.
type XMLAttribution struct {
	Title            string          `xml:"Title"`
	OnlineResourceEl *OnlineResource `xml:"OnlineResource"`
	LogoURL          *XMLLogo        `xml:"LogoURL"`
}

// URL returns the attribution URL.
func (a *XMLAttribution) URL() string {
	if a.OnlineResourceEl != nil {
		return a.OnlineResourceEl.Href
	}
	return ""
}

// XMLLogo represents a logo.
type XMLLogo struct {
	Format           string          `xml:"Format"`
	OnlineResourceEl *OnlineResource `xml:"OnlineResource"`
}

// URL returns the logo URL.
func (l *XMLLogo) URL() string {
	if l.OnlineResourceEl != nil {
		return l.OnlineResourceEl.Href
	}
	return ""
}

// XMLMetadataURL is the XML representation of metadata URL.
type XMLMetadataURL struct {
	Type             string          `xml:"type,attr"`
	Format           string          `xml:"Format"`
	OnlineResourceEl *OnlineResource `xml:"OnlineResource"`
}

// URL returns the metadata URL.
func (m *XMLMetadataURL) URL() string {
	if m.OnlineResourceEl != nil {
		return m.OnlineResourceEl.Href
	}
	return ""
}

// GetLayers returns a flattened list of all named layers.
func (c *Capabilities) GetLayers() []Layer {
	if c.Capability.Layer == nil {
		return nil
	}
	var layers []Layer
	collectLayers(c.Capability.Layer, &layers, nil)
	return layers
}

// GetLayer finds a layer by name.
func (c *Capabilities) GetLayer(name string) *Layer {
	layers := c.GetLayers()
	for i := range layers {
		if layers[i].Name == name {
			return &layers[i]
		}
	}
	return nil
}

// GetLayerNames returns all layer names.
func (c *Capabilities) GetLayerNames() []string {
	layers := c.GetLayers()
	names := make([]string, 0, len(layers))
	for _, l := range layers {
		if l.Name != "" {
			names = append(names, l.Name)
		}
	}
	return names
}

// SupportsFormat checks if the service supports a given GetMap format.
func (c *Capabilities) SupportsFormat(format string) bool {
	if c.Capability.Request.GetMap == nil {
		return false
	}
	for _, f := range c.Capability.Request.GetMap.Formats {
		if strings.EqualFold(f, format) {
			return true
		}
	}
	return false
}

// collectLayers recursively collects layers.
func collectLayers(xmlLayer *XMLLayer, layers *[]Layer, parentCRS []string) {
	layer := convertXMLLayer(xmlLayer, parentCRS)
	if layer.Name != "" {
		*layers = append(*layers, layer)
	}

	// Inherit CRS from parent
	mergedCRS := append(parentCRS, xmlLayer.CRS...)

	for i := range xmlLayer.Layers {
		collectLayers(&xmlLayer.Layers[i], layers, mergedCRS)
	}
}

// convertXMLLayer converts an XMLLayer to a Layer.
func convertXMLLayer(xl *XMLLayer, parentCRS []string) Layer {
	layer := Layer{
		Name:      xl.Name,
		Title:     xl.Title,
		Abstract:  xl.Abstract,
		Queryable: xl.Queryable == "1" || strings.ToLower(xl.Queryable) == "true",
		Opaque:    xl.Opaque == "1" || strings.ToLower(xl.Opaque) == "true",
		CRS:       append(parentCRS, xl.CRS...),
		MinScale:  xl.MinScale,
		MaxScale:  xl.MaxScale,
	}

	// Convert bounding boxes
	layer.BoundingBox = make(map[string]*BBox)
	for _, bb := range xl.BoundingBoxes {
		layer.BoundingBox[bb.CRS] = &BBox{
			MinX: bb.MinX,
			MinY: bb.MinY,
			MaxX: bb.MaxX,
			MaxY: bb.MaxY,
		}
	}

	// Add geographic bounding box
	if xl.EXBoundingBox != nil {
		layer.BoundingBox["EPSG:4326"] = &BBox{
			MinX: xl.EXBoundingBox.WestLon,
			MinY: xl.EXBoundingBox.SouthLat,
			MaxX: xl.EXBoundingBox.EastLon,
			MaxY: xl.EXBoundingBox.NorthLat,
		}
	}

	// Convert styles
	for _, xs := range xl.Styles {
		style := Style{
			Name:     xs.Name,
			Title:    xs.Title,
			Abstract: xs.Abstract,
		}
		if xs.LegendURL != nil {
			style.LegendURL = &LegendURL{
				Format: xs.LegendURL.Format,
				Width:  xs.LegendURL.Width,
				Height: xs.LegendURL.Height,
				URL:    xs.LegendURL.URL(),
			}
		}
		layer.Styles = append(layer.Styles, style)
	}

	// Convert dimensions
	for _, xd := range xl.Dimensions {
		layer.Dimensions = append(layer.Dimensions, Dimension{
			Name:           xd.Name,
			Units:          xd.Units,
			Default:        xd.Default,
			Values:         strings.TrimSpace(xd.Values),
			MultipleValues: xd.MultipleValues == "1" || strings.ToLower(xd.MultipleValues) == "true",
			NearestValue:   xd.NearestValue == "1" || strings.ToLower(xd.NearestValue) == "true",
		})
	}

	// Convert attribution
	if xl.Attribution != nil {
		layer.Attribution = &Attribution{
			Title: xl.Attribution.Title,
			URL:   xl.Attribution.URL(),
		}
		if xl.Attribution.LogoURL != nil {
			layer.Attribution.LogoURL = xl.Attribution.LogoURL.URL()
			layer.Attribution.LogoType = xl.Attribution.LogoURL.Format
		}
	}

	// Convert metadata URLs
	for _, xm := range xl.MetadataURLs {
		layer.MetadataURL = append(layer.MetadataURL, MetadataURL{
			Type:   xm.Type,
			Format: xm.Format,
			URL:    xm.URL(),
		})
	}

	return layer
}

// DescribeLayerResponse represents the response from DescribeLayer.
type DescribeLayerResponse struct {
	XMLName           xml.Name           `xml:"DescribeLayerResponse"`
	Version           string             `xml:"version,attr"`
	LayerDescriptions []LayerDescription `xml:"LayerDescription"`
}

// LayerDescription describes a single layer.
type LayerDescription struct {
	Name     string `xml:"name,attr"`
	WFS      string `xml:"wfs,attr"`
	OwsType  string `xml:"owsType,attr"`
	OwsURL   string `xml:"owsURL,attr"`
	TypeName string `xml:"TypeName>typeName"`
}
