# Security audit and hardening candidate — 2026-11-19

Status: **under review; not deployed to production**.

## Scope and evidence
- Repository `herzadigital/RouteGlass` (Go; no PHP dependency in `go.mod`).
- Public HTTPS header inspection of `lg-mnl.herza.cloud` and `lg-kul.herza.cloud`: both returned HTTP 200.
- Production CSP permits `script-src 'unsafe-inline'` and `style-src 'unsafe-inline'`, whereas the v1.4.3 documented baseline uses SHA-256 hashes. Production X-Frame-Options is SAMEORIGIN, while documented baseline is DENY.
- Internal OS, service, PHP package inventory, Nginx configuration and router credential verification **not yet assessed**.

## Source finding RG-SEC-001 — option injection in diagnostics (candidate fix)
Prior `validTarget` rejected shell metacharacters but permitted leading dashes, e.g. `-V` and `--help`. Diagnostic tools use `exec.CommandContext` without a shell, but a leading dash can still be parsed by the invoked tool as an option. The candidate fix accepts IP literals and DNS-label hostnames only. Unit tests cover common malicious and valid inputs. **Not yet independently built/tested on CI or deployed**.

## Recommended follow-up
1. Run `go test ./...`, `go vet ./...`, `govulncheck ./...` and review all Go stdlib/toolchain advisories on the build host.
2. Audit `httpapi` endpoints for methods, request concurrency, public error messages and diagnostic process limits.
3. Check trusted reverse-proxy handling of `X-Real-IP` and avoid trusting user-supplied headers from direct connections.
4. Compare live Nginx config to documented CSP and frame policy; retain compatibility with Cloudflare Challenge where configured.
5. On each production VM, inspect installed packages, PHP exposure if any, TLS, firewall, systemd sandbox, loopback-only Go listener and router backend credentials. Do not copy secrets into audit reports.
6. Stage, test, back up, and roll back before applying changes to production.

## Deployment
No production configuration or binary has been changed by this repository audit.

## Validation update — 2026-11-19
- On the separate Ansible controller build workspace, installed distro Go 1.22.2 and ran `go test ./...`: PASS (diagnostics regression tests passed; other packages have no test files).
- `go vet ./...`: PASS.
- `go list -m all`: only the RouteGlass module, no external Go module dependencies listed. Go standard-library/toolchain vulnerabilities remain in scope.
- `govulncheck`: not yet executed; a clean `go test` result does not establish CVE-free status.
- Draft Cloudflare CSP staging baseline: [NGINX-CLOUDFLARE-CSP-STAGING.md](NGINX-CLOUDFLARE-CSP-STAGING.md).
- No production VM internal configuration audited or modified yet.
