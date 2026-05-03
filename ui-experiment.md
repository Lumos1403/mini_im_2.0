\---

description: Frontend UI/UX experiment workflow for IM2. Use this for large frontend visual/layout experiments while preserving backend/API/store contracts.

\---



You are in IM2 frontend UI experiment mode.



Goal:

\- Improve frontend visual design, layout, interaction quality, and component structure.

\- You may make larger frontend changes than ship-lite.

\- You must preserve existing backend contracts and business behavior.



Allowed focus:

\- frontend UI/UX

\- AuthLayout / ChatLayout structure

\- Chat.vue decomposition

\- visual components

\- particle backgrounds

\- button hover states

\- global search placement

\- right-side panel simplification

\- frontend state cleanup for logout/token failure



Hard limits:

1\. Do not modify backend code.

2\. Do not modify database migrations.

3\. Do not modify WebSocket protocol.

4\. Do not change backend API fields or paths.

5\. Do not introduce a large UI framework.

6\. Do not implement frontend-only permission rules.

7\. Do not use nickname/avatar/title to identify conversations or users.

8\. Use conversation\_id, user\_id, group\_id as stable identifiers.

9\. Do not create a second WebSocket implementation.

10\. Do not bypass existing api/http.ts, stores, or token handling.

11\. Do not submit frontend/dist or frontend/node\_modules changes.

12\. Do not read secrets, .env files, private keys, or tokens.



Frontend architecture:

\- API calls live in frontend/src/api.

\- State lives in Pinia stores.

\- Views compose components.

\- WebSocket events stay centralized in ws store.

\- UI components should not directly create raw axios instances.

\- Components should not directly manage localStorage token logic.



For UI-X:

\- P0 is login-state isolation.

\- Unauthenticated users must never see chat data.

\- Logout/token failure must clear auth/chat/friend/group/search/ws state.

\- WebSocket must connect only after login and disconnect on logout/token failure.

\- AuthLayout and ChatLayout/AppShell separation is allowed and encouraged.

\- Chat.vue may be decomposed into layout components.

\- Right-side panel should only show current conversation context.

\- Global search should move to top/global entry.

\- Login page may use strong particle/magnetic visual effects.

\- Chat page may use subtle mouse-reactive ambience only.



Execution style:

\- You may propose and implement larger frontend diffs.

\- Prefer cohesive frontend structure over tiny patches.

\- Still avoid unrelated business logic rewrites.

\- Before coding, provide a plan.

\- After coding, run npm run build.

\- Report changed files, behavior changes, build result, git diff --stat, risks, and manual test checklist.

