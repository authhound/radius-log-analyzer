---
title: "FreeRADIUS rlm_ldap bind failures — admin bind vs user bind, and which one broke"
description: "Troubleshooting 'LDAP bind failed' errors in FreeRADIUS: wrong bind DN credentials, LDAPS/StartTLS certificate trust, unreachable directory, and the difference between the service bind and per-user authentication binds."
format: freeradius
signature: "rlm_ldap: ... bind failed / Bind credentials incorrect"
---

**The log lines** (varies by version):

```
rlm_ldap (ldap): Opening additional connection (0), 1 of 32 pending slots used
rlm_ldap (ldap): ERROR: Bind credentials incorrect: Invalid credentials
```

**What it actually means:** FreeRADIUS talks to your directory in two distinct ways, and "bind failed" means opposite things depending on **which bind** failed. Sorting that out first saves the afternoon:

## 1. The service (admin) bind failed → nobody can authenticate

At startup and when opening pool connections, `rlm_ldap` binds with the `identity`/`password` configured in `mods-available/ldap`. If *that* bind fails, every user fails, instantly, and the error appears when connections open — not per-request.

**Causes, in order:** the bind account's password changed or expired (service accounts with password-expiry policies are a classic); the DN is wrong (directories move OUs; `CN=svc-radius,OU=Old,...` stops resolving); the account got locked by something else spraying bad binds.

**Check:** take the exact `identity` DN and password from the config and test them directly:

```
ldapsearch -H ldaps://dc01.corp.example.com -D "CN=svc-radius,OU=Service,DC=corp,DC=example,DC=com" -W -b "" -s base
```

If that fails, FreeRADIUS is off the hook — fix the account/DN and you're done.

## 2. A per-user bind failed → that's just a wrong password

If your setup authenticates by re-binding as the user (`Auth-Type LDAP` / edir-style), a failed bind for `uid=jane,...` is the LDAP way of saying "wrong password" — normal reject traffic, not an infrastructure problem. Don't chase the module config for these; treat them like any credential failure.

**Tell them apart:** service-bind failures mention your configured admin DN and appear at connection setup; user-bind failures mention the end user's DN inside a request's `authenticate` section.

## 3. The bind never happened — transport problems

`Connection refused`, timeouts, or TLS errors before any bind:

- **LDAPS/StartTLS trust:** the directory's certificate isn't trusted by the CA bundle FreeRADIUS uses (`tls { ca_file }` in the ldap module). After a DC certificate renewal this breaks silently. Test with `openssl s_client -connect dc01:636 -CAfile <your ca_file>`.
- **Firewall/port:** 389 (StartTLS) vs 636 (LDAPS) — confirm which one the config uses and that it's open *from the RADIUS host*.
- **A DC was decommissioned** and the config still points at it by name or IP. Prefer a load-balanced or multi-server `server` line.

## A note on MSCHAPv2 + LDAP

If binds succeed but PEAP users still fail, you've likely hit the hash-format wall — LDAP directories storing SSHA can't back MSCHAPv2. That's a different failure with its own page: [MS-CHAP2-Response is incorrect](/kb/freeradius-mschap2-response-incorrect).
