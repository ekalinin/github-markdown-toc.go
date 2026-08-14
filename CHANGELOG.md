# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

Planned release: 2.1.0.

### Added

- `--insert` writes the generated TOC directly into a document, replacing everything
  between a `<!--ts-->` and `<!--te-->` marker pair. A backup copy is kept next to the
  file unless `--no-backup` is also passed.
- `--skip-header` ignores everything up to and including `<!--te-->` when building the
  TOC, so a document's own title heading is not picked up as an entry.
- `-` is now accepted as an explicit marker for reading Markdown from STDIN.
- `token.txt`, read from next to the executable, is now the last fallback for a GitHub
  token, after `--token` and `GH_TOC_TOKEN`.
- A Docker image is published to `ghcr.io/ekalinin/github-markdown-toc.go`.

### Security

- `--debug` no longer prints the GitHub token. Up to and including 2.0.1, a token passed
  through `--token` or `GH_TOC_TOKEN` was written to the debug output as part of the
  dumped configuration. **If you ran `gh-md-toc --debug` in CI or any environment with
  shared logs, rotate that token.** The debug output now only reports whether a token was
  configured. ([#60](https://github.com/ekalinin/github-markdown-toc.go/pull/60))
- CI and releases now build with Go 1.26.6. Go 1.26.5 carried four standard-library
  vulnerabilities that `govulncheck` reports as reachable from this code:
  [GO-2026-6218](https://pkg.go.dev/vuln/GO-2026-6218) (`net/url`),
  [GO-2026-6090](https://pkg.go.dev/vuln/GO-2026-6090) (`crypto/tls`),
  [GO-2026-5972](https://pkg.go.dev/vuln/GO-2026-5972) (`encoding/asn1`) and
  [GO-2026-5026](https://pkg.go.dev/vuln/GO-2026-5026) (`net/http`). Binaries you built
  yourself with Go 1.26.5 or earlier should be rebuilt.

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
- Building from source now requires Go 1.26 or newer. `go.mod` declares `go 1.26`, and CI
  no longer tests against 1.21.x, so the toolchain used for tests, builds and releases is
  the same one that ships the binaries.
  ([#58](https://github.com/ekalinin/github-markdown-toc.go/pull/58),
  [#84](https://github.com/ekalinin/github-markdown-toc.go/pull/84))
- Multi-document runs now prefix links with the document path, which is what the
  "Multiple files" and "Combo" sections of the README always documented but the tool
  never actually did.
- `--version` now also reports the OS, architecture and Go version used to build the
  binary. The bare version number stays on the first line, so scripts that parse it
  keep working.
- `--hide-footer` gains a second meaning under `--insert`: it also suppresses the
  signature comment written into the file, not just the printed footer.

### Fixed

- `--debug` writes its HTML dump next to the document you named. With `--skip-header`
  the dump used to be named after an internal temporary copy and was left behind in
  the temp directory.
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
- GitHub rate-limit responses now explain that a token raises the limit, instead of
  surfacing a bare HTTP status.
- Remote Markdown documents now render links against their source URL instead of the
  path of the temporary file they were downloaded to.

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
