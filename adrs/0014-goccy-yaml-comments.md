# ADR 0014: Use goccy/go-yaml so profile comments survive saves

Plan ID: ADR5.1 | Date: 2026-08-28 | Status: accepted

## Decision

Profiles are read strictly (unknown keys error) and written with
goccy/go-yaml's comment map. Saves are append-only so existing comment
paths stay valid, use block style only, and write through a temp-file
rename so an interrupted save cannot destroy the file.

## Why

Profiles are hand-commented specs. yaml.v3 is archived and drops comments
on struct marshal; goccy preserves them. Blank lines do not survive a save
and flow-style comments are lost, hence block style everywhere.
