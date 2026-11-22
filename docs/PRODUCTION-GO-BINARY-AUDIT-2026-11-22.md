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

## Follow-up — recovered local source and isolated candidate builds

On LG-Cyber1, a local source tree was recovered at `/root/herza-lookingglass-src` (module `herza-lookingglass`, x/crypto v0.31.0). On LG-KUL, a local source tree was recovered at `/root/herza-lookingglass-vyos` (module `github.com/herzadigital/herza-lookingglass-vyos`, x/crypto v0.24.0). Both pass `go test ./...` with no test files; exact correspondence to installed binary source revisions has **not** been proven. No repository provenance or Git commit is available in these source directories.

Local root-only full-application recovery archives were created and integrity-tested on the two pilot hosts (including deployed binary, `/etc/herza-lookingglass`, systemd unit and local source):
- LG-Cyber1: `/root/herza-app-recovery/20261122T024836Z/app-recovery.tar.gz` (approximately 28 MB)
- LG-KUL: `/root/herza-app-recovery/20261122T024844Z/app-recovery.tar.gz` (approximately 9.3 MB)
These archives **contain credentials/secrets** and must not be published to GitHub or moved to an untrusted destination. No off-host backup or restore test has yet been confirmed.

Isolated staging copies, not active production binary paths, were built using **Go 1.26.11** and upgraded `golang.org/x/crypto v0.59.0`, `golang.org/x/sys v0.50.0`. On both pilot hosts, `go test ./...`, `go vet ./...`, and `go build` passed. Binary-mode `govulncheck v1.10.0` returned **No vulnerabilities found** on both patched candidates using its database updated 2026-11-11. Candidate paths: `/root/herza-security-staging/cyber1-patched` and `/root/herza-security-staging/kul-patched`. Neither was deployed or run as the production service. This scanner result is not a guarantee against all vulnerabilities.

The original source rebuilt with Go 1.26.11 but without updating x/crypto still reported 11 matching-symbol findings in both variants, demonstrating that upgrading the Go toolchain alone was insufficient. Application-level regression and functional tests are still missing; source revision parity, complete credential-safe off-host backup, and tested rollback remain **release gates**.

## Follow-up validation — 22 November 2026

Read-only local API smoke checks of the **currently deployed** service:
- LG-Cyber1: `/` 200, `/api/client-ip` 200, `/api/diagnostics` with no parameters 400; `/api/health` and `/api/lookup` 404 (not routes of this Juniper variant).
- LG-KUL: `/` 200, `/api/health` 200, `/api/client-ip` 200, `/api/diagnostics` without parameters 400, `/api/lookup` without parameters 400 (VyOS variant).
- This is **not** a functional acceptance of the new candidate builds or real network diagnostics. Do not treat the 404s on Juniper as service failures when routes are absent by design.
- Both pilot app recovery archives were re-hashed locally: Cyber1 `7b7a12440a84ca2613c8c2967872aac1a7d71ba044ba5b9e0f9131927857210d`; KUL `5ab531766e48cd5e62f0ae99f74cb76b79fe653745e8f0a7a49ab8e6eabdee18`. **Neither is off-host**. Source copies and patched candidates remain on the respective production host in separate staging directories; neither is deployed.
- Exact production source-to-binary revision parity remains unproven. Source trees have no local Git history. Do not assign v1.4.5 to these production binaries.

**Release blocker:** an approved encrypted off-host backup target, restore test, and end-to-end staging validation are still needed before changing production binary. No secrets or recovery archives have been committed to GitHub.

## Production Go security pilot — 22 November 2026 (local-backup exception)

User approved **local backup first** instead of off-host backup. This is an accepted temporary risk, **not** off-host disaster recovery or tested restore.

Two pilot production services have been updated to isolated Go 1.26.11 / x/crypto v0.59.0 candidate binaries after local archive checksum validation, canary HTTP checks, and pre-change binary copy:
- **LG-Cyber1:** deployed binary SHA-256 `1df7495995b97c4b60927bc4892d5552dc0681078938d36eec266ab43fb52746`. Canary on loopback port 18080 returned `/` 200, `/api/client-ip` 200, `/api/diagnostics` without parameters 400. After atomic binary replacement and systemd restart, production `/` 200 and `/api/client-ip` 200, services active. Real `ping` API to `1.1.1.1` returned successful replies; invalid option-like target `-I` returned HTTP 400. Rollback binary: `/root/herza-app-recovery/20261122T024836Z/binary-before-pilot`.
- **LG-KUL:** deployed binary SHA-256 `240187461d08896aa54e3b297a9793c0a6282b633aaf32da6fdda2915bb41cd1`. Canary loopback port 18080 returned `/api/health` 200, `/api/client-ip` 200, `/api/diagnostics` without parameters 400; bare `/` 404 at the Go backend, whereas NGINX serves the public homepage as HTTP 200. After atomic binary replacement and systemd restart, production homepage, health, client-IP endpoints and service checks passed. Rollback binary: `/root/herza-app-recovery/20261122T024844Z/binary-before-pilot`.
- **No rollback was actually exercised**; rollback binaries are retained. No full BGP authentication/diagnostic acceptance has been completed. LG-Cyber1 historical journal showed TELKOM BGP SSH authentication failure on Nov 18, prior to this pilot; this needs separate investigation.

Other eight production hosts remain on Go 1.22.2. Source tree hash checks show exact source code parity between Cyber1, Kalibata and Tifa; KUL and MNL also match. Cibitung, SUB1/SUB2, NCIX have distinct source/build variants requiring per-host rebuild and functional checks; SG binary hash equals SUB1/SUB2 but local source was not found. **Do not label all ten as patched or publish v1.4.5 Stable yet.**
