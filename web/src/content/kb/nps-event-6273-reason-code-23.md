---
title: "NPS Event 6273 Reason Code 23 — 'an error occurred during EAP' means your NPS certificate"
description: "Decoding NPS Reason Code 23: why 'check EAP log files' almost always means the NPS server's own certificate is expired, missing its private key, or failing revocation checks."
format: nps
signature: "Event 6273 · Reason Code 23"
rule_id: nps/reason-code-23
---

**The official text:** *"An error occurred during the Network Policy Server use of the Extensible Authentication Protocol (EAP). Check EAP log files for EAP errors."*

**What it actually means:** NPS hit an internal error while running the EAP conversation. The text points you at "EAP log files" that few people have ever enabled; in the field, Reason 23 is almost always the **NPS server's own certificate** — not anything the client sent.

## The three certificate failures behind Reason 23

### 1. Expired, or missing its private key

On the NPS server, open `certlm.msc` → Personal → Certificates:

- Is there a certificate with the **Server Authentication** EKU that is not expired?
- Does it show the key icon (private key present)? A cert restored from a `.cer` export, or copied between servers without the key, looks fine and cannot sign anything.

The signature move of this failure: **Wi-Fi died for everyone overnight and nothing was changed** — because the change was a certificate quietly reaching its notAfter date.

### 2. Renewed, but the policy still points at the old one

NPS network policies bind to a specific certificate. After a renewal (especially manual renewals with a new key), the policy can keep referencing the vanished cert. Open the policy → Constraints → Authentication Methods → PEAP/EAP-TLS → Edit, reselect the current certificate, save. Do this even if the dialog "looks right" — reselecting rebuilds the binding.

### 3. Revocation checking fails

NPS validates its own chain, and if the CRL distribution point is unreachable from the NPS server (firewalled CA web server, decommissioned CA, expired CRL), EAP setup errors out. Check the **CAPI2 operational log** (Event Viewer → Applications and Services → Microsoft → Windows → CAPI2 — enable it if disabled) and test the CRL URL with `certutil -urlfetch -verify <cert>` on the NPS box.

## If it's genuinely not the certificate

Then the "check EAP log files" advice becomes real: enable NPS tracing with `netsh ras set tracing * enabled`, reproduce, and read `%windir%\tracing\` — the EAP trace names the failing component. But do the ten-minute certificate check first; it wins the bet most days.
