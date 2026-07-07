---
title: "FreeRADIUS 'No known good password' — the server has nothing to check against"
description: "What 'Failed to find known good password' in radiusd -X means: no authorize module supplied a password, the stored hash is incompatible with the auth method, or the lookup silently missed."
format: freeradius
signature: "ERROR: No \"known good\" password"
rule_id: freeradius/no-known-good-password
---

**The log lines:**

```
(3) pap: WARNING: No "known good" password found for the user.
(3) pap: WARNING: Authentication will fail unless a "known good" password is available
```

**What it actually means:** FreeRADIUS reached the authentication stage holding the user's *submitted* password — but with **nothing to compare it against**. No module in the `authorize` phase produced a `Cleartext-Password`, `NT-Password`, or other control attribute for this user. Authentication cannot succeed no matter what the user types; the problem is entirely server/backend-side.

## Follow the authorize section — it names the culprit

Scroll up in the `radiusd -X` output to the `authorize {}` section for this request. You're looking for which module was supposed to supply the password, and why it didn't:

### 1. No module matched the user at all

- **users file:** the entry doesn't match (typo in the name, wrong `DEFAULT` ordering, or a preceding entry with `Fall-Through = no` swallowing it). The debug shows `files:` lines — did any entry match?
- **SQL:** the query ran but returned zero rows. The debug prints the *exact* SQL executed — copy it and run it by hand against the database. Nine times out of ten the username stored differs from the one sent (case, domain prefix, realm suffix).
- **LDAP:** the search found no entry, or found it but didn't map a password attribute. Check the `ldap:` debug lines for the search filter and what attributes came back.

### 2. A password came back — in a format the auth method can't use

Subtler: the backend *did* return a password, but as an incompatible hash. The compatibility matrix is unforgiving:

| Stored format | PAP | MSCHAPv2 (PEAP inner) |
|---|---|---|
| Cleartext | ✔ | ✔ |
| NT hash | ✔ | ✔ |
| SSHA / SHA / MD5 / crypt / bcrypt | ✔ | ✘ never |

If your directory stores SSHA (as it should for security) and clients use PEAP-MSCHAPv2, this warning is structural — see the [MSCHAPv2 write-up](/kb/freeradius-mschap2-response-incorrect) for the escape routes (ntlm_auth against AD, TTLS-PAP, or EAP-TLS).

### 3. The password attribute exists but under the wrong name

In SQL schemas: the `radcheck` row says `Password` instead of `Cleartext-Password` (an ancient schema convention), or the operator is `=` where `:=` is needed. In LDAP: the mapping in `mods-available/ldap` doesn't include the attribute your directory actually uses.

## The five-minute check

`grep` the debug for what the authorize phase produced: if you never see a `Cleartext-Password` or `NT-Password` control attribute being added for the user, the backend lookup is your bug — fix the entry/query/mapping, not the client.
