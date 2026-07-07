---
title: "NPS Event 6273 Reason Code 48 — no network policy matched (group and port-type traps)"
description: "Why 'The connection request did not match any configured network policy' happens even when the user and password are right: group scoping, computer vs user objects, NAS Port Type conditions, and policy ordering."
format: nps
signature: "Event 6273 · Reason Code 48"
rule_id: nps/reason-code-48
---

**The official text:** *"The connection request did not match any configured network policy."*

**What it actually means:** the credentials might even be fine — NPS never got that far in the way you intended. The request was compared against your network policies top-to-bottom, none of them claimed it, and NPS's default answer is deny.

Reason 48 is a *scoping* problem: the request and the policy disagree about who/what/where this connection is.

## The usual suspects

### 1. Group membership isn't what you think

The most common cause. The policy's condition says "Windows Groups: WiFi-Users" and the account isn't in it — or is in it via a nested group that isn't evaluating the way you expect, or lives in another domain.

The subtle version: **object type mismatch**. If the device performs *computer* authentication (`host/LAPTOP-042` in the User field), NPS is matching the **computer object** against your group condition. A "User Groups" condition can never match a computer, and vice versa. Mixed user-and-computer environments should use "Windows Groups" with both kinds of members, or separate policies.

**Check:** read the exact account name in the 6273 event, then verify that specific object's membership in the specific group the policy names.

### 2. A condition doesn't match what the NAS actually sends

`NAS Port Type = Wireless - IEEE 802.11` is in nearly every wireless policy template. But some controllers and most switches send other values (`Ethernet`, `Virtual`), and VPN appliances send `Virtual (VPN)`. If the event's NAS Port-Type differs from the policy condition by one word, nothing matches.

**Check:** compare the `NAS Port-Type` field in the event against the policy condition, literally, word for word.

### 3. Policy ordering

NPS evaluates network policies in order and **the first match wins** — even if it's a deny policy, and even if a better policy sits right below it. A broad policy added "temporarily" at the top swallows requests meant for the specific one.

**Check:** in the NPS console, walk the policy list from the top with the event's attributes in hand and ask "would this one match?" for each. The first "yes" is what actually happened — the event's `Network Policy Name` field shows `-` when nothing matched at all.

## A note on disabled policies

A policy that was "just disabled for a minute" during troubleshooting is invisible to matching. If Reason 48 appeared suddenly after maintenance, check the enabled checkbox before anything else.
