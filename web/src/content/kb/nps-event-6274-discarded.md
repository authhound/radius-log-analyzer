---
title: "NPS Event 6274 — 'discarded the request' is not a reject (and why that matters)"
description: "The difference between NPS event 6273 (denied) and 6274 (discarded), and the causes of discards: unverifiable packets, malformed requests, and accounting problems."
format: nps
signature: "Event 6274 · NPS discarded the request"
---

**The event text:** *"Network Policy Server discarded the request for a user."*

**What it actually means:** NPS didn't say *no* — it said *nothing*. A **6273 (denied)** means the request was evaluated and rejected: the NAS gets an Access-Reject, the client fails fast. A **6274 (discarded)** means NPS threw the packet away without answering at all: the NAS sees a **timeout**, retries, and may fail over to a secondary RADIUS server. If users report "Wi-Fi takes forever and then fails" rather than failing instantly, you're in discard territory.

The discard reasons cluster into three families:

## 1. The packet couldn't be trusted

The most common discard: the message failed verification, so answering would be unsafe. The `Reason Code` / `Reason` line in the 6274 event says the message was malformed or its signature/authenticator couldn't be verified. This is almost always the **shared secret** disagreeing between the NAS and the NPS RADIUS Client entry — the discard-flavored sibling of [Reason Code 262](/kb/nps-event-6273-reason-code-262).

**Check:** re-enter the shared secret on both sides; confirm the request is matching the RADIUS Client entry you think it is (source IP in the event vs. your entries).

## 2. The packet was malformed or unsupported

Vendor quirks: NAS firmware sending malformed attributes, unsupported RADIUS extensions, or (rarely) genuinely corrupted UDP. If the secrets check out, capture a failing exchange and compare it with a working NAS of the same model — firmware upgrades on the NAS fix a surprising share of these.

## 3. Server-side trouble answering at all

If discards appear for *all* clients at once, look at the NPS server rather than the packets:

- **NPS not registered in AD** — it can't read dial-in properties, so requests can't be processed. Fix: right-click NPS (top node in the console) → *Register server in Active Directory* (adds the server to the `RAS and IAS Servers` group).
- **Accounting failure discards** — if NPS is configured to log accounting to a SQL server or a full disk and "discard connection requests if logging fails" is enabled, an accounting outage silently discards *authentications*. Check Accounting configuration and the free space/SQL connectivity.
- Event log neighbors (event 13, 4402 in the System/NPS logs) often name the server-side cause precisely.

## Triage in one line

One NAS discarding → shared secret or that NAS's firmware. Everything discarding → NPS registration or accounting/logging backend. Instant failures instead of slow ones → you're actually looking at 6273, different playbook.
