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
		ID        string `json:"id,omitempty"`
		LegacyID  string `json:"legacy_id,omitempty"`
		Status    string `json:"status,omitempty"`
		Rating    any    `json:"rating,omitempty"`
		Total     string `json:"total,omitempty"`
		CreatedAt string `json:"created_at,omitempty"`
		Actions   struct {
			AddAllItemsToCart struct {
				Label           string `json:"label,omitempty"`
				InProgressLabel string `json:"in_progress_label,omitempty"`
				OrderUUID       string `json:"order_uuid,omitempty"`
				SourceType      string `json:"source_type,omitempty"`
			} `json:"add_all_items_to_cart,omitempty"`
			Rating struct {
				URL   string `json:"url,omitempty"`
				Label string `json:"label,omitempty"`
			} `json:"rating,omitempty"`
			ReportProblem struct {
				URL   string `json:"url,omitempty"`
				Label string `json:"label,omitempty"`
			} `json:"report_problem,omitempty"`
		} `json:"actions,omitempty"`
		OrderDeliveries []struct {
			ID          string `json:"id,omitempty"`
			OrderID     string `json:"order_id,omitempty"`
			Description string `json:"description,omitempty"`
			Base62ID    string `json:"base62_id,omitempty"`
			Status      string `json:"status,omitempty"`
			DeliveredAt string `json:"delivered_at,omitempty"`
			Retailer    struct {
				ID   string `json:"id,omitempty"`
				Name string `json:"name,omitempty"`
				Slug string `json:"slug,omitempty"`
				Logo struct {
					URL        string `json:"url,omitempty"`
					Alt        string `json:"alt,omitempty"`
					Responsive struct {
						Template string `json:"template,omitempty"`
						Defaults struct {
							Width int `json:"width,omitempty"`
						} `json:"defaults,omitempty"`
					} `json:"responsive,omitempty"`
					Sizes []any `json:"sizes,omitempty"`
				} `json:"logo,omitempty"`
				BackgroundColor string `json:"background_color,omitempty"`
			} `json:"retailer,omitempty"`
			OrderItems []struct {
				Qty  float64 `json:"qty,omitempty"`
				Item struct {
					ID                      string   `json:"id,omitempty"`
					LegacyID                int      `json:"legacy_id,omitempty"`
					ProductID               string   `json:"product_id,omitempty"`
					Name                    string   `json:"name,omitempty"`
					Attributes              []string `json:"attributes,omitempty"`
					EbtAttributes           any      `json:"ebt_attributes,omitempty"`
					ShowFullBleedImage      any      `json:"show_full_bleed_image,omitempty"`
					PriceAffix              any      `json:"price_affix,omitempty"`
					PriceAffixAria          any      `json:"price_affix_aria,omitempty"`
					SecondaryPriceAffix     string   `json:"secondary_price_affix,omitempty"`
					SecondaryPriceAffixAria string   `json:"secondary_price_affix_aria,omitempty"`
					Size                    string   `json:"size,omitempty"`
					SizeAria                string   `json:"size_aria,omitempty"`
					ImageList               []struct {
						URL        string `json:"url,omitempty"`
						Alt        string `json:"alt,omitempty"`
						Responsive struct {
							Template string `json:"template,omitempty"`
							Defaults struct {
								Width  int    `json:"width,omitempty"`
								Fill   string `json:"fill,omitempty"`
								Format string `json:"format,omitempty"`
							} `json:"defaults,omitempty"`
						} `json:"responsive,omitempty"`
						Sizes []any `json:"sizes,omitempty"`
					} `json:"image_list,omitempty"`
					Image struct {
						URL        string `json:"url,omitempty"`
						Alt        string `json:"alt,omitempty"`
						Responsive struct {
							Template string `json:"template,omitempty"`
							Defaults struct {
								Width  int    `json:"width,omitempty"`
								Fill   string `json:"fill,omitempty"`
								Format string `json:"format,omitempty"`
							} `json:"defaults,omitempty"`
						} `json:"responsive,omitempty"`
						Sizes []any `json:"sizes,omitempty"`
					} `json:"image,omitempty"`
					WeightsAndMeasuresV2Enabled any `json:"weights_and_measures_v2_enabled,omitempty"`
					VariableAttributesMap       any `json:"variable_attributes_map,omitempty"`
					ProductPagePath             any `json:"product_page_path,omitempty"`
					ClickAction                 struct {
						Type string `json:"type,omitempty"`
						Data struct {
							Container struct {
								Title            string `json:"title,omitempty"`
								Path             string `json:"path,omitempty"`
								InitialStep      any    `json:"initial_step,omitempty"`
								Modules          []any  `json:"modules,omitempty"`
								DataDependencies []any  `json:"data_dependencies,omitempty"`
							} `json:"container,omitempty"`
							TrackingParams struct {
							} `json:"tracking_params,omitempty"`
							TrackingEventNames struct {
							} `json:"tracking_event_names,omitempty"`
						} `json:"data,omitempty"`
					} `json:"click_action,omitempty"`
					WineRatingBadge any    `json:"wine_rating_badge,omitempty"`
					Weekly          any    `json:"weekly,omitempty"`
					WeeklyOrderID   any    `json:"weekly_order_id,omitempty"`
					V4ItemID        string `json:"v4_item_id,omitempty"`
					QtyAttributes   struct {
						Initial                  int    `json:"initial,omitempty"`
						Increment                int    `json:"increment,omitempty"`
						Min                      int    `json:"min,omitempty"`
						Max                      int    `json:"max,omitempty"`
						Unit                     any    `json:"unit,omitempty"`
						UnitAria                 any    `json:"unit_aria,omitempty"`
						MaxReachedLabel          string `json:"max_reached_label,omitempty"`
						MinReachedLabel          any    `json:"min_reached_label,omitempty"`
						MinWeightExp             bool   `json:"min_weight_exp,omitempty"`
						HideUnitStepperIcon      bool   `json:"hide_unit_stepper_icon,omitempty"`
						QuantityType             any    `json:"quantity_type,omitempty"`
						Editable                 bool   `json:"editable,omitempty"`
						QtyEnforcedLabel         any    `json:"qty_enforced_label,omitempty"`
						VariableWeightDisclaimer any    `json:"variable_weight_disclaimer,omitempty"`
						Select                   struct {
							Options       []int `json:"options,omitempty"`
							DefaultOption int   `json:"default_option,omitempty"`
							CustomOption  struct {
								Label string `json:"label,omitempty"`
							} `json:"custom_option,omitempty"`
						} `json:"select,omitempty"`
					} `json:"qty_attributes,omitempty"`
					QtyAttributesPerUnit        any `json:"qty_attributes_per_unit,omitempty"`
					DeliveryPromotionAttributes any `json:"delivery_promotion_attributes,omitempty"`
				} `json:"item,omitempty"`
			} `json:"order_items,omitempty"`
		} `json:"order_deliveries,omitempty"`
		Actions0 struct {
			AddAllItemsToCart struct {
				Label           string `json:"label,omitempty"`
				InProgressLabel string `json:"in_progress_label,omitempty"`
				OrderUUID       string `json:"order_uuid,omitempty"`
				SourceType      string `json:"source_type,omitempty"`
			} `json:"add_all_items_to_cart,omitempty"`
			Rating struct {
				URL   string `json:"url,omitempty"`
				Label string `json:"label,omitempty"`
			} `json:"rating,omitempty"`
		} `json:"actions,omitempty"`
		Actions1 struct {
			AddAllItemsToCart struct {
				Label           string `json:"label,omitempty"`
				InProgressLabel string `json:"in_progress_label,omitempty"`
				OrderUUID       string `json:"order_uuid,omitempty"`
				SourceType      string `json:"source_type,omitempty"`
			} `json:"add_all_items_to_cart,omitempty"`
			Rating struct {
				URL   string `json:"url,omitempty"`
				Label string `json:"label,omitempty"`
			} `json:"rating,omitempty"`
		} `json:"actions,omitempty"`
		Actions2 struct {
			AddAllItemsToCart struct {
				Label           string `json:"label,omitempty"`
				InProgressLabel string `json:"in_progress_label,omitempty"`
				OrderUUID       string `json:"order_uuid,omitempty"`
				SourceType      string `json:"source_type,omitempty"`
			} `json:"add_all_items_to_cart,omitempty"`
			Rating struct {
				URL   string `json:"url,omitempty"`
				Label string `json:"label,omitempty"`
			} `json:"rating,omitempty"`
		} `json:"actions,omitempty"`
	} `json:"orders,omitempty"`
	Meta struct {
		Pagination struct {
			Total    int `json:"total,omitempty"`
			PerPage  int `json:"per_page,omitempty"`
			Page     int `json:"page,omitempty"`
			NextPage int `json:"next_page,omitempty"`
		} `json:"pagination,omitempty"`
	} `json:"meta,omitempty"`
}
