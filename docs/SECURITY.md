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

## Secrets

Do not commit private SSH keys, MikroTik passwords, exported certificates containing private material, or `/etc/routeglass/` contents to Git.

## Updates

Keep Ubuntu, Go, Nginx, Certbot, RouterOS, and Junos patched according to your operational maintenance policy.
