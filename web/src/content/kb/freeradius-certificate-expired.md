---
title: "FreeRADIUS 'certificate expired' TLS alert — server cert, client cert, or a wrong clock"
description: "Diagnosing expired-certificate failures in radiusd -X: everyone failing at once means the server certificate; one device means a client cert or its clock. Commands to check both."
format: freeradius
signature: "TLS Alert read:fatal:certificate expired"
rule_id: freeradius/tls-cert-expired
---

**The log line:**

```
(4) eap_tls: ERROR: TLS Alert read:fatal:certificate expired
```

**What it actually means:** a certificate in the EAP-TLS/PEAP handshake is past its `notAfter` date, and one side refused to continue. This failure has the most recognizable fingerprint in all of RADIUS troubleshooting: **it worked yesterday, and today it fails for everyone at the same minute** — because certificates expire on a schedule, not gradually.

## Who rejected whom?

The alert direction narrows it immediately:

- **`TLS Alert read:...`** — the server *read* the alert; the **client sent it**. The client is rejecting **your server certificate**. Expect every device to fail at once.
- **`TLS Alert write:...`** (or a verify error like `certificate verify failed`) — the **server** is rejecting the **client's certificate** (EAP-TLS). Expect one user or one batch of devices (issued together, expiring together).

## The three causes

### 1. The RADIUS server certificate expired

**Check right now:**

```
openssl x509 -enddate -noout -in /etc/freeradius/certs/server.pem
```

(Use the path from `certificate_file` in `mods-available/eap`.) If it's expired: renew, deploy, restart radiusd. Then put the date in your calendar for next time — RADIUS server certs are the single most-forgotten expiry in SMB networks because nothing warns you until the morning it breaks.

### 2. A client certificate expired (EAP-TLS)

One device failing while the fleet works — or a whole cohort failing together if they were enrolled the same week. Check the certificate on the failing device and reissue through whatever enrollment path you use (AD CS autoenrollment, SCEP/Intune, manual). While you're there, check how the *rest* of the fleet's expiry dates cluster.

### 3. Nothing is expired — a clock is wrong

A device that sat in a drawer for months, a VM with broken NTP, or a factory-reset appliance thinks it's 2020 (or 2038) and judges a perfectly valid certificate as outside its validity window.

**Check:** the date/time on the failing device, and on the RADIUS server itself, before reissuing anything. It takes ten seconds and regularly saves an afternoon of unnecessary CA work.
