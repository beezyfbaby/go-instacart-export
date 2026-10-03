package instacart

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// Client reads the undocumented consumer website API, not the Developer Platform API.
// SessionToken is the value of the _instacart_session_id cookie. Never log it.
type Client struct {
	SessionToken string
	// HTTPClient optionally supplies a transport. Redirects are always rejected.
	HTTPClient *http.Client
}

func (c *Client) getPage(ctx context.Context, page int) (OrdersResponse, error) {
	var result OrdersResponse
	cookie := &http.Cookie{Name: "_instacart_session_id", Value: c.SessionToken}
	if c.SessionToken == "" {
		return result, fmt.Errorf("INSTACART_SESSION_TOKEN is required")
	}
	if err := cookie.Valid(); err != nil {
		return result, fmt.Errorf("invalid session cookie value")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.instacart.com/v3/orders?page="+strconv.Itoa(page), nil)
	if err != nil {
		return result, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Client-Identifier", "web")
	req.Header.Set("User-Agent", "instacart-export")
	req.Header.Set("Referer", "https://www.instacart.com/store/account/orders")
	req.AddCookie(cookie)
	client := http.Client{Timeout: 30 * time.Second}
	if c.HTTPClient != nil {
		client = *c.HTTPClient
	}
	if client.Timeout <= 0 {
		client.Timeout = 30 * time.Second
	}
	client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := client.Do(req)
	if err != nil {
		return result, fmt.Errorf("request orders page: %w", err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusUnauthorized, http.StatusForbidden:
		return result, fmt.Errorf("HTTP %d: session expired or access blocked; sign in normally and refresh your session cookie", resp.StatusCode)
	case http.StatusTooManyRequests:
		return result, fmt.Errorf("HTTP 429: rate limited; wait before running again (Retry-After: %q)", resp.Header.Get("Retry-After"))
	default:
		return result, fmt.Errorf("HTTP %d: consumer orders endpoint unavailable or changed", resp.StatusCode)
	}
	const limit = 16 << 20
	body, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return result, fmt.Errorf("read orders: %w", err)
	}
	if len(body) > limit {
		return result, fmt.Errorf("orders response exceeds 16 MiB")
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return result, fmt.Errorf("unexpected orders JSON; consumer API may have changed: %w", err)
	}
	if result.Orders == nil || result.Meta.Pagination == nil || result.Meta.Pagination.NextPage == nil {
		return result, fmt.Errorf("missing orders or pagination fields; consumer API may have changed")
	}
	return result, nil
}

// OrdersResponse contains only fields needed for export. Unknown fields are ignored.
type OrdersResponse struct {
	Orders []apiOrder `json:"orders"`
	Meta   struct {
		Pagination *pagination `json:"pagination"`
	} `json:"meta"`
}

type pagination struct {
	// RawMessage distinguishes a missing next_page from an explicit null terminator.
	NextPage json.RawMessage `json:"next_page"`
}

type apiOrder struct {
	ID              string `json:"id"`
	Status          string `json:"status"`
	Total           string `json:"total"`
	CreatedAt       string `json:"created_at"`
	OrderDeliveries []struct {
		DeliveredAt string `json:"delivered_at"`
		Retailer    struct {
			Name string `json:"name"`
		} `json:"retailer"`
		OrderItems []struct {
			Qty  float64 `json:"qty"`
			Item struct {
				ID        string `json:"id"`
				ProductID string `json:"product_id"`
				Name      string `json:"name"`
			} `json:"item"`
		} `json:"order_items"`
	} `json:"order_deliveries"`
}
