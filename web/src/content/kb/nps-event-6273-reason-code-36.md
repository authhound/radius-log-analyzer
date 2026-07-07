---
title: "NPS Event 6273 Reason Code 36 — account locked out (usually by the Wi-Fi itself)"
description: "Why NPS rejects with 'the account is locked out', how a stale cached Wi-Fi password causes the lockout loop, and the neighboring reason codes 34, 35, 37, 38 for disabled/expired accounts."
format: nps
signature: "Event 6273 · Reason Code 36"
---

**What it means:** the account exists, the policies matched — but Active Directory has the account **locked out**, so NPS refuses before checking anything else. Unlocking it in ADUC fixes the symptom in ten seconds. The real question is *what locked it*, because with 802.1X the answer is very often: **the Wi-Fi itself**.

## The lockout loop

Here's the mechanism that makes Reason 36 special in wireless networks:

1. The user changes their AD password (or it expires and they update it at their desk).
2. Their phone / second laptop / tablet still has the **old password cached** in its Wi-Fi profile.
3. The device retries the network automatically — every few seconds, forever, from the parking lot, at midnight.
4. Each retry is a bad-password attempt against AD. The lockout threshold trips in minutes.
5. Now even the *correct* password fails everywhere with Reason 36, the helpdesk unlocks the account, and the loop trips it again within the hour.

The fingerprint: the same account locks repeatedly, event timestamps continue while the user is demonstrably not typing anything, and the lockouts started right after a password change.

**Fix:** find the device with the stale credential and make it forget the network. The DC's security log (event 4740 on the PDC emulator shows the lockout source, and 4625s name the calling workstation) or your WLC's client list by username will point at it. Then check `Calling Station Identifier` in the 6273 events — that MAC is your culprit device.

## The neighboring reason codes

These four travel together and mean exactly what they say — all are account-state, not password, problems:

| Reason | Meaning | Fix location |
|---|---|---|
| 34 | Account **disabled** | ADUC → enable (ask why it was disabled first) |
| 35 | Account **expired** | ADUC → Account tab → account expiry date |
| **36** | Account **locked out** | Unlock + find the source (above) |
| 37 | Outside allowed **logon hours** | ADUC → Account tab → Logon Hours |
| 38 | Password **must change at next logon** | User can't change it over 802.1X — reset at a desktop first |

Reason 38 deserves a note: "user must change password at next logon" is unfulfillable through PEAP Wi-Fi — the device just fails until the user logs on somewhere that can show the change-password prompt. New-starter laptops shipped with must-change passwords hit this constantly.
