# GeoProxy codebase summary

## Purpose
GeoProxy is a TCP-only proxy that accepts or blocks connections based on source IP geolocation from ip-api.com. It can optionally send or receive HAProxy PROXY protocol headers.

## Entry points
- `main.go`: CLI flags, config validation, starts server instances, and initializes the LRU cache.
- `geoproxy.yaml`: Default configuration file read by `config.ReadConfig`.

## Main flow
1. `main.go` loads YAML config and validates scheduling and proxy-protocol settings.
2. For each server block, `server.ServerConfig.StartServer` listens on the configured address and accepts client connections.
3. `handler.ClientHandler` checks allow/deny lists, schedule windows, and ip-api lookups, then either forwards traffic or rejects the connection.
4. `handler.TransferData` proxies data and, when enabled, writes a PROXY protocol header to the backend before copying the stream.

## Packages
- `config/`: YAML parsing and validation (IP validation for `trustedProxies`, schedule parsing checks).
- `server/`: Listener and dialer abstractions and server loop; wraps listener in `proxyproto.Listener` when receiving PROXY protocol.
- `handler/`: Per-connection logic and data transfer; supports always-allow/deny, country/region rules, and time/date/day gating.
- `ipapi/`: HTTP client and response cache; global LRU cache with size-based eviction.
- `common/`: Shared helpers (set helpers, CIDR checks, date/time range helpers).

## Proxy protocol notes
- `recvProxyProtocol` enables parsing PROXY headers on inbound connections.
- `trustedProxies` must be set when `recvProxyProtocol` is enabled. It must be a list of plain IPv4/IPv6 addresses (no CIDRs). GeoProxy rejects non-trusted upstreams and requires trusted upstreams to send a PROXY header.

## ip-api transport notes
- It's OK to use HTTP for free ip-api accounts because the free service only supports HTTP.
- When `apiKey` is set, GeoProxy forces `https://pro.ip-api.com/json/` and does not allow overriding the endpoint.

## Tests
- Run all tests with `go test ./...`.
- Tests should run after every task before any code can be comitted
- Code that fails unit tests should never be comitted
- Nothing should be comitted to the repo without running `go test ./...`.
- Test coverage should be 80%
