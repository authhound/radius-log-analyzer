---
title: "NPS Event 6273 Reason Code 8 — 'the user account does not exist' (but it does)"
description: "Why Windows NPS says the specified user account does not exist even when you can see it in AD: username format mismatches, host/ machine identities, and cross-domain lookups."
format: nps
signature: "Event 6273 · Reason Code 8"
rule_id: nps/reason-code-8
---

**The official text:** *"The specified user account does not exist."*

**What it actually means:** NPS took the identity string the device sent, tried to resolve it to an account, and found nothing. The password was never checked — the lookup itself failed. Nine times out of ten the account *does* exist; the device is just naming it in a way NPS can't resolve.

## Read the event before touching AD

The `Account Name` field in the 6273 event shows the exact string NPS searched for. Compare it, character by character, with what you expect. The usual mismatches:

### 1. Username format

`jsmith` vs `CORP\jsmith` vs `jsmith@corp.example.com` are three different lookups. A bare name relies on default-domain assumptions; a UPN must match the UPN attribute (which is not always `sam@domain`); and users typing their email address works only when the email happens to equal the UPN. A stray trailing space — invisible in most device UIs — also lands here.

**Fix:** re-enter the username on the device in the format that your working clients use.

### 2. It's a machine identity, not a user

If `Account Name` starts with `host/` (e.g. `host/LAPTOP-042.corp.example.com`), the device performed **computer authentication**. Reason 8 then means that computer object doesn't exist in the domain — typical for BYOD devices, machines re-imaged outside the domain, or workgroup laptops copying a corporate Wi-Fi profile.

**Fix:** either join the machine to the domain, or switch the profile to user authentication.

### 3. The account lives in another domain or forest

NPS resolves accounts through its own domain's view of the world. Users from a child domain, a trusted forest, or a recently migrated OU fail if the trust path is broken or the name lacks its domain qualifier.

**Fix:** check the domain prefix in the event and verify the trust between that domain and the NPS server's domain (`nltest /trusted_domains` on the NPS box is a quick start).

## The pattern worth noticing

One user failing → format typo on their device. A whole class of devices failing (all BYOD, all one department) → machine-auth or domain-scope problem. That grouping tells you where to look before you open a single console.
