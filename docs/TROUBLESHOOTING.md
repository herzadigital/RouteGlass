# Troubleshooting

## RouteGlass service is inactive

```bash
systemctl status routeglass --no-pager -l
journalctl -u routeglass -n 100 --no-pager
```

## Nginx returns 502

Check the backend first:

```bash
curl -v http://127.0.0.1:8080/
ss -lntp | grep 8080
systemctl status routeglass
```

Then check Nginx:

```bash
nginx -t
systemctl status nginx
journalctl -u nginx -n 100 --no-pager
```

## Ping reports missing CAP_NET_RAW

```bash
getcap $(command -v ping)
systemctl cat routeglass
```

Re-run the installer or restore the capability:

```bash
sudo setcap cap_net_raw+ep $(command -v ping)
sudo systemctl restart routeglass
```

## MikroTik BGP lookup fails

Check reachability and TLS validation:

```bash
nc -zv ROUTER_IP 8729
openssl s_client -connect ROUTER_IP:8729 -CAfile /etc/routeglass/certs/mikrotik-ca.crt -verify_return_error
```

Confirm the configured address and user in `/etc/routeglass/config.json` without exposing the password file contents.

## Juniper BGP lookup fails

Check TCP connectivity and SSH key authentication:

```bash
nc -zv ROUTER_IP SSH_PORT
sudo -u routeglass ssh -i /etc/routeglass/ssh/juniper_routeglass \
  -o UserKnownHostsFile=/etc/routeglass/ssh/juniper_known_hosts \
  -o StrictHostKeyChecking=yes USER@ROUTER_IP 'show route 1.1.1.1'
```

If the router host key changed intentionally, regenerate the pinned known-hosts file only after independently verifying the new host key.

## HTTPS verification fails

Confirm DNS and port access:

```bash
getent ahosts YOUR_DOMAIN
ss -lntp | grep -E ':80|:443'
```

Review Certbot:

```bash
journalctl -u nginx -n 100 --no-pager
cat /var/log/letsencrypt/letsencrypt.log
```

## Important Logs

```bash
journalctl -u routeglass -f
journalctl -u nginx -f
```
