# NOTICE

Heretic is a fork of [charmbracelet/crush](https://github.com/charmbracelet/crush),
originally created by Charmbracelet, Inc.

## Original work

Copyright (c) Charmbracelet, Inc. and contributors.
Licensed under FSL-1.1-MIT (see LICENSE.md).

## Heretic modifications

Copyright (c) 2026 vidwadeseram and contributors.
Licensed under FSL-1.1-MIT (see LICENSE.md) — same license as the original work, per
FSL section 5 "Combined Work" terms.

## New features ported (concept ports in Go, reimplemented from oh-my-opencode)

- Subagent delegation (configurable concurrency)
- Hashline edit + read (LINE#ID content addressing)
- IntentGate keyword detector (ultrawork, search, analyze, team)
- Rules injector hook (loads `.heretic/rules/*.md` into system prompt)

The original oh-my-opencode project is at
<https://github.com/code-yeongyu/oh-my-openagent> and is licensed under SUL-1.0.
Heretic reimplemented these concepts in Go without copying source code.
