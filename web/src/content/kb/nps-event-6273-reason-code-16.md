---
title: "NPS Event 6273 Reason Code 16 explained — and why it's often not the password"
description: "What 'Authentication failed due to a user credentials mismatch' actually means in Windows NPS, the three causes that produce it, and the exact fields in the event that tell you which one you have."
format: nps
signature: "Event 6273 · Reason Code 16"
rule_id: nps/reason-code-16
---

**The official text:** *"Authentication failed due to a user credentials mismatch. Either the user name provided does not map to an existing user account or the password was incorrect."*

**What it actually means:** NPS got all the way through policy matching and started verifying the credentials — and the verification failed. Your RADIUS client entry, shared secret, connection request policy, and network policy are all fine. The problem is the identity or the proof the device sent.

That sounds like "wrong password," and sometimes it is. But Reason Code 16 is the great garbage bin of NPS failures, and a large share of them have nothing to do with what the user typed.

## Check these three things, in this order

### 1. What identity did the device actually send?

Open the 6273 event and read the **User** section. This is the single highest-value field in the event:

- `CORP\jsmith` — a user authentication, as expected.
- `host/LAPTOP-042.corp.example.com` — this is **computer** authentication. The device authenticated as the machine account, not the user. If your network policy's group condition contains user groups, machine auth will fail with Reason 16 (or 48) even though "the password" was never wrong.
- `jsmith@corp.example.com` vs `jsmith` — UPN vs sAMAccountName. If the name format can't be resolved the way NPS looks it up, you get a credentials mismatch with a perfectly correct password.

The fix is to align the Wi-Fi/802.1X profile's authentication mode ("user authentication", "computer authentication", or "user or computer") with what your policies expect.

### 2. Is the device sending a stale saved password?

Windows, Android, and iOS all cache 802.1X credentials per network. After an AD password change, the device happily keeps sending the old password until someone tells it otherwise — the user isn't typing anything at all. The symptom: one user fails everywhere after a password change, and account lockouts often follow (each reconnect attempt burns a bad-password count).

**Fix:** forget/remove the Wi-Fi network on the device and reconnect with the current credentials.

### 3. Did NTLM hardening break PEAP-MSCHAPv2?

PEAP-MSCHAPv2 validation depends on the NT hash verification path. Security work that disables NTLM — `LmCompatibilityLevel` changes, NTLM blocking GPOs, or Credential Guard on current Windows builds — can break PEAP-MSCHAPv2 for the whole domain at once. The giveaway: **everyone** starts getting Reason 16 after a security baseline rollout, with no password changes anywhere.

Check the domain controllers' NTLM audit logs and 4625 events at the same timestamps. If this is your cause, the durable fix is migrating to certificate-based EAP-TLS rather than weakening NTLM policy.

## Still stuck?

Reason 16 is a symptom with at least three diseases. The fastest way to tell them apart is to read the full event — the `Authentication Type`, `EAP Type`, and `Account Name` fields narrow it down decisively.
