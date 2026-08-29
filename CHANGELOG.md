# Changelog

All notable changes to RouteGlass are documented in this file.

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
