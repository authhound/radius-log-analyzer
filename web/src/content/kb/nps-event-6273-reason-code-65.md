---
title: "NPS Event 6273 Reason Code 65 — denied by the account's dial-in 'Network Access Permission'"
description: "Why one user gets rejected by NPS while everyone else connects: the AD account's Dial-in tab set to Deny access, the msNPAllowDialin attribute, and the policy checkbox that overrides it."
format: nps
signature: "Event 6273 · Reason Code 65"
rule_id: nps/reason-code-65
---

**The official text:** *"The Network Access Permission setting in the dial-in properties of the user account in Active Directory is set to Deny access to the user."*

**What it actually means:** policies matched, credentials were fine — the user's own AD account said no. Every AD account carries a **Network Access Permission** setting on its Dial-in tab (a relic of the RAS dial-up era that still gates 802.1X and VPN through NPS), and this account has it set to **Deny access**. Per-account setting, so the classic symptom is **one user failing while everyone around them connects fine**.

## Fixes, from quickest to most durable

### 1. Fix the one account

Active Directory Users and Computers → the user → **Dial-in** tab → Network Access Permission → select **"Control access through NPS Network Policy"** (the sane default that delegates the decision to your policies). Done in ten seconds.

### 2. Find out how many accounts are affected

Deny flags rarely travel alone — they arrive in herds via migration scripts, old security baselines, or a well-meaning bulk edit years ago. The attribute behind the checkbox is `msNPAllowDialin` (`FALSE` = deny, `TRUE` = allow, *not set* = control through policy):

```powershell
Get-ADUser -Filter 'msNPAllowDialin -eq $false' -Properties msNPAllowDialin |
  Select-Object SamAccountName
```

If that returns hundreds of rows, clear the attribute in bulk (`Set-ADUser -Clear msNPAllowDialin`) rather than clicking through tabs.

### 3. Make NPS ignore the flag entirely

If you want network policies to be the single source of truth, open the matched network policy → Properties → check **"Ignore user account dial-in properties"**. With that set, stale per-account flags can never bite you again. (Requirement: the policy's grant/deny then does all the work — make sure your conditions are right before flipping it.)

## One caveat

Option 3 also ignores per-account *Allow* flags and callback settings — irrelevant in most modern networks, but worth knowing if you still run legacy VPN setups that depend on them.
