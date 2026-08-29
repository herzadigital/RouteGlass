# RouteGlass v1.4.2

RouteGlass is an Open Source web-based network looking glass platform for network operators. It provides Ping, MTR, Traceroute, optional downloadable test files, visitor IP visibility, and optional BGP route lookup backed by MikroTik RouterOS v7, Juniper Junos, or VyOS / FRRouting.

## Highlights

- Ubuntu 24.04 LTS installer
- Nginx reverse proxy
- Optional Let's Encrypt HTTPS
- Ping, MTR and Traceroute
- Optional Test Files with installer-defined sizes
- Visitor public IP display
- Optional rate limiting
- One BGP Lookup option in Network Diagnostics
- Select exactly one BGP backend during installation:
  - MikroTik RouterOS v7 via API-SSL
  - Juniper Junos via SSH key authentication
  - VyOS / FRRouting via SSH key authentication
- MikroTik CA validation and SSH host-key pinning for Juniper and VyOS
- Non-root RouteGlass service with only `CAP_NET_RAW` for ICMP

## Quick Start

```bash
sudo apt update && sudo apt upgrade -y
unzip routeglass-v1.4.2-open-source-release.zip
cd routeglass-v1.4.2-open-source-release
chmod +x installer/install.sh
sudo ./installer/install.sh
```

For a production deployment, read [docs/INSTALLATION.md](docs/INSTALLATION.md) before running the installer.

## BGP Backend Selection

The installer uses a single decision flow:

```text
Enable BGP Lookup? [y/N]: y

Select BGP Router Backend:

1) MikroTik RouterOS v7
   Connection: API-SSL

2) Juniper Junos
   Connection: SSH key authentication

3) VyOS / FRRouting
   Connection: SSH key authentication

Select backend [1-3]:
```

Only one backend is configured per RouteGlass instance. If BGP Lookup is disabled, Ping, MTR and Traceroute remain available.

## Juniper Default Port

The Juniper SSH prompt is intentionally generic:

```text
Juniper SSH port [22]:
```

Port `22` is the default because RouteGlass is distributed as a general-purpose Open Source platform. Operators using a non-standard SSH port can enter their own value.

## Documentation

- [Installation Guide](docs/INSTALLATION.md)
- [Architecture](docs/ARCHITECTURE.md)
- [Security](docs/SECURITY.md)
- [Test Files](docs/TEST-FILES.md)
- [MikroTik BGP Backend](docs/MIKROTIK.md)
- [Juniper BGP Backend](docs/JUNIPER.md)
- [VyOS BGP Backend](docs/VYOS.md)
- [Troubleshooting](docs/TROUBLESHOOTING.md)

## License and Notices

See [LICENSE](LICENSE), [NOTICE](NOTICE), [ATTRIBUTION.md](ATTRIBUTION.md), and [TRADEMARKS.md](TRADEMARKS.md).

## Release Notes: v1.4.2

- Hardened public BGP error handling to prevent router management details from being exposed to Looking Glass visitors.
- BGP backend failures now return only `BGP query is temporarily unavailable...` to the public UI and API.
- Detailed backend errors remain available in the RouteGlass service log for operator troubleshooting.
- Public BGP backend failures now use HTTP `503 Service Unavailable`.
- Sanitized browser-side BGP request failures so transport exception details are not rendered to visitors.

## Release Notes: v1.4.1

- Added VyOS / FRRouting as a supported BGP lookup backend.
- Added SSH key authentication and host-key verification for VyOS.
- Added IPv4 and IPv6 route lookup support using FRRouting `vtysh` JSON output.
- Updated the installer to support three BGP backend options.
- Updated RouteGlass documentation for VyOS deployment and configuration.

## Release Notes: v1.4.0

- Finalized the installer UX around one `Enable BGP Lookup?` option.
- Added explicit backend selection for MikroTik RouterOS v7 or Juniper Junos.
- Changed the Juniper SSH default port from an environment-specific value to the standard port `22`.
- Expanded installation, backend, security, and troubleshooting documentation for public distribution.
- Prepared as the Open Source distribution release.

## VyOS backend

RouteGlass supports VyOS through SSH and FRRouting `vtysh` JSON output for IPv4 and IPv6 BGP lookups. See `docs/VYOS.md`.
