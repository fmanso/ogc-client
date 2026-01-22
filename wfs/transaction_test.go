package wfs

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTransactionInsert(t *testing.T) {
	var receivedBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)

		if r.Method != http.MethodPost {
			t.Errorf("expected POST method, got %s", r.Method)
		}

		// Check content type
		contentType := r.Header.Get("Content-Type")
		if !strings.Contains(contentType, "xml") {
			t.Errorf("expected XML content type, got %s", contentType)
		}

		w.Header().Set("Content-Type", "application/xml")
		io.WriteString(w, `<?xml version="1.0" encoding="UTF-8"?>
<wfs:TransactionResponse version="2.0.0" xmlns:wfs="http://www.opengis.net/wfs/2.0">
  <wfs:TransactionSummary>
    <wfs:totalInserted>1</wfs:totalInserted>
    <wfs:totalUpdated>0</wfs:totalUpdated>
    <wfs:totalDeleted>0</wfs:totalDeleted>
  </wfs:TransactionSummary>
</wfs:TransactionResponse>`)
	}))
	defer server.Close()

	client := NewClient(server.URL, nil, nil, "")

	resp, err := client.Transaction().
		Insert("roads").
		Feature(map[string]interface{}{
			"name":    "Test Road",
			"highway": "primary",
		}).
		Add().
		Execute(context.Background())

	if err != nil {
		t.Fatalf("Transaction() error = %v", err)
	}

	// Verify the request body
	if !strings.Contains(receivedBody, "<wfs:Insert>") {
		t.Error("expected request body to contain Insert element")
	}
	if !strings.Contains(receivedBody, "Test Road") {
		t.Error("expected request body to contain feature data")
	}

	// Verify the response
	if resp.TransactionSummary.TotalInserted != 1 {
		t.Errorf("expected 1 inserted, got %d", resp.TransactionSummary.TotalInserted)
	}
}

func TestTransactionUpdate(t *testing.T) {
	var receivedBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)

		w.Header().Set("Content-Type", "application/xml")
		io.WriteString(w, `<?xml version="1.0" encoding="UTF-8"?>
<wfs:TransactionResponse version="2.0.0" xmlns:wfs="http://www.opengis.net/wfs/2.0">
  <wfs:TransactionSummary>
    <wfs:totalInserted>0</wfs:totalInserted>
    <wfs:totalUpdated>5</wfs:totalUpdated>
    <wfs:totalDeleted>0</wfs:totalDeleted>
  </wfs:TransactionSummary>
</wfs:TransactionResponse>`)
	}))
	defer server.Close()

	client := NewClient(server.URL, nil, nil, "")

	resp, err := client.Transaction().
		Update("roads").
		Set("status", "closed").
		Filter(PropertyEquals("highway", "primary")).
		Add().
		Execute(context.Background())

	if err != nil {
		t.Fatalf("Transaction() error = %v", err)
	}

	// Verify the request body
	if !strings.Contains(receivedBody, "<wfs:Update") {
		t.Error("expected request body to contain Update element")
	}
	if !strings.Contains(receivedBody, "status") {
		t.Error("expected request body to contain property name")
	}
	if !strings.Contains(receivedBody, "fes:Filter") {
		t.Error("expected request body to contain Filter element")
	}

	// Verify the response
	if resp.TransactionSummary.TotalUpdated != 5 {
		t.Errorf("expected 5 updated, got %d", resp.TransactionSummary.TotalUpdated)
	}
}

func TestTransactionDelete(t *testing.T) {
	var receivedBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)

		w.Header().Set("Content-Type", "application/xml")
		io.WriteString(w, `<?xml version="1.0" encoding="UTF-8"?>
<wfs:TransactionResponse version="2.0.0" xmlns:wfs="http://www.opengis.net/wfs/2.0">
  <wfs:TransactionSummary>
    <wfs:totalInserted>0</wfs:totalInserted>
    <wfs:totalUpdated>0</wfs:totalUpdated>
    <wfs:totalDeleted>3</wfs:totalDeleted>
  </wfs:TransactionSummary>
</wfs:TransactionResponse>`)
	}))
	defer server.Close()

	client := NewClient(server.URL, nil, nil, "")

	resp, err := client.Transaction().
		Delete("roads").
		Filter(PropertyEquals("status", "deprecated")).
		Add().
		Execute(context.Background())

	if err != nil {
		t.Fatalf("Transaction() error = %v", err)
	}

	// Verify the request body
	if !strings.Contains(receivedBody, "<wfs:Delete") {
		t.Error("expected request body to contain Delete element")
	}

	// Verify the response
	if resp.TransactionSummary.TotalDeleted != 3 {
		t.Errorf("expected 3 deleted, got %d", resp.TransactionSummary.TotalDeleted)
	}
}

func TestTransactionMultipleOperations(t *testing.T) {
	var receivedBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)

		w.Header().Set("Content-Type", "application/xml")
		io.WriteString(w, `<?xml version="1.0" encoding="UTF-8"?>
<wfs:TransactionResponse version="2.0.0" xmlns:wfs="http://www.opengis.net/wfs/2.0">
  <wfs:TransactionSummary>
    <wfs:totalInserted>2</wfs:totalInserted>
    <wfs:totalUpdated>1</wfs:totalUpdated>
    <wfs:totalDeleted>1</wfs:totalDeleted>
  </wfs:TransactionSummary>
</wfs:TransactionResponse>`)
	}))
	defer server.Close()

	client := NewClient(server.URL, nil, nil, "")

	resp, err := client.Transaction().
		Insert("roads").
		Feature(map[string]interface{}{"name": "Road 1"}).
		Feature(map[string]interface{}{"name": "Road 2"}).
		Add().
		Update("roads").
		Set("status", "active").
		Filter(PropertyEquals("id", 100)).
		Add().
		Delete("roads").
		Filter(PropertyEquals("status", "deprecated")).
		Add().
		Execute(context.Background())

	if err != nil {
		t.Fatalf("Transaction() error = %v", err)
	}

	// Verify all operations are present
	if !strings.Contains(receivedBody, "<wfs:Insert>") {
		t.Error("expected request body to contain Insert element")
	}
	if !strings.Contains(receivedBody, "<wfs:Update") {
		t.Error("expected request body to contain Update element")
	}
	if !strings.Contains(receivedBody, "<wfs:Delete") {
		t.Error("expected request body to contain Delete element")
	}

	// Verify totals
	total := resp.TransactionSummary.TotalInserted +
		resp.TransactionSummary.TotalUpdated +
		resp.TransactionSummary.TotalDeleted
	if total != 4 {
		t.Errorf("expected total of 4 operations, got %d", total)
	}
}

func TestTransactionEmpty(t *testing.T) {
	client := NewClient("http://localhost:8080", nil, nil, "")

	_, err := client.Transaction().Execute(context.Background())

	if err == nil {
		t.Error("expected error for empty transaction")
	}
}

func TestInsertBuilder(t *testing.T) {
	tb := &TransactionBuilder{
		inserts: make([]InsertOperation, 0),
	}

	tb.Insert("roads").
		Feature(map[string]interface{}{"name": "Road 1"}).
		Feature(map[string]interface{}{"name": "Road 2"}).
		Handle("insert-1").
		SRSName("EPSG:4326").
		Add()

	if len(tb.inserts) != 1 {
		t.Fatalf("expected 1 insert operation, got %d", len(tb.inserts))
	}

	insert := tb.inserts[0]
	if insert.TypeName != "roads" {
		t.Errorf("expected typeName 'roads', got '%s'", insert.TypeName)
	}
	if len(insert.Features) != 2 {
		t.Errorf("expected 2 features, got %d", len(insert.Features))
	}
	if insert.Handle != "insert-1" {
		t.Errorf("expected handle 'insert-1', got '%s'", insert.Handle)
	}
	if insert.SRSName != "EPSG:4326" {
		t.Errorf("expected SRSName 'EPSG:4326', got '%s'", insert.SRSName)
	}
}

func TestUpdateBuilder(t *testing.T) {
	tb := &TransactionBuilder{
		updates: make([]UpdateOperation, 0),
	}

	tb.Update("roads").
		Set("status", "closed").
		Set("updated_at", "2024-01-01").
		Filter(PropertyEquals("id", 1)).
		Handle("update-1").
		Add()

	if len(tb.updates) != 1 {
		t.Fatalf("expected 1 update operation, got %d", len(tb.updates))
	}

	update := tb.updates[0]
	if update.TypeName != "roads" {
		t.Errorf("expected typeName 'roads', got '%s'", update.TypeName)
	}
	if len(update.Properties) != 2 {
		t.Errorf("expected 2 properties, got %d", len(update.Properties))
	}
	if update.Handle != "update-1" {
		t.Errorf("expected handle 'update-1', got '%s'", update.Handle)
	}
}

func TestDeleteBuilder(t *testing.T) {
	tb := &TransactionBuilder{
		deletes: make([]DeleteOperation, 0),
	}

	tb.Delete("roads").
		Filter(PropertyEquals("status", "deprecated")).
		Handle("delete-1").
		Add()

	if len(tb.deletes) != 1 {
		t.Fatalf("expected 1 delete operation, got %d", len(tb.deletes))
	}

	del := tb.deletes[0]
	if del.TypeName != "roads" {
		t.Errorf("expected typeName 'roads', got '%s'", del.TypeName)
	}
	if del.Filter == nil {
		t.Error("expected filter to be set")
	}
	if del.Handle != "delete-1" {
		t.Errorf("expected handle 'delete-1', got '%s'", del.Handle)
	}
}

func TestTransactionWithLockID(t *testing.T) {
	var receivedBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)

		w.Header().Set("Content-Type", "application/xml")
		io.WriteString(w, `<?xml version="1.0" encoding="UTF-8"?>
<wfs:TransactionResponse version="2.0.0" xmlns:wfs="http://www.opengis.net/wfs/2.0">
  <wfs:TransactionSummary>
    <wfs:totalInserted>1</wfs:totalInserted>
  </wfs:TransactionSummary>
</wfs:TransactionResponse>`)
	}))
	defer server.Close()

	client := NewClient(server.URL, nil, nil, "")

	_, err := client.Transaction().
		LockID("lock-123").
		ReleaseAction("ALL").
		Insert("roads").
		Feature(map[string]interface{}{"name": "Test"}).
		Add().
		Execute(context.Background())

	if err != nil {
		t.Fatalf("Transaction() error = %v", err)
	}

	if !strings.Contains(receivedBody, `lockId="lock-123"`) {
		t.Error("expected request body to contain lockId")
	}
	if !strings.Contains(receivedBody, `releaseAction="ALL"`) {
		t.Error("expected request body to contain releaseAction")
	}
}
