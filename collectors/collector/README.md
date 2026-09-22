# UTMStack Forwarder

The UTMStack Forwarder is a standalone log collection service that receives logs from external integrations and forwards them to the UTMStack backend in real time.

It runs as an independent Linux service (`UTMStackForwarder`) alongside the UTMStack Agent. Both connect directly to the backend — there is no communication between them.

---

## What it does

The Forwarder opens listeners on your server and waits for logs to arrive. As soon as a log comes in, it sends it to the UTMStack backend immediately — no local storage, no buffering.

It supports four types of listeners:

| Listener | Protocol | Typical use |
|---|---|---|
| **Syslog** | TCP, UDP, TLS | Firewalls, switches, routers, security appliances |
| **Netflow** | UDP | Network traffic data (v5, v9, IPFIX) |
| **File** | — | Log files on the local server (nginx, postgresql) |
| **HTTP/HTTPS** | HTTP, HTTPS | Webhooks, APIs, custom applications |

---

## How it works

### Boot sequence

When the service starts:

1. Reads the encrypted configuration (`collector-config.yml`)
2. Connects to the UTMStack backend
3. Starts a ping goroutine (heartbeat every 15 seconds)
4. Starts the log forwarding goroutine
5. Syncs the integration config file (`log-collector-config.json`)
6. Starts all enabled listeners

### Log flow

```
Syslog listener  ──────┐
Netflow listener ──────┤──► Internal queue (10,000 slots) ──► UTMStack backend
File listener    ──────┤
HTTP listener    ──────┘
```

Every listener pushes logs into a shared internal queue. A single goroutine drains that queue and streams logs to the backend over gRPC. If the backend is unreachable, logs arriving during the outage are dropped — there is no local buffer by design.

### Live configuration

The Forwarder watches `log-collector-config.json` for changes using `fsnotify`. When you run a CLI command to enable or disable an integration, the file changes and the relevant listener starts or stops automatically — **no service restart needed**.

---

## Installation

```bash
# Install the service
./utmstack_forwarder install <server_address> <utm_key> <yes|no>

# yes = skip TLS certificate validation
# no  = validate TLS certificate (recommended for production)
```

Example:
```bash
./utmstack_forwarder install 192.168.1.100 my-utm-key no
```

This will:
1. Verify connectivity to the backend (ports 9000 and 50051)
2. Register the forwarder with the UTMStack backend
3. Save the encrypted configuration
4. Install and start the `UTMStackForwarder` system service

```bash
# Uninstall
./utmstack_forwarder uninstall
```

---

## Managing integrations

### Enable a syslog integration

```bash
# UDP (most common for syslog)
./utmstack_forwarder enable-integration firewall-fortigate-traffic 7005 udp

# TCP
./utmstack_forwarder enable-integration firewall-fortigate-traffic 7005 tcp

# TCP with TLS (requires a certificate loaded first)
./utmstack_forwarder enable-integration firewall-fortigate-traffic 7005 tls
```

### Enable a netflow integration

```bash
./utmstack_forwarder enable-integration netflow 2055 udp
```

### Enable an HTTP integration

```bash
# No authentication — for trusted internal networks
./utmstack_forwarder enable-integration my-app 7999 http

# Bearer token authentication
./utmstack_forwarder enable-integration my-app 7999 http --auth bearer

# HMAC authentication — compatible with Meta, GitHub, Stripe, etc.
./utmstack_forwarder enable-integration meta   7999 http --auth hmac --signature-header X-Hub-Signature-256
./utmstack_forwarder enable-integration github 7999 http --auth hmac --signature-header X-Hub-Signature-256
./utmstack_forwarder enable-integration stripe 7999 http --auth hmac --signature-header Stripe-Signature
```

### Enable an HTTPS integration

```bash
# Load the TLS certificate first
./utmstack_forwarder load-tls-certs /path/to/cert.crt /path/to/key.key

# Then enable
./utmstack_forwarder enable-integration my-app 443 https --auth bearer
```

### Disable an integration

```bash
./utmstack_forwarder disable-integration firewall-fortigate-traffic udp
./utmstack_forwarder disable-integration my-app http
```

### Change a port

```bash
./utmstack_forwarder change-port firewall-fortigate-traffic udp 7010
./utmstack_forwarder change-port my-app http 8080
```

### Adding a custom integration type

If the integration name you want does not exist in the built-in catalog, the Forwarder creates it automatically:

```bash
./utmstack_forwarder enable-integration my-custom-firewall 7999 udp
# Output: New data type "my-custom-firewall" created.
#         Integration "my-custom-firewall" enabled on udp port 7999
```

The new type is saved in `log-collector-catalog.json`. You can manage custom types:

```bash
# List all types (built-in and custom)
./utmstack_forwarder list-datatypes

# Remove a custom type (must be disabled first)
./utmstack_forwarder remove-datatype my-custom-firewall
```

---

## HTTP/HTTPS authentication

### No authentication

The listener accepts every POST without any check. Use only on isolated networks.

```bash
./utmstack_forwarder enable-integration my-app 7999 http
```

Sending logs:
```bash
curl -X POST http://your-server:7999/logs \
  -d '{"event": "login", "user": "john"}'
```

---

### Bearer token

When enabled, the Forwarder generates a random token and displays it **once**. Store it securely.

```bash
./utmstack_forwarder enable-integration my-app 7999 http --auth bearer
# Output:
# Token for my-app: a3f8c2d1e4b57f9...
# Store this token securely — it won't be shown again.
```

Sending logs:
```bash
curl -X POST http://your-server:7999/logs \
  -H "Authorization: Bearer a3f8c2d1e4b57f9..." \
  -d '{"event": "login", "user": "john"}'
```

Configure this token as an HTTP header secret in your external platform.

---

### HMAC signature

Used by platforms like Meta, GitHub, and Stripe. The platform signs each request with a shared secret — the Forwarder verifies the signature without the secret ever travelling in the request.

```bash
./utmstack_forwarder enable-integration meta 7999 http \
  --auth hmac \
  --signature-header X-Hub-Signature-256
```

When enabled, a token file is generated at `certs/integration-http-meta.token`. Configure this token as the webhook secret in the external platform (Meta, GitHub, etc.).

The `--signature-header` flag specifies which header the platform uses to send the signature:

| Platform | Header |
|---|---|
| Meta / Facebook | `X-Hub-Signature-256` |
| GitHub | `X-Hub-Signature-256` |
| Stripe | `Stripe-Signature` |
| Custom | any header you choose |

---

## HTTP/HTTPS body format

The HTTP and HTTPS listeners accept **any format** — JSON, NDJSON, plain text, XML, or anything else. The body is forwarded as-is to the UTMStack event processor, which applies the corresponding filter for that DataType. No parsing or transformation happens in the Forwarder.

```bash
# JSON
curl -X POST http://your-server:7999/logs \
  -H "Content-Type: application/json" \
  -d '{"event":"login","user":"john"}'

# Plain syslog text
curl -X POST http://your-server:7999/logs \
  -H "Content-Type: text/plain" \
  -d 'Jan 10 12:00:01 fw01 kernel: packet blocked src=1.2.3.4'

# Any other format — the event processor handles it
curl -X POST http://your-server:7999/logs \
  --data-binary @/path/to/logfile.xml
```

The only limit is **8 MB per request**.

---

## TLS for syslog

```bash
# Load your certificates
./utmstack_forwarder load-tls-certs /path/to/cert.crt /path/to/key.key /path/to/ca.crt

# Enable TLS on a syslog integration
./utmstack_forwarder enable-integration firewall-cisco-asa 1470 tls
```

All TLS connections use TLS 1.2/1.3 only with AEAD cipher suites.

---

## Configuration files

All files are stored in the Forwarder's installation directory.

| File | Description |
|---|---|
| `collector-config.yml` | Encrypted credentials (server address, ID, key) |
| `collector-uuid.yml` | Unique installation identifier |
| `log-collector-config.json` | Integration settings (ports, protocols, enabled status) |
| `log-collector-catalog.json` | User-defined custom DataTypes |
| `certs/integration.crt` | TLS certificate for syslog TLS and HTTPS |
| `certs/integration.key` | TLS private key |
| `certs/integration-http-<name>.token` | Auth token for HTTP/HTTPS integrations |
| `logs/utmstack_collector.log` | Service log file |

---

## Built-in integration types

| Name | Protocol | Default port |
|---|---|---|
| `syslog` | UDP/TCP | 7014 |
| `vmware-esxi` | UDP/TCP | 7002 |
| `antivirus-esmc-eset` | UDP/TCP | 7003 |
| `antivirus-kaspersky` | UDP/TCP | 7004 |
| `firewall-cisco-asa` | UDP/TCP | 514/1470 |
| `firewall-cisco-firepower` | UDP/TCP | 514/1470 |
| `cisco-switch` | UDP/TCP | 514/1470 |
| `firewall-meraki` | UDP/TCP | 514/1470 |
| `firewall-fortigate-traffic` | UDP/TCP | 7005 |
| `firewall-paloalto` | UDP/TCP | 7006 |
| `firewall-mikrotik` | UDP/TCP | 7007 |
| `firewall-sophos-xg` | UDP/TCP | 7008 |
| `firewall-sonicwall` | UDP/TCP | 7009 |
| `deceptive-bytes` | UDP/TCP | 7010 |
| `antivirus-sentinel-one` | UDP/TCP | 7012 |
| `ibm-aix` | UDP/TCP | 7016 |
| `firewall-pfsense` | UDP/TCP | 7017 |
| `firewall-fortiweb` | UDP/TCP | 7018 |
| `suricata` | UDP/TCP | 7019 |
| `netflow` | UDP | 2055 |
| `nginx` | file | `/var/log/nginx/` |
| `postgresql` | file | `/var/log/postgresql/` |

---

## Service management

```bash
systemctl start   UTMStackForwarder
systemctl stop    UTMStackForwarder
systemctl restart UTMStackForwarder
systemctl status  UTMStackForwarder
```

---

## Platform support

Linux only (amd64, arm64).
