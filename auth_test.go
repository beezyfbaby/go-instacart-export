package instacart

import (
	"net/http"
	"strings"
	"testing"
)

func TestSessionCookieCombinations(t *testing.T) {
	for _, tc := range []struct{ name, legacy, host string }{{"legacy", "legacy%3D--test", ""}, {"host", "", "host%3D--test"}, {"both", "legacy-test", "host-test"}} {
		t.Run(tc.name, func(t *testing.T) {
			c := Client{SessionToken: tc.legacy, HostSessionToken: tc.host, HTTPClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				if r.URL.Scheme != "https" || r.URL.Host != "www.instacart.com" {
					t.Fatal("wrong credential destination")
				}
				want := map[string]string{"_instacart_session_id": tc.legacy, "__Host-instacart_sid": tc.host}
				count := 0
				for name, value := range want {
					cookie, err := r.Cookie(name)
					if value == "" {
						if err == nil {
							t.Fatal("sent empty cookie")
						}
						continue
					}
					count++
					if err != nil || cookie.Value != value {
						t.Fatalf("cookie %s not preserved", name)
					}
				}
				if len(r.Cookies()) != count {
					t.Fatal("unexpected cookies")
				}
				return response(200, `{"data":{"orderDeliveriesConnection":{"nodes":[],"pageInfo":{"hasNextPage":false}}}}`), nil
			})}}
			if _, err := FetchDeliveryHistory(t.Context(), c, ""); err != nil {
				t.Fatal(err)
			}
		})
	}
}
func TestInvalidHostCookie(t *testing.T) {
	c := Client{HostSessionToken: "secret;invalid", HTTPClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) { t.Fatal("sent invalid cookie"); return nil, nil })}}
	_, err := FetchDeliveryHistory(t.Context(), c, "")
	if err == nil || strings.Contains(err.Error(), "secret") {
		t.Fatal("bad credential validation")
	}
}
func TestAuthenticationErrorIsActionable(t *testing.T) {
	for _, message := range []string{"Not Authenticated", "Not Authorized"} {
		_, err := decodeHistoryResponse([]byte(`{"errors":[{"message":"` + message + `"}]}`))
		if err == nil || !strings.Contains(err.Error(), "INSTACART_HOST_SESSION_TOKEN") {
			t.Fatalf("unhelpful error: %v", err)
		}
	}
}
func TestSessionCookiesNeverFollowRedirect(t *testing.T) {
	calls := 0
	c := Client{HostSessionToken: "host-secret", SessionToken: "legacy-secret", HTTPClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		calls++
		if calls > 1 {
			t.Fatal("followed redirect with credentials")
		}
		r := response(302, "")
		r.Header.Set("Location", "https://other.example/graphql")
		return r, nil
	})}}
	if _, err := FetchDeliveryHistory(t.Context(), c, ""); err == nil || calls != 1 {
		t.Fatal("accepted redirect")
	}
}
