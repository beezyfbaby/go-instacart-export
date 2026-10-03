package instacart

import (
	"encoding/json"
	"strings"
	"testing"
)

// Synthetic fixture: no account IDs, purchases, or receipt tokens from users.
const sampleDelivery = `{"id":"delivery-a","legacyOrderId":"order-a","legacyOrderUuid":"uuid-a","workflowState":"delivered","serviceType":"delivery","isMulti":true,"createdAt":"2026-01-02T15:04:05Z","retailer":{"name":"Example Store"},"amounts":{"viewSection":{"totalLine":{"amountString":"$54.12"}}},"orderItems":[{"id":"line-a","currentItem":{"id":"item-a","name":"Example item","basketProduct":{"item":{"productId":"product-a"}}}}],"receiptUrl":"https://example.invalid/receipt?token=DO_NOT_EXPORT"}`

func TestDecodeDeliveryPage(t *testing.T) {
	input := `{"nodes":[` + sampleDelivery + `],"pageInfo":{"endCursor":"cursor-a","hasNextPage":true}}`
	page, err := DecodeDeliveryPage([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	if !page.HasNextPage || page.EndCursor != "cursor-a" || len(page.Deliveries) != 1 {
		t.Fatalf("bad pagination: %#v", page)
	}
	d := page.Deliveries[0]
	if d.OrderID != "order-a" || d.ID != "delivery-a" || !d.IsMulti || d.Total != "$54.12" || d.Retailer != "Example Store" || len(d.Items) != 1 || d.Items[0].ProductID != "product-a" {
		t.Fatalf("bad mapping: %#v", d)
	}
	serialized, _ := json.Marshal(page)
	if strings.Contains(string(serialized), "DO_NOT_EXPORT") || strings.Contains(string(serialized), "receipt") {
		t.Fatal("retained receipt URL")
	}
}
func TestDecodeDeliveryPageRejectsIncomplete(t *testing.T) {
	for _, input := range []string{
		`{}`, `{"nodes":[],"pageInfo":{}}`, `{"nodes":null,"pageInfo":{"hasNextPage":false}}`,
		`{"nodes":[],"pageInfo":{"hasNextPage":true,"endCursor":"a"}}`,
		`{"nodes":[` + sampleDelivery + `],"pageInfo":{"hasNextPage":true}}`,
		`{"nodes":[{}],"pageInfo":{"hasNextPage":false}}`,
		`{"nodes":[` + sampleDelivery + `,` + sampleDelivery + `],"pageInfo":{"hasNextPage":false}}`,
		`{"nodes":[` + strings.Replace(sampleDelivery, `"isMulti":true,`, "", 1) + `],"pageInfo":{"hasNextPage":false}}`,
		`{"nodes":[` + strings.Replace(sampleDelivery, `2026-01-02T15:04:05Z`, `invalid`, 1) + `],"pageInfo":{"hasNextPage":false}}`,
		`{"nodes":[` + strings.Replace(sampleDelivery, `"name":"Example item"`, `"name":""`, 1) + `],"pageInfo":{"hasNextPage":false}}`,
	} {
		if page, err := DecodeDeliveryPage([]byte(input)); err == nil || page != nil {
			t.Fatalf("accepted incomplete page: %s", input)
		}
	}
}
func TestDecodeDeliveryPageFinalEmpty(t *testing.T) {
	page, err := DecodeDeliveryPage([]byte(`{"nodes":[],"pageInfo":{"hasNextPage":false,"endCursor":null}}`))
	if err != nil || page.HasNextPage || len(page.Deliveries) != 0 {
		t.Fatalf("%v %v", page, err)
	}
}
