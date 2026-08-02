# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

Planned release: 2.1.0.

The generated table of contents is byte-identical to 2.0.1. Everything below is about
CLI behaviour, not about the output format.

### Security

- `--debug` no longer prints the GitHub token. Up to and including 2.0.1, a token passed
  through `--token` or `GH_TOC_TOKEN` was written to the debug output as part of the
  dumped configuration. **If you ran `gh-md-toc --debug` in CI or any environment with
  shared logs, rotate that token.** The debug output now only reports whether a token was
  configured. ([#60](https://github.com/ekalinin/github-markdown-toc.go/pull/60))

### Changed

- The CLI now exits with a non-zero status when a document fails to process. Previously it
  exited 0 even when the file did not exist or an extractor failed, so failures were
  invisible to callers. Scripts that relied on the old always-zero status need review.
  ([#61](https://github.com/ekalinin/github-markdown-toc.go/pull/61))
- Output order for multiple documents now always matches the order of the CLI arguments.
  Previously results were printed in completion order and varied between runs.
  ([#63](https://github.com/ekalinin/github-markdown-toc.go/pull/63))
- Parallel processing is bounded to 8 documents at a time. `--serial` still processes one
  at a time and now goes through the same aggregation path as the parallel mode.
  ([#63](https://github.com/ekalinin/github-markdown-toc.go/pull/63))
- An unsupported `--re-version` value now fails with an explicit error listing the
  supported versions, instead of silently producing an empty TOC and exiting 0.
  ([#60](https://github.com/ekalinin/github-markdown-toc.go/pull/60))
- `Ctrl-C` and `SIGTERM` now cancel in-flight requests instead of being ignored until the
  current operation finishes.
  ([#61](https://github.com/ekalinin/github-markdown-toc.go/pull/61))
- Error messages name the document that failed and the operation that failed on it.
  ([#61](https://github.com/ekalinin/github-markdown-toc.go/pull/61))
- Building from source now requires Go 1.21 or newer, matching the `log/slog` usage that
  was already in the code. Releases are built with Go 1.26.5.
  ([#58](https://github.com/ekalinin/github-markdown-toc.go/pull/58))

### Fixed

- `GH_TOC_URL` is honoured again when `--github-url` is not passed. The flag's non-empty
  default used to shadow the environment variable, so the variable had no effect.
  ([#60](https://github.com/ekalinin/github-markdown-toc.go/pull/60))
- Temporary files created for STDIN input and for downloaded remote Markdown are removed
  on every path, including error paths. Previously a remote run left one file behind.
  ([#62](https://github.com/ekalinin/github-markdown-toc.go/pull/62))
- HTTP requests carry a context and a 30 second timeout, so an unresponsive server no
  longer hangs the CLI indefinitely.
  ([#62](https://github.com/ekalinin/github-markdown-toc.go/pull/62))
- Only 2xx responses are treated as success. Other responses produce an error carrying the
  status code and a truncated response body, instead of being parsed as content.
  ([#62](https://github.com/ekalinin/github-markdown-toc.go/pull/62))
- Response bodies are capped at 10 MiB.
  ([#62](https://github.com/ekalinin/github-markdown-toc.go/pull/62))
- `Content-Type` is parsed with `mime.ParseMediaType`, so values carrying parameters such
  as `text/plain; charset=utf-8` are recognised correctly, and an unexpected media type is
  reported as an error.
  ([#62](https://github.com/ekalinin/github-markdown-toc.go/pull/62))
- A single shared HTTP client is reused instead of constructing one per request.
  ([#62](https://github.com/ekalinin/github-markdown-toc.go/pull/62))
- `gopkg.in/alecthomas/kingpin.v2` updated from v2.2.4 to v2.2.6, and the indirect module
  graph was tidied. The CLI surface is unchanged.
  ([#68](https://github.com/ekalinin/github-markdown-toc.go/pull/68))

## [2.0.1] - 2026-04-03

### Added

- Support for GitHub's `codeViewBlobRoute` payload when reading a table of contents from a
  blob page, alongside the existing `Payload.Blob` layout.

### Fixed

- Module paths corrected for the `/v2` module.
  ([#55](https://github.com/ekalinin/github-markdown-toc.go/pull/55))

## [2.0.0] - 2025-04-13

### Changed

- Rewritten on a clean/hexagonal architecture, splitting the CLI into `cmd`, application
  wiring, controller, use cases, core model and adapters.
  ([#52](https://github.com/ekalinin/github-markdown-toc.go/pull/52))
- Module path moved to `github.com/ekalinin/github-markdown-toc.go/v2`.

Releases before 2.0.0 are documented in the
[GitHub releases](https://github.com/ekalinin/github-markdown-toc.go/releases) and in the
git history.

[Unreleased]: https://github.com/ekalinin/github-markdown-toc.go/compare/v2.0.1...master
[2.0.1]: https://github.com/ekalinin/github-markdown-toc.go/compare/v2.0.0...v2.0.1
[2.0.0]: https://github.com/ekalinin/github-markdown-toc.go/compare/v1.4.0...v2.0.0
