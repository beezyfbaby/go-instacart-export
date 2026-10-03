package instacart

import "time"

// Item represents a purchased item in the delivery.
type Item struct {
	ID        string
	ProductID string
	Quantity  float64
	Name      string
}

// Delivery is a purchase at a particular retailer.
type Delivery struct {
	Retailer    string
	DeliveredAt time.Time
	Items       []*Item
}

// Order is the complete transaction.
type Order struct {
	ID         string
	Status     string
	Total      string
	CreatedAt  time.Time
	Deliveries []*Delivery
}
