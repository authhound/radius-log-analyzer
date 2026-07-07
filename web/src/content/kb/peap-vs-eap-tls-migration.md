---
title: "Migrating from PEAP-MSCHAPv2 to EAP-TLS without breaking Wi-Fi for a week"
description: "A staged migration plan from password-based PEAP to certificate-based EAP-TLS: why NTLM hardening is forcing the move, the dual-policy transition pattern, and the reason codes you'll see when a device is mid-migration."
format: general
signature: "PEAP-MSCHAPv2 → EAP-TLS migration"
---

PEAP-MSCHAPv2 — a password inside a TLS tunnel — has run enterprise Wi-Fi for two decades. It's now living on borrowed time, and most organizations discover this the hard way: a security baseline lands, and Wi-Fi breaks.

## Why the move is becoming mandatory

- **MSCHAPv2 depends on NTLM-era cryptography.** Validating it needs the NT hash path, so every NTLM-hardening initiative (disabling NTLM domain-wide, Credential Guard on current Windows — which blocks MSCHAPv2 SSO outright) collides with PEAP Wi-Fi. The failure surfaces as [NPS Reason Code 16](/kb/nps-event-6273-reason-code-16) storms with no password changes anywhere.
- **Passwords are the weak link anyway:** crackable captures from evil-twin APs, credential stuffing, users typing corporate passwords into personal devices. Certificates can't be phished by a fake SSID.
- **The password-change lockout loop** ([Reason Code 36](/kb/nps-event-6273-reason-code-36)) simply disappears when devices stop caching passwords.

## What EAP-TLS demands up front

Be honest about the entry price: **every device needs a certificate**, which means working enrollment — AD CS with autoenrollment for domain machines, SCEP/Intune/Jamf for mobile and BYOD. The certificate *distribution* problem is 90% of the migration; the RADIUS-side change is an afternoon.

## The dual-policy transition pattern

Never flip the whole network. Run both methods side by side and drain PEAP over weeks:

1. **Add EAP-TLS alongside PEAP** — either both methods on the existing policy, or (cleaner) a separate policy/rule scoped to a "TLS-Migrated" device group. FreeRADIUS handles both EAP types in one `eap` module config; NPS takes both in one policy's constraints.
2. **Enroll certificates in rings** — IT first, then a pilot department, then the fleet. Update the Wi-Fi profile (GPO/MDM) to prefer EAP-TLS per ring.
3. **Watch the meters, not the calendar** — count remaining PEAP authentications per week (NPS: `Authentication Type`/`EAP Type` fields in 6272/6273 events; FreeRADIUS: eap type in the logs). Drain to zero *before* the cutoff.
4. **Remove PEAP** from the policy, then celebrate: an entire class of tickets (stale cached passwords, lockout loops, MSCHAP hash incompatibilities) retires with it.

## The mid-migration failure signatures

Devices caught between rings produce a predictable set of errors — recognize them as migration traffic, not new bugs:

| Symptom | Meaning |
|---|---|
| [NPS Reason 22](/kb/nps-event-6273-reason-code-22) / [Reason 66](/kb/nps-event-6273-reason-code-66) | device asks for the method its ring no longer offers |
| [unknown CA alerts](/kb/freeradius-tls-alert-unknown-ca) | new trust chain not yet on the device |
| [certificate expired](/kb/freeradius-certificate-expired) (client) | early-ring certs aging out — check your renewal automation *now* |
| [Reason 16](/kb/nps-event-6273-reason-code-16) spikes | stragglers still on PEAP hitting NTLM hardening |

## The two rules that keep migrations un-broken

**Server certificate hygiene first:** EAP-TLS makes your RADIUS server cert load-bearing in both directions — get the chain complete and monitored before enrolling a single client. **And never disable server-cert validation on clients** to "simplify" the rollout; it converts your migration into an evil-twin vulnerability that outlives it.
