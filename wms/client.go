// Package wms provides a client for OGC Web Map Service (WMS) 1.3.0.
package wms

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	inthttp "github.com/fmanso/ogc-client/internal/http"
)

const (
	// Version is the WMS version supported by this client.
	Version = "1.3.0"

	// ServiceName is the WMS service identifier.
	ServiceName = "WMS"
)

// Client provides access to WMS operations.
type Client struct {
	http *inthttp.Client
}

// NewClient creates a new WMS client.
func NewClient(endpoint string, httpClient *http.Client, auth func(*http.Request), userAgent string) *Client {
	return &Client{
		http: inthttp.NewClient(endpoint, httpClient, auth, userAgent),
	}
}

// GetCapabilities retrieves the service capabilities document.
func (c *Client) GetCapabilities(ctx context.Context) (*Capabilities, error) {
	body, resp, err := c.http.Get("").
		Param("SERVICE", ServiceName).
		Param("VERSION", Version).
		Param("REQUEST", "GetCapabilities").
		DoAndRead(ctx)

	if err != nil {
		return nil, fmt.Errorf("GetCapabilities request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, parseWMSException(body, resp.StatusCode)
	}

	var caps Capabilities
	if err := xml.Unmarshal(body, &caps); err != nil {
		return nil, fmt.Errorf("failed to parse capabilities: %w", err)
	}

	return &caps, nil
}

// GetMap returns a GetMapRequest builder for requesting map images.
func (c *Client) GetMap() *GetMapRequest {
	return &GetMapRequest{
		client:      c,
		format:      "image/png",
		crs:         "EPSG:4326",
		width:       256,
		height:      256,
		transparent: true,
	}
}

// DescribeLayer retrieves layer schema information.
func (c *Client) DescribeLayer(ctx context.Context, layers ...string) (*DescribeLayerResponse, error) {
	if len(layers) == 0 {
		return nil, fmt.Errorf("at least one layer must be specified")
	}

	body, resp, err := c.http.Get("").
		Param("SERVICE", ServiceName).
		Param("VERSION", Version).
		Param("REQUEST", "DescribeLayer").
		Param("LAYERS", strings.Join(layers, ",")).
		DoAndRead(ctx)

	if err != nil {
		return nil, fmt.Errorf("DescribeLayer request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, parseWMSException(body, resp.StatusCode)
	}

	var descResp DescribeLayerResponse
	if err := xml.Unmarshal(body, &descResp); err != nil {
		return nil, fmt.Errorf("failed to parse DescribeLayer response: %w", err)
	}

	return &descResp, nil
}

// GetLegendGraphic returns a GetLegendGraphicRequest builder.
func (c *Client) GetLegendGraphic() *GetLegendGraphicRequest {
	return &GetLegendGraphicRequest{
		client: c,
		format: "image/png",
		width:  20,
		height: 20,
	}
}

// GetMapRequest builds a GetMap request.
type GetMapRequest struct {
	client       *Client
	layers       []string
	styles       []string
	crs          string
	bbox         *BBox
	width        int
	height       int
	format       string
	transparent  bool
	bgcolor      string
	time         string
	elevation    string
	sld          string
	sldBody      string
	cqlFilter    string
	vendorParams map[string]string
}

// Layers sets the layers to request.
func (r *GetMapRequest) Layers(layers ...string) *GetMapRequest {
	r.layers = layers
	return r
}

// Styles sets the styles for each layer.
func (r *GetMapRequest) Styles(styles ...string) *GetMapRequest {
	r.styles = styles
	return r
}

// CRS sets the coordinate reference system.
func (r *GetMapRequest) CRS(crs string) *GetMapRequest {
	r.crs = crs
	return r
}

// BBox sets the bounding box.
func (r *GetMapRequest) BBox(minX, minY, maxX, maxY float64) *GetMapRequest {
	r.bbox = &BBox{MinX: minX, MinY: minY, MaxX: maxX, MaxY: maxY}
	return r
}

// Size sets the output image dimensions.
func (r *GetMapRequest) Size(width, height int) *GetMapRequest {
	r.width = width
	r.height = height
	return r
}

// Format sets the output format (e.g., "image/png", "image/jpeg").
func (r *GetMapRequest) Format(format string) *GetMapRequest {
	r.format = format
	return r
}

// Transparent sets whether the background should be transparent.
func (r *GetMapRequest) Transparent(transparent bool) *GetMapRequest {
	r.transparent = transparent
	return r
}

// BgColor sets the background color (hex format without #).
func (r *GetMapRequest) BgColor(color string) *GetMapRequest {
	r.bgcolor = color
	return r
}

// Time sets the TIME parameter for time-enabled layers.
func (r *GetMapRequest) Time(time string) *GetMapRequest {
	r.time = time
	return r
}

// Elevation sets the ELEVATION parameter for elevation-enabled layers.
func (r *GetMapRequest) Elevation(elevation string) *GetMapRequest {
	r.elevation = elevation
	return r
}

// SLD sets the SLD URL for styling.
func (r *GetMapRequest) SLD(sldURL string) *GetMapRequest {
	r.sld = sldURL
	return r
}

// SLDBody sets the inline SLD document for styling.
func (r *GetMapRequest) SLDBody(sld string) *GetMapRequest {
	r.sldBody = sld
	return r
}

// CQLFilter sets a CQL filter for the request (GeoServer extension).
func (r *GetMapRequest) CQLFilter(filter string) *GetMapRequest {
	r.cqlFilter = filter
	return r
}

// VendorParam sets a vendor-specific parameter.
func (r *GetMapRequest) VendorParam(key, value string) *GetMapRequest {
	if r.vendorParams == nil {
		r.vendorParams = make(map[string]string)
	}
	r.vendorParams[key] = value
	return r
}

// Execute performs the GetMap request and returns the image data.
func (r *GetMapRequest) Execute(ctx context.Context) (*MapResponse, error) {
	if len(r.layers) == 0 {
		return nil, fmt.Errorf("at least one layer must be specified")
	}
	if r.bbox == nil {
		return nil, fmt.Errorf("bounding box must be specified")
	}

	req := r.client.http.Get("").
		Param("SERVICE", ServiceName).
		Param("VERSION", Version).
		Param("REQUEST", "GetMap").
		Param("LAYERS", strings.Join(r.layers, ",")).
		Param("CRS", r.crs).
		Param("BBOX", r.bbox.String()).
		Param("WIDTH", strconv.Itoa(r.width)).
		Param("HEIGHT", strconv.Itoa(r.height)).
		Param("FORMAT", r.format)

	if len(r.styles) > 0 {
		req.Param("STYLES", strings.Join(r.styles, ","))
	} else {
		req.Param("STYLES", "")
	}

	if r.transparent {
		req.Param("TRANSPARENT", "TRUE")
	}

	if r.bgcolor != "" {
		req.Param("BGCOLOR", r.bgcolor)
	}

	if r.time != "" {
		req.Param("TIME", r.time)
	}

	if r.elevation != "" {
		req.Param("ELEVATION", r.elevation)
	}

	if r.sld != "" {
		req.Param("SLD", r.sld)
	}

	if r.sldBody != "" {
		req.Param("SLD_BODY", r.sldBody)
	}

	if r.cqlFilter != "" {
		req.Param("CQL_FILTER", r.cqlFilter)
	}

	for k, v := range r.vendorParams {
		req.Param(k, v)
	}

	resp, err := req.Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetMap request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check if response is an exception
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "xml") || resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, parseWMSException(body, resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read map image: %w", err)
	}

	return &MapResponse{
		Data:        data,
		ContentType: contentType,
		Width:       r.width,
		Height:      r.height,
	}, nil
}

// Stream performs the GetMap request and returns a streaming reader.
func (r *GetMapRequest) Stream(ctx context.Context) (*MapStreamResponse, error) {
	if len(r.layers) == 0 {
		return nil, fmt.Errorf("at least one layer must be specified")
	}
	if r.bbox == nil {
		return nil, fmt.Errorf("bounding box must be specified")
	}

	req := r.client.http.Get("").
		Param("SERVICE", ServiceName).
		Param("VERSION", Version).
		Param("REQUEST", "GetMap").
		Param("LAYERS", strings.Join(r.layers, ",")).
		Param("CRS", r.crs).
		Param("BBOX", r.bbox.String()).
		Param("WIDTH", strconv.Itoa(r.width)).
		Param("HEIGHT", strconv.Itoa(r.height)).
		Param("FORMAT", r.format)

	if len(r.styles) > 0 {
		req.Param("STYLES", strings.Join(r.styles, ","))
	} else {
		req.Param("STYLES", "")
	}

	if r.transparent {
		req.Param("TRANSPARENT", "TRUE")
	}

	for k, v := range r.vendorParams {
		req.Param(k, v)
	}

	resp, err := req.Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetMap request failed: %w", err)
	}

	// Check if response is an exception
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "xml") || resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, parseWMSException(body, resp.StatusCode)
	}

	return &MapStreamResponse{
		Reader:      resp.Body,
		ContentType: contentType,
		response:    resp,
	}, nil
}

// MapResponse contains the result of a GetMap request.
type MapResponse struct {
	Data        []byte
	ContentType string
	Width       int
	Height      int
}

// MapStreamResponse provides streaming access to map image data.
type MapStreamResponse struct {
	Reader      io.Reader
	ContentType string
	response    *http.Response
}

// Close closes the underlying response.
func (r *MapStreamResponse) Close() error {
	if r.response != nil {
		return r.response.Body.Close()
	}
	return nil
}

// GetLegendGraphicRequest builds a GetLegendGraphic request.
type GetLegendGraphicRequest struct {
	client       *Client
	layer        string
	style        string
	format       string
	width        int
	height       int
	rule         string
	scale        float64
	sld          string
	sldBody      string
	vendorParams map[string]string
}

// Layer sets the layer name.
func (r *GetLegendGraphicRequest) Layer(layer string) *GetLegendGraphicRequest {
	r.layer = layer
	return r
}

// Style sets the style name.
func (r *GetLegendGraphicRequest) Style(style string) *GetLegendGraphicRequest {
	r.style = style
	return r
}

// Format sets the output format.
func (r *GetLegendGraphicRequest) Format(format string) *GetLegendGraphicRequest {
	r.format = format
	return r
}

// Size sets the legend dimensions.
func (r *GetLegendGraphicRequest) Size(width, height int) *GetLegendGraphicRequest {
	r.width = width
	r.height = height
	return r
}

// Rule sets a specific rule to render.
func (r *GetLegendGraphicRequest) Rule(rule string) *GetLegendGraphicRequest {
	r.rule = rule
	return r
}

// Scale sets the scale denominator.
func (r *GetLegendGraphicRequest) Scale(scale float64) *GetLegendGraphicRequest {
	r.scale = scale
	return r
}

// SLD sets the SLD URL.
func (r *GetLegendGraphicRequest) SLD(sldURL string) *GetLegendGraphicRequest {
	r.sld = sldURL
	return r
}

// SLDBody sets the inline SLD.
func (r *GetLegendGraphicRequest) SLDBody(sld string) *GetLegendGraphicRequest {
	r.sldBody = sld
	return r
}

// VendorParam sets a vendor-specific parameter.
func (r *GetLegendGraphicRequest) VendorParam(key, value string) *GetLegendGraphicRequest {
	if r.vendorParams == nil {
		r.vendorParams = make(map[string]string)
	}
	r.vendorParams[key] = value
	return r
}

// Execute performs the GetLegendGraphic request.
func (r *GetLegendGraphicRequest) Execute(ctx context.Context) (*LegendResponse, error) {
	if r.layer == "" {
		return nil, fmt.Errorf("layer must be specified")
	}

	req := r.client.http.Get("").
		Param("SERVICE", ServiceName).
		Param("VERSION", Version).
		Param("REQUEST", "GetLegendGraphic").
		Param("LAYER", r.layer).
		Param("FORMAT", r.format)

	if r.style != "" {
		req.Param("STYLE", r.style)
	}

	if r.width > 0 {
		req.Param("WIDTH", strconv.Itoa(r.width))
	}

	if r.height > 0 {
		req.Param("HEIGHT", strconv.Itoa(r.height))
	}

	if r.rule != "" {
		req.Param("RULE", r.rule)
	}

	if r.scale > 0 {
		req.Param("SCALE", fmt.Sprintf("%f", r.scale))
	}

	if r.sld != "" {
		req.Param("SLD", r.sld)
	}

	if r.sldBody != "" {
		req.Param("SLD_BODY", r.sldBody)
	}

	for k, v := range r.vendorParams {
		req.Param(k, v)
	}

	resp, err := req.Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetLegendGraphic request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check for exception
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "xml") || resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, parseWMSException(body, resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read legend image: %w", err)
	}

	return &LegendResponse{
		Data:        data,
		ContentType: contentType,
	}, nil
}

// LegendResponse contains the result of a GetLegendGraphic request.
type LegendResponse struct {
	Data        []byte
	ContentType string
}

// parseWMSException parses a WMS service exception from XML.
func parseWMSException(body []byte, statusCode int) error {
	// Try to parse as ServiceExceptionReport
	var seReport struct {
		XMLName    xml.Name `xml:"ServiceExceptionReport"`
		Exceptions []struct {
			Code    string `xml:"code,attr"`
			Locator string `xml:"locator,attr"`
			Text    string `xml:",chardata"`
		} `xml:"ServiceException"`
	}

	if err := xml.Unmarshal(body, &seReport); err == nil && len(seReport.Exceptions) > 0 {
		ex := seReport.Exceptions[0]
		return &WMSError{
			Code:       ex.Code,
			Message:    strings.TrimSpace(ex.Text),
			Locator:    ex.Locator,
			StatusCode: statusCode,
		}
	}

	// Return generic error
	return &WMSError{
		Code:       "UnknownError",
		Message:    string(body),
		StatusCode: statusCode,
	}
}

// WMSError represents a WMS service exception.
type WMSError struct {
	Code       string
	Message    string
	Locator    string
	StatusCode int
}

// Error returns the error message.
func (e *WMSError) Error() string {
	msg := "wms: "
	if e.Code != "" {
		msg += "[" + e.Code + "] "
	}
	msg += e.Message
	if e.Locator != "" {
		msg += " (locator: " + e.Locator + ")"
	}
	return msg
}
