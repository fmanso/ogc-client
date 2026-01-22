package geojson

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

// FeatureReader provides streaming access to GeoJSON features.
type FeatureReader struct {
	decoder *json.Decoder
	reader  io.Reader
	started bool
	done    bool
	err     error
}

// NewFeatureReader creates a new streaming feature reader.
func NewFeatureReader(r io.Reader) *FeatureReader {
	return &FeatureReader{
		decoder: json.NewDecoder(r),
		reader:  r,
	}
}

// Next advances to the next feature.
// Returns false when there are no more features or an error occurred.
func (r *FeatureReader) Next() bool {
	if r.done || r.err != nil {
		return false
	}

	if !r.started {
		// Find the start of the features array
		if err := r.findFeaturesArray(); err != nil {
			r.err = err
			r.done = true
			return false
		}
		r.started = true
	}

	// Check if we've reached the end of the array
	if !r.decoder.More() {
		r.done = true
		return false
	}

	return true
}

// Feature reads and returns the current feature.
func (r *FeatureReader) Feature() (*Feature, error) {
	if r.err != nil {
		return nil, r.err
	}

	var f Feature
	if err := r.decoder.Decode(&f); err != nil {
		r.err = err
		return nil, err
	}

	return &f, nil
}

// Err returns any error that occurred during reading.
func (r *FeatureReader) Err() error {
	return r.err
}

// findFeaturesArray finds and positions the decoder at the features array.
func (r *FeatureReader) findFeaturesArray() error {
	// Read the opening brace
	token, err := r.decoder.Token()
	if err != nil {
		return fmt.Errorf("expected opening brace: %w", err)
	}
	if delim, ok := token.(json.Delim); !ok || delim != '{' {
		return fmt.Errorf("expected '{', got %v", token)
	}

	// Find the "features" key
	for r.decoder.More() {
		// Read key
		token, err := r.decoder.Token()
		if err != nil {
			return err
		}

		key, ok := token.(string)
		if !ok {
			continue
		}

		if key == "features" {
			// Read the opening bracket of the features array
			token, err := r.decoder.Token()
			if err != nil {
				return err
			}
			if delim, ok := token.(json.Delim); !ok || delim != '[' {
				return fmt.Errorf("expected '[' for features array, got %v", token)
			}
			return nil
		}

		// Skip the value for other keys
		if err := skipValue(r.decoder); err != nil {
			return err
		}
	}

	return fmt.Errorf("features array not found")
}

// skipValue skips a JSON value.
func skipValue(dec *json.Decoder) error {
	token, err := dec.Token()
	if err != nil {
		return err
	}

	switch t := token.(type) {
	case json.Delim:
		switch t {
		case '{':
			for dec.More() {
				// Skip key
				if _, err := dec.Token(); err != nil {
					return err
				}
				// Skip value
				if err := skipValue(dec); err != nil {
					return err
				}
			}
			// Read closing brace
			_, err := dec.Token()
			return err
		case '[':
			for dec.More() {
				if err := skipValue(dec); err != nil {
					return err
				}
			}
			// Read closing bracket
			_, err := dec.Token()
			return err
		}
	}

	return nil
}

// FeatureWriter provides streaming output of GeoJSON features.
type FeatureWriter struct {
	writer  io.Writer
	buf     *bufio.Writer
	started bool
	count   int
	crs     *CRS
}

// NewFeatureWriter creates a new streaming feature writer.
func NewFeatureWriter(w io.Writer) *FeatureWriter {
	return &FeatureWriter{
		writer: w,
		buf:    bufio.NewWriter(w),
	}
}

// SetCRS sets the CRS for the feature collection.
func (w *FeatureWriter) SetCRS(crs *CRS) {
	w.crs = crs
}

// WriteFeature writes a feature to the output.
func (w *FeatureWriter) WriteFeature(f *Feature) error {
	if !w.started {
		if err := w.writeHeader(); err != nil {
			return err
		}
		w.started = true
	}

	if w.count > 0 {
		if _, err := w.buf.WriteString(",\n"); err != nil {
			return err
		}
	}

	data, err := json.Marshal(f)
	if err != nil {
		return err
	}

	if _, err := w.buf.Write(data); err != nil {
		return err
	}

	w.count++
	return nil
}

// Close finishes writing and flushes the buffer.
func (w *FeatureWriter) Close() error {
	if !w.started {
		if err := w.writeHeader(); err != nil {
			return err
		}
	}

	if _, err := w.buf.WriteString("\n  ]\n}"); err != nil {
		return err
	}

	return w.buf.Flush()
}

func (w *FeatureWriter) writeHeader() error {
	if _, err := w.buf.WriteString(`{"type":"FeatureCollection"`); err != nil {
		return err
	}

	if w.crs != nil {
		crsData, err := json.Marshal(w.crs)
		if err != nil {
			return err
		}
		if _, err := w.buf.WriteString(`,"crs":`); err != nil {
			return err
		}
		if _, err := w.buf.Write(crsData); err != nil {
			return err
		}
	}

	if _, err := w.buf.WriteString(`,"features":[` + "\n"); err != nil {
		return err
	}

	return nil
}

// Count returns the number of features written.
func (w *FeatureWriter) Count() int {
	return w.count
}

// ReadAll reads all features from a reader into a FeatureCollection.
func ReadAll(r io.Reader) (*FeatureCollection, error) {
	reader := NewFeatureReader(r)
	fc := NewFeatureCollection()

	for reader.Next() {
		f, err := reader.Feature()
		if err != nil {
			return nil, err
		}
		fc.AddFeature(*f)
	}

	if err := reader.Err(); err != nil && err != io.EOF {
		return nil, err
	}

	return fc, nil
}

// WriteAll writes a FeatureCollection to a writer.
func WriteAll(w io.Writer, fc *FeatureCollection) error {
	writer := NewFeatureWriter(w)

	for i := range fc.Features {
		if err := writer.WriteFeature(&fc.Features[i]); err != nil {
			return err
		}
	}

	return writer.Close()
}
