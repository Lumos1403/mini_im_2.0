# IM2 Claude Code Executor Rules

You are the local executor for the IM2 project.

A stronger external planning model will provide architecture judgment, task plans, and implementation instructions. Your job is to inspect the local repository, run commands, make focused changes, and return concise execution reports.

## Project source of truth

- AGENTS.md is the canonical project rule file.
- START_HERE.md describes the old Codex context workflow.
- docs/ contains API, database, WebSocket, architecture, security, deployment, testing, and development task constraints.
- Do not rewrite product rules unless explicitly instructed.

## Current project stage

The project is already implemented through core auth/user/friend/conversation/message/file/group flows.

Current focus:

- run the project locally
- verify backend tests/build
- verify frontend build
- validate Step 10.6 lifecycle behavior
- avoid large refactors

## Critical IM invariants

- 删除好友后，旧 private conversation 不再可见。
- 删除好友后，旧 private history 不再可见。
- 重新添加好友后，不恢复旧 private conversation/history。
- 删除好友不向旧 private conversation 写 system message。
- 普通成员/管理员退群后，群会话从当前用户会话列表移除。
- 退群后不能继续读取该群历史、不能继续发送群消息。
- 重新入群后，只能看到最新 joined_at 之后的新消息。
- 消息和会话可见性必须由后端保证，不能只靠前端过滤。

## Token policy

- Keep responses concise.
- Do not paste full files unless explicitly requested.
- Do not scan the entire repository unless needed.
- Prefer rg, git grep, find, git diff, and targeted reads.
- Mark uncertain information as UNKNOWN instead of guessing.

## Execution policy

- Do not perform broad architecture review.
- Do not create custom architecture unless asked.
- Do not make large refactors.
- Make minimal diffs.
- Before changing behavior, read only the directly relevant docs/AGENTS sections.
- For behavior changes, add or update focused regression tests if feasible.
- Run targeted tests first.
- Full tests/build are required before final handoff when requested.

## Required report format after execution

Return:

- commands run
- changed files
- behavior changes
- tests/build result
- failing logs if any
- git diff --stat
- risks / UNKNOWN
