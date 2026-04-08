---
name: changelog-entry
description: Draft a Keep-a-Changelog `[Unreleased]` section for go-glyph from git log since the last release tag. Invoke when preparing a release or after landing a notable change. User-only (mutates a release-critical file).
disable-model-invocation: true
---

# changelog-entry

You are preparing an `[Unreleased]` entry in `CHANGELOG.md` for **go-glyph**. The file follows [Keep a Changelog 1.1.0](https://keepachangelog.com/en/1.1.0/) and [SemVer](https://semver.org/).

## Procedure

1. **Find the last release tag**:
   ```bash
   git describe --tags --abbrev=0
   ```
   If the repo has no tags, fall back to `git log --oneline` and ask the user how far back to summarize.

2. **Collect the commits since that tag**:
   ```bash
   git log --no-merges --pretty=format:'%h %s' <tag>..HEAD
   ```
   Read `CHANGELOG.md` to see the existing `[Unreleased]` section — **never clobber entries already there**, only add to them.

3. **Bucket each commit** into the standard Keep-a-Changelog categories. Only emit buckets that have entries:
   - **Added** — new user-visible APIs, features, backends, examples.
   - **Changed** — behavior changes, dependency bumps, CI/build changes that affect contributors.
   - **Deprecated** — APIs marked deprecated this cycle.
   - **Removed** — APIs/files removed.
   - **Fixed** — bug fixes, correctness fixes.
   - **Security** — security-relevant fixes.

   Skip purely internal commits (test refactors, comment fixes, docs-only) *unless* they affect external behavior.

4. **Phrase entries in user-visible terms**. Not "refactor layoutImpl" but "Faster word-wrap for long paragraphs". Match the terse, imperative style of the existing changelog — see 1.6.1 / 1.6.2 / 1.6.3 for tone.

5. **Platform scope**: when a change only affects one backend or OS, say so in the entry: `Windows: …`, `Android: …`, `WASM: …`. This repo has strong per-platform history — don't hide it.

6. **Dependency bumps**: if multiple deps changed, consolidate into one "Update dependencies: X vN, Y vN, Z vN" line under **Changed**, matching the format already used in 1.6.3.

7. **Write to the file**: insert the new lines under the existing `## [Unreleased]` heading, preserving any entries the user has already added by hand. Never reorder or rewrite existing entries.

8. **Show the user the diff** of `CHANGELOG.md` after editing, and ask whether they want to:
   - Promote `[Unreleased]` to a numbered release (e.g., `[1.6.4] - <today>`).
   - Adjust any entries.

## Guardrails

- Do **not** run `git tag` or bump versions. This skill only drafts the prose.
- Do **not** guess at a version number. If the user asks to promote, ask whether it is a patch, minor, or major bump and let them choose.
- Do **not** invent changes. If a commit is too cryptic to categorize, include its short hash in a "Needs review" note at the end of your draft and let the user decide.
- If there are **zero** commits since the last tag, say so and exit without editing.
