# Security

## Principle of Least Privilege

The RouteGlass service runs as the dedicated `routeglass` system user. The systemd unit grants only `CAP_NET_RAW` for ICMP diagnostics and uses additional service hardening options.

## Router Backends

Only one BGP backend is enabled per instance.

### MikroTik

Use API-SSL, a dedicated read-only account, source-address restrictions, and CA validation.

### Juniper

Use a dedicated read-only account, SSH key authentication, source restrictions, and strict host-key checking with a dedicated pinned `known_hosts` file.

The default SSH port is `22` because RouteGlass is generic Open Source software. A non-standard port is optional and is not assumed by the installer.

## Public Management Exposure

A public router endpoint should not be broadly exposed simply because RouteGlass can connect to it. Restrict access by source IP, ACL, firewall filter, or private transport. The RouteGlass VM should be the only permitted source whenever feasible.

## Public Error Sanitization

RouteGlass does not expose raw BGP backend connection errors to public visitors. Router management addresses, SSH/API ports, usernames, and transport-level failure details may appear in backend errors and are therefore retained only in the RouteGlass service log. Public BGP failures return the generic message `BGP query is temporarily unavailable...` with HTTP status `503 Service Unavailable`.

## Secrets

Do not commit private SSH keys, MikroTik passwords, exported certificates containing private material, or `/etc/routeglass/` contents to Git.

## Updates

Keep Ubuntu, Go, Nginx, Certbot, RouterOS, and Junos patched according to your operational maintenance policy.

## Nginx Edge Hardening

RouteGlass is designed to expose its public HTTP/HTTPS service through Nginx while the Go application remains bound to `127.0.0.1:8080`.

The v1.4.3 installer applies the following hardened Nginx baseline:

- Nginx version suppression using `server_tokens off`.
- TLS 1.2 and TLS 1.3 only.
- HSTS after successful HTTPS provisioning.
- `X-Content-Type-Options: nosniff`.
- `X-Frame-Options: DENY`.
- `Referrer-Policy: no-referrer`.
- A restrictive `Permissions-Policy`.
- An enforced Content Security Policy.
- Hidden-file and hidden-directory rejection.
- A `64k` client request-body limit.
- Explicit reverse-proxy timeout values.

HSTS is configured without `includeSubDomains`. This is intentional because a RouteGlass deployment should not impose HSTS policy on unrelated sibling subdomains.

### Content Security Policy

The bundled RouteGlass interface contains inline CSS and JavaScript.

Rather than permitting all inline code with `unsafe-inline`, RouteGlass v1.4.3 authorizes only the exact bundled content using SHA-256 CSP hashes.

Current bundled hashes:

- Style: `sha256-cgoLLQ9MeU1lLfrU6mhX00KZaXphodN+43QBEnHyS30=`
- Script: `sha256-ebxcz1IiJ4hCxms3EQrRlUIsZNpp0krhYSJovdjBit0=`

These hashes are byte-sensitive. Any modification to the inline `<style>` or `<script>` content in `internal/httpapi/httpapi.go`, including whitespace or line-ending changes, may require new CSP hashes.

Before releasing a modified web interface, recalculate and validate the hashes against the exact inline text served to browsers.

### Rate Limiting

When rate limiting is enabled during installation, RouteGlass uses two separate Nginx zones:

- `routeglass_web`: 5 requests/second, burst 30.
- `routeglass_diag`: 30 requests/minute, burst 10.

The diagnostic limiter applies specifically to `/api/diagnostics` and `/api/bgp`.

This separation prevents expensive diagnostic operations from consuming the same request budget as ordinary web-interface traffic.

Requests rejected by these limiters return HTTP `429 Too Many Requests`.

### Hidden Files and Scanner Probes

Nginx rejects hidden paths before proxying them to RouteGlass.

This protects against common automated probes for `.env`, `.git/`, `.aws/`, configuration backups, credential files, editor swap files, and similar sensitive paths.

The Let's Encrypt ACME path under `/.well-known/acme-challenge/` remains explicitly permitted.

### Test Files

When public RouteGlass test files are enabled:

- files are served directly by Nginx;
- only `GET` and `HEAD` are accepted;
- RouteGlass security headers are applied to test-file responses;
- the download `Content-Disposition` header is preserved.

### Reverse Proxy Timeouts

The generated Nginx configuration uses `proxy_connect_timeout 5s`, `proxy_send_timeout 60s`, and `proxy_read_timeout 120s`.

The longer read timeout allows diagnostic operations such as traceroute, MTR, and backend BGP lookups to complete without depending on distribution-default proxy timeouts.
