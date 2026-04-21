# 🌐 go-host

A lightweight DNS server written in Go — with domain blocking, allowlisting, per-client rate limiting, and query logging.

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go&logoColor=white)
![License](https://img.shields.io/badge/license-MIT-green?style=flat)
![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey?style=flat)

---

## What is go-host?

go-host is a DNS server that sits between your clients and the internet. It forwards DNS queries to an upstream resolver (Cloudflare's `1.1.1.1` by default), while giving you full control over which domains are allowed, blocked, or rate-limited.

Think of it as a self-hosted, programmable DNS filter — great for homelabs, local networks, or anyone who wants control over DNS resolution without running a full Pi-hole stack.

---

## Features

- **DNS Forwarding** — Proxies queries to an upstream DNS server (defaults to `1.1.1.1:53`)
- **Domain Blocking** — Blocks domains listed in `blocked_domains` with an `NXDOMAIN` response
- **Domain Allowlisting** — Explicitly permit domains via `allow_domains`, bypassing the blocklist
- **Per-Client Rate Limiting** — Caps each client IP at 10 requests per minute to prevent abuse
- **Query Logging** — Logs every query with its domain, client IP, and resolution status (`Success`, `Blocked`, `Rate Limited`)
- **Zero Dependencies** — Pure Go standard library, no heavy frameworks

---

## Getting Started

### Prerequisites

- Go 1.21 or later
- Port 53 requires elevated privileges on Linux/macOS

### Installation

```bash
git clone https://github.com/pkdiv/go-host.git
cd go-host
```

### Running

```bash
# Linux/macOS (port 53 requires root)
sudo go run main.go

# Or build first
go build -o go-host .
sudo ./go-host
```

You should see:

```
DNS Server running on port 53
```

### Testing

Use `dig` to send a query to your local server:

```bash
dig @127.0.0.1 google.com
dig @127.0.0.1 pkdiv.com
```

To test blocking, add a domain to `blocked_domains` and query it:

```bash
echo "ads.example.com" >> blocked_domains
dig @127.0.0.1 ads.example.com
# Expected: NXDOMAIN
```

---

## Configuration

go-host uses two plain-text files for domain control. One domain per line.

### `blocked_domains`

Domains listed here will be blocked and return `NXDOMAIN` to the client.

```
ads.example.com
tracker.analytics.io
malware-site.net
```

### `allow_domains`

Domains listed here are always allowed through, even if they appear on the blocklist. Useful for whitelisting specific subdomains.

```
safe.example.com
internal.corp
```

### Upstream DNS

The upstream resolver is set in `main.go`:

```go
UpstreamDNS := "1.1.1.1:53"
```

Swap this out for any DNS server you prefer, such as `8.8.8.8:53` (Google) or `9.9.9.9:53` (Quad9).

---

## How It Works

```
Client
  │
  ▼
go-host (UDP :53)
  │
  ├─ Is domain in allow_domains?  ──► Forward to upstream DNS
  │
  ├─ Is domain in blocked_domains? ──► Return NXDOMAIN
  │
  ├─ Is client rate-limited?  ──► Return SERVFAIL
  │
  └─ Forward to upstream DNS (1.1.1.1)
        │
        ▼
     Response returned to client
```

Every query is logged with the domain, client IP, and outcome.

---

## Using as Your System DNS

To route all traffic on your machine through go-host:

**Linux** — edit `/etc/resolv.conf`:
```
nameserver 127.0.0.1
```

**macOS** — System Settings → Network → DNS → set to `127.0.0.1`

**Windows** — Network Adapter Settings → IPv4 Properties → Preferred DNS: `127.0.0.1`

---

## Roadmap

- [ ] Configurable upstream DNS via config file or CLI flag
- [ ] DNS-over-TLS (DoT) support
- [ ] Web UI for query log viewing and blocklist management
- [ ] Support for blocklist subscriptions (e.g. hosts-format files)
- [ ] Metrics endpoint (Prometheus-compatible)
- [ ] Docker image

---

## Contributing

Contributions are welcome! Feel free to open an issue or submit a pull request.

1. Fork the repo
2. Create your branch: `git checkout -b feature/my-feature`
3. Commit your changes: `git commit -m 'Add my feature'`
4. Push and open a PR
