# Instacart CSV Export

Export personal Instacart delivery history to CSV using Go.

## Compatibility status

Targets Go 1.27.1. The CLI uses the consumer website's `PersonalOrderHistory`
GraphQL GET request observed on October 3, 2026, with the response parser checked
against a user-provided page of 10 deliveries. **An authenticated end-to-end run
has not yet been verified.** Request construction and pagination are tested with
synthetic responses. Session-cookie authentication may require further adjustment
based on a local run; the client does not bypass browser challenges.

This is an undocumented consumer interface, not the official Instacart Developer
Platform API. Developer Platform API keys cannot replace the browser session
cookie. The website's query hash or schema can change independently of this tool.

## Install the draft version

Install [Go 1.27.1 or newer](https://go.dev/dl/). In PowerShell:

```powershell
go install github.com/beezyfbaby/go-instacart-export/cmd/instacart-export@modernize-go-exporter
if ($LASTEXITCODE -ne 0) { throw 'Installation failed' }
```

From a local checkout, build with:

```powershell
go build -o instacart-export.exe ./cmd/instacart-export
```

## Run in PowerShell 7

Sign in normally at Instacart. Developer Tools > Application > Cookies >
https://www.instacart.com contains `_instacart_session_id`. Use its value only.
The prompt below hides input and keeps the value out of command history.

```powershell
$goBin = go env GOBIN
if (-not $goBin) { $goBin = Join-Path (go env GOPATH) 'bin' }
$exporter = Join-Path $goBin 'instacart-export.exe'
$destination = Join-Path $PWD ('instacart-deliveries-' + (Get-Date -Format 'yyyyMMdd-HHmmss') + '.csv')
$env:INSTACART_SESSION_TOKEN = Read-Host 'Instacart session cookie' -MaskInput
try {
    & $exporter -output $destination
    if ($LASTEXITCODE -ne 0) { throw 'Export failed. Review the message above.' }
} finally {
    Remove-Item Env:INSTACART_SESSION_TOKEN -ErrorAction SilentlyContinue
}
```

For a locally built binary, use `.\instacart-export.exe` as `$exporter`.
Keep cookies and receipt links containing access tokens private. Do not share
raw HAR files, request headers, or Copy-as-cURL output.

Bash usage after installing the command:

```bash
read -r -s -p 'Instacart session cookie: ' INSTACART_SESSION_TOKEN
export INSTACART_SESSION_TOKEN
instacart-export -output data/deliveries.csv
unset INSTACART_SESSION_TOKEN
```

Options:
- `-output`: destination CSV; existing files are never overwritten.
- `-timeout`: maximum time for the complete export (default `5m`).
- `-query-hash`: override the observed PersonalOrderHistory SHA-256 query hash.
- `-h`: help.

Requests time out after 30 seconds each. Ctrl+C cancels the export. Files are
written only after every page has been retrieved and validated. Owner-only Unix
permissions are requested; Windows directory access follows its existing ACLs.

## CSV schema

**One row per delivery, not per parent order.** This intentionally changes the
old order-level CSV to avoid presenting a delivery total as a whole-order total.

| Field | Meaning |
| --- | --- |
| deliveryId | Unique delivery ID |
| orderId | Parent legacy order ID; may occur on several rows |
| status | Workflow state returned by Instacart |
| deliveryTotal | Original displayed delivery total, including currency formatting |
| createdAt | RFC 3339 timestamp, preserving the response's time zone |
| retailer | Retailer name |
| numItemLines | Number of returned item lines; not purchased unit quantity |
| isMulti | Multi-delivery indicator from the response |
| serviceType | Delivery/pickup classification from the response |

Newest deliveries appear first. Formula-like text is prefixed with an apostrophe
for spreadsheet safety. Receipt URLs/access tokens are discarded. The observed
history response does not provide purchased quantities, so none are invented.
Multi-delivery totals are not combined; verify their accounting meaning against
receipts before aggregating expenses.

## How requests work

`GET https://www.instacart.com/graphql` with:
- `operationName=PersonalOrderHistory`
- `variables={"first":10}` initially, omitting the optional cursor
- `variables={"first":10,"after":"<previous endCursor>"}` subsequently
- `extensions={"persistedQuery":{"version":1,"sha256Hash":"<hash>"}}`

The initial request's omitted cursor is a conventional first-page assumption
that still requires live verification. Subsequent request parameters match the
captured browser request. The observed query identifier is stored in
`PersonalOrderHistoryHash`. A query hash identifies a registered query; it is not
a credential. No personal order cursor is hard-coded.

The response must contain exactly one `orderDeliveriesConnection` under `data`.
Its `nodes` and `pageInfo` are validated. Pagination ends only when
`hasNextPage=false`. Repeated cursors, duplicate delivery IDs, empty advancing
pages, missing fields, and GraphQL errors fail without returning a partial export.
Any GraphQL errors cause failure even if partial `data` accompanies them.

## Local smoke test and troubleshooting

1. Run with a fresh session cookie and a new destination filename.
2. Compare row count, oldest/newest dates, and totals against browser history.
   Count deliveries separately from parent orders. Check more than the first 10.
3. If you get `401`/`403` or a redirect, sign in normally again. If still blocked,
   report the error without credentials. Additional authentication requirements
   have not been verified; this client does not bypass access restrictions.
4. For `429`, wait before retrying. The error reports `Retry-After` if present.
5. For `PersistedQueryNotFound`, obtain the current hash from a successful
   PersonalOrderHistory request in Developer Tools and pass `-query-hash`.
6. For other GraphQL/schema errors, supply the sanitized response and request
   parameters. A changed first-page contract or missing fields needs a parser or
   request update, not a guessed API version.

## Development and library API

```sh
go test -race -cover ./...
go vet ./...
go build ./cmd/instacart-export
```

CI runs Go checks on Linux, Windows, and macOS. GoReleaser uses configuration v2
and retains checksum signing; a release needs an appropriately configured signer.

Use `FetchDeliveryHistory(ctx, client, hash)` for the current consumer integration.
An empty hash selects the default. `DecodeDeliveryPage` accepts the connection
object itself and is also available for offline parsing.

Legacy `FetchOrders(ctx, client)` and `Order` models remain for library callers,
but use the old unverified `/v3/orders` endpoint; the CLI no longer calls them.
Legacy `FetchOrders` now returns an error and legacy `Item.Quantity` is float64.

## References

- [Go downloads](https://go.dev/dl/)
- [Instacart Developer Platform](https://docs.instacart.com/developer_platform_api/api/overview/)
- [Instacart Connect](https://docs.instacart.com/connect/api/)

The consumer request and schema come from user-provided browser captures, not
these public API documents. Public tests contain synthetic data only.

## License

[MIT © Rocky Gray](LICENSE)
