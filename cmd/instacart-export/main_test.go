package main

import (
	"bytes"
	"encoding/csv"
	instacart "github.com/beezyfbaby/go-instacart-export"
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
