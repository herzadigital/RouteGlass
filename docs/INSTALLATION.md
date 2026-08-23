# Installation Guide

## 1. Supported Platform

The installer supports **Ubuntu 24.04 LTS**. Run it as `root` or through `sudo`.

Recommended baseline:

- 1 vCPU or more
- 1 GB RAM or more
- Public IPv4 recommended
- Optional public IPv6
- A DNS name if HTTPS will be enabled
- TCP/80 and TCP/443 reachable when using Let's Encrypt

Use a clean VM for the first deployment whenever possible.

## 2. Prepare DNS

Create an A record pointing the RouteGlass hostname to the VM public IPv4. Add an AAAA record only if IPv6 is configured and reachable.

Example:

```text
lg.example.net -> 203.0.113.10
```

Verify before requesting Let's Encrypt:

```bash
getent ahosts lg.example.net
curl -I http://lg.example.net
```

## 3. Install

Upload or clone the release, then:

```bash
unzip routeglass-v1.4.0-open-source-release.zip
cd routeglass-v1.4.0-open-source-release
chmod +x installer/install.sh
sudo ./installer/install.sh
```

The installer stops on errors. A successful installation ends with:

```text
RouteGlass installation completed successfully
```

## 4. Installer Questions

### Basic identity

- **Company Name**: displayed in the RouteGlass interface.
- **ASN**: numbers only.
- **POP / Location Name**: descriptive POP location.
- **Test IPv4 / IPv6**: optional addresses displayed to visitors.
- **Domain Name**: the public hostname served by Nginx.

### Rate limiting

Enable if you want Nginx to limit repeated requests from the same source IP.

### Test Files

Enable with `y`, then enter sizes as numbers separated by spaces:

```text
100 500 1000
```

This creates sparse files named `100MB.bin`, `500MB.bin`, and `1000MB.bin` under `/testfiles/`.

### BGP Lookup

The installer asks:

```text
Enable BGP Lookup? [y/N]:
```

If enabled, select exactly one backend.

#### MikroTik

Provide:

- API-SSL address and port
- Read-only username
- Password
- Exported CA certificate path

The certificate is copied to `/etc/routeglass/certs/mikrotik-ca.crt` and the password is stored with mode `0600`.

#### Juniper

Provide:

- Router IP address or hostname
- SSH port, default `22`
- Read-only username
- Path to the private SSH key

Example:

```text
Juniper NETCONF/SSH address (IP or hostname): 192.0.2.1
Juniper SSH port [22]:
Juniper read-only username: routeglass
Path to Juniper private SSH key: /root/juniper_routeglass
```

The private key is copied into `/etc/routeglass/ssh/juniper_routeglass` with restrictive permissions. The installer obtains and pins the router host key in `/etc/routeglass/ssh/juniper_known_hosts`.

### Let's Encrypt

Choose `y` only after DNS is pointing to the VM and TCP/80 is reachable from the Internet.

## 5. Post-install Verification

```bash
systemctl status routeglass
systemctl status nginx
curl -sS http://127.0.0.1:8080/
```

For the public site:

```bash
curl -I http://your-domain.example
curl -Ik https://your-domain.example
```

## 6. Service Management

```bash
sudo systemctl restart routeglass
sudo systemctl restart nginx
sudo journalctl -u routeglass -f
```

## 7. Reinstallation

Back up `/etc/routeglass/` before reinstalling if you need to preserve backend credentials or keys. Use a clean VM for the safest validation of a release.
