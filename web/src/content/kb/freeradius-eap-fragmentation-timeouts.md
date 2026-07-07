---
title: "FreeRADIUS EAP stalls mid-handshake — 'Discarding duplicate request' and the fragment-size fix"
description: "Why 802.1X authentications hang partway through the TLS handshake: oversized RADIUS packets carrying certificate fragments get dropped, the NAS retransmits, and the client never answers. Diagnosis and the fragment_size fix."
format: freeradius
signature: "Discarding duplicate request from client ..."
rule_id: freeradius/eap-fragment-loss
---

**The log lines:**

```
(12) Discarding duplicate request from client wlc-01 port 32768 - ID: 118 ...
(12) Sending duplicate reply to client wlc-01 port 32768 - ID: 118
```

**What it actually means:** the conversation died in the middle. The server sent an Access-Challenge (usually carrying a chunk of its TLS certificate chain) and is now seeing the **same request retransmitted** — the NAS never got an answer the client acted on. The usual physics: your server certificate chain is several KB, EAP splits it across multiple large RADIUS packets, one of them exceeds the path MTU, gets fragmented at the IP layer, and **a firewall drops the fragments**. The client is waiting for bytes that will never arrive; the NAS retries; everything times out.

This is the classic "some devices authenticate, others hang forever, nothing is *rejected*" failure — rejects are instant, fragment loss is silent.

## The fix that works most of the time

In `mods-available/eap`, inside the `tls-config` section:

```
fragment_size = 1024
```

This caps each EAP-TLS fragment so the resulting RADIUS packet stays comfortably under a 1500-byte MTU even with headers and attributes. Restart radiusd and retest. It costs a couple of extra round trips and fixes the problem outright in most environments.

While you're there, **trim the chain**: `certificate_file` should contain the server cert plus intermediates only — no root (clients must already have it), no leftover old certs. Every KB removed is a fragment that can't be dropped.

## Confirming it's really fragment loss

- `tcpdump -i any -n udp port 1812 and host <nas-ip>` on the RADIUS server: you'll see large Access-Challenge packets leave (look for IP fragments) and the same Access-Request arrive again and again.
- If you can capture on the NAS side too, watch where the large packets vanish — that hop is your fragment-dropping firewall or tunnel with a reduced MTU (IPsec and GRE paths are frequent offenders).

## The other cause of this exact log pattern

Sometimes nothing is dropped — the **supplicant is waiting on a human**. A certificate-trust prompt is sitting on the user's screen; until someone clicks, the client answers nothing and the NAS retransmits identically. Watch a failing device while the retransmissions happen (Windows: WLAN-AutoConfig event log). If a trust dialog is the blocker, your real issue is certificate trust distribution, not MTU — fix the profile so the prompt never appears.
