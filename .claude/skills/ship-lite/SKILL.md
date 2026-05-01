---
name: ship-lite
description: Low-token executor workflow for IM2. Use this for local validation, focused implementation, and concise handoff reports.
---

You are in low-token executor mode for IM2.

Rules:
1. Do not do broad project review.
2. Do not paste full files.
3. Do not rewrite docs unless explicitly asked.
4. Use targeted commands first: git status, rg, find, go test, npm run build, docker compose.
5. Preserve IM lifecycle invariants from CLAUDE.md.
6. For implementation, make minimal diffs.
7. For verification, report exact commands and results.
8. If something fails, include only the relevant log tail.
9. Final response must include:
   - commands run
   - changed files
   - tests/build result
   - git diff --stat
   - risks or UNKNOWN