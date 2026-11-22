# Production Go binary and backup audit — 2026-11-22

Status: **ACTION REQUIRED — Go production binary security findings**. NGINX package patch is complete on all ten VMs, but this is not an application security clearance.

## Inventory and binary provenance
Root-authorized RDC inspection found **Go 1.22.2 embedded in all inspected production binary builds**. All ten host binaries were hashed; these are different binary families and builds:
- `herza-lookingglass` with `golang.org/x/crypto v0.31.0`: LG-Cyber1, LG-Kalibata, LG-Tifa (same SHA-256 `d97cb625b5431e1dfb88b03c53f6dee036e882dd9c96e4f51f49766f888c1bf5`), LG-SUB1 and LG-SUB2 (same SHA-256 `073c83ce84556a007859a0854b6ef6e9e73aca09a6406adb3ea7930a3c56ce45`), LG-NCIX (SHA-256 `ca790303935e68594c8df80be58dc55b70ed7330ac04449ddc3284a4f12f98a8`), LG-Cibitung (SHA-256 `bc8b12ac6e336b29471df4bf6b1c1bf32c784d7bd12782bcc26b5f790beff5f1`).
- LG-SG binary SHA-256 `073c83ce84556a007859a0854b6ef6e9e73aca09a6406adb3ea7930a3c56ce45`, **identical** to LG-SUB1/SUB2; Go CLI was not installed on SG, so module attribution is by identical binary hash, not an independent local `go version -m`.
- `github.com/herzadigital/herza-lookingglass-vyos` with `golang.org/x/crypto v0.24.0`: LG-KUL and LG-MNL (same SHA-256 `24b48d93ab7b78568b6c53cb7f551276ba6742ede93dd3298a8a51acf2d2d1ee`).
- Public `herzadigital/RouteGlass` is a **different** Go module; its local Go 1.26.11 candidate binary scanned with `govulncheck -mode=binary` and returned `No vulnerabilities found`. **This does not clear any production binary or authorize swapping binaries.**

## Production vulnerability scans
Using `govulncheck v1.10.0` with vulnerability DB updated 2026-11-11:
- LG-Cyber1 **actual deployed binary**: `govulncheck -mode=binary` exit 3; 49 vulnerabilities with matching symbols in one module and Go standard library. Additional 20 package-import and 26 module dependency findings were not identified as called. Report stored locally at `/root/herza-security-tools/govulncheck-production.txt`.
- LG-KUL **actual deployed binary**: exit 3; 49 vulnerabilities with matching symbols in one module and Go standard library; additional 21 package-import and 26 module dependency findings not identified as called. Report at the same path on LG-KUL.
- Binary mode identifies matching vulnerable symbols, **not proof of reachable execution paths or exploitation**. The current scan database predates 22 November; update when a newer database becomes available. Other distinct production hashes require scans or carefully verified binary-equivalence analysis.

## Functional verification
- All ten hosts: HTTPS homepage 200 and NGINX and Looking Glass services active after NGINX patch. Read-only Ansible tests of `https://127.0.0.1/` returned HTTP 200 on ten hosts, local response times approximately 0.009–0.018 seconds. This is a local service test, not an external browser test.
- `/api/diagnostics` without required test parameters returns HTTP 400 on all ten hosts; LG-Cyber1 response `{"error":"unsupported test type"}`. This is not an end-to-end ping/traceroute/BGP acceptance test. No intrusive diagnostic test was run.

## Backup audit
- Each of ten hosts has a local `/root/herza-nginx-patch/<UTC timestamp>/config-backup.tar.gz` from NGINX patch. The LG-Cyber1 archive was successfully listed with `tar -tzf`.
- On LG-Cyber1 and Ansible controller, no NFS/CIFS backup mounts or active dedicated off-host backup service were found by the targeted checks. **Off-host backup has not been verified**; do not describe local backups as disaster-recovery ready.
- Prior NGINX backups include NGINX config, Let's Encrypt certs, and Looking Glass systemd unit. **They do not include the application binary or its complete `/etc/herza-lookingglass` configuration.** A full app rollback requires both and a tested restore process. Treat app config as sensitive; never commit secrets to GitHub.

## Remediation gates
1. Recover and authenticate the exact source for both `herza-lookingglass` and `herza-lookingglass-vyos` builds, including per-site differences. Do not infer that an inaccessible GitHub repo exists.
2. Prepare reproducible patched builds using a supported fixed Go toolchain and reviewed dependencies. Scan both source and binaries; inspect affected symbols and behavior.
3. Back up binary, configuration and credentials securely off-host, verify integrity and restoration. Prepare service rollback.
4. Run staging end-to-end ping/traceroute/BGP and browser tests, then one-host pilot; roll out gradually with per-host acceptance and rollback.
5. Re-scan every distinct binary build, verify NGINX and app health, and record versioned release evidence.
