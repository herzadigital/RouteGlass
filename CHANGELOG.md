# Changelog

> Versions 1.4.4 and 1.4.5 below are proposed release milestones documented on the security review branch. No GitHub release/tag or production deployment is implied.

## [1.4.5] - 2026-11-24 (unreleased)

### Production validation
- Updated production NGINX to Ubuntu package 1.24.0-2ubuntu7.21 and Go binaries to Go 1.26.11; production dependencies include golang.org/x/crypto v0.59.0.
- Replaced Juniper text route parsing with structured XML in production variants, isolating active route attributes from advertised metrics and rejecting RPC/SSH errors.
- Validated TELKOM local preference 220, actual AS path and communities, static routes, IPv6 and no-route responses on Cyber1.
- Deployed XML parser to all eight Juniper LGs; final BGP, traceroute, MTR and service smoke tests passed on all ten production LGs, including both VyOS nodes.
- Source and binary vulnerability scans passed for Cyber1 candidate; remaining Juniper candidate binaries passed binary scans.
- Public repository source reconciliation, installer and all-backend release QA remain required. Production deployment is not proof of public release equivalence. No stable tag or package published yet.

## [1.4.4] - 2026-11-19 (planned; not released) — Security review

- Restrict diagnostic targets to IP literals and DNS-label hostnames to prevent leading-dash argument injection.
- Add regression tests for option-like and malformed diagnostic targets.
- Record production CSP/frame-policy drift and remaining audit work in [Security Audit](docs/SECURITY-AUDIT-2026-11.md).
- Go unit tests and go vet passed on the separate controller build workspace (Go 1.22.2). Binary-mode govulncheck on a Go 1.26.11 local candidate returned no vulnerabilities; staging acceptance remains pending. Not released or deployed.
- Added a staging-only Nginx CSP baseline compatible with Cloudflare Challenge origins (requires browser validation).


All notable changes to RouteGlass are documented in this file.

## [1.4.3] - 2026-08-30

### Security
- Hardened the default Nginx deployment for public-facing RouteGlass installations.
- Disabled Nginx version disclosure with `server_tokens off`.
- Restricted TLS protocol support to TLS 1.2 and TLS 1.3.
- Disabled server-side cipher preference so modern TLS negotiation is used consistently.
- Added `Strict-Transport-Security: max-age=31536000` after successful HTTPS provisioning.
- HSTS intentionally does not use `includeSubDomains`, preventing a RouteGlass deployment from imposing HSTS policy on unrelated sibling services.
- Added `X-Content-Type-Options: nosniff`.
- Added `X-Frame-Options: DENY`.
- Added `Referrer-Policy: no-referrer`.
- Added `Permissions-Policy` restrictions for camera, microphone, and geolocation.
- Added an enforced Content Security Policy for the bundled web interface.
- Inline CSS and JavaScript are authorized with exact SHA-256 hashes instead of `unsafe-inline`.
- Added hidden-file protection at the Nginx layer for paths such as `.env`, `.git`, and `.aws`.
- Preserved `/.well-known/acme-challenge/` access for Let's Encrypt validation.
- Added a `64k` client request-body limit.
- Hardened test-file delivery so only `GET` and `HEAD` are accepted.
- Ensured test-file responses receive the same security headers as the main application.
- Added explicit reverse-proxy connect, send, and read timeouts.

### Rate Limiting
- Split optional Nginx rate limiting into two independent protection zones.
- Normal web traffic uses `routeglass_web` at `5 requests/second` with a burst allowance of `30`.
- Diagnostic and BGP API traffic uses `routeglass_diag` at `30 requests/minute` with a burst allowance of `10`.
- Rate-limited requests return HTTP `429 Too Many Requests`.
- Diagnostic/BGP rate limiting is isolated from normal page traffic.
- Rate limiting remains optional during installation.

### Compatibility
- No RouteGlass API semantics or BGP backend behavior were changed.
- `/api/diagnostics` remains the primary diagnostics endpoint.
- `/api/bgp` remains available for backward compatibility.
- RouteGlass continues to bind its application backend to `127.0.0.1:8080`.
- IPv4 and IPv6 public access use the same Nginx security policy.
- Juniper, MikroTik RouterOS v7, and VyOS/FRRouting integrations are unchanged.

### Installer
- Updated the installer banner to v1.4.3.
- Updated Nginx template rendering for separate web and diagnostic rate-limit policies.
- Added centralized RouteGlass security-header configuration.
- Added post-Certbot HSTS activation.
- Added hardened test-file location generation.
- Added explicit Nginx proxy timeout configuration.

### Validation
- Validated the hardened configuration on Ubuntu 24.04.
- Verified TLS 1.0 and TLS 1.1 rejection.
- Verified TLS 1.2 and TLS 1.3 operation.
- Verified HTTP-to-HTTPS redirection.
- Verified HTTPS operation over IPv4 and IPv6.
- Verified enforced CSP with the bundled CSS and JavaScript SHA-256 hashes.
- Verified hidden-file probes return HTTP `404`.
- Verified diagnostic rate limiting returns HTTP `429` without rate-limiting the normal homepage.
- Verified hardened test-file downloads remain functional.
- Verified rendered Nginx configurations with rate limiting enabled and disabled using `nginx -t`.

### Operator Note
- CSP hashes are tied to the exact inline CSS and JavaScript in `internal/httpapi/httpapi.go`.
- Any future change to those inline blocks, including whitespace, requires the CSP hashes to be recalculated and validated before release.

## [1.4.2] - 2026-08-29

### Security
- Sanitized public BGP backend errors to prevent disclosure of router management IP addresses, SSH/API ports, usernames, and low-level transport details.
- Public BGP failures now return only `BGP query is temporarily unavailable...`.
- Detailed backend errors are logged server-side for administrators.
- BGP backend failures return HTTP `503 Service Unavailable`.
- Browser-side BGP fetch failures no longer display raw exception details.

### Documentation
- Added public error-sanitization guidance to the Security documentation.
- Updated release documentation and Quick Start references for v1.4.2.

## [1.4.1]
- Added VyOS / FRRouting BGP backend support.
- Added SSH key authentication and host-key verification for VyOS.
- Added IPv4 and IPv6 route lookup support using FRRouting `vtysh` JSON output.
- Updated installer and documentation for three supported BGP backends.

## [1.4.0]
- Finalized single BGP backend selection during installation.
- Added MikroTik RouterOS v7 and Juniper Junos backend selection.
- Expanded installation, backend, security, and troubleshooting documentation for public distribution.
