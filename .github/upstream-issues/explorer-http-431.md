# Explorer links with long queries fail with HTTP 431 before routing

## What happened

Opening an Explorer URL containing a sufficiently long SQL or LogchefQL query
returns `431 Request Header Fields Too Large`. A short URL works again after
removing the query parameters. The same failure can affect an otherwise short
API request when the browser supplies the long Explorer URL as its `Referer`.

This was reproduced directly against Logchef, bypassing the reverse proxy.
An 8 KiB query string and an 8 KiB Referer each caused 431 in v2.6.1. With the
proposed patch, the same direct HTTP requests return 200.

## Expected behavior

Explorer links containing realistic queries should load, and subsequent API
calls should accept their Referer, while retaining a finite request-header limit.

## Steps to reproduce

1. Start Logchef v2.6.1 and open Explorer.
2. Use a query that produces a URL longer than 4 KiB, for example an SQL string
   literal containing 8,192 ASCII characters.
3. Open that Explorer URL directly: the server returns 431 before route handling.
4. Alternatively, request `/api/v1/health` with a `Referer` header containing an
   Explorer URL plus 8,192 characters: this also returns 431.
5. Remove the long URL parameters or Referer: the request succeeds.

## Environment

Logchef v2.6.1, upstream commit `a0b1cf58a39b2fdede2cf427bd7c782bb90d511a`.
Go/Fiber backend. Reproduction does not require a particular log backend,
browser, authentication provider, or reverse proxy.

## Cause and proposed solution

`internal/server/server.go` leaves Fiber's `ReadBufferSize` at its 4 KiB default.
The HTTP parser applies this limit before Logchef routing. It affects the
request line and headers, including cookies and forwarded headers.

Set an explicit, bounded `ReadBufferSize` of 64 KiB. This preserves existing URL
state and shared links. It does not increase the request-body limit or relax
authentication. Reverse proxies keep their own independent URL/header limits.
Queries exceeding the new bound should still be rejected; this is a bounded
compatibility fix rather than an unlimited-header setting.

## Regression coverage

The proposed branch `fix/explorer-request-headers` tests the raw Fiber HTTP
parser with ordinary Explorer links, long SQL, percent-encoded Unicode,
long Referer, combined cookie/header overhead, unchanged API authentication,
and rejection of request lines/headers above the finite limit. Handler-only
tests would miss this pre-routing failure. `just check` passes natively.
