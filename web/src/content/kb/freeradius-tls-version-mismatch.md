---
title: "FreeRADIUS TLS handshake failures with old devices — tls_min_version and legacy supplicants"
description: "Diagnosing generic TLS handshake failures in radiusd -X: printers and embedded devices stuck on TLS 1.0/1.1 against servers enforcing TLS 1.2+, cipher mismatches, and users cancelling trust prompts."
format: freeradius
signature: "eap_tls: ERROR: TLS_accept: Error in error / SSL_read failed"
rule_id: freeradius/tls-handshake-failed
---

**The log lines:**

```
(5) eap_tls: ERROR: TLS_accept: Error in error
(5) eap_tls: ERROR: Failed in __FUNCTION__ (SSL_read)
(5) eap_tls: ERROR: TLS receive handshake failed during operation
```

**What it actually means:** the TLS handshake broke off without a specific alert like "unknown CA" or "certificate expired" (if you have one of those, follow that trail instead — it's more precise). A generic handshake failure means the two ends **couldn't even agree on how to talk**: protocol version, cipher suite, or the client simply walked away mid-conversation.

## Cause 1: TLS version mismatch (the printer problem)

Modern FreeRADIUS builds and hardened configs set `tls_min_version = "1.2"` — correct for laptops and phones, fatal for the long tail of embedded devices: older printers, scanners, barcode readers, medical devices, and door controllers whose supplicants top out at TLS 1.0 or 1.1. The fingerprint: **all your computers work; one category of device never connects.**

**Check:** in `mods-available/eap`, look at `tls_min_version` (and `tls_max_version`). Then find what the device supports — its 802.1X or "EAP" settings page, or a packet capture showing its ClientHello version.

**Fix, in order of preference:** firmware-update the device (many gained TLS 1.2 late in life); if not possible, weigh lowering `tls_min_version = "1.0"` — that weakens the setting for *everyone* on that virtual server, so consider giving legacy devices their own SSID/virtual server with relaxed TLS and tighter network segmentation instead. OpenSSL itself may also need `CipherString = DEFAULT@SECLEVEL=1` on newer distros to allow TLS 1.0 at all.

## Cause 2: the user cancelled the trust prompt

On unmanaged devices (BYOD, iOS especially), connecting shows a "trust this certificate?" dialog. Cancel it and the supplicant aborts the handshake — the server just sees the conversation die. The fingerprint: **intermittent, user-shaped failures** — the same device works when a patient person taps Trust.

**Check:** reproduce on a device you can watch. The durable fix is distributing a proper Wi-Fi profile (configurator/MDM) so no prompt appears.

## Cause 3: no common cipher suite

Rarer, same shape as cause 1: the ClientHello offers only ciphers the server's `cipher_list` (or the distro's OpenSSL policy) refuses. The debug lines just above the failure show the negotiation details.

**Check:** `cipher_list` in the eap TLS config, and your distro's system-wide crypto policy (`update-crypto-policies --show` on RHEL-family) which can override what FreeRADIUS asks for.
