#!/usr/bin/env bash
set -Eeuo pipefail
trap 'rc=$?; echo; echo "[ERROR] Installation failed at line $LINENO (exit code $rc)."; echo "No success message means installation is incomplete."; exit $rc' ERR

ok(){ echo "[OK] $1"; }
step(){ echo; echo "==> $1"; }
fail(){ echo "[ERROR] $1"; exit 1; }

[ "${EUID}" -eq 0 ] || fail "Run as root: sudo ./installer/install.sh"
command -v apt-get >/dev/null || fail "This installer currently supports Debian/Ubuntu systems only."
. /etc/os-release
[ "${ID:-}" = "ubuntu" ] || fail "Supported OS: Ubuntu. Detected: ${PRETTY_NAME:-unknown}"

echo "================================================"
echo " RouteGlass v1.4.2 - Open Source Release Installer"
echo " Author: Hernan Gibran Ramadhan"
echo " Developed by PT Herza Digital Indonesia"
echo "================================================"
echo

read -rp "Company Name: " COMPANY
read -rp "ASN (numbers only): " ASN
read -rp "POP / Location Name: " LOCATION
read -rp "Test IPv4 (optional): " IPV4
read -rp "Test IPv6 (optional): " IPV6
read -rp "Domain Name (e.g. lg.example.com): " DOMAIN
read -rp "Enable rate limiting? [y/N]: " RL
read -rp "Enable Test Files? [y/N]: " TF
TEST_FILES=()
if [[ "${TF,,}" == "y" || "${TF,,}" == "yes" ]]; then
  read -rp "Test file sizes in MB (numbers only, separated by spaces, e.g. 100 200 500 1000): " TF_INPUT
  for size in $TF_INPUT; do
    [[ "$size" =~ ^[0-9]+$ ]] || fail "Test file size must contain numbers only: $size"
    [ "$size" -gt 0 ] || fail "Test file size must be greater than zero."
    TEST_FILES+=("$size")
  done
  [ "${#TEST_FILES[@]}" -gt 0 ] || fail "At least one test file size is required when Test Files are enabled."
fi
read -rp "Enable BGP Lookup? [y/N]: " BGP_ENABLE
MK_ENABLED=false
JN_ENABLED=false
VY_ENABLED=false
MK_ADDR=""; MK_USER=""; MK_PASS=""; MK_CA_SOURCE=""
JN_ADDR=""; JN_PORT="22"; JN_USER=""; JN_KEY_SOURCE=""
VY_ADDR=""; VY_PORT="22"; VY_USER=""; VY_KEY_SOURCE=""
if [[ "${BGP_ENABLE,,}" == "y" || "${BGP_ENABLE,,}" == "yes" ]]; then
  echo
  echo "Select BGP Router Backend:"
  echo
  echo "1) MikroTik RouterOS v7"
  echo "   Connection: API-SSL"
  echo
  echo "2) Juniper JunOS"
  echo "   Connection: SSH key authentication"
  echo
  echo "3) VyOS"
  echo "   Connection: SSH + FRRouting vtysh"
  echo
  read -rp "Select backend [1-3]: " BGP_BACKEND
  case "$BGP_BACKEND" in
    1)
      MK_ENABLED=true
      read -rp "MikroTik API-SSL address (IP:port): " MK_ADDR
      read -rp "MikroTik read-only username: " MK_USER
      read -rsp "MikroTik password: " MK_PASS; echo
      read -rp "Path to exported MikroTik CA certificate: " MK_CA_SOURCE
      [[ -n "$MK_ADDR" && -n "$MK_USER" && -n "$MK_PASS" && -f "$MK_CA_SOURCE" ]] || fail "Invalid MikroTik configuration or CA file."
      ;;
    2)
      JN_ENABLED=true
      read -rp "Juniper NETCONF/SSH address (IP or hostname): " JN_ADDR
      read -rp "Juniper SSH port [22]: " JN_PORT; JN_PORT="${JN_PORT:-22}"
      read -rp "Juniper read-only username: " JN_USER
      read -rp "Path to Juniper private SSH key: " JN_KEY_SOURCE
      [[ -n "$JN_ADDR" && -n "$JN_USER" && -f "$JN_KEY_SOURCE" && "$JN_PORT" =~ ^[0-9]+$ && "$JN_PORT" -ge 1 && "$JN_PORT" -le 65535 ]] || fail "Invalid Juniper configuration or SSH key file."
      ;;
    3)
      VY_ENABLED=true
      read -rp "VyOS SSH address (IP or hostname): " VY_ADDR
      read -rp "VyOS SSH port [22]: " VY_PORT; VY_PORT="${VY_PORT:-22}"
      read -rp "VyOS read-only username: " VY_USER
      read -rp "Path to VyOS private SSH key: " VY_KEY_SOURCE
      [[ -n "$VY_ADDR" && -n "$VY_USER" && -f "$VY_KEY_SOURCE" && "$VY_PORT" =~ ^[0-9]+$ && "$VY_PORT" -ge 1 && "$VY_PORT" -le 65535 ]] || fail "Invalid VyOS configuration or SSH key file."
      ;;
    *) fail "Invalid BGP backend selection. Choose 1, 2 or 3." ;;
  esac
fi
read -rp "Install Let's Encrypt SSL now? [y/N]: " SSL

[ -n "$COMPANY" ] || fail "Company Name is required."
[ -n "$DOMAIN" ] || fail "Domain Name is required."
[[ "$ASN" =~ ^[0-9]+$ ]] || fail "ASN must contain numbers only."

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
[ -f "$ROOT/go.mod" ] || fail "Run installer from the RouteGlass source directory."
[ -f "$ROOT/installer/routeglass.service" ] || fail "Installer files are incomplete."

step "1/11 Installing dependencies"
export DEBIAN_FRONTEND=noninteractive
apt-get update
apt-get install -y nginx golang-go certbot python3-certbot-nginx mtr-tiny traceroute iputils-ping libcap2-bin curl openssh-client
command -v nginx >/dev/null
command -v go >/dev/null
echo "[INFO] Using $(go version)"
GO_MINOR="$(go env GOVERSION | sed -n 's/^go\([0-9]*\)\.\([0-9]*\).*$/\1 \2/p')"
read -r GO_MAJOR GO_MIN <<< "$GO_MINOR"
if [ "${GO_MAJOR:-0}" -lt 1 ] || { [ "${GO_MAJOR:-0}" -eq 1 ] && [ "${GO_MIN:-0}" -lt 22 ]; }; then
  fail "Go 1.22 or newer is required. Detected: $(go version)"
fi
command -v certbot >/dev/null
command -v ping >/dev/null
ok "Dependencies installed"

step "2/11 Creating RouteGlass user and directories"
id routeglass >/dev/null 2>&1 || useradd --system --home /opt/routeglass --shell /usr/sbin/nologin routeglass
install -d -o routeglass -g routeglass /opt/routeglass/bin /etc/routeglass
ok "User and directories ready"

step "3/11 Building RouteGlass"
cd "$ROOT"
go build -o /opt/routeglass/bin/routeglass-server ./cmd/routeglass-server
[ -x /opt/routeglass/bin/routeglass-server ] || fail "RouteGlass binary was not created."
chown routeglass:routeglass /opt/routeglass/bin/routeglass-server
ok "RouteGlass binary built"

step "4/11 Configuring diagnostics"
PING_BIN="$(command -v ping)"
setcap cap_net_raw+ep "$PING_BIN"
getcap "$PING_BIN" | grep -q "cap_net_raw"
ok "Ping capability configured"

RATE=false
RATE_LINE=""
if [[ "${RL,,}" == "y" || "${RL,,}" == "yes" ]]; then
  RATE=true
  RATE_LINE="        limit_req zone=routeglass burst=10 nodelay;"
fi

step "5/11 Writing RouteGlass configuration"
TEST_FILES_CSV="$(IFS=,; echo "${TEST_FILES[*]:-}")"
python3 - "$COMPANY" "$ASN" "$LOCATION" "$IPV4" "$IPV6" "$RATE" "$TEST_FILES_CSV" <<'PY'
import json, sys
files=[int(x) for x in sys.argv[7].split(",") if x]
cfg={"company_name":sys.argv[1],"asn":sys.argv[2],"location":sys.argv[3],
     "test_ipv4":sys.argv[4],"test_ipv6":sys.argv[5],
     "rate_limit":sys.argv[6].lower()=="true","test_files":files,"mikrotik":{"enabled":False},"juniper":{"enabled":False},"vyos":{"enabled":False}}
with open("/etc/routeglass/config.json","w") as f: json.dump(cfg,f,indent=2)
PY
chmod 0640 /etc/routeglass/config.json
if [[ "$MK_ENABLED" == "true" ]]; then
 install -d -m 0755 /etc/routeglass/certs
 MIKROTIK_CA_DEST="/etc/routeglass/certs/mikrotik-ca.crt"
 CA_SOURCE="$(readlink -f "$MK_CA_SOURCE")"; CA_DEST="$(readlink -f "$MIKROTIK_CA_DEST" 2>/dev/null || printf '%s' "$MIKROTIK_CA_DEST")"
 if [ "$CA_SOURCE" != "$CA_DEST" ]; then install -m 0644 "$MK_CA_SOURCE" "$MIKROTIK_CA_DEST"; else chmod 0644 "$MIKROTIK_CA_DEST"; fi
 printf '%s' "$MK_PASS" > /etc/routeglass/mikrotik-password
 chown routeglass:routeglass /etc/routeglass/mikrotik-password
 chmod 0600 /etc/routeglass/mikrotik-password
 python3 - "$MK_ADDR" "$MK_USER" <<'PY2'
import json,sys
p='/etc/routeglass/config.json'; c=json.load(open(p))
c['mikrotik']={'enabled':True,'address':sys.argv[1],'username':sys.argv[2],'password_file':'/etc/routeglass/mikrotik-password','ca_file':'/etc/routeglass/certs/mikrotik-ca.crt','timeout_seconds':8}
json.dump(c,open(p,'w'),indent=2)
PY2
fi
if [[ "$JN_ENABLED" == "true" ]]; then
 install -d -m 0755 /etc/routeglass/ssh
 install -m 0600 "$JN_KEY_SOURCE" /etc/routeglass/ssh/juniper_routeglass
 chown -R routeglass:routeglass /etc/routeglass/ssh
 # Pin the current SSH host key into a dedicated known_hosts file.
 ssh-keyscan -p "$JN_PORT" -H "$JN_ADDR" > /etc/routeglass/ssh/juniper_known_hosts 2>/dev/null || fail "Unable to retrieve Juniper SSH host key."
 chown routeglass:routeglass /etc/routeglass/ssh/juniper_known_hosts; chmod 0644 /etc/routeglass/ssh/juniper_known_hosts
 python3 - "$JN_ADDR" "$JN_PORT" "$JN_USER" <<'PY3'
import json,sys
p='/etc/routeglass/config.json'; c=json.load(open(p))
c['juniper']={'enabled':True,'address':sys.argv[1],'port':int(sys.argv[2]),'username':sys.argv[3],'private_key_file':'/etc/routeglass/ssh/juniper_routeglass','known_hosts_file':'/etc/routeglass/ssh/juniper_known_hosts','timeout_seconds':8}
json.dump(c,open(p,'w'),indent=2)
PY3
fi
if [[ "$VY_ENABLED" == "true" ]]; then
 install -d -m 0755 /etc/routeglass/ssh
 install -m 0600 "$VY_KEY_SOURCE" /etc/routeglass/ssh/vyos_routeglass
 chown -R routeglass:routeglass /etc/routeglass/ssh
 ssh-keyscan -p "$VY_PORT" -H "$VY_ADDR" > /etc/routeglass/ssh/vyos_known_hosts 2>/dev/null || fail "Unable to retrieve VyOS SSH host key."
 chown routeglass:routeglass /etc/routeglass/ssh/vyos_known_hosts; chmod 0644 /etc/routeglass/ssh/vyos_known_hosts
 python3 - "$VY_ADDR" "$VY_PORT" "$VY_USER" <<'PY4'
import json,sys
p='/etc/routeglass/config.json'; c=json.load(open(p))
c['vyos']={'enabled':True,'address':sys.argv[1],'port':int(sys.argv[2]),'username':sys.argv[3],'private_key_file':'/etc/routeglass/ssh/vyos_routeglass','known_hosts_file':'/etc/routeglass/ssh/vyos_known_hosts','timeout_seconds':8}
json.dump(c,open(p,'w'),indent=2)
PY4
fi
chown routeglass:routeglass /etc/routeglass/config.json
python3 -m json.tool /etc/routeglass/config.json >/dev/null
ok "Configuration created"

step "6/11 Preparing Test Files"
rm -rf /var/www/routeglass-testfiles
if [ "${#TEST_FILES[@]}" -gt 0 ]; then
  install -d -m 0755 -o root -g root /var/www/routeglass-testfiles
  for size in "${TEST_FILES[@]}"; do
    file="/var/www/routeglass-testfiles/${size}MB.bin"
    truncate -s "${size}M" "$file"
    chmod 0644 "$file"
  done
  ok "Test files created: ${TEST_FILES[*]} MB"
else
  ok "Test files disabled"
fi

step "7/11 Installing and starting systemd service"
cp "$ROOT/installer/routeglass.service" /etc/systemd/system/routeglass.service
systemctl daemon-reload
systemctl enable routeglass
systemctl restart routeglass
sleep 2
systemctl is-active --quiet routeglass || { journalctl -u routeglass -n 80 --no-pager; fail "RouteGlass service is not active."; }
ss -lntp | grep -q '127.0.0.1:8080' || fail "RouteGlass is not listening on 127.0.0.1:8080."
curl --fail --silent --show-error --max-time 5 http://127.0.0.1:8080/ | grep -q "RouteGlass" || fail "RouteGlass backend verification failed."
PID="$(systemctl show -p MainPID --value routeglass)"
[ "$PID" -gt 0 ] || fail "Could not determine RouteGlass service PID."
grep -q "0000000000002000" "/proc/$PID/status" || fail "RouteGlass service does not have CAP_NET_RAW."
ok "RouteGlass service, loopback backend, and ICMP capability verification passed"

step "8/11 Configuring Nginx"
if [ "$RATE" = true ]; then
cat >/etc/nginx/conf.d/routeglass-rate-limit.conf <<'EOF'
limit_req_zone $binary_remote_addr zone=routeglass:10m rate=30r/m;
EOF
else
rm -f /etc/nginx/conf.d/routeglass-rate-limit.conf
fi

TEST_FILES_LOCATION=""
if [ "${#TEST_FILES[@]}" -gt 0 ]; then
  TEST_FILES_LOCATION=$'    location /testfiles/ {\n        alias /var/www/routeglass-testfiles/;\n        default_type application/octet-stream;\n        add_header Content-Disposition "attachment";\n    }'
fi
python3 - "$ROOT/installer/routeglass.nginx.conf.template" "$DOMAIN" "$RATE_LINE" "$TEST_FILES_LOCATION" > /etc/nginx/sites-available/routeglass <<'PY'
import sys

template_path, domain, rate_line, test_files_location = sys.argv[1:]
with open(template_path, "r", encoding="utf-8") as f:
    content = f.read()

content = content.replace("__DOMAIN__", domain)
content = content.replace("__RATE_LIMIT__", rate_line)
content = content.replace("__TEST_FILES_LOCATION__", test_files_location)
sys.stdout.write(content)
PY
ln -sf /etc/nginx/sites-available/routeglass /etc/nginx/sites-enabled/routeglass
rm -f /etc/nginx/sites-enabled/default
nginx -t
systemctl enable nginx
systemctl restart nginx
systemctl is-active --quiet nginx || fail "Nginx is not active."
ok "Nginx configured"

step "9/11 Verifying local HTTP"
curl --fail --silent --show-error --max-time 10 -H "Host: $DOMAIN" http://127.0.0.1/ | grep -q "RouteGlass" || fail "Local HTTP verification failed."
ok "HTTP RouteGlass verification passed"

if [[ "${SSL,,}" == "y" || "${SSL,,}" == "yes" ]]; then
 step "10/11 Installing Let's Encrypt SSL"
 read -rp "Email address for Let's Encrypt: " EMAIL
 [ -n "$EMAIL" ] || fail "Email is required for Let's Encrypt."
 certbot --nginx -d "$DOMAIN" --non-interactive --agree-tos --email "$EMAIL" --redirect
 nginx -t
 systemctl restart nginx
 ss -lnt | grep -qE '[:.]443 ' || fail "Nginx is not listening on port 443."
 curl --fail --silent --show-error --max-time 15 --resolve "$DOMAIN:443:127.0.0.1" "https://$DOMAIN/" | grep -q "RouteGlass" || fail "Local HTTPS verification failed."
 ok "HTTPS RouteGlass verification passed"
else
 step "10/11 SSL skipped"
 echo "[INFO] HTTP is active. HTTPS can be enabled later with certbot."
fi

step "11/11 Final verification"
systemctl is-active --quiet routeglass || fail "Final check: RouteGlass service inactive."
systemctl is-active --quiet nginx || fail "Final check: Nginx inactive."
ss -lntp | grep -q "127.0.0.1:8080" || fail "Final check: RouteGlass backend port 8080 missing."
if [[ "${SSL,,}" == "y" || "${SSL,,}" == "yes" ]]; then
  curl --fail --silent --show-error --max-time 10 --resolve "$DOMAIN:443:127.0.0.1" "https://$DOMAIN/" | grep -q "RouteGlass" || fail "Final HTTPS check failed."
else
  curl --fail --silent --show-error --max-time 10 -H "Host: $DOMAIN" http://127.0.0.1/ | grep -q "RouteGlass" || fail "Final HTTP check failed."
fi
ok "Final verification passed"

echo
echo "================================================"
echo " RouteGlass installation completed successfully"
echo "================================================"
echo "HTTP:  http://$DOMAIN"
if [[ "${SSL,,}" == "y" || "${SSL,,}" == "yes" ]]; then echo "HTTPS: https://$DOMAIN"; fi
echo "Service: systemctl status routeglass"
echo "Logs:    journalctl -u routeglass -f"
