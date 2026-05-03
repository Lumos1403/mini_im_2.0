# Backend-for-Frontend Contract Pack

本契约包面向后续前端 UI/UX 优化 agent。它只描述前端需要遵守的后端契约和当前前端消费方式，不包含后端实现细节，不包含任何真实 token、密码、密钥或测试账号。

## A. 基础信息

- API base path：`/api`。前端开发环境 Vite 已把 `/api` 代理到 `http://localhost:8081`；也可通过 `VITE_API_BASE_URL` 指定 Axios baseURL。
- backend 默认端口：`8081`，来自 `SERVER_PORT` 默认值。
- WebSocket 地址：本地 `ws://localhost:8081/ws?token=<access_token>`；前端可用 `VITE_WS_URL` 覆盖。
- 认证方式：HTTP 使用 `Authorization: Bearer <access_token>`；WebSocket 使用 query 参数 `token=<access_token>`。
- Access Token 放在哪里：当前前端保存在 `localStorage.access_token`，Axios request interceptor 每次请求自动加 Bearer。
- Refresh Token 使用方式：当前前端保存在 `localStorage.refresh_token`，调用 `POST /api/auth/refresh` 时放入 JSON body 的 `refresh_token`。注意：当前 `frontend/src/api/http.ts` 只加 Authorization，不会自动 refresh。
- 统一成功响应结构：`{ code: 0, message: "success", data: ... }`。
- 统一失败响应结构：`{ code: number, message: string, data: null }`。前端应优先显示 `message`，鉴权类错误另走清理流程。
- 分页响应结构：普通分页为 `{ list, total, page, page_size }`；聊天历史 cursor 分页为 `{ list, next_cursor, has_more, limit }`。
- 文件下载例外：`GET /api/files/:file_id/download` 成功返回附件流，不是统一 JSON；失败仍是统一 JSON。
- 常见错误码展示：`20001/20002` token 无效或过期，应尝试 refresh；`20003` refresh 无效，应退出登录并清空状态；`30002` 用户名已存在，注册页展示；`30003` 账号或密码错误，登录页展示；`40002/40004` 已是好友/申请待处理，好友面板展示；`40003` 被对方拉黑，好友申请展示；`50002` 消息不可撤回；`50004` 消息无权访问，应清空当前无权会话；`50005` 消息内容无效；`60002` 群权限不足；`60009` 群成员被禁言；`70002` 文件无权限；`70003` 文件过大。

## B. 登录 / 注册 / 退出 / 刷新契约

- `POST /api/auth/register`：不需要登录。request：`username`、`password`、`nickname`。`response.data`：`user_id`、`username`、`nickname`。当前前端封装：`frontend/src/api/auth.ts register`，store：`auth.register`。注册成功不自动写 token，当前 UI 提示用户去登录。失败时注册页展示后端 `message`。业务文档要求注册后默认 Agent 好友，但当前可见契约无法确认，见 UNKNOWN。
- `POST /api/auth/login`：不需要登录。request：`username`、`password`。`response.data`：`access_token`、`refresh_token`、`expires_in`、`user{ user_id, username, nickname, avatar_url }`。封装：`authApi.login`，store：`auth.login` 调用 `setAuth` 写入 localStorage 和 `auth.user`。登录成功后应进入 `/chat` 并连接 WebSocket。
- `POST /api/auth/refresh`：不需要 Access Token。request：`refresh_token`。`response.data`：新 `access_token`、新 `refresh_token`、`expires_in`。封装：`authApi.refresh`，store：`auth.refresh`。成功后必须覆盖本地两个 token。失败或返回 `20003` 时必须断开 WS、清空所有用户态并跳到登录层。
- `POST /api/auth/logout`：需要登录。request：只需要 Authorization header。`response.data`：空对象。封装：`authApi.logout`，store：`auth.logout` 当前只 `clearAuth`，不会自动断开 WS 或清空 chat/friend/group。UI 重构时 logout 入口必须先调用后端，最终无论成功失败都清空本地状态。
- `GET /api/users/me`：需要登录。`response.data`：当前用户基础信息。封装：`authApi.getMe`，store：`auth.loadMe`。用于刷新页面恢复用户信息。
- 未登录访问 `/chat`：前端应跳转登录层。当前 `frontend/src/router/index.ts` 没有路由守卫，UI 重构必须补齐。
- logout、token 失效或 refresh 失败后：清理 `auth` token/user，断开 WebSocket，清空会话列表、当前会话、消息列表、好友列表、群成员/申请、搜索面板、本地弹窗和右侧详情面板。

## C. 会话和消息契约

- `GET /api/conversations`：需要登录。query：`page` 默认 1，`page_size` 默认 20、最大 100。`response.data` 普通分页，item 字段：`conversation_id`、`conversation_type(private/group)`、`title`、`avatar_url`、`peer_user_id`、`peer_nickname`、`peer_avatar_url`、`group_id`、`group_no`、`group_status`、`last_message|null`、`unread_count`、`is_pinned`、`is_muted`。当前代码中 `last_message` 可能为 `null`，前端必须兼容。封装：`frontend/src/api/conversation.ts listConversations`；store：`chat.loadConversationList`。
- `GET /api/conversations/:conversation_id/messages`：需要登录。query：`cursor` 可空，`limit` 默认 30、最大 100。首次不传 cursor；加载更早消息传当前最旧的 `message_id`。`response.data`：`list`、`next_cursor`、`has_more`、`limit`。message 字段：`client_msg_id`、`message_id`、`conversation_id`、`sender_id`、`sender_nickname`、`sender_avatar_url`、`sender_group_role`、`message_type`、`content`、`extra_json`、`send_status`、`created_at`。封装：`listMessages`；store：`chat.loadCurrentMessages`、`chat.loadOlderMessages`。
- `DELETE /api/conversations/:conversation_id/messages`：需要登录。清空当前用户视角。响应空对象。封装：`clearConversationMessages`；store：`chat.clearCurrentConversation`。UI 成功后清空当前消息、`hasMore=false`、当前会话 `last_message=null`、`unread_count=0`。
- `DELETE /api/conversations/:conversation_id/messages/:message_id`：需要登录。删除当前用户视角的单条消息。响应空对象。封装：`deleteMessage`；store：`chat.deleteVisibleMessage`。UI 成功后只从当前用户消息列表移除，不影响对方。
- `POST /api/messages/:message_id/recall`：需要登录。只能撤回自己 5 分钟内、`send_status=sent` 的消息。`response.data`：`message_id`、`editable_until`。封装：`recallMessage`；store：`chat.recallVisibleMessage`。发送方本地显示“你撤回了一条消息”和“重新编辑”；接收方通过 WS 移除原消息且无提示。
- `GET /api/messages/:message_id/recall-edit-cache`：需要登录，仅发送者且缓存有效期内可读。`response.data`：`message_id`、`content`。封装：`getRecallEditCache`；store：`chat.reEditMessage`，成功后填回输入框。
- UI 关键规则：`conversation_id` 是定位会话的唯一可靠字段。禁止用 `nickname`、`avatar_url`、`title` 匹配会话；好友兜底只能用 `peer_user_id`，群聊只能用 `group_id/conversation_id`。消息发送主要走 WebSocket，不走 HTTP。前端本地先插入 `sending`；收到 `ack` 后改为 `sent` 并替换服务端 `message_id`；收到 `failed` 后改为 `failed/failed_blocked`，`failed_blocked` 显示红色感叹号和后端失败文案。清空、删除只影响当前用户视角；撤回会让双方原消息消失。

## D. WebSocket 契约

- 连接地址：`ws://localhost:8081/ws?token=<access_token>`；生产用 `wss`。
- token 传递：query 参数 `token`，必须是 Access Token。无效或过期会被拒绝连接。
- envelope 字段：`seq`、`type`、`data`、`timestamp`。`seq` 由前端生成，用于匹配 ack/failed；`timestamp` 毫秒。
- `chat.message.send` request data：`conversation_id`、`client_msg_id`、`message_type`、`content`、`extra_json`。`client_msg_id` 必填，建议 UUID，最大 64；text 内容 trim 后不能为空；file 内容为 `file_id`。
- `chat.message.ack` data：`client_msg_id`、`message_id`、`conversation_id`、`send_status`、`server_time`。
- `chat.message.receive` data：`client_msg_id`、`message_id`、`conversation_id`、`sender_id`、`sender_nickname`、`sender_avatar_url`、`sender_group_role`、`message_type`、`content`、`extra_json`、`send_status`、`created_at`。私聊昵称/头像/角色可能为空；群聊应包含发送者群身份。
- `chat.message.failed` data：`client_msg_id`、可选 `message_id`、`conversation_id`、`send_status`、`code`、`message`、可选 `server_time`。常见 code：`failed_blocked`、`not_friends`、`conversation_not_found`、`duplicate_client_msg_id_conflict`、`group_member_muted`、`group_dissolved`。
- `chat.message.recalled` data：`message_id`、`conversation_id`、`recalled_by`、`recalled_at`。前端只移除原消息；接收方不显示提示。
- ping/pong：服务端会发 WebSocket 协议级 ping；客户端库自动 pong。业务层也支持 `type=ping` 返回 `type=pong`。
- 前端入口：`frontend/src/stores/ws.ts connect`、`disconnect`、`sendChatMessage`、`handleEnvelope`。`handleEnvelope` 分发到 `chat.applyAck`、`applyFailed`、`applyReceive`、`applyRecalled`。
- 安全规则：`sender_id` 必须以后端 WebSocket 鉴权结果为准，前端不能传或伪造。发送消息必须等待 ack/failed 后更新最终状态。logout、token 失效、切换用户后必须主动断开 WebSocket。

## E. 好友契约

- `GET /api/users/search`：需要登录。query：`keyword`、`page`、`page_size`。返回分页，item：`user_id`、`username`、`nickname`、`avatar_url`、`bio`。封装：`api/user.ts searchUsers`；store：`friend.search`。UI 用于搜 user_id 或昵称，不展示敏感字段。
- `POST /api/friends/requests`：需要登录。body：`to_user_id`、`message`。返回：`request_id`、`status`。封装：`createFriendRequest`；store：`friend.sendFriendRequest`，群成员弹窗也复用它。
- `GET /api/friends/requests`：需要登录。query：`direction=received|sent`、`page`、`page_size`。返回分页，item：`request_id`、`from_user_id`、`to_user_id`、`user`、`message`、`status`、`created_at`、`updated_at`。封装：`listFriendRequests`；store：`friend.loadReceivedRequests`。
- `POST /api/friends/requests/:request_id/accept` / `reject`：需要登录。无 body，响应空对象。封装：`acceptFriendRequest`、`rejectFriendRequest`；store：`friend.acceptRequest`、`friend.rejectRequest`。同意后应刷新好友和会话。
- `GET /api/friends`：需要登录。返回分页，item：`friend_user_id`、`nickname`、`avatar_url`、`bio`、`conversation_id`、`is_blocked_by_me`、`created_at`、`updated_at`。封装：`listFriends`；store：`friend.loadFriends`。打开聊天优先使用 `conversation_id`。
- `DELETE /api/friends/:user_id`：需要登录。响应空对象。封装：`deleteFriend`；store：`friend.removeFriend` 已调用 `chat.removeConversation(conversationID)` 和 `chat.removePrivateConversationByPeer(userID)`，再刷新好友/会话。
- `POST /api/friends/:user_id/block` / `DELETE /api/friends/:user_id/block`：需要登录。响应空对象。封装：`blockFriend`、`unblockFriend`；store：`friend.block`、`friend.unblock`。好友列表拉黑状态以后端 `is_blocked_by_me` 为准。
- 删除好友 UI 规则：成功后必须移除好友和对应 private conversation；如果当前打开该会话，清空当前聊天窗口或切到无选中状态。删除后旧会话、旧历史、旧搜索结果不能继续展示；重新添加好友后从新 `conversation_id` 的空白会话开始。

## F. 群聊契约

- `POST /api/groups`：body：`name`、可选 `avatar_url`。返回：`group_id`、`group_no`、`conversation_id`。封装：`createGroup`；store：`group.create` 会刷新会话并选中。
- `GET /api/groups/search`：query：`keyword`。返回分页，item：`group_id`、`group_no`、`conversation_id`、`owner_id`、`name`、`avatar_url`、`max_members`、`allow_member_invite`、`status`、`is_member`。封装：`searchGroups`；store：`group.search`。
- `POST /api/groups/:group_id/join-requests`：body：`message`。返回：`request_id`、`status`。封装：`createJoinRequest`；store：`group.applyJoin`。
- `GET /api/groups/:group_id/join-requests`：query：`page`、`page_size`，仅群主/管理员。返回分页，item：`request_id`、`group_id`、`user_id`、`user`、`message`、`status`、`handled_by`、`created_at`、`updated_at`。封装：`listJoinRequests`；store：`group.loadJoinRequests`。
- `POST /api/groups/join-requests/:request_id/accept` / `reject`：返回：`request_id`、`group_id`、`conversation_id`、`user_id`、`status`。封装：`acceptJoinRequest`、`rejectJoinRequest`；store：`group.accept`、`group.reject`。
- `GET /api/groups/:group_id/members`：返回分页，成员字段固定为：`user_id`、`nickname`、`avatar_url`、`bio`、`role(owner/admin/member)`、`mute_until`、`joined_at`、`status`、`friendship_status(self/friend/not_friend/pending_sent/pending_received)`。封装：`listGroupMembers`；store：`group.loadMembers`。
- 管理接口：`POST /api/groups/:group_id/admins/:user_id`、`DELETE /api/groups/:group_id/admins/:user_id`、`POST /api/groups/:group_id/members/:user_id/mute` body `mute_until`、`DELETE /api/groups/:group_id/members/:user_id/mute`、`PUT /api/groups/:group_id/settings` body 可选 `allow_member_invite`、`max_members`、`DELETE /api/groups/:group_id`、`POST /api/groups/:group_id/leave`。前端封装分别是 `setGroupAdmin`、`unsetGroupAdmin`、`muteGroupMember`、`unmuteGroupMember`、`updateGroupSettings`、`dissolveGroup`、`leaveGroup`；store 对应 `setAdmin`、`unsetAdmin`、`mute`、`unmute`、`saveInviteSetting`、`saveMaxMembers`、`dissolve`、`leave`。
- UI 规则：`sender_group_role` 由后端返回，前端只能使用该字段显示群主/管理员标识；缺失按 `member` 处理。退出群聊后当前用户会话列表移除该群；若当前打开该群，应清空聊天窗口或切到无选中。重新入群后只能看到本次 `joined_at` 之后消息。

## G. 文件契约

- `POST /api/files/upload`：需要登录，`multipart/form-data` 字段 `file`。返回：`file_id`、`original_name`、`file_size`、`mime_type`。默认最大 50MB，可由配置调整。封装：`api/file.ts uploadFile`；store：`chat.prepareUploadedFileMessage`。
- `GET /api/files/:file_id/download`：需要登录，成功返回附件流。封装：`downloadFile`；store：`chat.downloadVisibleFile`。下载必须带登录态，未授权展示 `70002` 文案。
- 文件消息 WebSocket：`message_type="file"`，`content` 必须是 `file_id` 字符串；前端可带 `extra_json{ file_id,file_name,file_size,mime_type }` 做本地 optimistic 展示，但后端会按上传记录生成最终 `extra_json`，前端不要信任自己传入的文件名和大小作为最终事实。
- 文件消息显示字段：优先读 `extra_json.file_id/file_name/file_size/mime_type`，缺失时 `content` 作为下载 file_id 兜底。当前 `Chat.vue` 展示文件名、大小/MIME、下载按钮。
- 当前群聊是否允许文件上传：不允许。`Chat.vue canUploadFile` 已在 group conversation 下禁用文件按钮；后端群消息当前也只接受 text。
- 文件搜索结果：当前 `SearchPanel` 点击只进入 `conversation_id` 对应会话，不直接下载。UI 重构建议保持“进入会话”，不从搜索结果绕过上下文直接下载。

## H. 搜索契约

- `GET /api/search/messages?keyword=&page=&page_size=`：需要登录。`keyword` 必填且 trim 后非空；`page` 默认 1；`page_size` 默认 20、最大 100。返回分页。item：`message_id`、`conversation_id`、`conversation_type`、`sender_id`、`sender_nickname`、`sender_avatar_url`、`message_type`、`content`、`created_at`。封装：`api/search.ts searchMessages`；组件：`components/search/SearchPanel.vue`。
- `GET /api/search/files?keyword=&page=&page_size=`：需要登录。query 同上。返回分页。item：`file_id`、`original_name`、`file_size`、`mime_type`、`uploader_id`、`uploader_nickname`、`message_id`、`conversation_id`、`conversation_type`、`created_at`。封装：`searchFiles`；组件：`SearchPanel.vue`。
- UI 规则：搜索结果点击只需要进入 `conversation_id` 对应会话，不强制定位 `message_id`。权限过滤由后端完成，前端不能做伪权限过滤，也不能展示已缓存但后端不再返回的旧结果。删除好友、退群、重新入群、清空、删除、撤回后的不可见内容不会由后端返回。

## I. 前端状态清理契约

- auth store：`frontend/src/stores/auth.ts` 有 `clearAuth`，会清空 token 和 user。`logout` 只调用后端并 clearAuth，缺少联动清理其他 store。
- ws store：`frontend/src/stores/ws.ts` 有 `disconnect`，应在 logout、token 失效、refresh 失败、切换用户前调用。
- chat store：有 `resetActiveConversation`、`removeConversation`、`removePrivateConversationByPeer`、`removeGroupConversationByGroupID`、`stopRecallNoticeExpiryTimer`，缺少 `resetAll`。UI 重构应补充清空 `conversations`、`activeConversationID`、`messages`、`draft`、`hasMore`、`seenMessageKeys`、error 和定时器的方法。
- friend store：只有 `clearMessages`，缺少 `resetAll`。应补充清空 `friends`、`receivedRequests`、`searchResults`、loading/operating/error/notice。
- group store：有 `clearMessages`、`closeMemberDrawer`、`closeMemberProfile`，缺少 `resetAll`。应补充清空 `searchResults`、`joinRequests`、`members`、`membersGroupID`、抽屉/弹窗、选中成员、好友申请 loading 和所有状态文案。
- search panel state：当前是 `SearchPanel.vue` 本地 state，不在 store。logout/切换用户应关闭面板并重置 keyword、tab、results、page、total、error。
- 当前会话/消息列表：必须由 chat reset 清空，不能等下一次接口失败。
- 群成员弹窗/资料弹窗/右侧详情面板：必须关闭并清空选中对象，避免切换用户后显示上一用户数据。

## J. 前端当前结构摘要

- 登录/注册页面：`frontend/src/views/Login.vue`、`Register.vue`，样式为 scoped CSS，登录成功跳 `/chat`，注册成功仅提示再登录。
- 当前 `Chat.vue` 职责过重：会话列表、聊天头、消息列表、输入框、文件上传、清空、搜索抽屉、好友面板、群面板、群成员抽屉和资料弹窗都在这里组合。
- 会话列表：在 `Chat.vue` 左侧直接渲染，数据来自 `chat.conversations`，点击 `chat.selectConversation(conversation_id)`。
- 消息列表：在 `Chat.vue` 中部直接渲染，支持加载更早、sending、failed 红色感叹号、文件卡片、删除、撤回、群身份 badge。
- 好友面板：`components/friend/FriendPanel.vue`，调用 `friend` store，包含好友列表、申请、搜索用户、打开聊天、删除、拉黑。
- 群面板：`components/group/GroupPanel.vue`，包含创建、搜索、入群、审批、成员管理、设置、解散、退出；成员抽屉/资料弹窗是 `GroupMemberDrawer`、`GroupMemberProfileModal`。
- 搜索组件：`components/search/SearchPanel.vue`，本地维护搜索状态，调用 `api/search.ts`，点击结果 emit `open-conversation`。
- 当前右侧面板问题：右侧固定同时塞入 FriendPanel 和 GroupPanel，和群成员抽屉/搜索抽屉/资料弹窗并列，状态分散；UI 重构时应把“当前详情面板”做成单一受控状态。
- 样式主要分布：各 Vue 文件内 scoped CSS；`App.vue` 有顶栏全局壳样式；没有统一设计系统或全局 reset。
- 当前路由：`router/index.ts` 没有登录守卫，`App.vue` 顶部导航总是显示登录/注册/资料/聊天。

## K. 后续 UI 优化不能破坏的行为清单

- 未登录不能看到聊天数据，访问 `/chat` 应进入登录层。
- 登录后建立唯一 WebSocket 连接。
- logout、token 失效、refresh 失败后断开 WebSocket 并清空所有用户态数据。
- 发送消息必须走 `sending -> ack(sent)` 或 `sending -> failed/failed_blocked`。
- `failed_blocked` 必须显示红色感叹号和拒收含义。
- 删除好友后旧 private conversation 不可见，旧历史和旧搜索结果不可继续展示。
- 退群后旧 group conversation 不可见，重新入群只看本次之后消息。
- 搜索结果只显示后端返回内容，不展示本地旧缓存。
- 文件下载必须走鉴权接口，不能拼公开 URL。
- 群成员身份显示只使用 `sender_group_role`。
- 所有会话定位使用 `conversation_id`，禁止用昵称、头像、标题匹配。

## L. UNKNOWN

- 业务规则要求注册后自动拥有默认 Agent 好友；当前 `AgentService.EnsureDefaultAgentFriend` 未暴露可确认的前端可见结果，无法保证登录后一定有 Agent 会话。
- 文档要求 Axios 自动 refresh，但当前 `http.ts` 未实现 401/业务错误码自动刷新；具体拦截策略需 UI 重构时补齐。
- `GET /api/conversations` DTO 有 `last_message` 字段，但当前 service 未明显填充，运行时可能长期为 `null`；前端必须兼容。
- 消息类型模型包含 `emoji`、`system`，但当前 WebSocket 发送校验只确认 `text` 和 `file`，群聊当前只允许 `text`。
- WebSocket 自动重连、跨多标签页连接策略、切换用户时的统一 reset 入口当前未实现。
- 搜索结果是否要支持定位到具体 `message_id` 当前无后端/前端完整契约，本轮只确认进入会话。
