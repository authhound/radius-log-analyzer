---
title: "FreeRADIUS 'Login incorrect' — how to read radiusd -X to find the real reason"
description: "'Login incorrect' is the summary line, not the diagnosis. How to walk a FreeRADIUS debug log backwards from the reject to the module that actually failed, with the patterns to grep for."
format: freeradius
signature: "Login incorrect (...): [user] (from client ...)"
rule_id: freeradius/login-incorrect-generic
---

**The log line:**

```
(8) Login incorrect (mschap: ERROR: MS-CHAP2-Response is incorrect): [alice] (from client office-ap port 0 via TLS tunnel)
```

**What it actually means:** almost nothing by itself — and that's the point of this page. `Login incorrect` is FreeRADIUS's *summary* line, printed for every deliberate reject regardless of cause. Treating it as the diagnosis is the most common dead end in FreeRADIUS troubleshooting. The actual reason is always earlier in the debug output; here's how to walk to it.

## 1. Read the parentheses first

Since FreeRADIUS 3, the summary line includes the failing module's own message in parentheses — `(mschap: ERROR: ...)`, `(pap: ...)`, `(eap_peap: ...)`, `(rlm_ldap: ...)`. That names the module and often the whole answer. Frequent parentheticals and where they lead:

| Parenthetical contains | Real problem | Deep dive |
|---|---|---|
| `mschap: ... MS-CHAP2-Response is incorrect` | wrong password / hash format / AD account state | [MSCHAPv2 page](/kb/freeradius-mschap2-response-incorrect) |
| `No "known good" password` | backend never supplied a password | [known-good password page](/kb/freeradius-no-known-good-password) |
| `eap: ... EAP sub-module failed` | look further up for the TLS alert | [unknown CA](/kb/freeradius-tls-alert-unknown-ca), [expired cert](/kb/freeradius-certificate-expired) |
| `rlm_ldap` / `sql` errors | backend lookup or bind failure | [LDAP bind page](/kb/freeradius-ldap-bind-failed) |

## 2. Walk up to the last `authenticate` section

Find the request number — the `(8)` prefix — and read *only* lines with that prefix (in a busy debug, interleaved requests are the great confuser; `grep '^(8)'` is your friend). Locate the `authenticate {` block and read which module returned `reject` or `fail`:

```
(8)     [mschap] = reject
(8)   } # authenticate = reject
```

Everything above that module's own ERROR lines is context; the first ERROR from that module is the diagnosis.

## 3. If no module errored — it's your policy

Sometimes nothing "failed": an unlang policy deliberately rejected (a group check, a NAS filter, a time restriction). Search the request's lines for `Post-Auth-Type REJECT` and for `reject` inside `policy` / `authorize` sections — the surrounding lines name the condition that fired.

## The habit that makes this fast

Never debug from `/var/log/radius/radius.log` — its one-line summaries are exactly the trap this page describes. Reproduce the failure once under `radiusd -X` (or `raddebug` on a running server) and you get every decision the server made, in order. Then the reject is never mysterious: it's the last module that said no, for the reason printed right above it.
