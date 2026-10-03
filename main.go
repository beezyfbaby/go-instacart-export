package instacart

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"
)

func parseTime(value string) (time.Time, error) {
	for _, layout := range []string{time.RFC3339Nano, "Jan 2, 2006, 3:04 PM", time.DateOnly} {
		if t, err := time.Parse(layout, strings.Join(strings.Fields(value), " ")); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported timestamp format")
}

func extractOrders(resp OrdersResponse) ([]*Order, error) {
	orders := make([]*Order, 0, len(resp.Orders))
	for _, o := range resp.Orders {
		if o.ID == "" || o.Total == "" || o.Status == "" || o.OrderDeliveries == nil {
			return nil, fmt.Errorf("order missing required fields; consumer API may have changed")
		}
		created, err := parseTime(o.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("order created_at: %w", err)
		}
		order := &Order{ID: o.ID, Status: o.Status, Total: o.Total, CreatedAt: created}
		for _, d := range o.OrderDeliveries {
			delivery := &Delivery{Retailer: d.Retailer.Name}
			if d.DeliveredAt != "" {
				delivery.DeliveredAt, err = parseTime(d.DeliveredAt)
				if err != nil {
					return nil, fmt.Errorf("delivery delivered_at: %w", err)
				}
			}
			if d.Retailer.Name == "" || d.OrderItems == nil {
				return nil, fmt.Errorf("delivery missing retailer or items; consumer API may have changed")
			}
			for _, i := range d.OrderItems {
				delivery.Items = append(delivery.Items, &Item{ID: i.Item.ID, ProductID: i.Item.ProductID, Quantity: i.Qty, Name: i.Item.Name})
			}
			order.Deliveries = append(order.Deliveries, delivery)
		}
		orders = append(orders, order)
	}
	return orders, nil
}

// FetchOrders retrieves orders newest first. Errors never return a partial export.
// Breaking change: callers must now pass a context and handle the returned error.
func FetchOrders(ctx context.Context, client Client) ([]*Order, error) {
	orders := make([]*Order, 0)
	seen := make(map[string]bool)
	for page := 1; ; {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		resp, err := client.getPage(ctx, page)
		if err != nil {
			return nil, fmt.Errorf("page %d: %w", page, err)
		}
		batch, err := extractOrders(resp)
		if err != nil {
			return nil, fmt.Errorf("page %d: %w", page, err)
		}
		for _, order := range batch {
			if seen[order.ID] {
				return nil, fmt.Errorf("duplicate order across pages; retry when order history is stable")
			}
			seen[order.ID] = true
			orders = append(orders, order)
		}
		var next int
		if err := json.Unmarshal(resp.Meta.Pagination.NextPage, &next); err != nil {
			return nil, fmt.Errorf("invalid next_page: %w", err)
		}
		if next == 0 {
			break
		}
		if next <= page || len(batch) == 0 {
			return nil, fmt.Errorf("invalid pagination progression at page %d", page)
		}
		page = next
	}
	slices.SortStableFunc(orders, func(a, b *Order) int { return b.CreatedAt.Compare(a.CreatedAt) })
	return orders, nil
}
