---
title: "RADIUS server not responding / timing out — the silent failures, in order of likelihood"
description: "When the NAS reports RADIUS timeouts instead of rejects: silent drops from unknown clients and bad secrets, dropped UDP fragments mid-EAP, dead services, and firewalled or wrong ports — with the checks for each."
format: general
signature: "RADIUS timeout / server not responding"
---

A *reject* means the RADIUS server said no. A *timeout* means it said **nothing** — and RADIUS servers say nothing deliberately more often than people expect. If your switch, AP, or controller reports "RADIUS server not responding," walk this list in order; it's sorted by real-world frequency.

## 1. The server is ignoring you on purpose (unknown client)

Both major servers **silently drop** packets from source IPs they don't recognize — answering unknown senders would be an amplification vector. No reject, no log on the NAS side, just a timeout.

- FreeRADIUS: `Ignoring request ... from unknown client` in `radiusd -X` — [full write-up](/kb/freeradius-unknown-client).
- NPS: nothing in the event log at all for unregistered clients (the deafening silence *is* the clue), or [event 6274 discards](/kb/nps-event-6274-discarded).

**Check:** does the server's client list contain the **exact source IP** the packets arrive from? Multi-homed NAS devices and NAT rewrite source addresses constantly.

## 2. The shared secret is wrong (also a silent drop)

An unverifiable packet is dropped, not rejected — FreeRADIUS logs [invalid Message-Authenticator](/kb/freeradius-invalid-message-authenticator), NPS logs [Reason 262](/kb/nps-event-6273-reason-code-262) or a 6274 discard. From the NAS side, identical to a dead server.

## 3. Authentication *starts*, then times out mid-handshake

If simple test requests work but real 802.1X clients hang partway, you're not timing out on the server — you're losing **large EAP-TLS packets** carrying the certificate chain. Firewalls dropping UDP fragments is the classic. Fingerprint: some device types connect fine, others stall forever; the server log shows retransmitted duplicates. [Full write-up with the fragment_size fix](/kb/freeradius-eap-fragmentation-timeouts).

## 4. The service is actually down (or deaf)

Now — only now — suspect the server itself:

- **Is it listening?** `ss -ulnp | grep 1812` on the RADIUS host. FreeRADIUS crashed-on-config-error is common after edits: run `radiusd -XC` to syntax-check. On Windows, check the NPS service and whether the server is still [registered in AD](/kb/nps-event-6274-discarded).
- **Right port?** Modern RADIUS is **1812/1813 (UDP)**; ancient gear defaults to 1645/1646. A NAS pointed at 1645 against a server listening on 1812 is a permanent timeout.
- **Firewall path:** UDP 1812 open *from the NAS's source IP* to the server — host firewall included (`iptables`/`nftables`/Windows Firewall). Test from the NAS's network segment, not from your desk: `radtest` or `nc -u <server> 1812` from a machine on the same VLAN as the NAS.

## 5. It responds, just too slowly

Backend latency (a slow LDAP/AD lookup, a hung SQL pool, an unreachable DC being retried) can push responses past the NAS's timeout — the server *did* answer; the NAS had given up. Server logs show successful (late) authentications the NAS never accepted; NAS-side timeout counters climb while server-side failure counters don't. Fix the backend latency, or as a bandage, raise the NAS's RADIUS timeout above 5 seconds and add retries.

## The one diagnostic that settles it

`tcpdump -ni any udp port 1812` **on the RADIUS server** while a client fails. Packets arriving + nothing leaving = causes 1–2 (read the server's own log for which). Packets arriving + replies leaving = the loss is on the return path or the NAS gave up (causes 3, 5). No packets arriving = network path or wrong port (cause 4). Ten seconds of capture beats an hour of guessing.
