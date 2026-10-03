package instacart

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func response(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}
}
func orderJSON(id, date string) string {
	return fmt.Sprintf(`{"id":%q,"status":"delivered","total":"12.34","created_at":%q,"order_deliveries":[{"retailer":{"name":"Store"},"order_items":[{"qty":1.5,"item":{"id":"item","product_id":"product","name":"Fruit","qty_attributes":{"increment":0.25}}}]}]}`, id, date)
}
func pageJSON(orders, next string) string {
	return `{"orders":[` + orders + `],"meta":{"pagination":{"next_page":` + next + `}}}`
}

func TestFetchOrders(t *testing.T) {
	calls := 0
	c := Client{SessionToken: "secret", HTTPClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.URL.Scheme != "https" || r.URL.Host != "www.instacart.com" {
			t.Fatal("wrong destination")
		}
		cookie, err := r.Cookie("_instacart_session_id")
		if err != nil || cookie.Value != "secret" {
			t.Fatal("missing cookie")
		}
		switch r.URL.Query().Get("page") {
		case "1":
			return response(200, pageJSON(orderJSON("older", "Jan 2, 2024,  3:04 PM"), "2")), nil
		case "2":
			return response(200, pageJSON(orderJSON("newer", "2025-02-03T10:00:00Z"), "null")), nil
		default:
			t.Fatal("unexpected extra page")
			return nil, nil
		}
	})}}
	orders, err := FetchOrders(t.Context(), c)
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 || len(orders) != 2 || orders[0].ID != "newer" || orders[0].Deliveries[0].Items[0].Quantity != 1.5 {
		t.Fatalf("incorrect orders: %#v", orders)
	}
}

func TestFetchFailures(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
	}{
		{"unauthorized", 401, "secret body"}, {"blocked", 403, "secret body"}, {"rate limited", 429, "secret body"}, {"server", 500, "secret body"}, {"redirect", 302, ""},
		{"html", 200, "<html>login</html>"}, {"changed schema", 200, `{"data":{}}`}, {"null orders", 200, `{"orders":null,"meta":{"pagination":{"next_page":null}}}`},
		{"missing pagination", 200, `{"orders":[]}`}, {"missing next", 200, `{"orders":[],"meta":{"pagination":{}}}`},
		{"loop", 200, pageJSON(orderJSON("x", "2025-01-01"), "1")}, {"negative", 200, pageJSON(orderJSON("x", "2025-01-01"), "-1")},
		{"empty advancing", 200, pageJSON("", "2")}, {"invalid next", 200, pageJSON("", `"two"`)},
		{"invalid date", 200, pageJSON(orderJSON("x", "tomorrow"), "null")}, {"missing fields", 200, pageJSON(`{"id":"x"}`, "null")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			c := Client{SessionToken: "secret", HTTPClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) { calls++; return response(tc.status, tc.body), nil })}}
			orders, err := FetchOrders(t.Context(), c)
			if err == nil || orders != nil || calls != 1 {
				t.Fatalf("orders=%v err=%v calls=%d", orders, err, calls)
			}
			if strings.Contains(err.Error(), "secret") {
				t.Fatal("leaked secret")
			}
		})
	}
}
func TestTerminalPages(t *testing.T) {
	for _, end := range []string{"null", "0"} {
		t.Run(end, func(t *testing.T) {
			c := Client{SessionToken: "secret", HTTPClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) { return response(200, pageJSON("", end)), nil })}}
			orders, err := FetchOrders(t.Context(), c)
			if err != nil || len(orders) != 0 {
				t.Fatalf("%v %v", orders, err)
			}
		})
	}
}
func TestDuplicateOrders(t *testing.T) {
	c := Client{SessionToken: "secret", HTTPClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		next := "2"
		if r.URL.Query().Get("page") == "2" {
			next = "null"
		}
		return response(200, pageJSON(orderJSON("same", "2025-01-01"), next)), nil
	})}}
	if _, err := FetchOrders(t.Context(), c); err == nil {
		t.Fatal("accepted duplicate orders")
	}
}
func TestCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := FetchOrders(ctx, Client{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("%v", err)
	}
	c := Client{SessionToken: "secret", HTTPClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) { <-r.Context().Done(); return nil, r.Context().Err() })}}
	ctx, cancel = context.WithTimeout(t.Context(), time.Millisecond)
	defer cancel()
	if _, err := FetchOrders(ctx, c); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("%v", err)
	}
}
func TestInvalidCookie(t *testing.T) {
	for _, token := range []string{"", "bad;cookie", "bad\nheader"} {
		c := Client{SessionToken: token, HTTPClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) { t.Fatal("sent invalid cookie"); return nil, nil })}}
		if _, err := FetchOrders(t.Context(), c); err == nil {
			t.Fatal("accepted invalid cookie")
		}
	}
}
