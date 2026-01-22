package wfs

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
)

// TransactionBuilder builds WFS-T transaction requests.
type TransactionBuilder struct {
	client        *Client
	inserts       []InsertOperation
	updates       []UpdateOperation
	deletes       []DeleteOperation
	lockID        string
	releaseAction string
}

// InsertOperation represents an insert operation.
type InsertOperation struct {
	TypeName    string
	Features    []map[string]interface{}
	Handle      string
	InputFormat string
	SRSName     string
}

// UpdateOperation represents an update operation.
type UpdateOperation struct {
	TypeName   string
	Properties map[string]interface{}
	Filter     Filter
	CQLFilter  string
	Handle     string
	SRSName    string
}

// DeleteOperation represents a delete operation.
type DeleteOperation struct {
	TypeName  string
	Filter    Filter
	CQLFilter string
	Handle    string
}

// Insert adds an insert operation.
func (t *TransactionBuilder) Insert(typeName string) *InsertBuilder {
	return &InsertBuilder{
		transaction: t,
		typeName:    typeName,
		features:    make([]map[string]interface{}, 0),
	}
}

// Update adds an update operation.
func (t *TransactionBuilder) Update(typeName string) *UpdateBuilder {
	return &UpdateBuilder{
		transaction: t,
		typeName:    typeName,
		properties:  make(map[string]interface{}),
	}
}

// Delete adds a delete operation.
func (t *TransactionBuilder) Delete(typeName string) *DeleteBuilder {
	return &DeleteBuilder{
		transaction: t,
		typeName:    typeName,
	}
}

// LockID sets the lock ID for the transaction.
func (t *TransactionBuilder) LockID(lockID string) *TransactionBuilder {
	t.lockID = lockID
	return t
}

// ReleaseAction sets the release action (ALL or SOME).
func (t *TransactionBuilder) ReleaseAction(action string) *TransactionBuilder {
	t.releaseAction = action
	return t
}

// Execute performs the transaction.
func (t *TransactionBuilder) Execute(ctx context.Context) (*TransactionResponse, error) {
	if len(t.inserts) == 0 && len(t.updates) == 0 && len(t.deletes) == 0 {
		return nil, fmt.Errorf("transaction has no operations")
	}

	// Build the transaction XML
	xmlBody, err := t.buildXML()
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction XML: %w", err)
	}

	// Send the request
	body, resp, err := t.client.http.Post("").
		Param("SERVICE", ServiceName).
		BodyXML(xmlBody).
		DoAndRead(ctx)

	if err != nil {
		return nil, fmt.Errorf("transaction request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, parseWFSException(body, resp.StatusCode)
	}

	// Parse the response
	var txResp TransactionResponse
	if err := xml.Unmarshal(body, &txResp); err != nil {
		return nil, fmt.Errorf("failed to parse transaction response: %w", err)
	}

	return &txResp, nil
}

// buildXML builds the transaction XML document.
func (t *TransactionBuilder) buildXML() ([]byte, error) {
	var sb strings.Builder

	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	sb.WriteString("\n")
	sb.WriteString(`<wfs:Transaction version="2.0.0" service="WFS"`)
	sb.WriteString(` xmlns:wfs="http://www.opengis.net/wfs/2.0"`)
	sb.WriteString(` xmlns:fes="http://www.opengis.net/fes/2.0"`)
	sb.WriteString(` xmlns:gml="http://www.opengis.net/gml/3.2"`)
	sb.WriteString(` xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"`)

	if t.lockID != "" {
		sb.WriteString(fmt.Sprintf(` lockId="%s"`, escapeXML(t.lockID)))
	}
	if t.releaseAction != "" {
		sb.WriteString(fmt.Sprintf(` releaseAction="%s"`, escapeXML(t.releaseAction)))
	}

	sb.WriteString(">\n")

	// Write inserts
	for _, insert := range t.inserts {
		sb.WriteString(t.buildInsertXML(insert))
	}

	// Write updates
	for _, update := range t.updates {
		sb.WriteString(t.buildUpdateXML(update))
	}

	// Write deletes
	for _, del := range t.deletes {
		sb.WriteString(t.buildDeleteXML(del))
	}

	sb.WriteString("</wfs:Transaction>")

	return []byte(sb.String()), nil
}

// buildInsertXML builds the XML for an insert operation.
func (t *TransactionBuilder) buildInsertXML(insert InsertOperation) string {
	var sb strings.Builder

	sb.WriteString("  <wfs:Insert")
	if insert.Handle != "" {
		sb.WriteString(fmt.Sprintf(` handle="%s"`, escapeXML(insert.Handle)))
	}
	if insert.InputFormat != "" {
		sb.WriteString(fmt.Sprintf(` inputFormat="%s"`, escapeXML(insert.InputFormat)))
	}
	if insert.SRSName != "" {
		sb.WriteString(fmt.Sprintf(` srsName="%s"`, escapeXML(insert.SRSName)))
	}
	sb.WriteString(">\n")

	// Write each feature
	for _, feature := range insert.Features {
		sb.WriteString(fmt.Sprintf("    <%s>\n", insert.TypeName))
		for prop, val := range feature {
			sb.WriteString(fmt.Sprintf("      <%s>%v</%s>\n", prop, escapeXML(fmt.Sprintf("%v", val)), prop))
		}
		sb.WriteString(fmt.Sprintf("    </%s>\n", insert.TypeName))
	}

	sb.WriteString("  </wfs:Insert>\n")
	return sb.String()
}

// buildUpdateXML builds the XML for an update operation.
func (t *TransactionBuilder) buildUpdateXML(update UpdateOperation) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("  <wfs:Update typeName=\"%s\"", escapeXML(update.TypeName)))
	if update.Handle != "" {
		sb.WriteString(fmt.Sprintf(` handle="%s"`, escapeXML(update.Handle)))
	}
	if update.SRSName != "" {
		sb.WriteString(fmt.Sprintf(` srsName="%s"`, escapeXML(update.SRSName)))
	}
	sb.WriteString(">\n")

	// Write properties to update
	for prop, val := range update.Properties {
		sb.WriteString("    <wfs:Property>\n")
		sb.WriteString(fmt.Sprintf("      <wfs:ValueReference>%s</wfs:ValueReference>\n", escapeXML(prop)))
		sb.WriteString(fmt.Sprintf("      <wfs:Value>%v</wfs:Value>\n", escapeXML(fmt.Sprintf("%v", val))))
		sb.WriteString("    </wfs:Property>\n")
	}

	// Write filter
	if update.Filter != nil {
		filterXML := update.Filter.ToXML()
		// Remove the outer Filter element and namespace declarations
		filterXML = strings.Replace(filterXML, `<fes:Filter xmlns:fes="http://www.opengis.net/fes/2.0">`, "<fes:Filter>", 1)
		filterXML = strings.Replace(filterXML, `<fes:Filter xmlns:fes="http://www.opengis.net/fes/2.0" xmlns:gml="http://www.opengis.net/gml/3.2">`, "<fes:Filter>", 1)
		sb.WriteString("    ")
		sb.WriteString(filterXML)
		sb.WriteString("\n")
	} else if update.CQLFilter != "" {
		// For CQL filter, we need to use the GeoServer vendor parameter approach
		// or convert CQL to Filter (complex, not implemented here)
		sb.WriteString(fmt.Sprintf("    <!-- CQL_FILTER: %s -->\n", escapeXML(update.CQLFilter)))
	}

	sb.WriteString("  </wfs:Update>\n")
	return sb.String()
}

// buildDeleteXML builds the XML for a delete operation.
func (t *TransactionBuilder) buildDeleteXML(del DeleteOperation) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("  <wfs:Delete typeName=\"%s\"", escapeXML(del.TypeName)))
	if del.Handle != "" {
		sb.WriteString(fmt.Sprintf(` handle="%s"`, escapeXML(del.Handle)))
	}
	sb.WriteString(">\n")

	// Write filter
	if del.Filter != nil {
		filterXML := del.Filter.ToXML()
		// Remove the outer Filter element namespace declarations
		filterXML = strings.Replace(filterXML, `<fes:Filter xmlns:fes="http://www.opengis.net/fes/2.0">`, "<fes:Filter>", 1)
		filterXML = strings.Replace(filterXML, `<fes:Filter xmlns:fes="http://www.opengis.net/fes/2.0" xmlns:gml="http://www.opengis.net/gml/3.2">`, "<fes:Filter>", 1)
		sb.WriteString("    ")
		sb.WriteString(filterXML)
		sb.WriteString("\n")
	}

	sb.WriteString("  </wfs:Delete>\n")
	return sb.String()
}

// InsertBuilder builds an insert operation.
type InsertBuilder struct {
	transaction *TransactionBuilder
	typeName    string
	features    []map[string]interface{}
	handle      string
	inputFormat string
	srsName     string
}

// Feature adds a feature to insert.
func (b *InsertBuilder) Feature(properties map[string]interface{}) *InsertBuilder {
	b.features = append(b.features, properties)
	return b
}

// Features adds multiple features to insert.
func (b *InsertBuilder) Features(features ...map[string]interface{}) *InsertBuilder {
	b.features = append(b.features, features...)
	return b
}

// Handle sets the operation handle.
func (b *InsertBuilder) Handle(handle string) *InsertBuilder {
	b.handle = handle
	return b
}

// InputFormat sets the input format.
func (b *InsertBuilder) InputFormat(format string) *InsertBuilder {
	b.inputFormat = format
	return b
}

// SRSName sets the spatial reference system.
func (b *InsertBuilder) SRSName(srs string) *InsertBuilder {
	b.srsName = srs
	return b
}

// Add finalizes the insert operation and returns the transaction builder.
func (b *InsertBuilder) Add() *TransactionBuilder {
	b.transaction.inserts = append(b.transaction.inserts, InsertOperation{
		TypeName:    b.typeName,
		Features:    b.features,
		Handle:      b.handle,
		InputFormat: b.inputFormat,
		SRSName:     b.srsName,
	})
	return b.transaction
}

// UpdateBuilder builds an update operation.
type UpdateBuilder struct {
	transaction *TransactionBuilder
	typeName    string
	properties  map[string]interface{}
	filter      Filter
	cqlFilter   string
	handle      string
	srsName     string
}

// Set sets a property value to update.
func (b *UpdateBuilder) Set(property string, value interface{}) *UpdateBuilder {
	b.properties[property] = value
	return b
}

// Properties sets multiple property values.
func (b *UpdateBuilder) Properties(properties map[string]interface{}) *UpdateBuilder {
	for k, v := range properties {
		b.properties[k] = v
	}
	return b
}

// Filter sets the filter for features to update.
func (b *UpdateBuilder) Filter(filter Filter) *UpdateBuilder {
	b.filter = filter
	return b
}

// CQLFilter sets a CQL filter.
func (b *UpdateBuilder) CQLFilter(filter string) *UpdateBuilder {
	b.cqlFilter = filter
	return b
}

// Handle sets the operation handle.
func (b *UpdateBuilder) Handle(handle string) *UpdateBuilder {
	b.handle = handle
	return b
}

// SRSName sets the spatial reference system.
func (b *UpdateBuilder) SRSName(srs string) *UpdateBuilder {
	b.srsName = srs
	return b
}

// Add finalizes the update operation and returns the transaction builder.
func (b *UpdateBuilder) Add() *TransactionBuilder {
	b.transaction.updates = append(b.transaction.updates, UpdateOperation{
		TypeName:   b.typeName,
		Properties: b.properties,
		Filter:     b.filter,
		CQLFilter:  b.cqlFilter,
		Handle:     b.handle,
		SRSName:    b.srsName,
	})
	return b.transaction
}

// DeleteBuilder builds a delete operation.
type DeleteBuilder struct {
	transaction *TransactionBuilder
	typeName    string
	filter      Filter
	cqlFilter   string
	handle      string
}

// Filter sets the filter for features to delete.
func (b *DeleteBuilder) Filter(filter Filter) *DeleteBuilder {
	b.filter = filter
	return b
}

// CQLFilter sets a CQL filter.
func (b *DeleteBuilder) CQLFilter(filter string) *DeleteBuilder {
	b.cqlFilter = filter
	return b
}

// Handle sets the operation handle.
func (b *DeleteBuilder) Handle(handle string) *DeleteBuilder {
	b.handle = handle
	return b
}

// Add finalizes the delete operation and returns the transaction builder.
func (b *DeleteBuilder) Add() *TransactionBuilder {
	b.transaction.deletes = append(b.transaction.deletes, DeleteOperation{
		TypeName:  b.typeName,
		Filter:    b.filter,
		CQLFilter: b.cqlFilter,
		Handle:    b.handle,
	})
	return b.transaction
}
