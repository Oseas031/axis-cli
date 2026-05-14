# Skills Package Boundary

> Referenced by CLAUDE.md §4

## Owns

- On-demand knowledge skill discovery from `.axis/skills/`
- Skill loading (frontmatter parsing + content retrieval)
- Skill validation (directory structure, naming, frontmatter completeness)
- Layer 1 metadata injection (skill name/description list for System Prompt)
- `load_skill` tool (Agent-initiated Layer 2 full content retrieval)

## Must NOT Do

- **Never push skill content into provider prompts automatically** — Layer 1 injects metadata only; full content requires explicit Agent `load_skill` call
- **Never modify scheduler or contract semantics** — skills are informational, not behavioral
- **Never perform network access** — skills are local files only, no remote fetching
- **Never run background work** — discovery and loading are synchronous, on-demand operations
- **Never bypass path safety** — all skill paths must be validated within `.axis/skills/` boundary
