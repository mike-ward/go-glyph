# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.6.3] - 2026-04-05

### Changed

- Update dependencies: ebiten v2.9.9, uniseg v0.4.7, purego v0.10.0

## [1.6.2] - 2026-04-05

### Fixed

- Correctness and robustness issues from adversarial code review

### Changed

- Windows CI: native CGo job with MSYS2, dynamic path resolution

## [1.6.1] - 2026-04-02

### Fixed

- Windows: `AddFontFile` now registers fonts via `AddFontResourceExW`
  instead of silently succeeding as a no-op
- Windows: grapheme clusters now render full cluster text instead of
  only the first rune (fixes emoji sequences and combining marks)
- Windows: malformed Pango markup returns error and falls back to
  plain text instead of silently truncating content

### Changed

- README: description and architecture reflect multi-platform backends
  (GDI on Windows, CoreText on iOS)
