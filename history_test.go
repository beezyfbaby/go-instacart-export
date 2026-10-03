package instacart

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
)

func historyResponse(node, cursor string, more bool) string {
	b, _ := json.Marshal(map[string]any{"data": map[string]any{"orderDeliveriesConnection": map[string]any{"nodes": []json.RawMessage{json.RawMessage(node)}, "pageInfo": map[string]any{"endCursor": cursor, "hasNextPage": more}}}})
	return string(b)
}
func TestFetchDeliveryHistoryRequest(t *testing.T) {
	calls := 0
	c := Client{SessionToken: "secret", HTTPClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Method != "GET" || r.URL.Host != "www.instacart.com" || r.URL.Path != "/graphql" {
			t.Fatal("wrong request")
		}
		q := r.URL.Query()
		if q.Get("operationName") != "PersonalOrderHistory" {
			t.Fatal("wrong operation")
		}
		var v struct {
			First int
			After *string
		}
		if err := json.Unmarshal([]byte(q.Get("variables")), &v); err != nil {
			t.Fatal(err)
		}
		var e struct {
			PersistedQuery struct {
				Version    int
				Sha256Hash string
			}
		}
		if err := json.Unmarshal([]byte(q.Get("extensions")), &e); err != nil {
			t.Fatal(err)
		}
		if v.First != 10 || e.PersistedQuery.Version != 1 || e.PersistedQuery.Sha256Hash != PersonalOrderHistoryHash {
			t.Fatal("wrong query parameters")
		}
		cookie, err := r.Cookie("_instacart_session_id")
		if err != nil || cookie.Value != "secret" {
			t.Fatal("missing session")
		}
		if calls == 1 {
			if v.After != nil {
				t.Fatal("first page must not start at captured second-page cursor")
			}
			return response(200, historyResponse(sampleDelivery, "opaque+/=cursor", true)), nil
		}
		if calls != 2 || v.After == nil || *v.After != "opaque+/=cursor" {
			t.Fatal("incorrect next page")
		}
		return response(200, historyResponse(strings.ReplaceAll(sampleDelivery, "delivery-a", "delivery-b"), "end", false)), nil
	})}}
	result, err := FetchDeliveryHistory(t.Context(), c, "")
	if err != nil || len(result) != 2 || calls != 2 {
		t.Fatalf("result=%v err=%v calls=%d", result, err, calls)
	}
	// Both deliveries intentionally share a parent order ID and must remain separate.
	if result[0].OrderID != result[1].OrderID {
		t.Fatal("changed parent IDs")
	}
}
func TestHistoryErrors(t *testing.T) {
	for _, body := range []string{`<html>login</html>`, `{}`, `{"data":null}`, `{"data":{"orderDeliveriesConnection":null}}`, `{"errors":[{"message":"secret account error"}],"data":{}}`, `{"errors":[{"message":"PersistedQueryNotFound"}]}`, `{"data":{"a":{"orderDeliveriesConnection":{}},"b":{"orderDeliveriesConnection":{}}}}`} {
		c := Client{SessionToken: "secret", HTTPClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) { return response(200, body), nil })}}
		got, err := FetchDeliveryHistory(t.Context(), c, "")
		if err == nil || got != nil {
			t.Fatal("accepted invalid response")
		}
		if strings.Contains(err.Error(), "secret") {
			t.Fatal("leaked server message")
		}
	}
}
func TestHistoryRepeatedCursorAndDuplicate(t *testing.T) {
	for _, duplicate := range []bool{true, false} {
		calls := 0
		c := Client{SessionToken: "secret", HTTPClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			calls++
			n := sampleDelivery
			if !duplicate && calls == 2 {
				n = strings.ReplaceAll(n, "delivery-a", "delivery-b")
			}
			return response(200, historyResponse(n, "same", true)), nil
		})}}
		if got, err := FetchDeliveryHistory(t.Context(), c, ""); err == nil || got != nil || calls != 2 {
			t.Fatalf("calls=%d err=%v", calls, err)
		}
	}
}
func TestHistoryEmptyAndCancellation(t *testing.T) {
	c := Client{SessionToken: "secret", HTTPClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return response(200, `{"data":{"orderDeliveriesConnection":{"nodes":[],"pageInfo":{"endCursor":null,"hasNextPage":false}}}}`), nil
	})}}
	if got, err := FetchDeliveryHistory(t.Context(), c, ""); err != nil || len(got) != 0 {
		t.Fatalf("%v %v", got, err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := FetchDeliveryHistory(ctx, c, ""); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := FetchDeliveryHistory(t.Context(), c, "invalid"); err == nil {
		t.Fatal("accepted bad hash")
	}
}
func TestHistoryHashOverride(t *testing.T) {
	hash := strings.Repeat("a", 64)
	c := Client{SessionToken: "secret", HTTPClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if !strings.Contains(r.URL.Query().Get("extensions"), hash) {
			t.Fatal("ignored override")
		}
		return response(401, ""), nil
	})}}
	if _, err := FetchDeliveryHistory(t.Context(), c, hash); err == nil {
		t.Fatal("ignored authentication failure")
	}
}
