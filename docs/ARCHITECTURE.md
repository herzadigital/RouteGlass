# Architecture

```text
Internet Visitor
      |
      v
    Nginx
      |
      v
RouteGlass HTTP Service (127.0.0.1:8080)
      |
      +-- Ping / MTR / Traceroute
      |
      +-- BGP Lookup
             |
             +-- MikroTik RouterOS v7 / API-SSL
             |
             +-- Juniper Junos / SSH key authentication
```

Only one BGP backend is configured per installation. Nginx is the public-facing component; the RouteGlass application listens only on loopback.
