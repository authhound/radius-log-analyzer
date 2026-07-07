---
title: "NPS Event 6273 Reason Code 22 — 'EAP type cannot be processed' (check the server cert first)"
description: "Why NPS rejects clients with 'the EAP type cannot be processed by the server' — expired NPS certificates, PEAP vs EAP-TLS mismatches, and corrupted EAP policy config."
format: nps
signature: "Event 6273 · Reason Code 22"
rule_id: nps/reason-code-22
---

**The official text:** *"The client could not be authenticated because the Extensible Authentication Protocol (EAP) Type cannot be processed by the server."*

**What it actually means:** the client proposed an EAP method that NPS couldn't run *for this connection*. The wording blames the client's choice of EAP type — but the most common real cause is on the server: **if the NPS certificate is broken, PEAP can't initialize at all**, and every client gets Reason 22 no matter what it asks for.

## Triage by blast radius

**Everyone failing at once?** Start with the NPS server certificate:

- Open the network policy → Constraints → Authentication Methods → select PEAP (or EAP-TLS) → **Edit**. If the dialog errors out or shows no certificate, you've found it.
- On the NPS server, open `certlm.msc` → Personal. You need a non-expired certificate with the **Server Authentication** EKU and its private key present (the little key icon). Certificates renewed by autoenrollment sometimes drop out of the policy selection — reselect and save.

This exact failure has a famous schedule: it fires the morning after the NPS cert quietly expires, breaking Wi-Fi for the whole org at once.

**One device class failing?** Then it's a genuine method mismatch:

- The event's `EAP Type` field shows what the client asked for. Compare it with the EAP types listed in the matched policy's constraints. Classic case: devices provisioned for EAP-TLS (certificate auth) hitting a policy that only offers PEAP-MSCHAPv2 — common mid-migration.
- Fix on whichever side is wrong: add the method to the policy, or fix the device's 802.1X profile.

**Failing since a config change or import?** The EAP configuration blob inside a network policy can corrupt — a known aftermath of certificate renewals and `netsh nps import` operations. The repair is blunt but effective: in the policy's authentication settings, remove the EAP type, re-add it, reselect the certificate, save.

## What Reason 22 is *not*

It is not a credentials problem (the password was never reached) and not a policy-matching problem (a policy *did* match — check `Network Policy Name` in the event). Don't reset passwords or reorder policies for this one.
