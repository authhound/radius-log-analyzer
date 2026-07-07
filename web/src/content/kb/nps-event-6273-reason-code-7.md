---
title: "NPS Event 6273 Reason Code 7 — 'the specified domain does not exist'"
description: "Why NPS says the domain does not exist: usernames with the wrong realm, personal email addresses typed as identities, NetBIOS vs DNS domain names, and DC locator failures."
format: nps
signature: "Event 6273 · Reason Code 7"
---

**The official text:** *"The specified domain does not exist."*

**What it actually means:** NPS split the identity the device sent into domain + username, went looking for that domain, and couldn't find it. The user and password were never evaluated — the *domain half of the username* failed first.

## Read the Account Name — the answer is usually right there

The 6273 event's `Account Name` field shows the exact string NPS parsed. Common shapes of this failure:

### 1. The user typed an identity with a foreign realm

The classic: a user enters their **personal email address** (`jane@gmail.com`) as the Wi-Fi username, and NPS dutifully goes looking for a domain called `gmail.com`. Also common: a corporate UPN suffix that exists in email but was never added as a UPN suffix in AD.

**Fix:** re-enter the username on the device — bare `jane` or `CORP\jane` or the real UPN. If a whole class of users does this, your onboarding instructions are the actual bug.

### 2. NetBIOS vs DNS name confusion

`CORP\jane` works while `corp.example.com\jane` doesn't (or vice versa) depending on how the request travels. Devices and VPN clients sometimes hardcode one form.

**Check:** compare the domain portion in the event against both the NetBIOS name and the DNS name of your domain. Standardize the profile on the form that works.

### 3. The domain exists — but NPS can't reach it

If the domain in the event is genuinely yours and correct, the failure is on the NPS server's side: DNS misconfiguration, no reachable domain controller, or a broken trust to the child/partner domain the user belongs to.

**Check:** on the NPS server, `nltest /dsgetdc:<domain>` — if that fails, you have a DC locator/DNS problem, not an authentication problem. For cross-domain users, verify the trust with `nltest /trusted_domains`.

### 4. A realm that should have been stripped

In proxy/roaming setups (eduroam-style, or MSPs forwarding between tenants), the realm suffix is routing information, not a Windows domain. If a request that should be forwarded is instead processed locally, NPS tries to treat `@partner.org` as a domain and lands here.

**Check:** the Connection Request Policy that matched (named in the event) — should it have forwarded this realm to a remote RADIUS server group, or stripped the suffix before local processing?
