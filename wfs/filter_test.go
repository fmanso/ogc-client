package wfs

import (
	"strings"
	"testing"
)

func TestPropertyIsEqualTo(t *testing.T) {
	filter := PropertyIsEqualTo("highway", "primary")

	t.Run("ToXML", func(t *testing.T) {
		xml := filter.ToXML()
		if !strings.Contains(xml, "PropertyIsEqualTo") {
			t.Error("expected XML to contain PropertyIsEqualTo")
		}
		if !strings.Contains(xml, "highway") {
			t.Error("expected XML to contain property name")
		}
		if !strings.Contains(xml, "primary") {
			t.Error("expected XML to contain value")
		}
	})

	t.Run("ToCQL", func(t *testing.T) {
		cql := filter.ToCQL()
		if cql != "highway = 'primary'" {
			t.Errorf("expected 'highway = 'primary'', got '%s'", cql)
		}
	})
}

func TestPropertyIsNotEqualTo(t *testing.T) {
	filter := PropertyIsNotEqualTo("status", "closed")
	cql := filter.ToCQL()
	if !strings.Contains(cql, "<>") {
		t.Errorf("expected CQL to contain <>, got '%s'", cql)
	}
}

func TestPropertyIsLessThan(t *testing.T) {
	filter := PropertyIsLessThan("population", 1000)
	cql := filter.ToCQL()
	if !strings.Contains(cql, "<") {
		t.Errorf("expected CQL to contain <, got '%s'", cql)
	}
}

func TestPropertyIsGreaterThan(t *testing.T) {
	filter := PropertyIsGreaterThan("population", 1000)
	cql := filter.ToCQL()
	if !strings.Contains(cql, ">") {
		t.Errorf("expected CQL to contain >, got '%s'", cql)
	}
}

func TestPropertyIsLike(t *testing.T) {
	filter := PropertyIsLike("name", "*road*")

	t.Run("ToXML", func(t *testing.T) {
		xml := filter.ToXML()
		if !strings.Contains(xml, "PropertyIsLike") {
			t.Error("expected XML to contain PropertyIsLike")
		}
		if !strings.Contains(xml, "wildCard") {
			t.Error("expected XML to contain wildCard attribute")
		}
	})

	t.Run("ToCQL", func(t *testing.T) {
		cql := filter.ToCQL()
		if !strings.Contains(cql, "LIKE") {
			t.Errorf("expected CQL to contain LIKE, got '%s'", cql)
		}
		if !strings.Contains(cql, "%road%") {
			t.Errorf("expected CQL to convert wildcards, got '%s'", cql)
		}
	})
}

func TestPropertyIsNull(t *testing.T) {
	filter := PropertyIsNull("description")

	t.Run("ToCQL", func(t *testing.T) {
		cql := filter.ToCQL()
		if cql != "description IS NULL" {
			t.Errorf("expected 'description IS NULL', got '%s'", cql)
		}
	})
}

func TestPropertyIsBetween(t *testing.T) {
	filter := PropertyIsBetween("year", 2000, 2020)

	t.Run("ToXML", func(t *testing.T) {
		xml := filter.ToXML()
		if !strings.Contains(xml, "PropertyIsBetween") {
			t.Error("expected XML to contain PropertyIsBetween")
		}
		if !strings.Contains(xml, "LowerBoundary") {
			t.Error("expected XML to contain LowerBoundary")
		}
	})

	t.Run("ToCQL", func(t *testing.T) {
		cql := filter.ToCQL()
		if cql != "year BETWEEN 2000 AND 2020" {
			t.Errorf("expected 'year BETWEEN 2000 AND 2020', got '%s'", cql)
		}
	})
}

func TestAnd(t *testing.T) {
	filter := And(
		PropertyEquals("highway", "primary"),
		PropertyGreaterThan("lanes", 2),
	)

	t.Run("ToXML", func(t *testing.T) {
		xml := filter.ToXML()
		if !strings.Contains(xml, "<fes:And>") {
			t.Error("expected XML to contain And element")
		}
	})

	t.Run("ToCQL", func(t *testing.T) {
		cql := filter.ToCQL()
		if !strings.Contains(cql, " AND ") {
			t.Errorf("expected CQL to contain AND, got '%s'", cql)
		}
	})
}

func TestOr(t *testing.T) {
	filter := Or(
		PropertyEquals("highway", "primary"),
		PropertyEquals("highway", "secondary"),
	)

	t.Run("ToCQL", func(t *testing.T) {
		cql := filter.ToCQL()
		if !strings.Contains(cql, " OR ") {
			t.Errorf("expected CQL to contain OR, got '%s'", cql)
		}
	})
}

func TestNot(t *testing.T) {
	filter := Not(PropertyEquals("closed", true))

	t.Run("ToXML", func(t *testing.T) {
		xml := filter.ToXML()
		if !strings.Contains(xml, "<fes:Not>") {
			t.Error("expected XML to contain Not element")
		}
	})

	t.Run("ToCQL", func(t *testing.T) {
		cql := filter.ToCQL()
		if !strings.HasPrefix(cql, "NOT") {
			t.Errorf("expected CQL to start with NOT, got '%s'", cql)
		}
	})
}

func TestBBOXFilter(t *testing.T) {
	filter := BBOXFilter("the_geom", -122.5, 37.5, -122.0, 38.0, "EPSG:4326")

	t.Run("ToXML", func(t *testing.T) {
		xml := filter.ToXML()
		if !strings.Contains(xml, "<fes:BBOX>") {
			t.Error("expected XML to contain BBOX element")
		}
		if !strings.Contains(xml, "EPSG:4326") {
			t.Error("expected XML to contain SRS")
		}
		if !strings.Contains(xml, "lowerCorner") {
			t.Error("expected XML to contain lowerCorner")
		}
	})

	t.Run("ToCQL", func(t *testing.T) {
		cql := filter.ToCQL()
		if !strings.HasPrefix(cql, "BBOX") {
			t.Errorf("expected CQL to start with BBOX, got '%s'", cql)
		}
	})
}

func TestIntersects(t *testing.T) {
	filter := Intersects("the_geom", "POINT(0 0)")

	t.Run("ToXML", func(t *testing.T) {
		xml := filter.ToXML()
		if !strings.Contains(xml, "<fes:Intersects>") {
			t.Error("expected XML to contain Intersects element")
		}
	})

	t.Run("ToCQL", func(t *testing.T) {
		cql := filter.ToCQL()
		if !strings.Contains(cql, "INTERSECTS") {
			t.Errorf("expected CQL to contain INTERSECTS, got '%s'", cql)
		}
	})
}

func TestDWithin(t *testing.T) {
	filter := DWithin("the_geom", "POINT(0 0)", 1000, "meters")

	t.Run("ToXML", func(t *testing.T) {
		xml := filter.ToXML()
		if !strings.Contains(xml, "<fes:DWithin>") {
			t.Error("expected XML to contain DWithin element")
		}
		if !strings.Contains(xml, "Distance") {
			t.Error("expected XML to contain Distance element")
		}
	})

	t.Run("ToCQL", func(t *testing.T) {
		cql := filter.ToCQL()
		if !strings.Contains(cql, "DWITHIN") {
			t.Errorf("expected CQL to contain DWITHIN, got '%s'", cql)
		}
	})
}

func TestResourceIDFilter(t *testing.T) {
	filter := ResourceIDFilter("feature.1", "feature.2", "feature.3")

	t.Run("ToXML", func(t *testing.T) {
		xml := filter.ToXML()
		if !strings.Contains(xml, "ResourceId") {
			t.Error("expected XML to contain ResourceId element")
		}
		if !strings.Contains(xml, "feature.1") {
			t.Error("expected XML to contain feature ID")
		}
	})

	t.Run("ToCQL", func(t *testing.T) {
		cql := filter.ToCQL()
		if !strings.Contains(cql, "IN") {
			t.Errorf("expected CQL to contain IN, got '%s'", cql)
		}
	})
}

func TestCQLFilter(t *testing.T) {
	filter := CQLFilter("highway = 'primary' AND lanes > 2")

	if filter.ToCQL() != "highway = 'primary' AND lanes > 2" {
		t.Error("CQLFilter should return the raw CQL string")
	}

	if filter.ToXML() != "" {
		t.Error("CQLFilter.ToXML should return empty string")
	}
}

func TestComplexFilter(t *testing.T) {
	// Build a complex filter
	filter := And(
		Or(
			PropertyEquals("highway", "primary"),
			PropertyEquals("highway", "secondary"),
		),
		PropertyGreaterThan("lanes", 1),
		BBOXFilter("the_geom", -122.5, 37.5, -122.0, 38.0, "EPSG:4326"),
	)

	cql := filter.ToCQL()

	// Verify the CQL contains expected parts
	if !strings.Contains(cql, "OR") {
		t.Error("expected CQL to contain OR")
	}
	if !strings.Contains(cql, "AND") {
		t.Error("expected CQL to contain AND")
	}
	if !strings.Contains(cql, "BBOX") {
		t.Error("expected CQL to contain BBOX")
	}
}

func TestPoint(t *testing.T) {
	p := NewPoint(10.5, 20.5, "EPSG:4326")

	if p.X != 10.5 {
		t.Errorf("expected X=10.5, got %f", p.X)
	}
	if p.Y != 20.5 {
		t.Errorf("expected Y=20.5, got %f", p.Y)
	}
	if p.SRS != "EPSG:4326" {
		t.Errorf("expected SRS=EPSG:4326, got %s", p.SRS)
	}
}

func TestEscapeXML(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"<tag>", "&lt;tag&gt;"},
		{"a&b", "a&amp;b"},
		{"'quote'", "&apos;quote&apos;"},
		{"\"double\"", "&quot;double&quot;"},
	}

	for _, tt := range tests {
		result := escapeXML(tt.input)
		if result != tt.expected {
			t.Errorf("escapeXML(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
