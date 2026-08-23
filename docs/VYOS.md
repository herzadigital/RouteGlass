# VyOS BGP Lookup Backend

RouteGlass supports VyOS routers running FRRouting (FRR) through SSH. The RouteGlass service authenticates with a dedicated SSH key and executes read-only `vtysh` show commands.

## Router-side principle

Use a dedicated RouteGlass account with the minimum permissions required to execute the relevant `vtysh` commands. Do not use the VyOS root account. Restrict SSH access with an ACL or firewall to the RouteGlass server IP where possible.

## Commands used

For IPv4, RouteGlass queries the route covering the target and then requests the BGP entry for the selected prefix:

- `vtysh -c 'show ip route TARGET json'`
- `vtysh -c 'show bgp ipv4 unicast PREFIX json'`

For IPv6:

- `vtysh -c 'show ipv6 route TARGET json'`
- `vtysh -c 'show bgp ipv6 unicast PREFIX json'`

JSON output is used to avoid brittle parsing of terminal formatting.

## Installer inputs

Select backend `3` and provide:

- VyOS SSH address
- SSH port (default `22`)
- dedicated read-only username
- path to the private SSH key on the server running the installer

The installer copies the key to `/etc/routeglass/ssh/vyos_routeglass`, pins the current SSH host key in `/etc/routeglass/ssh/vyos_known_hosts`, and writes the connection details to `/etc/routeglass/config.json`.

## Security notes

- Use key authentication only.
- Restrict the VyOS account and source IP.
- Do not expose a special SSH port merely for RouteGlass; the installer defaults to standard port 22.
- Host-key verification is enabled. If the router SSH host key changes, update the dedicated known_hosts file deliberately.
