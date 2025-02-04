package instacart

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

// Client is the HTTP client for the Instacart orders API
type Client struct {
	SessionToken string
}

func (c *Client) getPage(page int) OrdersResponse {

	req, err := http.NewRequest("GET", "https://www.instacart.com/v3/orders?page="+strconv.Itoa(page), nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Authority", "www.instacart.com")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Client-Identifier", "web")
	req.Header.Set("User-Agent", "Instacart Orders To CSV Client")
	req.Header.Set("Dnt", "1")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", "https://www.instacart.com/store/account/orders")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	cookie := "_instacart_session_id=" + c.SessionToken + ";"
	req.Header.Set("Cookie", cookie)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	var ordersResp OrdersResponse

	if err := json.NewDecoder(resp.Body).Decode(&ordersResp); err != nil {
		log.Fatal(err)
	}

	return ordersResp
}

// OrdersResponse is the response from the orders API
// auto-generated from: https://mholt.github.io/json-to-go/
//   - Updated Actions to be map[string]struct
//   - Updated .orders.order_deliveries.order_items.qty to be float
//   - Updated .orders.order_deliveries.order_items.item.qty_attributes.increment  to be float
//   - Updated .orders.order_deliveries.order_items.item.qty_attributes.min  to be float
//   - Updated .orders.order_deliveries.order_items.item.qty_attributes.max  to be float
//   - Updated .orders.order_deliveries.order_items.item.qty_attributes.select.options to be float
//   - Updated .orders.rating to be float
type OrdersResponse struct {
	Orders []struct {
		ID        string `json:"id"`
		LegacyID  int    `json:"legacy_id"`
		Status    string `json:"status"`
		Rating    any    `json:"rating"`
		Total     string `json:"total"`
		CreatedAt string `json:"created_at"`
		Actions   struct {
			AddAllItemsToCart struct {
				Label           string `json:"label"`
				InProgressLabel string `json:"in_progress_label"`
				OrderUUID       string `json:"order_uuid"`
				SourceType      string `json:"source_type"`
			} `json:"add_all_items_to_cart"`
			Rating struct {
				URL   string `json:"url"`
				Label string `json:"label"`
			} `json:"rating"`
			ReportProblem struct {
				URL   string `json:"url"`
				Label string `json:"label"`
			} `json:"report_problem"`
		} `json:"actions"`
		OrderDeliveries []struct {
			ID          string `json:"id"`
			OrderID     string `json:"order_id"`
			Description string `json:"description"`
			Base62ID    string `json:"base62_id"`
			Status      string `json:"status"`
			DeliveredAt string `json:"delivered_at"`
			Retailer    struct {
				ID   string `json:"id"`
				Name string `json:"name"`
				Slug string `json:"slug"`
				Logo struct {
					URL        string `json:"url"`
					Alt        string `json:"alt"`
					Responsive struct {
						Template string `json:"template"`
						Defaults struct {
							Width int `json:"width"`
						} `json:"defaults"`
					} `json:"responsive"`
					Sizes []any `json:"sizes"`
				} `json:"logo"`
				BackgroundColor string `json:"background_color"`
			} `json:"retailer"`
			OrderItems []struct {
				Qty  float64 `json:"qty"`
				Item struct {
					ID                      string   `json:"id"`
					LegacyID                int      `json:"legacy_id"`
					ProductID               string   `json:"product_id"`
					Name                    string   `json:"name"`
					Attributes              []string `json:"attributes"`
					EbtAttributes           any      `json:"ebt_attributes"`
					ShowFullBleedImage      any      `json:"show_full_bleed_image"`
					PriceAffix              any      `json:"price_affix"`
					PriceAffixAria          any      `json:"price_affix_aria"`
					SecondaryPriceAffix     string   `json:"secondary_price_affix"`
					SecondaryPriceAffixAria string   `json:"secondary_price_affix_aria"`
					Size                    string   `json:"size"`
					SizeAria                string   `json:"size_aria"`
					ImageList               []struct {
						URL        string `json:"url"`
						Alt        string `json:"alt"`
						Responsive struct {
							Template string `json:"template"`
							Defaults struct {
								Width  int    `json:"width"`
								Fill   string `json:"fill"`
								Format string `json:"format"`
							} `json:"defaults"`
						} `json:"responsive"`
						Sizes []any `json:"sizes"`
					} `json:"image_list"`
					Image struct {
						URL        string `json:"url"`
						Alt        string `json:"alt"`
						Responsive struct {
							Template string `json:"template"`
							Defaults struct {
								Width  int    `json:"width"`
								Fill   string `json:"fill"`
								Format string `json:"format"`
							} `json:"defaults"`
						} `json:"responsive"`
						Sizes []any `json:"sizes"`
					} `json:"image"`
					WeightsAndMeasuresV2Enabled any `json:"weights_and_measures_v2_enabled"`
					VariableAttributesMap       any `json:"variable_attributes_map"`
					ProductPagePath             any `json:"product_page_path"`
					ClickAction                 struct {
						Type string `json:"type"`
						Data struct {
							Container struct {
								Title            string `json:"title"`
								Path             string `json:"path"`
								InitialStep      any    `json:"initial_step"`
								Modules          []any  `json:"modules"`
								DataDependencies []any  `json:"data_dependencies"`
							} `json:"container"`
							TrackingParams struct {
							} `json:"tracking_params"`
							TrackingEventNames struct {
							} `json:"tracking_event_names"`
						} `json:"data"`
					} `json:"click_action"`
					WineRatingBadge any    `json:"wine_rating_badge"`
					Weekly          any    `json:"weekly"`
					WeeklyOrderID   any    `json:"weekly_order_id"`
					V4ItemID        string `json:"v4_item_id"`
					QtyAttributes   struct {
						Initial                  int    `json:"initial"`
						Increment                int    `json:"increment"`
						Min                      int    `json:"min"`
						Max                      int    `json:"max"`
						Unit                     any    `json:"unit"`
						UnitAria                 any    `json:"unit_aria"`
						MaxReachedLabel          string `json:"max_reached_label"`
						MinReachedLabel          any    `json:"min_reached_label"`
						MinWeightExp             bool   `json:"min_weight_exp"`
						HideUnitStepperIcon      bool   `json:"hide_unit_stepper_icon"`
						QuantityType             any    `json:"quantity_type"`
						Editable                 bool   `json:"editable"`
						QtyEnforcedLabel         any    `json:"qty_enforced_label"`
						VariableWeightDisclaimer any    `json:"variable_weight_disclaimer"`
						Select                   struct {
							Options       []int `json:"options"`
							DefaultOption int   `json:"default_option"`
							CustomOption  struct {
								Label string `json:"label"`
							} `json:"custom_option"`
						} `json:"select"`
					} `json:"qty_attributes"`
					QtyAttributesPerUnit        any `json:"qty_attributes_per_unit"`
					DeliveryPromotionAttributes any `json:"delivery_promotion_attributes"`
				} `json:"item"`
			} `json:"order_items"`
		} `json:"order_deliveries"`
	} `json:"orders"`
	Meta struct {
		Pagination struct {
			Total    int `json:"total"`
			PerPage  int `json:"per_page"`
			Page     int `json:"page"`
			NextPage int `json:"next_page"`
		} `json:"pagination"`
	} `json:"meta"`
}
