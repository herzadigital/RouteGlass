# Staging-only Nginx CSP candidate (Cloudflare Challenge)

**Do not deploy to production without testing the exact HTML, scripts and Cloudflare Challenge flow.**

The default RouteGlass v1.4.3 CSP uses hashes for its bundled inline script and style. LG-MNL and LG-KUL currently allow unsafe-inline and Cloudflare Challenge. The candidate below combines hash authorization with Cloudflare's challenge origin. This is an **example for staging**, not a complete Nginx server configuration.

```nginx
# Put this in the appropriate HTTPS server block; do not create duplicate CSP headers.
add_header X-Frame-Options "DENY" always;
add_header Content-Security-Policy "default-src 'self'; base-uri 'self'; object-src 'none'; frame-ancestors 'none'; form-action 'self'; img-src 'self' data:; font-src 'self' data:; style-src 'self' 'sha256-cgoLLQ9MeU1lLfrU6mhX00KZaXphodN+43QBEnHyS30='; script-src 'self' 'sha256-ebxcz1IiJ4hCxms3EQrRlUIsZNpp0krhYSJovdjBit0=' https://challenges.cloudflare.com; connect-src 'self' https://challenges.cloudflare.com; frame-src https://challenges.cloudflare.com;" always;
```

## Acceptance gates
1. Confirm the exact served inline script/style SHA-256 values match the bundled hashes. If custom templates differ, recompute hashes before deployment.
2. Confirm Cloudflare Turnstile/Challenge execution, including any inline bootstrap or additional origins required by the actual integration. Use CSP Report-Only first to collect violations.
3. Check that Nginx configuration has only one effective CSP policy and validate with `nginx -t`.
4. Verify GET/HEAD homepage, BGP/diagnostic endpoints, test-file downloads, rate limits, 4xx/5xx error pages, TLS 1.2/1.3, IPv4/IPv6.
5. Compare pre/post headers and monitor errors; retain config backup and tested rollback.
6. Roll out to one staging host, then one production pilot, only after approval.

The current policy is not proof of exploitable XSS. This candidate is a defense-in-depth improvement.
