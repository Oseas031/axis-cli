# Documentation Changelog

> Append-only record of documentation changes. Parseable format: `## [date] operation | subject`

## [2026-05-15] research | Bystander Effect in Multi-Agent Reasoning
- Added: `docs/research/bystander-effect-multi-agent-reasoning-2026-05-15.md` — 多Agent认知偷懒量化 + Axis防御策略
- Added vigil-cf6700: CandidatePool 强制异构策略 (P1)
- Added vigil-cc7aed: Subagent反调隔离强化 (P1)
- Added vigil-ad418d: Prompt装配Lead Anchor防御 (P1)
- Added vigil-604e20: Self-Judgement Sovereignty Gap检测 (P2)
- Completed: vigil-80ce6c

## [2026-05-15] research | Harness Engineering (Claude Code + Codex comparison)
- Added: `reports/research/harness-engineering-claude-code-guide-2026-05-15.md` — Book1 full text (PDF→MD)
- Added: `reports/research/harness-comparison-claude-code-vs-codex-2026-05-15.md` — Book2 full text (PDF→MD)
- Added: `reports/research/harness-books-digest-2026-05-15.md` — Research digest with Axis mapping
- Completed: vigil-b1d12b (Claude Code vs Codex Harness 设计哲学比较)
- Completed: vigil-e06388 (Harness Engineering — Claude Code 设计指南)

## [2026-05-15] architecture | Harness-informed design items
- Added vigil-378ebb: Prompt 分层装配链 (P1)
- Added vigil-07c165: 权限三态化 ask 语义 (P1)
- Added vigil-8a284d: Compact 语义恢复 (P1)
- Added vigil-d6f1e8: 中断账本闭合 (P1)

## [2026-05-15] fix | vigil CLI --tags alias
- Modified: `cmd/axis/vigil_cmd.go` — added `--tags` as hidden alias for `--tag`, supports comma-separated



## [2026-05-15] init | Knowledge Base Infrastructure
- Added: `docs/PURPOSE.md` — knowledge base direction and intent
- Added: `docs/CHANGELOG.md` — documentation timeline (this file)
- Added: `docs/WIKI-SCHEMA.md` — machine-parseable Agent editing rules (constraints + triggers + acceptance)
- Added: `docs/research/llm-wiki-knowledge-base-2026-05-15.md` — LLM Wiki pattern research
- Added: `.axis/skills/docs-knowledge-base/SKILL.md` — knowledge base maintenance skill
- Added: `docs/lessons/` — 5 structured lessons with executable verification
- Added: `docs/architecture/README.md`, `docs/specs/README.md`, `docs/research/README.md` — sub-indexes
- Added: `docs/status/history/` — milestone completions + architecture diagnosis (split from current-progress)
- Added: `internal/boundary_test.go` — 3 boundary verification tests (kernel/memory/contextpack)
- Added: `cmd/axis/docs.go` — `axis docs lint` command (orphan/dead-link/no-frontmatter detection)
- Modified: `docs/README.md` — compressed to top-level router (~225 tok, -86%)
- Modified: `docs/status/current-progress.md` — split from 28KB to 3.3KB (-88%)
- Modified: `CLAUDE.md` — added §2.1 RDM exemption for routine doc maintenance
- Modified: 5x `BOUNDARY.md` — added executable verification sections
- Modified: 6x architecture docs — added YAML frontmatter
- Moved: `docs/deprecated/` → `archive/deprecated-docs/` (-31% docs volume)

## [2026-05-14] research | Multiple papers
- Added: `docs/research/harness-categorical-architecture-2026-05-14.md`
- Added: `docs/research/token-hallucination-detection-2026-05-14.md`
- Added: `docs/research/symbolic-equivalence-partitioning-2026-05-14.md`
- Added: `docs/research/variational-posterior-guidance-2026-05-14.md`
- Added: `docs/research/executable-agentic-memory-2026-05-14.md`
- Added: `docs/research/gemma4-mtp-evaluation-2026-05-14.md`

## [2026-05-14] architecture | Agent evaluation + design principles
- Added: `docs/architecture/agent-evaluation-principles.md`
- Added: `docs/architecture/agent-design-first-principles.md`
- Modified: `docs/architecture/dialectical-development-methodology.md`

## [2026-05-14] spec | Coding Agent + Actor Comm + Vigil + CLI Execution
- Added: `docs/specs/coding-agent/` (requirements + design + tasks + first-principles)
- Added: `docs/specs/actor-comm/design.md`
- Added: `docs/specs/vigil/` (requirements + design + tasks)
- Added: `docs/specs/cli-execution-semantics/` (requirements + design + tasks)
