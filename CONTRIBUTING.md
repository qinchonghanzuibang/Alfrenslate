# Contributing to Alfrenslate

Alfrenslate is a Go application packaged as an Alfred 5 Workflow. Contributions should preserve the one-keyword product model, local direction detection, provider isolation, stable ordering, and credential privacy.

## Requirements

- Go 1.23 or newer
- macOS with `plutil`, `zip`, `unzip`, and `shasum`
- `rsvg-convert` only when regenerating `workflow/icon.png` from `assets/icon.svg`
- Alfred 5 with Powerpack for interactive Workflow testing

## Layout

- `cmd/alfrenslate`: CLI and Alfred entry points
- `internal/detect`: local target-language decision
- `internal/prompt`: versioned LLM translation prompt and minimal output cleanup
- `internal/providers`: shared transport and seven provider implementations
- `internal/translate`: concurrent orchestration
- `internal/cache`: atomic local cache
- `internal/alfred`: Script Filter JSON
- `workflow`: importable Workflow definition and architecture launcher
- `scripts`: package and secret verification

## Commands

```bash
make format
make test
make vet
make build
make package
make verify
```

`make package` creates `Alfrenslate.alfredworkflow` and `Alfrenslate.alfredworkflow.sha256`. `make verify` runs formatting, mocked automated tests, vet, both macOS builds, plist validation, package extraction checks, executable checks, and secret scanning.

## Testing

Default tests use local `httptest.Server` instances and never need provider credentials. Provider requests, signatures, parsing, error handling, retries, timeouts, cancellation, concurrency, caching, Alfred JSON, and packaging are checked without paid APIs.

Optional live tests are deliberately separate:

```bash
ALFRENSLATE_LIVE_TEST=1 go test ./... -tags=live
```

Set provider variables exactly as documented and enable only the providers you intend to charge. Never add real credentials to source, fixtures, screenshots, command transcripts, or pull requests.

## Provider changes

Use the provider's current official REST documentation. Add mock tests for method, URL, headers, body, target mapping, auto source detection, success parsing, missing fields, malformed responses, status errors, timeout, cancellation, retry behavior, and redaction. Do not use unofficial web endpoints or scrape translation websites.

## Workflow testing

After `make package`, import the generated file into Alfred 5 and verify `ts`, forced directions, keyboard modifiers, Universal Action, management, provider status, cache clearing, and Accessibility fallback. Test both an arm64 Mac and Rosetta/x86_64 when available.

## Pull requests

Use Conventional Commits. Explain behavior, tests, privacy/security impact, and release implications. Keep generated or local data out of the diff.
