package wms

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const testCapabilitiesXML = `<?xml version="1.0" encoding="UTF-8"?>
<WMS_Capabilities version="1.3.0" xmlns="http://www.opengis.net/wms">
  <Service>
    <Name>WMS</Name>
    <Title>GeoServer Web Map Service</Title>
    <Abstract>A compliant implementation of WMS</Abstract>
    <KeywordList>
      <Keyword>WMS</Keyword>
      <Keyword>GEOSERVER</Keyword>
    </KeywordList>
  </Service>
  <Capability>
    <Request>
      <GetCapabilities>
        <Format>text/xml</Format>
      </GetCapabilities>
      <GetMap>
        <Format>image/png</Format>
        <Format>image/jpeg</Format>
      </GetMap>
    </Request>
    <Layer queryable="1">
      <Name>roads</Name>
      <Title>Roads Layer</Title>
      <Abstract>All roads in the area</Abstract>
      <CRS>EPSG:4326</CRS>
      <CRS>EPSG:3857</CRS>
      <EX_GeographicBoundingBox>
        <westBoundLongitude>-180</westBoundLongitude>
        <eastBoundLongitude>180</eastBoundLongitude>
        <southBoundLatitude>-90</southBoundLatitude>
        <northBoundLatitude>90</northBoundLatitude>
      </EX_GeographicBoundingBox>
      <BoundingBox CRS="EPSG:4326" minx="-180" miny="-90" maxx="180" maxy="90"/>
      <Style>
        <Name>line</Name>
        <Title>Default Line Style</Title>
      </Style>
    </Layer>
  </Capability>
</WMS_Capabilities>`

func TestGetCapabilities(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request parameters
		if r.URL.Query().Get("SERVICE") != "WMS" {
			t.Errorf("expected SERVICE=WMS, got %s", r.URL.Query().Get("SERVICE"))
		}
		if r.URL.Query().Get("VERSION") != "1.3.0" {
			t.Errorf("expected VERSION=1.3.0, got %s", r.URL.Query().Get("VERSION"))
		}
		if r.URL.Query().Get("REQUEST") != "GetCapabilities" {
			t.Errorf("expected REQUEST=GetCapabilities, got %s", r.URL.Query().Get("REQUEST"))
		}

		w.Header().Set("Content-Type", "application/xml")
		io.WriteString(w, testCapabilitiesXML)
	}))
	defer server.Close()

	client := NewClient(server.URL, nil, nil, "")
	caps, err := client.GetCapabilities(context.Background())
	if err != nil {
		t.Fatalf("GetCapabilities() error = %v", err)
	}

	if caps.Version != "1.3.0" {
		t.Errorf("expected version 1.3.0, got %s", caps.Version)
	}

	if caps.Service.Title != "GeoServer Web Map Service" {
		t.Errorf("unexpected service title: %s", caps.Service.Title)
	}
}

func TestGetLayers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		io.WriteString(w, testCapabilitiesXML)
	}))
	defer server.Close()

	client := NewClient(server.URL, nil, nil, "")
	caps, err := client.GetCapabilities(context.Background())
	if err != nil {
		t.Fatalf("GetCapabilities() error = %v", err)
	}

	layers := caps.GetLayers()
	if len(layers) != 1 {
		t.Errorf("expected 1 layer, got %d", len(layers))
	}

	if layers[0].Name != "roads" {
		t.Errorf("expected layer name 'roads', got '%s'", layers[0].Name)
	}

	if !layers[0].Queryable {
		t.Error("expected layer to be queryable")
	}
}

func TestGetLayer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		io.WriteString(w, testCapabilitiesXML)
	}))
	defer server.Close()

	client := NewClient(server.URL, nil, nil, "")
	caps, err := client.GetCapabilities(context.Background())
	if err != nil {
		t.Fatalf("GetCapabilities() error = %v", err)
	}

	layer := caps.GetLayer("roads")
	if layer == nil {
		t.Fatal("GetLayer() returned nil for existing layer")
	}

	if layer.Title != "Roads Layer" {
		t.Errorf("expected title 'Roads Layer', got '%s'", layer.Title)
	}

	nonExistent := caps.GetLayer("nonexistent")
	if nonExistent != nil {
		t.Error("expected nil for non-existent layer")
	}
}

func TestSupportsFormat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		io.WriteString(w, testCapabilitiesXML)
	}))
	defer server.Close()

	client := NewClient(server.URL, nil, nil, "")
	caps, err := client.GetCapabilities(context.Background())
	if err != nil {
		t.Fatalf("GetCapabilities() error = %v", err)
	}

	if !caps.SupportsFormat("image/png") {
		t.Error("expected PNG format to be supported")
	}

	if !caps.SupportsFormat("IMAGE/JPEG") {
		t.Error("expected JPEG format to be supported (case insensitive)")
	}

	if caps.SupportsFormat("image/gif") {
		t.Error("expected GIF format to not be supported")
	}
}

func TestGetMapRequest(t *testing.T) {
	pngData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A} // PNG header

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request parameters
		q := r.URL.Query()
		if q.Get("REQUEST") != "GetMap" {
			t.Errorf("expected REQUEST=GetMap, got %s", q.Get("REQUEST"))
		}
		if q.Get("LAYERS") != "roads,buildings" {
			t.Errorf("expected LAYERS=roads,buildings, got %s", q.Get("LAYERS"))
		}
		if q.Get("CRS") != "EPSG:4326" {
			t.Errorf("expected CRS=EPSG:4326, got %s", q.Get("CRS"))
		}
		if q.Get("WIDTH") != "512" {
			t.Errorf("expected WIDTH=512, got %s", q.Get("WIDTH"))
		}
		if q.Get("HEIGHT") != "512" {
			t.Errorf("expected HEIGHT=512, got %s", q.Get("HEIGHT"))
		}
		if q.Get("FORMAT") != "image/png" {
			t.Errorf("expected FORMAT=image/png, got %s", q.Get("FORMAT"))
		}
		if q.Get("TRANSPARENT") != "TRUE" {
			t.Errorf("expected TRANSPARENT=TRUE, got %s", q.Get("TRANSPARENT"))
		}

		w.Header().Set("Content-Type", "image/png")
		w.Write(pngData)
	}))
	defer server.Close()

	client := NewClient(server.URL, nil, nil, "")
	resp, err := client.GetMap().
		Layers("roads", "buildings").
		CRS("EPSG:4326").
		BBox(-122.5, 37.5, -122.0, 38.0).
		Size(512, 512).
		Format("image/png").
		Transparent(true).
		Execute(context.Background())

	if err != nil {
		t.Fatalf("GetMap() error = %v", err)
	}

	if resp.ContentType != "image/png" {
		t.Errorf("expected content type image/png, got %s", resp.ContentType)
	}

	if len(resp.Data) != len(pngData) {
		t.Errorf("expected %d bytes, got %d", len(pngData), len(resp.Data))
	}
}

func TestGetMapRequestValidation(t *testing.T) {
	client := NewClient("http://localhost:8080", nil, nil, "")

	t.Run("no layers", func(t *testing.T) {
		_, err := client.GetMap().
			BBox(-122.5, 37.5, -122.0, 38.0).
			Size(256, 256).
			Execute(context.Background())

		if err == nil {
			t.Error("expected error for missing layers")
		}
	})

	t.Run("no bbox", func(t *testing.T) {
		_, err := client.GetMap().
			Layers("roads").
			Size(256, 256).
			Execute(context.Background())

		if err == nil {
			t.Error("expected error for missing bbox")
		}
	})
}

func TestServiceException(t *testing.T) {
	exceptionXML := `<?xml version="1.0" encoding="UTF-8"?>
<ServiceExceptionReport version="1.3.0">
  <ServiceException code="LayerNotDefined">
    Layer does not exist: invalid_layer
  </ServiceException>
</ServiceExceptionReport>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.ogc.se_xml")
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, exceptionXML)
	}))
	defer server.Close()

	client := NewClient(server.URL, nil, nil, "")
	_, err := client.GetCapabilities(context.Background())

	if err == nil {
		t.Fatal("expected error for service exception")
	}

	wmsErr, ok := err.(*WMSError)
	if !ok {
		t.Fatalf("expected WMSError, got %T", err)
	}

	if wmsErr.Code != "LayerNotDefined" {
		t.Errorf("expected code 'LayerNotDefined', got '%s'", wmsErr.Code)
	}

	if !strings.Contains(wmsErr.Message, "invalid_layer") {
		t.Errorf("expected message to contain 'invalid_layer', got '%s'", wmsErr.Message)
	}
}

func TestBBox(t *testing.T) {
	bbox := &BBox{MinX: -122.5, MinY: 37.5, MaxX: -122.0, MaxY: 38.0}

	t.Run("String", func(t *testing.T) {
		str := bbox.String()
		if !strings.Contains(str, "-122.5") {
			t.Errorf("expected string to contain -122.5, got %s", str)
		}
	})

	t.Run("Contains", func(t *testing.T) {
		if !bbox.Contains(-122.25, 37.75) {
			t.Error("expected point to be contained")
		}
		if bbox.Contains(-123.0, 37.75) {
			t.Error("expected point to not be contained")
		}
	})

	t.Run("Intersects", func(t *testing.T) {
		other := &BBox{MinX: -122.3, MinY: 37.7, MaxX: -121.5, MaxY: 38.5}
		if !bbox.Intersects(other) {
			t.Error("expected boxes to intersect")
		}

		nonIntersecting := &BBox{MinX: 0, MinY: 0, MaxX: 1, MaxY: 1}
		if bbox.Intersects(nonIntersecting) {
			t.Error("expected boxes to not intersect")
		}
	})
}

func TestGetLegendGraphic(t *testing.T) {
	pngData := []byte{0x89, 0x50, 0x4E, 0x47}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("REQUEST") != "GetLegendGraphic" {
			t.Errorf("expected REQUEST=GetLegendGraphic, got %s", q.Get("REQUEST"))
		}
		if q.Get("LAYER") != "roads" {
			t.Errorf("expected LAYER=roads, got %s", q.Get("LAYER"))
		}

		w.Header().Set("Content-Type", "image/png")
		w.Write(pngData)
	}))
	defer server.Close()

	client := NewClient(server.URL, nil, nil, "")
	resp, err := client.GetLegendGraphic().
		Layer("roads").
		Format("image/png").
		Size(20, 20).
		Execute(context.Background())

	if err != nil {
		t.Fatalf("GetLegendGraphic() error = %v", err)
	}

	if resp.ContentType != "image/png" {
		t.Errorf("expected content type image/png, got %s", resp.ContentType)
	}
}

func TestDescribeLayer(t *testing.T) {
	describeXML := `<?xml version="1.0" encoding="UTF-8"?>
<DescribeLayerResponse version="1.3.0">
  <LayerDescription name="roads" wfs="http://localhost/wfs" owsType="wfs" owsURL="http://localhost/wfs">
    <TypeName>
      <typeName>roads</typeName>
    </TypeName>
  </LayerDescription>
</DescribeLayerResponse>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("REQUEST") != "DescribeLayer" {
			t.Errorf("expected REQUEST=DescribeLayer, got %s", q.Get("REQUEST"))
		}

		w.Header().Set("Content-Type", "application/xml")
		io.WriteString(w, describeXML)
	}))
	defer server.Close()

	client := NewClient(server.URL, nil, nil, "")
	resp, err := client.DescribeLayer(context.Background(), "roads")

	if err != nil {
		t.Fatalf("DescribeLayer() error = %v", err)
	}

	if len(resp.LayerDescriptions) != 1 {
		t.Errorf("expected 1 layer description, got %d", len(resp.LayerDescriptions))
	}

	if resp.LayerDescriptions[0].Name != "roads" {
		t.Errorf("expected name 'roads', got '%s'", resp.LayerDescriptions[0].Name)
	}
}
