package instacart

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"slices"
)

// PersonalOrderHistoryHash was observed in the consumer website on 2026-10-03.
// This is a query identifier, not an authentication credential.
const PersonalOrderHistoryHash = "ffde6dab3a502e1cffa3b908b616778300bd9c91b2ff95be1cb90444820e4e66"

// FetchDeliveryHistory retrieves all delivery pages, newest first. An empty hash
// uses the observed default. A changed website hash can be supplied explicitly.
// No partial history is returned on errors. Each record is one delivery, not
// necessarily one complete order.
func FetchDeliveryHistory(ctx context.Context, client Client, hash string) ([]HistoryDelivery, error) {
	if hash == "" {
		hash = PersonalOrderHistoryHash
	}
	decoded, err := hex.DecodeString(hash)
	if err != nil || len(decoded) != 32 {
		return nil, fmt.Errorf("query hash must contain 64 hexadecimal characters")
	}
	deliveries := make([]HistoryDelivery, 0)
	seenIDs := map[string]bool{}
	seenCursors := map[string]bool{}
	after := ""
	for pageNumber := 1; ; pageNumber++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		variables := map[string]any{"first": 10}
		// Initial-page request omits the optional cursor. Subsequent requests use the
		// exact cursor returned by pageInfo, never an assumed numeric increment.
		if after != "" {
			variables["after"] = after
		}
		variablesJSON, _ := json.Marshal(variables)
		extensionsJSON, _ := json.Marshal(map[string]any{"persistedQuery": map[string]any{"version": 1, "sha256Hash": hash}})
		query := url.Values{"operationName": {"PersonalOrderHistory"}, "variables": {string(variablesJSON)}, "extensions": {string(extensionsJSON)}}
		body, err := client.getJSON(ctx, "https://www.instacart.com/graphql?"+query.Encode())
		if err != nil {
			return nil, fmt.Errorf("history page %d: %w", pageNumber, err)
		}
		page, err := decodeHistoryResponse(body)
		if err != nil {
			return nil, fmt.Errorf("history page %d: %w", pageNumber, err)
		}
		for _, d := range page.Deliveries {
			if seenIDs[d.ID] {
				return nil, fmt.Errorf("duplicate delivery across history pages; retry when history is stable")
			}
			seenIDs[d.ID] = true
			deliveries = append(deliveries, d)
		}
		if !page.HasNextPage {
			break
		}
		if seenCursors[page.EndCursor] {
			return nil, fmt.Errorf("repeated history cursor; refusing an incomplete export")
		}
		seenCursors[page.EndCursor] = true
		after = page.EndCursor
	}
	slices.SortStableFunc(deliveries, func(a, b HistoryDelivery) int { return b.CreatedAt.Compare(a.CreatedAt) })
	return deliveries, nil
}

func decodeHistoryResponse(body []byte) (*DeliveryPage, error) {
	var envelope struct {
		Data   json.RawMessage `json:"data"`
		Errors []struct {
			Message    string `json:"message"`
			Extensions struct {
				Code string `json:"code"`
			} `json:"extensions"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("invalid GraphQL JSON: %w", err)
	}
	if len(envelope.Errors) > 0 {
		for _, e := range envelope.Errors {
			if e.Message == "PersistedQueryNotFound" || e.Extensions.Code == "PERSISTED_QUERY_NOT_FOUND" {
				return nil, fmt.Errorf("persisted query no longer registered; obtain the current PersonalOrderHistory hash from your browser and use -query-hash")
			}
		}
		// Do not echo arbitrary server messages that may contain account information.
		return nil, fmt.Errorf("GraphQL returned errors; check your session and the current PersonalOrderHistory request")
	}
	var data any
	if len(envelope.Data) == 0 || json.Unmarshal(envelope.Data, &data) != nil || data == nil {
		return nil, fmt.Errorf("GraphQL response missing data")
	}
	// The supplied cache excerpt starts at the connection, without its ancestors.
	// Locate exactly one named connection under data rather than guessing a root.
	var connections []any
	var visit func(any)
	visit = func(v any) {
		switch x := v.(type) {
		case map[string]any:
			for k, value := range x {
				if k == "orderDeliveriesConnection" {
					connections = append(connections, value)
				} else {
					visit(value)
				}
			}
		case []any:
			for _, value := range x {
				visit(value)
			}
		}
	}
	visit(data)
	if len(connections) != 1 || connections[0] == nil {
		return nil, fmt.Errorf("expected exactly one orderDeliveriesConnection in GraphQL data")
	}
	connection, _ := json.Marshal(connections[0])
	return DecodeDeliveryPage(connection)
}
