package main

import (
	"bytes"
	"encoding/csv"
	instacart "github.com/beezyfbaby/go-instacart-export"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestCSV(t *testing.T) {
	orders := []*instacart.Order{{ID: "1", Status: "delivered", Total: "12.34", CreatedAt: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC), Deliveries: []*instacart.Delivery{{Retailer: "Store, Inc.", Items: []*instacart.Item{{Quantity: 1.5}, {Quantity: 2}}}}}}
	data := extractOrdersData(orders)
	want := []string{"1", "delivered", "12.34", "2025-01-02", "Store, Inc.", "2"}
	if !reflect.DeepEqual(data[1], want) {
		t.Fatalf("%v", data)
	}
	path := filepath.Join(t.TempDir(), "nested", "orders.csv")
	if err := writeToCSV(path, data); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got, err := csv.NewReader(bytes.NewReader(body)).ReadAll()
	if err != nil || !reflect.DeepEqual(got, data) {
		t.Fatalf("%v %v", got, err)
	}
	if err := writeToCSV(path, [][]string{{"overwrite"}}); err == nil {
		t.Fatal("overwrote existing file")
	}
	after, _ := os.ReadFile(path)
	if !bytes.Equal(body, after) {
		t.Fatal("changed existing file")
	}
}
func TestFormulaSafety(t *testing.T) {
	for _, s := range []string{"=1+1", "+CMD", "-1", "@SUM(A1)", "  =1", "\ttext", "\rtext", "\ntext"} {
		if !strings.HasPrefix(spreadsheetSafe(s), "'") {
			t.Fatalf("unsafe %q", s)
		}
	}
	if spreadsheetSafe("12.34") != "12.34" {
		t.Fatal("changed normal total")
	}
}
func TestCLI(t *testing.T) {
	t.Setenv("INSTACART_SESSION_TOKEN", "")
	t.Setenv("INSTACART_HOST_SESSION_TOKEN", "")
	var out bytes.Buffer
	if err := run(t.Context(), []string{"-h"}, &out, &out); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"-timeout=0"}, {"unexpected"}, {}} {
		if err := run(t.Context(), args, &out, &out); err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
}

func TestDeliveryCSV(t *testing.T) {
	data := extractDeliveryData([]instacart.HistoryDelivery{{ID: "delivery-1", OrderID: "order-1", Status: "delivered", Total: "$54.12", CreatedAt: time.Date(2026, 1, 2, 15, 4, 5, 0, time.UTC), Retailer: "=FORMULA()", Items: []instacart.HistoryItem{{Name: "Item"}}, IsMulti: true, ServiceType: "delivery"}})
	want := []string{"delivery-1", "order-1", "delivered", "$54.12", "2026-01-02T15:04:05Z", "'=FORMULA()", "1", "true", "delivery"}
	if !reflect.DeepEqual(data[1], want) {
		t.Fatalf("wrong CSV mapping: %v", data)
	}
	if data[0][0] != "deliveryId" || data[0][3] != "deliveryTotal" {
		t.Fatal("ambiguous totals or identifiers")
	}
}

type testTransport func(*http.Request) (*http.Response, error)

func (f testTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestCLIHistoryEndToEnd(t *testing.T) {
	t.Setenv("INSTACART_SESSION_TOKEN", "test-session")
	original := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = original })
	failed := false
	http.DefaultTransport = testTransport(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/graphql" {
			t.Fatal("CLI used legacy endpoint")
		}
		body := `{"data":{"orderDeliveriesConnection":{"nodes":[],"pageInfo":{"hasNextPage":false,"endCursor":null}}}}`
		if failed {
			body = `{"errors":[{"message":"authentication failed"}]}`
		}
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
	})
	path := filepath.Join(t.TempDir(), "history.csv")
	var out bytes.Buffer
	if err := run(t.Context(), []string{"-output", path}, &out, &out); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil || !strings.HasPrefix(string(b), "deliveryId,orderId,") {
		t.Fatalf("bad output: %s %v", b, err)
	}
	failed = true
	path = filepath.Join(t.TempDir(), "failed.csv")
	if err := run(t.Context(), []string{"-output", path}, &out, &out); err == nil {
		t.Fatal("ignored GraphQL error")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("created CSV after failed fetch")
	}
}
