---
title: "FreeRADIUS 'Ignoring request from unknown client' — why the NAS times out instead of failing"
description: "What it means when radiusd -X shows 'Ignoring request to auth address from unknown client': the missing clients.conf entry, multi-homed NAS source IPs, and why the client side only sees timeouts."
format: freeradius
signature: "Ignoring request to auth address ... from unknown client"
rule_id: freeradius/unknown-client
---

**The log line:**

```
Ignoring request to auth address * port 1812 bound to server default
  from unknown client 192.168.10.99 port 49152 proto udp
```

**What it actually means:** the packet reached FreeRADIUS, but no `client {}` block in `clients.conf` covers its source IP — so the server dropped it **without any reply**. That silence is the defining symptom: the NAS doesn't get a reject, it gets nothing, retries, and eventually reports a *timeout* or "RADIUS server not responding." Users just see Wi-Fi that spins forever.

If you're debugging "the RADIUS server isn't responding" from the NAS side, this log line is the first thing to look for — the server is responding fine to everyone it knows; it's deliberately ignoring this one address.

## The three causes

### 1. There's simply no client entry yet

New access point, new switch, or a NAS whose management IP changed. Add a block with the **exact source IP from the log line** (not the IP you think the NAS has):

```
client wlc-02 {
    ipaddr = 192.168.10.99
    secret = <shared secret>
}
```

Then restart radiusd (or send SIGHUP) — clients.conf is read at startup, not per-request.

### 2. The NAS sends from a different interface than you whitelisted

Multi-homed controllers and switches pick a source interface by routing table, which is often the uplink IP rather than the management IP you configured. NAT between NAS and server does the same thing wholesale.

**Check:** compare the source IP in the log line with the address in your client entry. On the NAS, look for a *"RADIUS source interface"* setting and pin it to the management interface — that's the durable fix; whitelisting whatever IP happens to show up is the fragile one.

### 3. The entry exists but isn't loaded

Config never reloaded after the edit, the entry sits in a file that isn't included, or it belongs to a virtual server that isn't listening on this address.

**Check:** `radiusd -X` prints the full client list at startup — search it for your IP. If it's missing there, the config on disk isn't what the server is running.

## Note for wide networks

If you have many NAS devices on one subnet, a range entry (`ipaddr = 192.168.10.0/24`) beats twenty host entries — but give the whole range **one** secret and never let a broader range overlap a host entry with a different secret, or you'll trade this error for shared-secret mismatches that are far harder to spot.
