# Changelog

All notable changes to Alfrenslate are documented here. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project uses [Semantic Versioning](https://semver.org/).

## [1.0.1] - 2026-08-01

### Fixed

- Alfred Workflow Configuration now renders correctly and saves every supported provider setting.
- Added the actual text-only Universal Action integration for **Translate with Alfrenslate**.
- Added safe macOS Gatekeeper installation and troubleshooting guidance.
- Added stronger parsed-plist Workflow Configuration, object, and connection graph tests.

## [1.0.0] - 2026-08-01

### Added

- Automatic Chinese-to-English and English/other-to-Simplified-Chinese direction detection.
- One daily `ts` keyword with `--to zh`, `--to en`, `-t zh`, and `-t en` overrides.
- DeepSeek, Generic OpenAI-Compatible, DeepL, Youdao, Baidu Translate, Google Cloud Translation, and Microsoft Translator providers.
- Concurrent multi-provider comparison with deterministic success and failure ordering.
- Alfred Universal Action, management entry, provider diagnostics, clipboard, Large Type, and copy-then-paste actions.
- Atomic local translation cache with TTL, safe clearing, and credential-free cache keys.
- Native macOS arm64 and x86_64 binaries in one installable Workflow.
- Credential redaction, privacy documentation, mocked automated tests, packaging verification, CI, and tagged release automation.

[1.0.1]: https://github.com/qinchonghanzuibang/Alfrenslate/releases/tag/v1.0.1
[1.0.0]: https://github.com/qinchonghanzuibang/Alfrenslate/releases/tag/v1.0.0
