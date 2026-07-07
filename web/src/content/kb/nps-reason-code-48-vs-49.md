---
title: "NPS Reason Code 48 vs 49 — which policy layer rejected you, and why it matters"
description: "The two 'no policy matched' rejects in Windows NPS compared: connection request policies vs network policies, how to tell them apart in seconds, and which console tab to open for each."
format: nps
signature: "Event 6273 · Reason Codes 48 & 49"
---

Both codes say "nothing matched," and they get confused for each other constantly — but they point at **different policy layers**, and troubleshooting the wrong layer wastes hours. NPS processes every request through two gates, in order:

```
RADIUS request
   │
   ▼
1. Connection Request Policies (CRPs)   → none matched? ── Reason 49
   "does this server handle this request, and how?"
   conditions: NAS address, port type, day/time, username realm
   │
   ▼
2. Network Policies                     → none matched? ── Reason 48
   "is this user/device allowed on the network?"
   conditions: groups, auth methods, NAS port type, ...
```

## Reason 49 — rejected at the front door

No **Connection Request Policy** claimed the request. Credentials, group memberships, and network policies were never consulted — don't touch them.

Look at *how the request arrived*: NAS IP, port type, time of day, username realm. The most common cause is the default catch-all CRP ("Use Windows authentication for all users") having been deleted, disabled, or reordered during a cleanup. In proxy setups, it's a realm that no forwarding rule covers.

**Full write-up: [Reason Code 49](/kb/nps-event-6273-reason-code-49)**

## Reason 48 — rejected at the second door

A CRP matched fine; no **Network Policy** did. Now it *is* about who is connecting: the account isn't in the policy's AD group (or is the wrong object type — computer vs user), a condition like `NAS Port Type = Wireless` doesn't match what the NAS sends, or a broader policy higher in the list swallowed the request.

**Full write-up: [Reason Code 48](/kb/nps-event-6273-reason-code-48)**

## Tell them apart in the event itself

Open the 6273 event and read the Authentication Details:

| Field | Reason 49 | Reason 48 |
|---|---|---|
| Connection Request Policy Name | `-` (nothing matched) | a real policy name |
| Network Policy Name | `-` (never evaluated) | `-` (evaluated, none matched) |

That first field is the discriminator: **if a CRP name is present, you're past layer 1 — work the network policies.** If it's `-`, stay at the CRP layer no matter how suspicious your group memberships look.

## One habit that prevents both

After any policy change, take one known-good and one known-bad request and walk the two lists top-to-bottom asking "would this match?" — NPS is strictly first-match-wins at both layers, and almost every 48/49 mystery is an ordering or condition literalism problem, not a Windows bug.
