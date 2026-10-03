package instacart

import (
	"encoding/json"
	"fmt"
	"time"
)

// DeliveryPage is one consumer orderDeliveriesConnection page. It is not a
// complete order history when HasNextPage is true. Delivery totals must not be
// treated as whole-order totals for multi-delivery orders.
type DeliveryPage struct {
	Deliveries  []HistoryDelivery
	EndCursor   string
	HasNextPage bool
}

// HistoryDelivery is a delivery-level record from the consumer website cache.
// Receipt URLs and their access tokens are intentionally not retained.
type HistoryDelivery struct {
	ID          string
	OrderID     string
	OrderUUID   string
	Status      string
	ServiceType string
	IsMulti     bool
	CreatedAt   time.Time
	Retailer    string
	Total       string
	Items       []HistoryItem
}

// HistoryItem contains the item metadata present in the order-history response.
// This response does not provide purchased quantities; none are inferred.
type HistoryItem struct {
	ID            string
	CurrentItemID string
	ProductID     string
	Name          string
}

// DecodeDeliveryPage decodes the JSON object that is the value of
// orderDeliveriesConnection. Callers must extract that object from its envelope.
// This parser does not fetch pages or assume a GraphQL operation name/hash.
func DecodeDeliveryPage(data []byte) (*DeliveryPage, error) {
	var wire struct {
		Nodes *[]struct {
			ID          string `json:"id"`
			OrderID     string `json:"legacyOrderId"`
			OrderUUID   string `json:"legacyOrderUuid"`
			Status      string `json:"workflowState"`
			ServiceType string `json:"serviceType"`
			IsMulti     *bool  `json:"isMulti"`
			CreatedAt   string `json:"createdAt"`
			Retailer    struct {
				Name string `json:"name"`
			} `json:"retailer"`
			Amounts struct {
				ViewSection struct {
					TotalLine struct {
						AmountString string `json:"amountString"`
					} `json:"totalLine"`
				} `json:"viewSection"`
			} `json:"amounts"`
			Items *[]struct {
				ID          string `json:"id"`
				CurrentItem *struct {
					ID            string `json:"id"`
					Name          string `json:"name"`
					BasketProduct struct {
						Item struct {
							ProductID string `json:"productId"`
						} `json:"item"`
					} `json:"basketProduct"`
				} `json:"currentItem"`
			} `json:"orderItems"`
		} `json:"nodes"`
		PageInfo *struct {
			EndCursor   *string `json:"endCursor"`
			HasNextPage *bool   `json:"hasNextPage"`
		} `json:"pageInfo"`
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return nil, fmt.Errorf("decode delivery page: %w", err)
	}
	if wire.Nodes == nil || wire.PageInfo == nil || wire.PageInfo.HasNextPage == nil {
		return nil, fmt.Errorf("delivery page missing nodes or pageInfo.hasNextPage")
	}
	page := &DeliveryPage{HasNextPage: *wire.PageInfo.HasNextPage, Deliveries: make([]HistoryDelivery, 0, len(*wire.Nodes))}
	if wire.PageInfo.EndCursor != nil {
		page.EndCursor = *wire.PageInfo.EndCursor
	}
	if page.HasNextPage && (page.EndCursor == "" || len(*wire.Nodes) == 0) {
		return nil, fmt.Errorf("advancing delivery page requires a cursor and nonempty nodes")
	}
	seen := make(map[string]bool)
	for index, n := range *wire.Nodes {
		if n.ID == "" || n.OrderID == "" || n.Status == "" || n.IsMulti == nil || n.Retailer.Name == "" || n.Amounts.ViewSection.TotalLine.AmountString == "" || n.Items == nil {
			return nil, fmt.Errorf("delivery %d missing required fields", index)
		}
		if seen[n.ID] {
			return nil, fmt.Errorf("duplicate delivery at index %d", index)
		}
		seen[n.ID] = true
		created, err := time.Parse(time.RFC3339Nano, n.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("delivery %d has invalid createdAt", index)
		}
		delivery := HistoryDelivery{ID: n.ID, OrderID: n.OrderID, OrderUUID: n.OrderUUID, Status: n.Status, ServiceType: n.ServiceType, IsMulti: *n.IsMulti, CreatedAt: created, Retailer: n.Retailer.Name, Total: n.Amounts.ViewSection.TotalLine.AmountString, Items: make([]HistoryItem, 0, len(*n.Items))}
		for itemIndex, item := range *n.Items {
			if item.ID == "" || item.CurrentItem == nil || item.CurrentItem.Name == "" {
				return nil, fmt.Errorf("delivery %d item %d missing item data", index, itemIndex)
			}
			delivery.Items = append(delivery.Items, HistoryItem{ID: item.ID, CurrentItemID: item.CurrentItem.ID, ProductID: item.CurrentItem.BasketProduct.Item.ProductID, Name: item.CurrentItem.Name})
		}
		page.Deliveries = append(page.Deliveries, delivery)
	}
	return page, nil
}
