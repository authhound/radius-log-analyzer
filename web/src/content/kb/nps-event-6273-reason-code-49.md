---
title: "NPS Event 6273 Reason Code 49 — no connection request policy matched"
description: "What it means when a RADIUS request fails to match any NPS connection request policy: deleted default CRPs, day/time and NAS conditions, and proxy realm rules."
format: nps
signature: "Event 6273 · Reason Code 49"
rule_id: nps/reason-code-49
---

**The official text:** *"The connection attempt did not match any connection request policy."*

**What it actually means:** the request was rejected one layer **earlier** than most people look. NPS processes every request through Connection Request Policies (CRPs) *before* network policies or credentials come into play. CRPs decide "does this server handle this request at all, and how" — they filter on how and where the request arrived (NAS address, port type, day and time, username realm), never on group membership or passwords.

Reason 49 therefore means: no CRP claimed the request. Your network policies were never consulted.

## The usual causes

### 1. The default CRP is gone or disabled

A fresh NPS install ships with **"Use Windows authentication for all users"** — a CRP with no conditions that matches everything, 24/7. It's routinely deleted during cleanups or lost in config imports. If your CRP list has no enabled catch-all and your specific CRPs don't cover this request, everything falls through.

**Check:** NPS console → Policies → Connection Request Policies. Confirm at least one enabled policy whose conditions cover this request.

### 2. A condition is narrower than reality

CRP templates often include a **Day and time restriction** or a specific **NAS IPv4 Address**. A request at 2 AM against a business-hours CRP, or from a controller whose IP changed, matches nothing.

**Check:** compare the event's `NAS IPv4 Address` and timestamp with each CRP's conditions, top to bottom — first match wins, same as network policies.

### 3. Proxy/realm routing doesn't cover this username

In RADIUS proxy setups, CRPs route by realm (the `@corp.example.com` suffix or `CORP\` prefix). A user with an unexpected suffix — or no suffix — falls through the realm-matching CRPs. The `User Name` field of the event shows the raw string that was matched against.

**Fix:** add a CRP (or realm-stripping rule) for the missing case, or fix the identity format on the client.

## How to tell 49 from 48 at a glance

Both mean "nothing matched," one layer apart: **49 = connection request policy** (where/how the request arrived), **48 = network policy** (who is allowed). If you're seeing 49, don't touch group conditions or network policy order — the request never got that far.
