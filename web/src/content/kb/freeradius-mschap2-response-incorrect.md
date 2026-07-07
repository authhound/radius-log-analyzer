---
title: "FreeRADIUS 'MS-CHAP2-Response is incorrect' — wrong password, or a password your backend can't check"
description: "What the mschap FAILED error in radiusd -X really means: the TLS tunnel is fine, and the inner MSCHAPv2 check failed — because of the password, the stored hash format, or the AD account state."
format: freeradius
signature: "mschap: ERROR: MS-CHAP2-Response is incorrect"
rule_id: freeradius/mschap-wrong-password
---

**The log lines:**

```
(8) mschap: ERROR: MS-CHAP2-Response is incorrect
(8) eap_peap: ERROR: Tunneled authentication was rejected
(8) Login incorrect (mschap: ERROR: MS-CHAP2-Response is incorrect): [alice]
```

**What it actually means:** everything up to the password worked. The client trusted your certificate, the TLS tunnel came up, the inner EAP conversation ran — and the MSCHAPv2 proof the client sent doesn't match what the server computed from its copy of the password. Certificates, shared secrets, and clients.conf are all exonerated.

Three things produce a failed proof:

## 1. The password (or the username variant) is simply wrong

Including the sneaky versions: a stale cached Wi-Fi credential after a password change, or the same user existing as `alice`, `CORP\alice`, and `alice@corp.example.com` where only one resolves to the right account.

**Check:** take the exact `User-Name` from the debug output and test those credentials on the server itself — `radtest` for a quick loop, or `ntlm_auth --username=alice --password=...` if you authenticate against AD. If it fails there too, it's genuinely the credential.

## 2. Your backend can't produce an NT hash — so MSCHAPv2 can *never* succeed

This is the one that eats weekends. MSCHAPv2 is mathematically bound to the **NT hash** of the password. The server must have either the cleartext password or the NT hash. If your LDAP directory stores passwords as SSHA, MD5-crypt, or bcrypt — as security best practice says it should — FreeRADIUS cannot verify MSCHAPv2 against them. Every user fails, always, with this exact error.

**Check:** look earlier in the debug output for warnings like `No Cleartext-Password` / `No NT-Password`. If present, your options are: store NT hashes (unpalatable), authenticate via `ntlm_auth` against AD/Samba, switch the inner method to TTLS-PAP (the directory verifies the cleartext), or move to EAP-TLS and stop fighting password formats.

## 3. The AD account is locked, disabled, or expired

When FreeRADIUS delegates to AD via `ntlm_auth`/winbind, account-state problems surface as a generic MSCHAP failure rather than a helpful message.

**Check:** run `ntlm_auth` manually on the RADIUS host and read the `NT_STATUS_*` code it returns — `NT_STATUS_ACCOUNT_LOCKED_OUT` and `NT_STATUS_PASSWORD_EXPIRED` name the problem precisely.

## Triage shortcut

One user failing → cause 1 or 3. Every user failing, since forever or since a directory migration → cause 2, and no password reset will ever fix it.
