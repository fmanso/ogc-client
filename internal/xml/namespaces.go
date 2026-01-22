// Package xml provides XML parsing utilities for OGC services.
package xml

// Common OGC XML namespaces.
const (
	// WMS namespaces
	NSWMS130 = "http://www.opengis.net/wms"
	NSWMS111 = ""

	// WFS namespaces
	NSWFS20 = "http://www.opengis.net/wfs/2.0"
	NSWFS11 = "http://www.opengis.net/wfs"

	// Common OGC namespaces
	NSOWS       = "http://www.opengis.net/ows/1.1"
	NSOWS20     = "http://www.opengis.net/ows/2.0"
	NSFilter    = "http://www.opengis.net/fes/2.0"
	NSGML       = "http://www.opengis.net/gml/3.2"
	NSGML32     = "http://www.opengis.net/gml/3.2"
	NSGML31     = "http://www.opengis.net/gml"
	NSXLink     = "http://www.w3.org/1999/xlink"
	NSXSI       = "http://www.w3.org/2001/XMLSchema-instance"
	NSXMLSchema = "http://www.w3.org/2001/XMLSchema"

	// GeoServer specific
	NSGeoServer = "http://geoserver.org/wfs"
)
