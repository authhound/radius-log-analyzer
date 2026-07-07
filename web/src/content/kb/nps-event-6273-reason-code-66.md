---
title: "NPS Event 6273 Reason Code 66 — auth method not enabled on the matching policy"
description: "Decoding 'the user attempted to use an authentication method that is not enabled on the matching network policy': EAP mismatches mid-migration, non-EAP fallbacks, and requests matching the wrong policy."
format: nps
signature: "Event 6273 · Reason Code 66"
rule_id: nps/reason-code-66
---

**The official text:** *"The user attempted to use an authentication method that is not enabled on the matching network policy."*

**What it actually means:** a network policy matched the request (so conditions, groups, and scoping are fine), but the *how* disagrees: the client wants to authenticate with a method the policy doesn't offer. Client and server are both healthy; they just haven't agreed on a protocol. Two fields in the event — `Authentication Type` and `EAP Type` — versus the policy's Constraints tab tell the entire story.

## The three flavors

### 1. EAP method mismatch (the migration classic)

Device asks for PEAP-MSCHAPv2, policy only lists EAP-TLS — or the reverse. This is the standard mid-migration failure when moving from password-based PEAP to certificate-based EAP-TLS: devices not yet re-profiled keep asking for the old method.

**Fix:** compare the event's `EAP Type` with the policy's EAP list. During migration, either run both methods on the policy temporarily, or (cleaner) separate policies with group conditions for migrated vs unmigrated devices.

### 2. The client fell back to a non-EAP method

If `Authentication Type` shows `MS-CHAPv2`, `CHAP`, or `PAP` rather than EAP/PEAP, the client isn't doing 802.1X EAP at all — typical for VPN appliances, older switch fallbacks, and test tools. Those checkboxes ("less secure authentication methods") are deliberately off in modern policies.

**Fix:** decide consciously. Enabling PAP/CHAP to make the error go away weakens the whole policy; usually the right move is fixing the client to do EAP, or giving the legacy device its own tightly-scoped policy.

### 3. The request matched a *different* policy than you think

You added the method to the policy you always edit — and the error persists, because the request matches an earlier, stricter policy. NPS policy matching is first-match-wins.

**Check:** the event's `Network Policy Name` field names the policy that actually matched. If it isn't the one you've been editing, fix conditions or ordering — don't loosen the wrong policy's methods.

## Quick sanity path

Read `Network Policy Name` → open that exact policy → Constraints → Authentication Methods → compare with the event's `Authentication Type`/`EAP Type`. That's the whole diagnosis; the only decisions are on which side to change.
