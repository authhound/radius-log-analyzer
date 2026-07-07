---
title: "FreeRADIUS 'TLS Alert read:fatal:unknown CA' — the client doesn't trust your server cert"
description: "What the unknown CA alert in radiusd -X debug output means, why the client (not the server) is rejecting the handshake, and the three trust problems that cause it."
format: freeradius
signature: "eap_peap: ERROR: TLS Alert read:fatal:unknown CA"
rule_id: freeradius/tls-unknown-ca
---

**The log line:**

```
(6) eap_peap: ERROR: TLS Alert read:fatal:unknown CA
(6) eap_peap: TLS_accept: Error in error
(6) eap_peap: ERROR: Failed in __FUNCTION__ (SSL_read): ... tlsv1 alert unknown ca
```

**What it actually means:** read the direction carefully — the server *read* the alert. The **client device sent it**. Your supplicant (the phone, the laptop, the printer) looked at the certificate your RADIUS server presented during the PEAP/EAP-TLS handshake and said: *"I don't trust whoever signed this."* Then it hung up.

The server never got far enough to see a username or password. Nothing on the inner-auth side (users, passwords, LDAP, AD) is involved.

## Three causes, in order of likelihood

### 1. The client doesn't trust your CA

The device is missing the CA certificate that signed the RADIUS server cert — or has it, but the 802.1X profile doesn't reference it. This is the standard failure after:

- switching from the FreeRADIUS default snake-oil certs to real ones,
- migrating to a new internal CA,
- switching from a public CA to a private one (or the reverse).

Note that on Windows and Android, having the CA in the system store is **not** sufficient — the Wi-Fi profile itself pins which CAs are acceptable. iOS asks the user to trust the cert on first connect; users tap "cancel" more often than you'd hope, which also produces this alert.

**Check:** on one failing device, confirm the CA cert is installed *and* selected in the network profile.

### 2. Your server sends an incomplete chain

If your server cert was issued by an intermediate CA and the server only presents the leaf, clients that trust the root but have never seen the intermediate can't build a chain — "unknown CA."

**Check:** `certificate_file` in `mods-available/eap` must contain the server certificate **followed by all intermediates** (PEM concatenated, leaf first). Verify what's actually presented with `eapol_test` or `openssl s_client`.

### 3. The name changed at renewal

Profiles that validate the server name (as they should) will reject a renewed certificate whose CN/SAN no longer matches the name the profile expects — some supplicants report this as unknown CA rather than a name mismatch.

**Check:** compare the certificate's `subjectAltName` with the server name configured in the client profiles.

## Why "just disable validation" is the wrong fix

Turning off server certificate validation on clients makes the error go away and turns your network into an evil-twin honeypot's dream. The certificate check is the *only* thing telling the client it's talking to your RADIUS server and not an attacker's. Fix the trust chain instead.
