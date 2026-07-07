---
title: "FreeRADIUS 'invalid Message-Authenticator! (Shared secret is incorrect.)' explained"
description: "Why FreeRADIUS drops packets with an invalid Message-Authenticator, how to find which clients.conf entry is actually being used, and the NAT trap that keeps this error alive after you've fixed the secret."
format: freeradius
signature: "Received packet ... with invalid Message-Authenticator! (Shared secret is incorrect.)"
rule_id: freeradius/shared-secret-mismatch
---

**The log line:**

```
Dropping packet without response because of error:
Received packet from 192.168.10.5 with invalid Message-Authenticator!
(Shared secret is incorrect.)
```

**What it actually means:** the RADIUS request failed its integrity check. The NAS (switch/AP/controller) signed the packet with one shared secret; FreeRADIUS verified it with another. They disagree, so the server drops the packet **without replying at all** — which is why the NAS side shows timeouts, not rejects, and why clients just spin.

Nothing about the user, their password, EAP, or your policies is involved. This failure happens before all of that.

## The three ways this happens

### 1. The secrets are actually different

Someone typo'd, someone pasted with a trailing newline, or one side still has last year's secret. Both input fields show dots, so you can't compare by looking.

**Fix:** re-type (never paste) the secret on the NAS *and* in the matching `client {}` block in `clients.conf`, restart or `SIGHUP` radiusd, retest. If the NAS firmware is old, keep the secret under 32 characters — some devices truncate silently.

### 2. The packet matched a different client block than you think

FreeRADIUS selects the `client {}` entry by the packet's source IP, and more-specific definitions coexist with broad ones. A stray `client 0.0.0.0/0` (often added "temporarily" for testing) or an overlapping subnet entry means the secret being checked isn't from the block you keep editing.

**Fix:** run `radiusd -X`, find the incoming packet, and read which client FreeRADIUS says it is. Then grep `clients.conf` (and `sites-enabled/`, and any SQL-backed client store) for overlapping definitions.

### 3. NAT is rewriting the source address

If the NAS sits behind NAT or a load balancer, packets arrive from the translated address — which matches a different (or no) client entry. You fixed the right secret on the right client block; the packets just aren't attributed to it.

**Fix:** compare the source IP in the debug line with the NAS's real address. If they differ, add a client entry for the NAT address, or fix routing so RADIUS traffic keeps its source.

## The NPS equivalent

Windows NPS reports the identical disease as **Event 6273, Reason Code 262** — *"The supplied message is incomplete. The signature was not verified."* Same fix, different console.
