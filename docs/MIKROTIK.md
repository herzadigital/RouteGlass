# MikroTik RouterOS v7 BGP Backend

## Security Model

RouteGlass uses API-SSL with a dedicated read-only account and validates the router certificate using the CA certificate supplied during installation.

## Required Router Settings

- Enable `api-ssl`.
- Restrict source access to the RouteGlass VM IP address.
- Use a dedicated read-only account.
- Use a certificate signed by a trusted or dedicated CA.
- Keep unrelated management services disabled or separately restricted.

## Pre-install Tests

From the RouteGlass VM:

```bash
nc -zv ROUTER_IP 8729
openssl s_client -connect ROUTER_IP:8729 -CAfile /path/to/router-ca.crt -verify_return_error
```

The TLS verification should complete successfully before installing RouteGlass.

## Validation

After installation:

```bash
curl -sS 'http://127.0.0.1:8080/api/diagnostics?type=bgp&target=1.1.1.1'
```
