# Instacart CSV Export

Export personal Instacart order history to CSV using a Go command-line tool.

## Compatibility status

**Go modernization is tested with Go 1.27.1. Live Instacart compatibility is NOT verified.**

This project uses the undocumented consumer website endpoint
`https://www.instacart.com/v3/orders`, authenticated by a browser session cookie.
It does **not** use the official Instacart Developer Platform API. The public
Developer Platform documentation reviewed on October 3, 2026 does not provide a
replacement for exporting a consumer's full personal order history. Its API key
cannot replace this tool's session cookie. Retailer Connect integrations are a
different product and are not a drop-in replacement either.

Do not interpret successful unit tests as confirmation that the current website
still accepts this endpoint or returns this schema. A live smoke test with your
own account is required. If Instacart changed its endpoint, authentication, or
response format, the tool stops with an error rather than silently exporting an
empty report. It does not bypass login challenges or access restrictions.

References:
- [Go downloads](https://go.dev/dl/)
- [Instacart Developer Platform API](https://docs.instacart.com/developer_platform_api/api/overview/)
- [Instacart Developer Platform FAQ](https://docs.instacart.com/developer_platform_api/faq/)
- [Instacart Connect APIs](https://docs.instacart.com/connect/api/)

## Build and install

Install [Go 1.27.1 or newer](https://go.dev/dl/). From this checkout:

```sh
go build -o instacart-export ./cmd/instacart-export
```

On Windows PowerShell:

```powershell
go build -o instacart-export.exe ./cmd/instacart-export
```

After the changes have been merged into the default branch, install directly:

```sh
go install github.com/beezyfbaby/go-instacart-export/cmd/instacart-export@main
```

## Usage

Sign in to Instacart normally in your browser. In Developer Tools, inspect the
`_instacart_session_id` cookie for `www.instacart.com`. Use only its value, not a
whole Cookie header. Keep it private. Do not paste it into issues, chat, commits,
or screenshots. Session cookies grant account access.

For PowerShell 7, enter it without displaying it or saving it in command history:

```powershell
$env:INSTACART_SESSION_TOKEN = Read-Host 'Instacart session cookie' -MaskInput
try {
    .\instacart-export.exe -output .\data\orders.csv
} finally {
    Remove-Item Env:INSTACART_SESSION_TOKEN
}
```

For Bash:

```bash
read -r -s -p 'Instacart session cookie: ' INSTACART_SESSION_TOKEN
export INSTACART_SESSION_TOKEN
./instacart-export -output data/orders.csv
unset INSTACART_SESSION_TOKEN
```

Options:
- `-output`: destination file, default `data/instacart_orders_<timestamp>.csv`.
  Existing files are never overwritten. Use a new filename for each run.
- `-timeout`: maximum duration for the entire export, default `5m`.
- `-h`: help.

The HTTP client also has a 30-second per-request timeout. Ctrl+C cancels the run.
CSV output is written only after all pages have been fetched and validated.
Output files use owner-only permissions where the OS supports Unix permissions;
on Windows, protect the directory with your account's normal access controls.

## CSV fields

| Field | Meaning |
| --- | --- |
| id | Order ID |
| status | Order status |
| total | Original total string, without floating-point conversion |
| createdAt | Order date, YYYY-MM-DD |
| retailers | Retailer names separated by a pipe |
| numItems | Number of item lines across deliveries, not sum of quantities |

Orders are newest first. Formula-like cell values receive a leading apostrophe
for spreadsheet safety. Fractional item quantities are preserved in the Go model.
Timestamps support RFC 3339, date-only, and the original English website format.
Website timestamps without a time zone are treated as UTC; date-only CSV output
retains their calendar date.

## Troubleshooting and live verification

1. Run with a fresh session cookie and a new output filename.
2. Compare the exported order count, newest/oldest dates, and several totals
   against your browser's order history.
3. `401`/`403` or redirect: sign in normally again. If access is still blocked,
   stop; this client cannot resolve browser challenges.
4. `429`: wait before trying again; the error reports `Retry-After` when supplied.
5. Unexpected JSON, missing fields, or `404`: the private interface may have
   changed. Do not guess a version number or substitute a Developer API key.
   For an adapter update, provide the request URL/method and a **redacted response
   JSON** from an orders request in Developer Tools. Remove names, addresses,
   payment data, cookies, tokens, and other identifying data. Do not share a raw HAR.

Pagination currently requires an `orders` array and
`meta.pagination.next_page` with an integer page number, `0`, or `null`. Missing
pagination is treated as an error to avoid silently exporting only one page.
Repeated/backward pages, duplicate order IDs, and empty advancing pages fail
explicitly. No partial results are returned on failure.

## Development

```sh
go test -race ./...
go vet ./...
go build ./cmd/instacart-export
```

Tests use synthetic HTTP responses and cover pagination, schema changes,
authentication errors, cancellation, fractional quantities, CSV escaping,
formula safety, and overwrite protection. CI runs on Linux, Windows, and macOS.
GoReleaser configuration uses version 2; run `goreleaser check` before releasing.

### Library API changes

`FetchOrders` now takes a context and returns an error:

```go
orders, err := instacart.FetchOrders(ctx, instacart.Client{SessionToken: token})
```

Handle `err` before using `orders`. `Item.Quantity` is now `float64`, preserving
weighted-item quantities. The response model includes only fields used by the
exporter, avoiding failures caused by irrelevant website UI schema changes.

## License

[MIT © Rocky Gray](LICENSE)
