# RADIUS Log Analyzer

Paste FreeRADIUS debug output (`radiusd -X`) or a Windows NPS event log entry — get back a plain-English diagnosis:

1. **What happened** — the failure stage: shared secret, certificate validation, inner auth, policy match, EAP fragmentation, …
2. **Top 3 likely causes**, ranked by real-world frequency.
3. **One concrete next check** for each cause.

Use it in the browser at **[the web analyzer](https://authhound.com/analyzer)** or offline via the CLI.

## Privacy: your log never leaves your machine

The web analyzer parses **entirely in your browser** — open DevTools → Network and watch nothing get sent while you analyze. The CLI is fully offline. This repo is the proof: the browser engine ([parser/](parser/)) and the CLI run the **same rule files** ([rules/](rules/)), and a shared fixture corpus ([testdata/](testdata/)) pins both implementations to identical behaviour in CI. The website at [authhound.com](https://authhound.com) consumes this repo unmodified — the parser running in your browser is the code you're looking at.

## CLI

```console
$ radiusd -X 2>&1 | radius-analyze
$ radius-analyze nps-event.txt
$ radius-analyze -json debug.log | jq .matches[0].rule_id
```

Install: grab a binary from [Releases](../../releases), or:

```console
$ go install github.com/authhound/radius-log-analyzer/cmd/radius-analyze@latest
```

Example output:

```
Format detected : Windows NPS event log
reason_code     : 16
authentication_type: PEAP

DIAGNOSIS — Reason Code 16 — credential mismatch (it's not always the password)
Stage: inner_auth

NPS matched your policies and got as far as checking the credentials — and
the check failed. ...

Likely causes, most likely first:
  1. A stale saved password on the device ...
     -> Next check: On the failing device, forget/delete the Wi-Fi network ...
```

## Supported input formats

| Format | How to capture it |
|---|---|
| FreeRADIUS debug | Run the server in debug mode: `radiusd -X` (or `freeradius -X`) and copy the output around the failing request. |
| Windows NPS events | Event Viewer → Security log → event 6273/6274 → right-click → **Copy details as text**. |

## How diagnosis works

Every failure signature is a small JSON rule in [rules/](rules/) (schema: [rules/schema.json](rules/schema.json)):

```json
{
  "id": "nps/reason-code-262",
  "failure_stage": "shared_secret",
  "match": { "any": ["Reason Code:\\s*262\\b"] },
  "plain_english": "…what actually happened…",
  "causes": [ { "cause": "…", "next_check": "…" } ]
}
```

The rules encode years of vendor-side RADIUS ticket pattern-matching. Both engines (Go and TypeScript) load the same files, match them line-by-line, and rank results by specificity.

**Contributing a rule** is the most valuable contribution possible: add the JSON rule, a redacted sample log in `testdata/`, and its `.expected.json`. `go test ./...` and `cd parser && npm test` must both pass — that's the whole review bar for parity.

## Repository layout

```
rules/        Shared diagnosis rules (JSON) — the single source of truth
testdata/     Shared fixtures: sample log -> expected diagnosis, run by BOTH engines
cmd/          Go CLI (radius-analyze)
internal/     Go parser + rules engine
parser/       TypeScript parser — the exact code the web analyzer runs in your browser
```

The [authhound.com](https://authhound.com) website lives in a separate repo and pulls this one in as a submodule, so the deployed parser is always a pinned commit of this public code.

## Development

```console
# Go engine + CLI
$ go test ./...
$ go run ./cmd/radius-analyze testdata/freeradius/mschap-fail.log

# TypeScript engine (same fixtures)
$ cd parser && npm install
$ npm test
```

## License

[MIT](LICENSE)
