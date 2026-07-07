---
title: "NPS Event 6273 Reason Code 262 — 'the signature was not verified' means shared secret"
description: "Why 'The supplied message is incomplete. The signature was not verified' in Windows NPS is a shared secret mismatch, and how to find which RADIUS client entry is at fault."
format: nps
signature: "Event 6273 · Reason Code 262"
rule_id: nps/reason-code-262
---

**The official text:** *"The supplied message is incomplete. The signature was not verified."*

**What it actually means:** the RADIUS packet from your switch, AP, or controller failed cryptographic verification. In practice this is a **shared secret mismatch** between the NAS and the RADIUS Client entry in NPS — the two ends are signing with different keys, so every message looks forged.

The critical thing to understand: the user's credentials were **never evaluated**. No amount of password resetting, account unlocking, or policy editing will change anything. This failure happens before authentication begins.

## Why it happens

### 1. The secrets genuinely differ

The classic: someone re-entered the secret on one side and typo'd it, or pasted it with a trailing space or newline. Because both fields are write-only blobs of dots, you cannot eyeball-compare them.

**Fix:** re-enter the secret on **both** sides — the NAS and the NPS RADIUS Clients entry. Type it, don't paste it. Keep it under 32 characters and alphanumeric if you want to rule out NAS firmware quirks with length or special characters.

### 2. The request is matching the wrong RADIUS Client entry

NPS picks the RADIUS Client entry by the packet's **source IP**. If the NAS's IP changed (DHCP renewal, controller failover, new management VLAN) or the packets arrive through NAT, the request can land on a *different* client entry — one with a different secret. The secret you keep re-checking is the right one; it's just not the one being used.

**Fix:** read the `Client IP Address` and `Client Friendly Name` fields in the 6273 event and confirm they name the entry you think they do.

### 3. Overlapping client entries

If you have both a subnet-range entry (say `10.20.0.0/24`) and a host entry (`10.20.0.12`) with different secrets, which one wins is not what you'd guess on every NPS version. Overlaps are a time bomb.

**Fix:** audit your RADIUS Clients list for overlapping addresses and keep exactly one entry per NAS.

## The FreeRADIUS equivalent

If you run FreeRADIUS too: the same failure shows up in `radiusd -X` as *"Received packet ... with invalid Message-Authenticator! (Shared secret is incorrect.)"* — same disease, different log line.
