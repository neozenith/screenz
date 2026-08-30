---
type: Architecture Decision
title: "Distribute via CI-built artifacts on the GitHub Releases page"
description: "Distribution machinery must match the audience that exists today"
tags: [distribution, release, ci]
status: accepted
accepted_on: 2026-08-28
plan_id: ADR6.1
provenance: "The screen-organiser planning session; distribution research into Gatekeeper and quarantine behaviour"
enforced_in:
  - ".github/workflows/release.yml"
  - "Makefile (build-all, dist, release)"
  - "docs/install.md"
generated: { by: human:neozenith, at: 2026-08-28T00:00:00Z }
---

> **Lens**: Distribution machinery must match the audience that exists today. CI-built Releases artifacts serve a single user; tap or store machinery waits for a real external audience.

## Relates to

- Split 2026-08-29: the TCC-client decision moved to [ADR-0017](0017-terminal-app-is-the-tcc-client.md)

## Problem

### Symptom

screenz must install on client laptops that have no App Store access and no Apple developer tooling, the same bar Rectangle clears via a cask.

### Pain point

A Homebrew tap needs a second repo and formula sha256 rewrites on every release, and notarisation needs an Apple account.
All of that serves an audience of zero for a single-user tool.

## Decision

### The lens

- **Given**: a single-user tool whose binary cross-compiles CGO-free from any runner OS, and CLI downloads (`gh`, `curl`) that never set the quarantine xattr
- **We prefer**: tag-triggered CI builds uploaded to the GitHub Releases page, over a Homebrew tap or committed per-arch binaries
- **Because**: Releases already gives versioned, downloadable artifacts straight from CI with no extra machinery and no Apple account
- **Unless**: the tool grows an external audience, at which point a prebuilt tap formula is the upgrade path

### In practice

- A `v*` tag runs `make check` then `make dist` on ubuntu-latest and uploads both darwin tarballs plus checksums.txt to the release.
- The repo stays binary-free; installation is `gh release download` or `curl` per docs/install.md.

## Consequences

### Pros

- Versioned artifacts with checksums, no second repo, no notarisation.

### Cons

- Browser downloads still hit Gatekeeper; the runbook documents the one-time xattr clear.
