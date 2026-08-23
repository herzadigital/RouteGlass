# Juniper Junos BGP Backend

## Security Model

RouteGlass connects to Junos using a dedicated read-only account and SSH public-key authentication. The installer pins the router SSH host key and runtime uses strict host-key checking.

Do not reuse a super-user account.

## Recommended Connectivity

The preferred design is to restrict SSH access with routing and firewall policy so that only the RouteGlass VM can reach the router management or lookup endpoint. A private transport path is preferred when available, but the installer also supports a public address when access is tightly restricted.

## Router User

Create a dedicated user with permissions sufficient only for route inspection. Validate manually before installing RouteGlass:

```text
show route 1.1.1.1
```

The account must not be able to enter configuration mode.

## SSH Key

Generate a dedicated key pair on the RouteGlass VM:

```bash
ssh-keygen -t ed25519 -f /etc/routeglass/juniper_routeglass -C routeglass
```

Install the public key for the dedicated Junos user. Protect the private key:

```bash
chmod 600 /etc/routeglass/juniper_routeglass
```

During installation provide the private-key path, not the `.pub` file.

## Port

The default prompt is:

```text
Juniper SSH port [22]:
```

Enter a different port only if your Junos SSH service uses a non-standard port.

## Validation

From the RouteGlass VM:

```bash
nc -zv ROUTER_IP 22
ssh -i /etc/routeglass/juniper_routeglass USER@ROUTER_IP 'show route 1.1.1.1'
```

After installation:

```bash
curl -sS 'http://127.0.0.1:8080/api/diagnostics?type=bgp&target=1.1.1.1'
```
