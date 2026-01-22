// Package wfs provides a client for OGC Web Feature Service (WFS) 2.0.
package wfs

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
	// Version is the WFS version supported by this client.
	Version = "2.0.0"

	// ServiceName is the WFS service identifier.
	ServiceName = "WFS"
)

// Client provides access to WFS operations.
type Client struct {
	http *inthttp.Client
}

// NewClient creates a new WFS client.
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
		return nil, parseWFSException(body, resp.StatusCode)
	}

	var caps Capabilities
	if err := xml.Unmarshal(body, &caps); err != nil {
		return nil, fmt.Errorf("failed to parse capabilities: %w", err)
	}

	return &caps, nil
}

// DescribeFeatureType retrieves the schema for feature types.
func (c *Client) DescribeFeatureType(ctx context.Context, typeNames ...string) (*DescribeFeatureTypeResponse, error) {
	req := c.http.Get("").
		Param("SERVICE", ServiceName).
		Param("VERSION", Version).
		Param("REQUEST", "DescribeFeatureType")

	if len(typeNames) > 0 {
		req.Param("TYPENAMES", strings.Join(typeNames, ","))
	}

	body, resp, err := req.DoAndRead(ctx)
	if err != nil {
		return nil, fmt.Errorf("DescribeFeatureType request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, parseWFSException(body, resp.StatusCode)
	}

	var schema XSDSchema
	if err := xml.Unmarshal(body, &schema); err != nil {
		return nil, fmt.Errorf("failed to parse schema: %w", err)
	}

	return &DescribeFeatureTypeResponse{
		Schema: &schema,
		Raw:    body,
	}, nil
}

// GetFeature returns a GetFeatureRequest builder.
func (c *Client) GetFeature(typeNames ...string) *GetFeatureRequest {
	return &GetFeatureRequest{
		client:     c,
		typeNames:  typeNames,
		format:     FormatGeoJSON,
		count:      0, // No limit
		startIndex: 0,
	}
}

// GetPropertyValue returns a GetPropertyValueRequest builder.
func (c *Client) GetPropertyValue(typeName, valueReference string) *GetPropertyValueRequest {
	return &GetPropertyValueRequest{
		client:         c,
		typeName:       typeName,
		valueReference: valueReference,
	}
}

// Transaction returns a new TransactionBuilder.
func (c *Client) Transaction() *TransactionBuilder {
	return &TransactionBuilder{
		client:  c,
		inserts: make([]InsertOperation, 0),
		updates: make([]UpdateOperation, 0),
		deletes: make([]DeleteOperation, 0),
	}
}

// GetFeatureRequest builds a GetFeature request.
type GetFeatureRequest struct {
	client        *Client
	typeNames     []string
	format        string
	count         int
	startIndex    int
	sortBy        string
	propertyNames []string
	filter        Filter
	cqlFilter     string
	bbox          *BBox
	resourceID    string
	srsName       string
	vendorParams  map[string]string
}

// TypeNames sets the feature types to query.
func (r *GetFeatureRequest) TypeNames(names ...string) *GetFeatureRequest {
	r.typeNames = names
	return r
}

// OutputFormat sets the output format.
func (r *GetFeatureRequest) OutputFormat(format string) *GetFeatureRequest {
	r.format = format
	return r
}

// Count sets the maximum number of features to return.
func (r *GetFeatureRequest) Count(count int) *GetFeatureRequest {
	r.count = count
	return r
}

// MaxFeatures is an alias for Count.
func (r *GetFeatureRequest) MaxFeatures(count int) *GetFeatureRequest {
	return r.Count(count)
}

// StartIndex sets the starting index for paging.
func (r *GetFeatureRequest) StartIndex(index int) *GetFeatureRequest {
	r.startIndex = index
	return r
}

// SortBy sets the sort order.
func (r *GetFeatureRequest) SortBy(sortBy string) *GetFeatureRequest {
	r.sortBy = sortBy
	return r
}

// PropertyNames sets the properties to return.
func (r *GetFeatureRequest) PropertyNames(names ...string) *GetFeatureRequest {
	r.propertyNames = names
	return r
}

// Filter sets an OGC filter.
func (r *GetFeatureRequest) Filter(filter Filter) *GetFeatureRequest {
	r.filter = filter
	return r
}

// CQLFilter sets a CQL filter (GeoServer extension).
func (r *GetFeatureRequest) CQLFilter(filter string) *GetFeatureRequest {
	r.cqlFilter = filter
	return r
}

// BBox sets a bounding box filter.
func (r *GetFeatureRequest) BBox(minX, minY, maxX, maxY float64, srs string) *GetFeatureRequest {
	r.bbox = &BBox{
		LowerCorner: [2]float64{minX, minY},
		UpperCorner: [2]float64{maxX, maxY},
		CRS:         srs,
	}
	return r
}

// ResourceID sets specific resource IDs to retrieve.
func (r *GetFeatureRequest) ResourceID(id string) *GetFeatureRequest {
	r.resourceID = id
	return r
}

// SRSName sets the output coordinate reference system.
func (r *GetFeatureRequest) SRSName(srs string) *GetFeatureRequest {
	r.srsName = srs
	return r
}

// VendorParam sets a vendor-specific parameter.
func (r *GetFeatureRequest) VendorParam(key, value string) *GetFeatureRequest {
	if r.vendorParams == nil {
		r.vendorParams = make(map[string]string)
	}
	r.vendorParams[key] = value
	return r
}

// buildRequest builds the HTTP request.
func (r *GetFeatureRequest) buildRequest() *inthttp.Request {
	req := r.client.http.Get("").
		Param("SERVICE", ServiceName).
		Param("VERSION", Version).
		Param("REQUEST", "GetFeature").
		Param("OUTPUTFORMAT", r.format)

	if len(r.typeNames) > 0 {
		req.Param("TYPENAMES", strings.Join(r.typeNames, ","))
	}

	if r.count > 0 {
		req.Param("COUNT", strconv.Itoa(r.count))
	}

	if r.startIndex > 0 {
		req.Param("STARTINDEX", strconv.Itoa(r.startIndex))
	}

	if r.sortBy != "" {
		req.Param("SORTBY", r.sortBy)
	}

	if len(r.propertyNames) > 0 {
		req.Param("PROPERTYNAME", strings.Join(r.propertyNames, ","))
	}

	if r.filter != nil {
		req.Param("FILTER", r.filter.ToXML())
	}

	if r.cqlFilter != "" {
		req.Param("CQL_FILTER", r.cqlFilter)
	}

	if r.bbox != nil {
		bboxStr := fmt.Sprintf("%f,%f,%f,%f",
			r.bbox.MinX(), r.bbox.MinY(), r.bbox.MaxX(), r.bbox.MaxY())
		if r.bbox.CRS != "" {
			bboxStr += "," + r.bbox.CRS
		}
		req.Param("BBOX", bboxStr)
	}

	if r.resourceID != "" {
		req.Param("RESOURCEID", r.resourceID)
	}

	if r.srsName != "" {
		req.Param("SRSNAME", r.srsName)
	}

	for k, v := range r.vendorParams {
		req.Param(k, v)
	}

	return req
}

// Execute performs the GetFeature request and returns the raw response.
func (r *GetFeatureRequest) Execute(ctx context.Context) ([]byte, error) {
	if len(r.typeNames) == 0 {
		return nil, fmt.Errorf("at least one type name must be specified")
	}

	body, resp, err := r.buildRequest().DoAndRead(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetFeature request failed: %w", err)
	}

	// Check for exception
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "xml") && resp.StatusCode != http.StatusOK {
		return nil, parseWFSException(body, resp.StatusCode)
	}

	return body, nil
}

// Stream performs the GetFeature request and returns a streaming reader.
func (r *GetFeatureRequest) Stream(ctx context.Context) (*FeatureStream, error) {
	if len(r.typeNames) == 0 {
		return nil, fmt.Errorf("at least one type name must be specified")
	}

	resp, err := r.buildRequest().Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetFeature request failed: %w", err)
	}

	// Check for exception
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "xml") && resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, parseWFSException(body, resp.StatusCode)
	}

	return &FeatureStream{
		Reader:      resp.Body,
		ContentType: contentType,
		Format:      r.format,
		response:    resp,
	}, nil
}

// FeatureStream provides streaming access to feature data.
type FeatureStream struct {
	Reader      io.Reader
	ContentType string
	Format      string
	response    *http.Response
}

// Close closes the underlying response.
func (s *FeatureStream) Close() error {
	if s.response != nil {
		return s.response.Body.Close()
	}
	return nil
}

// Read implements io.Reader.
func (s *FeatureStream) Read(p []byte) (n int, err error) {
	return s.Reader.Read(p)
}

// GetPropertyValueRequest builds a GetPropertyValue request.
type GetPropertyValueRequest struct {
	client         *Client
	typeName       string
	valueReference string
	filter         Filter
	cqlFilter      string
	count          int
}

// Filter sets an OGC filter.
func (r *GetPropertyValueRequest) Filter(filter Filter) *GetPropertyValueRequest {
	r.filter = filter
	return r
}

// CQLFilter sets a CQL filter.
func (r *GetPropertyValueRequest) CQLFilter(filter string) *GetPropertyValueRequest {
	r.cqlFilter = filter
	return r
}

// Count sets the maximum number of values to return.
func (r *GetPropertyValueRequest) Count(count int) *GetPropertyValueRequest {
	r.count = count
	return r
}

// Execute performs the GetPropertyValue request.
func (r *GetPropertyValueRequest) Execute(ctx context.Context) (*GetPropertyValueResponse, error) {
	req := r.client.http.Get("").
		Param("SERVICE", ServiceName).
		Param("VERSION", Version).
		Param("REQUEST", "GetPropertyValue").
		Param("TYPENAMES", r.typeName).
		Param("VALUEREFERENCE", r.valueReference)

	if r.filter != nil {
		req.Param("FILTER", r.filter.ToXML())
	}

	if r.cqlFilter != "" {
		req.Param("CQL_FILTER", r.cqlFilter)
	}

	if r.count > 0 {
		req.Param("COUNT", strconv.Itoa(r.count))
	}

	body, resp, err := req.DoAndRead(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetPropertyValue request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, parseWFSException(body, resp.StatusCode)
	}

	// Parse the response
	response := &GetPropertyValueResponse{
		Values: make([]PropertyValue, 0),
	}

	// Parse XML value collection
	var valueCollection struct {
		Members []struct {
			Value string `xml:",chardata"`
		} `xml:"member"`
	}

	if err := xml.Unmarshal(body, &valueCollection); err == nil {
		for _, m := range valueCollection.Members {
			response.Values = append(response.Values, PropertyValue{
				Value: strings.TrimSpace(m.Value),
			})
		}
	}

	return response, nil
}

// parseWFSException parses a WFS service exception from XML.
func parseWFSException(body []byte, statusCode int) error {
	// Try to parse as OWS ExceptionReport
	var exReport struct {
		XMLName    xml.Name `xml:"ExceptionReport"`
		Exceptions []struct {
			Code    string   `xml:"exceptionCode,attr"`
			Locator string   `xml:"locator,attr"`
			Text    []string `xml:"ExceptionText"`
		} `xml:"Exception"`
	}

	if err := xml.Unmarshal(body, &exReport); err == nil && len(exReport.Exceptions) > 0 {
		ex := exReport.Exceptions[0]
		message := ""
		if len(ex.Text) > 0 {
			message = strings.Join(ex.Text, "; ")
		}
		return &WFSError{
			Code:       ex.Code,
			Message:    strings.TrimSpace(message),
			Locator:    ex.Locator,
			StatusCode: statusCode,
		}
	}

	// Try legacy ServiceExceptionReport format
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
		return &WFSError{
			Code:       ex.Code,
			Message:    strings.TrimSpace(ex.Text),
			Locator:    ex.Locator,
			StatusCode: statusCode,
		}
	}

	// Return generic error
	return &WFSError{
		Code:       "UnknownError",
		Message:    string(body),
		StatusCode: statusCode,
	}
}
