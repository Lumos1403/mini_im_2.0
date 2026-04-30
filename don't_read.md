提示词

# 通用

请不要依赖聊天历史。本轮是一个新的独立开发任务。

请先阅读：

1. AGENTS.md
2. START_HERE.md
3. docs/10_DEVELOPMENT_TASKS.md 中 Step X
4. 与本轮任务相关的 docs 文件

# 具体

## 第一轮提示词：让 Codex 初始化项目

你可以直接复制下面这段给 Codex。

你现在只完成 Step 1：项目初始化。

请先阅读：

1. AGENTS.md
2. docs/06_BACKEND_ARCHITECTURE.md
3. docs/07_FRONTEND_ARCHITECTURE.md
4. docs/10_DEVELOPMENT_TASKS.md 中 Step 1 的内容

当前任务目标：
初始化一个前后端分离的即时通讯系统项目骨架。

后端要求：

1. 使用 Go。
2. 使用 Gin。
3. 预留 gorilla/websocket 目录结构。
4. 创建标准后端目录：
   - cmd/server
   - internal/api/handler
   - internal/api/router
   - internal/middleware
   - internal/service
   - internal/repository/mysql
   - internal/repository/redis
   - internal/model
   - internal/ws
   - internal/config
   - internal/pkg/response
   - internal/pkg/errors
   - internal/pkg/logger
   - internal/pkg/jwt
   - internal/pkg/snowflake
   - migrations
5. 创建基础 main.go。
6. 创建健康检查接口：GET /api/health。
7. 创建统一响应结构。
8. 创建基础配置读取结构。
9. 暂时不要实现注册、登录、聊天、好友、群聊。

前端要求：

1. 使用 Vue 3 + Vite。
2. 使用 TypeScript。
3. 创建基础目录：
   - src/api
   - src/components
   - src/views
   - src/router
   - src/stores
   - src/utils
   - src/types
4. 创建基础页面：
   - Login.vue
   - Register.vue
   - Chat.vue
5. 暂时只搭建页面骨架，不实现真实业务。

Docker 要求：

1. 创建 docker-compose.yml。
2. 包含 mysql、redis 服务。
3. 暂时不要求完整生产部署。

严格限制：

1. 不要开发登录功能。
2. 不要开发聊天功能。
3. 不要开发好友功能。
4. 不要开发群聊功能。
5. 不要引入 Kafka、Elasticsearch、MinIO。
6. 不要把所有代码写在 main.go。

请先输出实现计划和预计修改/新增文件列表，等我确认后再开始写代码。

## 第二轮提示词：用户注册登录

你现在只完成 Step 2：用户注册登录与认证系统。

请先阅读：

1. AGENTS.md
2. START_HERE.md
3. docs/01_PROJECT_DEVELOPMENT_SPEC.md 中“用户与账号系统”部分
4. docs/03_DATABASE_DESIGN.md 中 users、user_profiles 相关表
5. docs/04_API_SPEC.md 中 auth 相关接口
6. docs/08_SECURITY_AND_CONCURRENCY.md 中 Token 和密码安全部分

当前任务目标：
实现用户注册、登录、刷新 Token、退出登录、鉴权中间件。

功能要求：

1. 注册字段：
   - username
   - password
   - nickname
2. user_id 使用雪花算法生成。
3. password 使用 bcrypt 加密保存。
4. 登录使用 username + password。
5. 登录成功后返回 Access Token 和 Refresh Token。
6. Access Token 使用 JWT，有效期 15 分钟。
7. Refresh Token 存 Redis，有效期 7 天。
8. 实现刷新 Token 接口。
9. 实现退出登录接口。
10. 实现 Gin 鉴权中间件。
11. 实现获取当前用户信息接口。
12. 注册成功后创建 user_profiles 记录。
13. 注册成功后暂时预留创建 Agent 好友的 service 方法，但本次不实现 Agent 聊天。

数据库要求：

1. 创建 users 表 migration。
2. 创建 user_profiles 表 migration。
3. 添加必要唯一索引：
   - users.user_id
   - users.username

接口要求：

1. POST /api/auth/register
2. POST /api/auth/login
3. POST /api/auth/refresh
4. POST /api/auth/logout
5. GET /api/users/me

严格限制：

1. 不要开发好友系统。
2. 不要开发聊天系统。
3. 不要开发 WebSocket。
4. 不要开发群聊。
5. 不要修改前端复杂页面，只需要能调用接口即可。

请先输出实现计划、数据库 migration 设计、预计修改文件，等我确认后再写代码。

## 第三轮提示词：用户资料模块 done

你现在只完成 Step 3：用户资料模块。

请先阅读：

1. AGENTS.md
2. START_HERE.md,这两个md文档会帮助你理解项目
3. docs/01_PROJECT_DEVELOPMENT_SPEC.md 中“用户资料”部分
4. docs/03_DATABASE_DESIGN.md 中 user_profiles 表
5. docs/04_API_SPEC.md 中 user profile 相关接口

当前任务目标：
实现用户查看和修改个人资料。

功能要求：

1. 用户可以查看自己的资料。
2. 用户可以修改：
   - avatar_url
   - nickname
   - gender
   - bio
3. 当前版本不做内容审核。
4. 但必须保留字段：
   - profile_status
   - profile_review_reason
5. 当前版本不限制修改频率。

接口要求：

1. GET /api/users/me/profile
2. PUT /api/users/me/profile

权限要求：

1. 必须登录。
2. 用户只能修改自己的资料。

前端要求：

1. 创建或完善 Profile.vue。
2. 能展示和修改头像、昵称、性别、个性签名。
3. 表单提交成功后刷新页面状态。

严格限制：

1. 不要开发好友功能。
2. 不要开发聊天功能。
3. 不要开发文件上传头像功能，avatar_url 暂时用字符串输入。
4. 不要修改认证逻辑，除非发现明显 bug。

请先输出实现计划和预计修改文件，等我确认后再写代码。

## 七、第四轮提示词：好友系统done

你现在只完成 Step 4：好友系统。

请先阅读：

1. AGENTS.md
2. START_HERE.md,这两个md文档会帮助你理解项目
3. docs/01_PROJECT_DEVELOPMENT_SPEC.md 中“好友系统”部分
4. docs/03_DATABASE_DESIGN.md 中 friend_requests、friendships、block_relations 表
5. docs/04_API_SPEC.md 中 friends 相关接口

当前任务目标：
实现搜索用户、好友申请、同意/拒绝、好友列表、删除好友、拉黑/解除拉黑。

功能要求：

1. 用户可以通过 user_id 精确搜索用户。
2. 用户可以通过 nickname 模糊搜索用户。
3. 添加好友必须发送好友申请。
4. 对方同意后建立好友关系。
5. 好友关系使用一条记录表示，user_id_1 小于 user_id_2。
6. 删除好友后双方好友列表都移除。
7. 删除好友后允许重新添加。
8. 删除好友时生成系统提示事件或预留 system message 创建逻辑。
9. 拉黑是单向关系。
10. A 拉黑 B 后，B 给 A 发消息时后续应返回 failed_blocked。
11. 本次只实现拉黑关系数据，不实现聊天发送失败逻辑。

数据库要求：

1. 创建 friend_requests migration。
2. 创建 friendships migration。
3. 创建 block_relations migration。
4. 添加必要唯一索引。

接口要求：

1. GET /api/users/search
2. POST /api/friends/requests
3. GET /api/friends/requests
4. POST /api/friends/requests/{id}/accept
5. POST /api/friends/requests/{id}/reject
6. GET /api/friends
7. DELETE /api/friends/{user_id}
8. POST /api/friends/{user_id}/block
9. DELETE /api/friends/{user_id}/block

严格限制：

1. 不要开发 WebSocket。
2. 不要开发消息发送。
3. 不要开发群聊。
4. 不要开发文件功能。
5. 不要跳过服务端权限校验。

请先输出实现计划、数据库表设计确认、预计修改文件，等我确认后再写代码。

## 第五轮提示词：会话系统done

你现在只完成 Step 5：会话系统。

请先阅读：

1. AGENTS.md
2. START_HERE.md,这两个md文档会帮助你理解项目
3. docs/01_PROJECT_DEVELOPMENT_SPEC.md 中“会话系统”部分
4. docs/03_DATABASE_DESIGN.md 中 conversations、conversation_members、conversation_user_states 表
5. docs/04_API_SPEC.md 中 conversations 相关接口

当前任务目标：
实现会话表、会话成员表、用户会话状态表，以及获取会话列表接口。

功能要求：

1. 支持 private 和 group 两种会话类型。
2. 当前重点实现 private 会话。
3. 好友关系建立后应能创建或恢复 private conversation。
4. conversation_members 保存会话成员。
5. conversation_user_states 保存用户视角下的会话状态。
6. 支持字段：
   - is_deleted
   - cleared_at
   - last_read_message_id
   - last_read_at
   - unread_count
   - is_pinned
   - is_muted
7. 实现获取当前用户会话列表接口。

接口要求：

1. GET /api/conversations

严格限制：

1. 不要开发 WebSocket 消息收发。
2. 不要开发群聊。
3. 不要开发消息删除和撤回。
4. 不要开发文件消息。
5. 只实现会话基础能力。

请先输出实现计划和预计修改文件，等我确认后再写代码。

## 第六轮提示词：WebSocket 基础done

你现在只完成 Step 6：WebSocket 基础设施。

请先阅读：

1. AGENTS.md
2. START_HERE.md,这两个md文档会帮助你理解项目
3. docs/05_WEBSOCKET_PROTOCOL.md
4. docs/06_BACKEND_ARCHITECTURE.md 中 ws 目录设计
5. docs/08_SECURITY_AND_CONCURRENCY.md 中 WebSocket 在线状态部分

当前任务目标：
实现 WebSocket 鉴权、连接管理、在线状态、心跳机制和基础事件格式。

功能要求：

1. WebSocket 地址：/ws?token=xxx
2. 连接时校验 Access Token。
3. 校验通过后获取 user_id。
4. 将 user_id 与连接注册到 ws Hub。
5. Redis 写入在线状态。
6. 断开连接后清理在线状态。
7. 实现 ping/pong 心跳。
8. 定义统一 WebSocket 消息结构：
   - seq
   - type
   - data
   - timestamp
9. 支持基础测试事件：
   - ping
   - pong
10. 暂时不实现真实聊天消息发送。

严格限制：

1. 不要开发消息入库。
2. 不要开发聊天业务。
3. 不要开发群聊。
4. 不要让 WebSocket 层直接写复杂数据库逻辑。
5. 在线状态必须写 Redis，不要只存在内存。

请先输出实现计划和预计修改文件，等我确认后再写代码。

## 第七轮提示词：单聊文本消息done

你现在只完成 Step 7：单聊文本消息。

请先阅读：

1. AGENTS.md
2. START_HERE.md,这两个md文档会帮助你理解项目
3. docs/01_PROJECT_DEVELOPMENT_SPEC.md 中“消息系统”部分
4. docs/03_DATABASE_DESIGN.md 中 messages、message_user_states 表
5. docs/05_WEBSOCKET_PROTOCOL.md 中 chat.message.send / ack / receive
6. docs/08_SECURITY_AND_CONCURRENCY.md 中消息并发设计部分

当前任务目标：
实现好友之间的单聊文本消息发送、入库、实时推送、历史消息分页加载。

功能要求：

1. 只实现 text 消息。
2. 用户必须是会话成员才能发送。
3. 单聊时必须检查好友关系。
4. 必须检查是否被对方拉黑。
5. 如果被拉黑：
   - 不推送给对方
   - 不进入对方聊天记录
   - 返回 failed_blocked
   - 前端显示红色感叹号
6. 正常消息写入 messages 表。
7. 在线接收方通过 WebSocket 收到 chat.message.receive。
8. 发送方收到 chat.message.ack。
9. 离线用户上线后通过历史消息接口分页拉取。
10. 聊天记录必须分页，禁止一次性加载全部。

接口要求：

1. GET /api/conversations/{id}/messages

WebSocket 事件：

1. chat.message.send
2. chat.message.ack
3. chat.message.receive
4. chat.message.failed

严格限制：

1. 不要开发文件消息。
2. 不要开发群聊消息。
3. 不要开发撤回。
4. 不要开发清空聊天记录。
5. 不要引入消息队列，先同步处理，但代码要预留 dispatcher/service 抽象。

请先输出实现计划、消息流转说明、预计修改文件，等我确认后再写代码。

## 第八轮提示词：消息删除、清空、撤回done

你现在只完成 Step 8：消息删除、清空聊天记录、5 分钟撤回、重新编辑缓存。

请先阅读：

1. AGENTS.md
2. START_HERE.md,这两个md文档会帮助你理解项目
3. docs/01_PROJECT_DEVELOPMENT_SPEC.md 中“消息删除、清空、撤回”部分
4. docs/03_DATABASE_DESIGN.md 中 messages、message_user_states、conversation_user_states 表
5. docs/04_API_SPEC.md 中 messages 相关接口
6. docs/05_WEBSOCKET_PROTOCOL.md 中撤回相关事件

当前任务目标：
实现用户对消息的个人删除、会话清空、5 分钟内撤回和重新编辑缓存。

功能要求：

1. 用户可以删除单条消息。
2. 删除只影响当前用户。
3. 删除后前端立即隐藏该消息。
4. 被当前用户删除的消息不参与该用户搜索。
5. 用户可以清空某个会话聊天记录。
6. 清空通过 conversation_user_states.cleared_at 实现。
7. 清空后只影响当前用户。
8. 用户只能撤回自己发送的消息。
9. 只能撤回发送后 5 分钟内的消息。
10. 撤回后双方聊天窗口中的原消息消失。
11. 接收方不显示撤回提示。
12. 发送方显示“你撤回了一条消息”。
13. 发送方可以点击重新编辑。
14. 重新编辑内容存 Redis，TTL 5 分钟。
15. 5 分钟后不可重新编辑。
16. 撤回消息不参与搜索。

接口要求：

1. DELETE /api/conversations/{id}/messages/{message_id}
2. DELETE /api/conversations/{id}/messages
3. POST /api/messages/{message_id}/recall
4. GET /api/messages/{message_id}/recall-edit-cache

严格限制：

1. 不要做群消息撤回的复杂特殊逻辑，先复用会话成员校验。
2. 不要物理删除 messages 表记录。
3. 不要影响对方的单向删除状态。
4. 不要实现导出聊天记录。

请先输出实现计划、数据状态变化说明、预计修改文件，等我确认后再写代码。

## 第九轮提示词：文件消息done

你现在只完成 Step 9：文件上传、文件下载鉴权、文件消息。

请先阅读：

1. AGENTS.md
2. START_HERE.md,这两个md文档会帮助你理解项目
3. docs/01_PROJECT_DEVELOPMENT_SPEC.md 中“文件消息系统”部分
4. docs/03_DATABASE_DESIGN.md 中 files 表
5. docs/04_API_SPEC.md 中 files 相关接口
6. docs/08_SECURITY_AND_CONCURRENCY.md 中文件安全部分

当前任务目标：
实现文件上传、文件下载鉴权、文件消息发送。

功能要求：

1. 支持各类文件上传。
2. 不提供在线预览。
3. 单文件大小限制为 50MB，可配置。
4. 文件存储到本地 uploads 目录。
5. Docker 部署时 uploads 目录需要可挂载。
6. files 表保存文件元信息。
7. 文件下载必须登录。
8. 只有文件上传者或文件所属会话成员可以下载。
9. 文件消息 message_type = file。
10. extra_json 中保存：
    - file_id
    - file_name
    - file_size
    - mime_type

接口要求：

1. POST /api/files/upload
2. GET /api/files/{file_id}/download

WebSocket 要求：

1. 支持发送 file 类型消息。
2. 文件消息复用已有消息发送流程。

严格限制：

1. 不要实现文件预览。
2. 不要引入 MinIO 或 OSS。
3. 不要允许未登录用户直接访问 uploads 静态目录。
4. 不要破坏 text 消息逻辑。

请先输出实现计划、文件鉴权方案、预计修改文件，等我确认后再写代码。

## Step 10：群聊基础功能

请不要依赖聊天历史。本轮是一个新的独立开发任务。

你现在只完成 Step 10：群聊基础功能。

请先阅读：

1. `AGENTS.md`
2. `START_HERE.md`
3. `docs/01_PROJECT_DEVELOPMENT_SPEC.md` 中“群聊系统”部分
4. `docs/03_DATABASE_DESIGN.md` 中 `groups`、`group_members`、`group_join_requests` 表
5. `docs/04_API_SPEC.md` 中群聊接口部分
6. `docs/05_WEBSOCKET_PROTOCOL.md` 中群消息事件部分
7. `docs/06_BACKEND_ARCHITECTURE.md` 中 group / message / ws 模块职责
8. `docs/08_SECURITY_AND_CONCURRENCY.md` 中群权限和消息权限相关内容
9. `docs/10_DEVELOPMENT_TASKS.md` 中 Step 10：群聊基础功能
10. `docs/11_TESTING_ACCEPTANCE.md` 中群聊验收标准

当前任务目标：

实现 Step 10 群聊基础功能，打通群聊主链路，包括：

1. 创建群聊。
2. 群号搜索。
3. 申请入群。
4. 群主或管理员审批入群。
5. 群成员管理。
6. 设置管理员和取消管理员。
7. 群成员禁言和解除禁言。
8. 修改群设置。
9. 群主解散群聊。
10. 群 text 消息发送、入库、实时推送和历史分页。

本轮重点是群聊基础能力和最小前端验证入口，不实现完整群成员 GUI，不实现群主 / 管理员 Badge，不实现群成员资料弹窗和群内添加好友按钮，这些放到 Step 10.5。

### 功能边界

本轮必须实现：

```txt
POST /api/groups
GET /api/groups/search
POST /api/groups/{group_id}/join-requests
GET /api/groups/{group_id}/join-requests
POST /api/groups/join-requests/{request_id}/accept
POST /api/groups/join-requests/{request_id}/reject
GET /api/groups/{group_id}/members
POST /api/groups/{group_id}/admins/{user_id}
DELETE /api/groups/{group_id}/admins/{user_id}
POST /api/groups/{group_id}/members/{user_id}/mute
DELETE /api/groups/{group_id}/members/{user_id}/mute
PUT /api/groups/{group_id}/settings
DELETE /api/groups/{group_id}
```

群消息必须复用现有：

```txt
chat.message.send
chat.message.ack
chat.message.receive
GET /api/conversations/{conversation_id}/messages
```

群消息发送时必须服务端校验：

```txt
当前用户是群成员
群未解散
当前用户未被禁言
message_type = text
sender_id 来自 WebSocket 鉴权上下文
```

### 数据库要求

请根据 `docs/03_DATABASE_DESIGN.md` 检查是否已有：

```txt
groups
group_members
group_join_requests
```

如果 migration 已存在，优先复用。

如果缺失字段或索引，先说明需要新增 migration，并列出原因和字段，再实现。

不要随意改动已有消息、单聊、文件相关表结构。

### 后端要求

1. 后端遵守 handler / service / repository 分层。
2. 群权限必须在 service 层校验。
3. 群主、管理员、普通成员权限必须严格按照文档实现。
4. 同意入群申请必须使用事务。
5. 创建群聊必须同时创建 group、group_members、conversation、conversation_members、conversation_user_states。
6. 群消息不得破坏现有单聊消息逻辑。
7. 群消息发送失败必须返回明确 `chat.message.failed`。
8. 群消息推送字段必须满足 `docs/05_WEBSOCKET_PROTOCOL.md`。
9. 群成员列表字段必须满足 `docs/04_API_SPEC.md`，为 Step 10.5 做准备。

### 前端最小要求

本轮只做最小可测试入口：

1. 可以创建群聊。
2. 可以搜索群号。
3. 可以申请入群。
4. 群主或管理员可以处理入群申请。
5. 可以进入群聊会话。
6. 可以发送群 text 消息。
7. 被禁言时发送失败并显示错误。
8. 群解散后不能继续发送。
9. 可以查看基础群成员列表。

完整群成员 GUI、身份 Badge、成员资料弹窗、群内添加好友放到 Step 10.5。

### 严格限制

1. 不要实现复杂群公告。
2. 不要实现群文件空间。
3. 不要实现语音通话。
4. 不要实现复杂邀请流程。
5. 不要重写好友系统。
6. 不要重写单聊消息逻辑。
7. 不要破坏文件消息、撤回、删除、清空逻辑。
8. 不要只靠前端隐藏按钮做权限控制。
9. 不要用 nickname 或 avatar 判断用户身份。
10. 不要引入 Kafka、Elasticsearch、微服务、Kubernetes。
11. 不要开发 Step 10.5 的完整 GUI 能力。

### 文档要求

如果本轮修改了接口、字段、WebSocket 事件或前端结构，必须同步更新：

```txt
docs/03_DATABASE_DESIGN.md
docs/04_API_SPEC.md
docs/05_WEBSOCKET_PROTOCOL.md
docs/07_FRONTEND_ARCHITECTURE.md
docs/10_DEVELOPMENT_TASKS.md
docs/11_TESTING_ACCEPTANCE.md
```

只更新实际发生变化的文档。

### 现在请先输出

在写代码前，请先输出：

1. 你理解的本轮任务目标。
2. 你准备阅读的文件。
3. 群聊数据模型和现有 migration 检查计划。
4. 群权限校验方案。
5. 群消息如何复用现有 message / conversation / ws 逻辑。
6. 预计新增或修改的文件。
7. 是否需要新增 migration。
8. 是否需要修改 API 文档或 WebSocket 文档。
9. 需要我确认的问题。

在我确认前，不要直接写代码。

## Step 10.5：群聊成员 GUI 和身份标识

请不要依赖聊天历史。本轮是一个新的独立开发任务。

你现在只完成 Step 10.5：群聊成员 GUI 和身份标识。

请先阅读：

1. `AGENTS.md`
2. `START_HERE.md`
3. `docs/01_PROJECT_DEVELOPMENT_SPEC.md` 中“群聊系统”部分
4. `docs/04_API_SPEC.md` 中群成员、群消息、好友申请相关接口
5. `docs/05_WEBSOCKET_PROTOCOL.md` 中群消息事件字段
6. `docs/07_FRONTEND_ARCHITECTURE.md` 中群聊交互、群成员 GUI、群身份标识部分
7. `docs/10_DEVELOPMENT_TASKS.md` 中 Step 10.5：群聊成员 GUI 和身份标识
8. `docs/11_TESTING_ACCEPTANCE.md` 中群成员 GUI 和身份标识验收
9. Step 10 已实现的群聊相关代码
10. 当前前端 chat store、conversation store、ws store、group store、Chat.vue 和群聊相关组件

当前任务目标：

在 Step 10 群聊基础能力已经完成的前提下，补齐群聊前端体验：

1. 群聊消息中显示群主 / 管理员身份标识。
2. 支持查看群成员列表。
3. 支持点击群成员查看资料。
4. 支持从群成员资料弹窗中添加好友。
5. 群内成员相关操作根据角色和好友状态显示正确按钮。

本阶段重点是前端 GUI 和响应字段接入，不重写 Step 10 群聊后端主链路。

### 功能要求

本轮必须实现：

```txt
群主 owner 消息昵称旁显示金色“群主”
管理员 admin 消息昵称旁显示绿色“管理员”
普通成员 member 不显示特殊标识
历史群消息和实时群消息都支持身份标识
群聊页面提供查看群成员入口
群成员列表展示头像、昵称、user_id、bio、role、mute_until、friendship_status
点击群成员头像或昵称打开资料弹窗
资料弹窗中可以对非好友成员发起好友申请
群内添加好友复用现有 POST /api/friends/requests
```

按钮状态必须遵守：

```txt
self：不显示添加好友
friend：显示已是好友或发消息
not_friend：显示添加好友
pending_sent：显示申请中
pending_received：提示对方已申请添加你
```

### 前端结构要求

1. 所有 HTTP 请求必须放在 `src/api`。
2. 群成员状态可以放在 `stores/group.ts` 或现有 chat store 中，但不能散落在页面组件里。
3. 建议新增组件：
   - `GroupRoleBadge`
   - `GroupMemberList`
   - `GroupMemberDrawer` 或 `GroupMemberModal`
   - `GroupMemberProfileModal`
4. `Chat.vue` 只负责组合组件，不要堆积大量业务逻辑。
5. WebSocket 群消息接收逻辑应复用现有消息处理流程。
6. 不要在多个页面重复写群消息解析逻辑。
7. 必须使用 `user_id`、`group_id`、`conversation_id` 做唯一标识。
8. 不允许使用 nickname 或 avatar_url 匹配用户或会话。

### 后端字段补齐要求

如果 Step 10 已经返回所需字段，本轮不需要改后端。

如果发现字段不足，可以只做 DTO / 查询补齐，不改核心群聊业务逻辑。

可能需要确认的字段：

```txt
GET /api/groups/{group_id}/members:
- user_id
- nickname
- avatar_url
- bio
- role
- mute_until
- joined_at
- status
- friendship_status

群消息历史:
- sender_nickname
- sender_avatar_url
- sender_group_role

chat.message.receive:
- sender_nickname
- sender_avatar_url
- sender_group_role
```

如果补充 API 响应字段，必须更新 `docs/04_API_SPEC.md`。

如果补充 WebSocket 字段，必须更新 `docs/05_WEBSOCKET_PROTOCOL.md`。

不允许新增 migration，除非发现现有数据库确实缺必要字段；如果需要 migration，必须先说明原因。

### 好友接口复用要求

群内添加好友必须复用现有好友申请接口：

```txt
POST /api/friends/requests
```

要求：

1. 不新增重复好友接口。
2. 不直接在前端修改好友关系。
3. 不允许前端伪造好友状态。
4. 添加好友成功后，可以刷新群成员列表或局部更新 `friendship_status`。
5. 如果后端返回已经是好友或已有待处理申请，前端必须正确展示状态。

### 严格限制

1. 不要重新实现 Step 10 群聊后端主链路。
2. 不要重写单聊消息逻辑。
3. 不要重写好友系统。
4. 不要实现复杂群公告。
5. 不要实现群文件空间。
6. 不要实现语音通话。
7. 不要引入新的大型 UI 框架。
8. 不要用昵称判断身份。
9. 不要用头像匹配用户。
10. 不要通过前端隐藏按钮代替服务端权限校验。
11. 不要修改数据库，除非先说明原因。
12. 不要破坏现有单聊、文件消息、撤回、删除、清空功能。

### 文档要求

如果本轮新增或修改了响应字段，需要更新：

```txt
docs/04_API_SPEC.md
docs/05_WEBSOCKET_PROTOCOL.md
```

如果新增了前端组件、store、页面结构，需要更新：

```txt
docs/07_FRONTEND_ARCHITECTURE.md
```

如果任务状态或拆分有变化，需要更新：

```txt
docs/10_DEVELOPMENT_TASKS.md
docs/11_TESTING_ACCEPTANCE.md
```

只更新实际发生变化的文档。

### 现在请先输出

在写代码前，请先输出：

1. 你理解的本轮任务目标。
2. 你准备阅读的文件。
3. 当前 Step 10 已实现字段是否满足 Step 10.5。
4. 是否需要补充后端 DTO 或接口响应字段。
5. 前端组件拆分方案。
6. 群成员列表和资料弹窗状态管理方案。
7. 群内添加好友如何复用现有好友接口。
8. 预计新增或修改的文件。
9. 是否需要修改 API 文档或 WebSocket 文档。
10. 需要我确认的问题。

在我确认前，不要直接写代码。

## Step 10.6：关系变更后的会话生命周期修复

请不要依赖聊天历史。本轮是一个新的独立开发任务。

你现在只完成 Step 10.6：删除好友、退出群聊后的会话列表和历史消息可见性修复。

本轮是 Step 10.5 人工测试后发现的数据一致性问题修复，不是前端页面重构任务。前端右侧功能栏臃肿、页面结构优化、完整 UI/UX 整理，放到搜索功能实现之后的统一前端整理阶段处理。本轮前端只允许做必要的状态刷新和最小交互修复。

\---

#### 请先阅读

1. `AGENTS.md`
2. `START\\\\\\\_HERE.md`
3. `docs/01\\\\\\\_PROJECT\\\\\\\_DEVELOPMENT\\\\\\\_SPEC.md` 中好友系统、会话系统、消息系统、群聊系统相关部分
4. `docs/03\\\\\\\_DATABASE\\\\\\\_DESIGN.md` 中以下表：

   - `friendships`
   - `conversations`
   - `conversation\\\\\\\_members`
   - `conversation\\\\\\\_user\\\\\\\_states`
   - `messages`
   - `message\\\\\\\_user\\\\\\\_states`
   - `groups`
   - `group\\\\\\\_members`
   - `group\\\\\\\_join\\\\\\\_requests`
5. `docs/04\\\\\\\_API\\\\\\\_SPEC.md` 中好友、会话、消息、群聊相关接口
6. `docs/05\\\\\\\_WEBSOCKET\\\\\\\_PROTOCOL.md` 中消息事件、会话同步相关说明
7. `docs/07\\\\\\\_FRONTEND\\\\\\\_ARCHITECTURE.md` 中当前聊天页、store、会话列表、群聊相关说明
8. `docs/08\\\\\\\_SECURITY\\\\\\\_AND\\\\\\\_CONCURRENCY.md` 中会话权限、消息权限、群权限相关内容
9. `docs/10\\\\\\\_DEVELOPMENT\\\\\\\_TASKS.md` 中 Step 4、Step 5、Step 7、Step 8、Step 10、Step 10.5 的相关记录
10. `docs/11\\\\\\\_TESTING\\\\\\\_ACCEPTANCE.md` 中好友、会话、消息、群聊相关验收标准
11. 当前后端好友、会话、消息、群聊相关代码
12. 当前前端聊天页、好友 store、会话 store、群聊 store、WebSocket store、Chat.vue 相关代码

\---

#### 当前人工测试发现的问题

###### 问题 1：删除好友后旧单聊会话和历史消息仍然存在

当前表现：

1. A 删除 B 为好友后，好友关系确实被删除。
2. 但是 A 与 B 的单聊会话仍然出现在页面或重新刷新后仍可恢复。
3. 旧的单聊消息仍然可以通过历史消息接口看到。
4. A 和 B 重新加回好友后，之前已经删除好友前的旧聊天记录仍然存在。

期望行为：

1. 删除好友后，双方好友关系删除。
2. 删除好友后，该 private conversation 应从双方会话列表中移除。
3. 删除好友后，该 private conversation 的旧消息不应再被任一方通过历史消息接口读取。
4. 删除好友后，该 private conversation 的旧消息不应再参与消息搜索或文件搜索。
5. 重新加回好友后，应创建新的单聊会话或保证旧会话历史不会恢复。
6. 重新加回好友后，聊天记录应从空白状态重新开始。

###### 问题 2：退出群聊后旧群会话仍然存在

当前表现：

1. 用户退出某个群聊后，页面上该群会话仍然存在，或者刷新后仍可能看到旧群会话。
2. 用户退出群聊后，仍可能看到该群历史消息。

期望行为：

1. 用户退出群聊后，只影响当前用户的群成员身份和当前用户会话列表。
2. 退出群聊后，该群会话应从当前用户会话列表移除。
3. 退出群聊后，当前用户不能继续通过历史消息接口读取该群消息。
4. 退出群聊后，当前用户不能继续发送该群消息。
5. 退出群聊后，群聊本身、群成员中的其他成员、群历史消息、其他成员的会话列表都不能被删除或破坏。
6. 群消息数据必须保留在数据库中，因为群聊是多人场景，不能因为某个成员退出而删除整个群的消息。
7. 如果该用户未来重新入群，默认只应看到重新入群之后产生的新消息，不应自动恢复退出前或退出期间的旧消息。

\---

#### 当前任务目标

修复关系变更后的会话生命周期问题，使删除好友、退出群聊后的会话列表、历史消息、搜索结果、发送权限符合真实聊天场景。

本轮重点是后端数据一致性和权限过滤，前端只做最小必要刷新。

\---

#### 功能要求：删除好友

删除好友时必须满足：

1. `DELETE /api/friends/{user\\\\\\\_id}` 仍然是删除好友入口，不要随意新增重复接口。
2. 删除好友必须继续删除双方好友关系。
3. 删除好友必须同时处理两人之间的 private conversation 生命周期。
4. 删除好友后，双方都不能再在会话列表中看到这段 private conversation。
5. 删除好友后，双方都不能再通过 `GET /api/conversations/{conversation\\\\\\\_id}/messages` 读取这段 private conversation 的旧消息。
6. 删除好友后，双方都不能再通过搜索接口搜到这段 private conversation 的旧消息或文件消息。
7. 删除好友后，双方都不能再向该 private conversation 发送消息。
8. 删除好友后，如果双方重新加好友，必须从新的空白聊天开始，旧消息不能恢复。
9. 删除好友相关数据库操作必须使用事务，避免好友关系删除成功但会话/消息状态未处理的半完成状态。
10. 如果现有 `AcceptFriendRequest` / `EnsurePrivateConversationInTx` 会复用旧 private conversation，必须修复该逻辑，避免重新加回好友后恢复旧聊天记录。

###### 删除好友时关于“物理删除”和“软删除”的要求

用户期望是：删除好友后，旧聊天记录不应再保留为可恢复状态。

请先阅读现有表结构和外键关系，再选择最安全的实现方式：

1. 如果可以安全物理删除该 private conversation 相关数据，可以删除：

   - 该 private conversation 的 `messages`
   - 相关 `message\\\\\\\_user\\\\\\\_states`
   - 相关 `conversation\\\\\\\_user\\\\\\\_states`
   - 相关 `conversation\\\\\\\_members`
   - 该 private `conversations` 记录
2. 如果物理删除会破坏文件消息、外键、审计或现有代码假设，则可以采用“终止旧会话 + 永久不可见”的软删除方案，但必须保证：

   - 会话列表永远不返回该旧 private conversation
   - 历史消息接口永远不返回该旧 private conversation 的旧消息
   - 搜索接口永远不返回该旧 private conversation 的旧消息和文件
   - 重新加好友不会恢复旧 conversation\_id 的旧消息
3. 不允许只在前端隐藏会话。
4. 不允许只修改 `conversation\\\\\\\_user\\\\\\\_states.is\\\\\\\_deleted` 但重新加回好友后又恢复旧消息。
5. 不允许破坏单条消息删除、清空会话、撤回、文件消息、群消息逻辑。

如果发现当前 docs 与本轮需求存在冲突，例如旧文档要求删除好友后保留系统提示消息，请先在实现计划中指出冲突，并说明本轮按“删除好友后旧单聊会话和旧历史不再恢复”的新需求调整哪些文档。

\---

#### 功能要求：退出群聊

如果当前项目已经有“退出群聊”接口，请修复现有接口。

如果当前项目还没有普通成员退出群聊接口，请新增一个最小接口。建议接口为：

```txt
POST /api/groups/{group\\\\\\\_id}/leave
```

也可以根据现有项目 REST 风格选择更合适的路径，但必须先在计划中说明原因，并同步更新 API 文档。

退出群聊必须满足：

1. 当前用户必须是该群当前有效成员。
2. 群主不能通过普通退出接口直接退出群聊。

   - 当前项目如果还没有群主转让能力，群主应只能解散群聊。
   - 不要在本轮新增群主转让功能。
3. 管理员和普通成员可以退出群聊。
4. 退出群聊不能删除 `groups` 记录。
5. 退出群聊不能删除其他成员的 `group\\\\\\\_members` 记录。
6. 退出群聊不能删除 `messages` 表中的群消息。
7. 退出群聊不能影响其他成员的会话列表和历史消息。
8. 退出群聊后，当前用户在 `group\\\\\\\_members` 中应变为非有效成员状态，或者用现有字段表达已退出状态。
9. 退出群聊后，当前用户的 `conversation\\\\\\\_user\\\\\\\_states` 应使该群会话从当前用户会话列表消失。
10. 退出群聊后，当前用户不能再读取该群历史消息。
11. 退出群聊后，当前用户不能再发送该群消息，应返回明确错误。
12. 如果当前用户未来重新入群，应该以新的入群时间作为可见起点，不能自动恢复退出前或退出期间的旧消息。
13. 如果现有 `group\\\\\\\_members.joined\\\\\\\_at` 可用于控制重新入群后的消息可见范围，可以复用并在重新入群时更新。
14. 如果现有表结构缺少必要字段，例如无法区分 active / left / removed，必须先说明是否需要新增 migration，列出原因和字段后再实现。

\---

#### 历史消息接口要求

必须检查：

```txt
GET /api/conversations/{conversation\\\\\\\_id}/messages
```

并确保：

1. 当前用户不是 private conversation 成员或该 private conversation 已因删除好友终止时，不能读取旧消息。
2. 当前用户不是群聊当前有效成员时，不能读取群消息。
3. 当前用户退出群聊后，不能读取退出前、退出后任何群消息。
4. 当前用户重新入群后，只能读取重新入群时间之后的新消息。
5. 当前用户清空会话、删除单条消息、消息撤回的过滤规则不能被破坏。
6. 文件消息的下载鉴权不能因为本轮修改被绕过。

\---

#### 会话列表接口要求

必须检查：

```txt
GET /api/conversations
```

并确保：

1. 删除好友后，双方不再看到对应 private conversation。
2. 退出群聊后，当前用户不再看到对应 group conversation。
3. 群聊其他成员仍然能看到该 group conversation。
4. 重新加好友后，如果创建新 private conversation，会话列表显示新的空白会话。
5. 重新入群后，会话列表可以重新显示该群，但旧消息不能恢复。

\---

#### 搜索接口前置要求

如果搜索功能当前尚未实现，只需要确保本轮新增的数据状态和查询规则为后续搜索预留正确过滤条件，并更新文档说明。

如果搜索功能当前已经实现或已有部分实现，则必须确保：

1. 删除好友后的旧 private conversation 消息不出现在消息搜索结果中。
2. 删除好友后的旧 private conversation 文件消息不出现在文件搜索结果中。
3. 退出群聊后的群消息不出现在退出用户的搜索结果中。
4. 群聊其他有效成员仍然可以搜索自己可见范围内的群消息。

\---

#### 前端最小修复要求

本轮不是前端页面重构任务。

只允许做以下最小前端修复：

1. 删除好友成功后：

   - 从好友列表移除该用户。
   - 从会话列表移除对应 private conversation。
   - 如果当前正在打开该 conversation，清空当前聊天窗口或切换到无选中状态。
   - 刷新会话列表和好友列表。
2. 退出群聊成功后：

   - 从会话列表移除对应 group conversation。
   - 如果当前正在打开该 group conversation，清空当前聊天窗口或切换到无选中状态。
   - 刷新会话列表和群相关状态。
3. 不重构右侧功能栏。
4. 不做 UI 美化。
5. 不新增复杂弹窗设计。
6. 不改变现有 Chat.vue 大布局。
7. 不把后端权限判断挪到前端。

如果必须新增一个“退出群聊”按钮才能完成最小测试入口，请只做最小按钮接入，不做页面结构重构。

\---

#### 后端重点检查文件

请根据实际仓库阅读，不要凭文件名猜测。以下是本轮重点检查方向，最终以你实际读到的文件为准：

```txt
backend/internal/service/friend\\\\\\\_service.go
backend/internal/service/group\\\\\\\_service.go
backend/internal/service/conversation\\\\\\\_service.go
backend/internal/service/message\\\\\\\_service.go
backend/internal/repository/mysql/friend\\\\\\\_repository.go
backend/internal/repository/mysql/group\\\\\\\_repository.go
backend/internal/repository/mysql/conversation\\\\\\\_repository.go
backend/internal/repository/mysql/message\\\\\\\_repository.go
backend/internal/api/handler/friend\\\\\\\_handler.go
backend/internal/api/handler/group\\\\\\\_handler.go
backend/internal/api/handler/conversation\\\\\\\_handler.go
backend/internal/api/handler/message\\\\\\\_handler.go
backend/internal/api/router/router.go
backend/internal/model/\\\\\\\*.go
backend/internal/pkg/errors/errors.go
backend/migrations/\\\\\\\*.sql
```

当前仓库中已经确认存在 `FriendService.DeleteFriend` 和 `FriendHandler.DeleteFriend`，并且 `GroupHandler` 当前需要检查是否已有普通成员退出群聊接口。不要创建重复逻辑。

\---

#### 前端重点检查文件

请根据实际仓库阅读，不要凭文件名猜测。以下是本轮重点检查方向，最终以你实际读到的文件为准：

```txt
frontend/src/views/Chat.vue
frontend/src/api/\\\\\\\*.ts
frontend/src/stores/\\\\\\\*.ts
frontend/src/types/\\\\\\\*.ts
frontend/src/components/\\\\\\\*\\\\\\\*/\\\\\\\*.vue
```

重点检查：

1. 删除好友的前端调用在哪里。
2. 会话列表状态在哪里维护。
3. 当前激活会话在哪里维护。
4. 群聊退出或群成员操作入口在哪里。
5. 删除好友或退出群聊后是否只是弹提示，没有同步移除会话状态。

\---

#### 数据库和事务要求

1. 删除好友 + 旧 private conversation 生命周期处理必须在同一事务内完成。
2. 退出群聊 + 当前用户 membership 状态 + 当前用户会话状态处理必须在同一事务内完成。
3. 不允许出现接口返回成功但刷新后旧会话又出现的情况。
4. 如果新增 migration，必须说明为什么现有字段不足。
5. migration 必须可重复按顺序执行，不要修改已发布旧 migration，除非项目当前明确允许重写历史 migration。
6. 不要随意删除群消息数据。
7. 不要随意删除文件实体或 uploads 文件，除非确认没有任何消息引用且本轮需求明确要求；本轮默认不处理物理文件清理。

\---

#### 权限要求

删除好友：

1. 只能删除自己与目标用户之间的好友关系。
2. 不能删除其他人的好友关系。
3. 删除好友后，双方都不能继续向旧 private conversation 发消息。
4. 删除好友后，双方不能读取旧 private conversation 历史。

退出群聊：

1. 只能退出自己所在的群。
2. 不能替别人退出群聊。
3. 群主不能通过普通退出接口退出。
4. 已退出成员不能再读取群消息。
5. 已退出成员不能再发送群消息。
6. 群聊其他成员不受影响。

\---

#### 严格限制

1. 不要重构右侧功能栏。
2. 不要做前端 UI 大改。
3. 不要开发搜索功能本身，除非当前项目已经有搜索代码且必须补过滤条件。
4. 不要开发群主转让。
5. 不要开发复杂群公告。
6. 不要开发群文件空间。
7. 不要开发语音通话。
8. 不要重写好友系统。
9. 不要重写消息系统。
10. 不要破坏单聊 text/file 消息。
11. 不要破坏撤回、重新编辑、单向删除、清空聊天记录。
12. 不要删除群聊数据库消息。
13. 不要只靠前端隐藏按钮实现权限控制。
14. 不要引入 Kafka、Elasticsearch、Meilisearch、MinIO、OSS、微服务、Kubernetes。
15. 不要凭空新增与现有风格不一致的接口。
16. 不要在 handler 中写复杂业务逻辑。
17. 不要让 WebSocket 层直接操作复杂数据库逻辑。

\---

#### 文档要求

如果本轮修改了业务规则、接口、字段、数据库状态或验收标准，必须同步更新相关文档。

重点可能涉及：

```txt
docs/01\\\\\\\_PROJECT\\\\\\\_DEVELOPMENT\\\\\\\_SPEC.md
docs/03\\\\\\\_DATABASE\\\\\\\_DESIGN.md
docs/04\\\\\\\_API\\\\\\\_SPEC.md
docs/05\\\\\\\_WEBSOCKET\\\\\\\_PROTOCOL.md
docs/07\\\\\\\_FRONTEND\\\\\\\_ARCHITECTURE.md
docs/08\\\\\\\_SECURITY\\\\\\\_AND\\\\\\\_CONCURRENCY.md
docs/10\\\\\\\_DEVELOPMENT\\\\\\\_TASKS.md
docs/11\\\\\\\_TESTING\\\\\\\_ACCEPTANCE.md
```

只更新实际发生变化的文档。

如果发现 `AGENTS.md` 中旧规则与本轮最新需求冲突，请先在计划中指出，不要直接覆盖。等我确认后再修改 `AGENTS.md`。

\---

#### 测试要求

至少需要给出并尽量执行以下测试。

###### 删除好友测试

1. A 和 B 是好友。
2. A 与 B 发送多条单聊 text 消息。
3. A 与 B 发送至少一条 file 消息，如果当前单聊文件消息已可用。
4. A 删除 B。
5. 验证 A 的好友列表没有 B。
6. 验证 B 的好友列表没有 A。
7. 验证 A 的会话列表没有 A-B private conversation。
8. 验证 B 的会话列表没有 A-B private conversation。
9. 直接请求旧 conversation\_id 的历史消息接口，应被拒绝或返回不可见结果，不能返回旧消息。
10. A 和 B 重新加好友。
11. 验证新的聊天窗口为空，不能看到删除前旧消息。
12. 发送新消息后，新消息正常显示。
13. 如果搜索功能已存在，验证旧消息和旧文件不出现在搜索结果。

###### 退出群聊测试

1. 创建群聊，至少 3 个成员。
2. 群内发送多条 text 消息。
3. 普通成员 C 退出群聊。
4. 验证 C 的会话列表不再显示该群。
5. 验证 C 不能读取该群历史消息。
6. 验证 C 不能发送该群消息。
7. 验证 A、B 等其他成员仍能看到群会话和群历史消息。
8. 验证数据库中的群消息没有因为 C 退出而被删除。
9. C 重新申请入群并被同意后，验证 C 默认只能看到重新入群之后的新消息。
10. 验证群主不能通过普通退出接口退出群聊。

###### 回归测试

1. `go test ./...`
2. `go build ./...`
3. 前端：`npm run build`
4. 手动验证现有单聊 text 消息不被破坏。
5. 手动验证现有文件消息不被破坏。
6. 手动验证撤回、删除单条消息、清空聊天记录不被破坏。
7. 手动验证群聊 text 消息不被破坏。

\---

#### 现在请先输出

在写代码前，请先输出：

1. 你理解的本轮任务目标。
2. 你准备阅读的文件。
3. 当前删除好友流程的实际代码路径。
4. 当前好友重新添加时 private conversation 创建或复用逻辑的实际代码路径。
5. 当前是否已有普通成员退出群聊接口。
6. 删除好友后的 private conversation 处理方案。
7. 退出群聊后的 group conversation 和 group\_members 处理方案。
8. 历史消息接口和会话列表接口需要补充哪些过滤规则。
9. 是否需要新增 migration；如果需要，说明字段、原因和兼容方案。
10. 预计新增或修改的后端文件。
11. 预计新增或修改的前端文件。
12. 预计需要更新的文档。
13. 测试计划。
14. 需要我确认的问题。

在我确认前，不要直接写代码。

## 第十一轮提示词：搜索功能

你现在只完成 Step 11：消息搜索和文件搜索。

请先阅读：

1. AGENTS.md
2. docs/01_PROJECT_DEVELOPMENT_SPEC.md 中“搜索系统”部分
3. docs/03_DATABASE_DESIGN.md 中 messages、files、message_user_states、conversation_user_states 表
4. docs/04_API_SPEC.md 中 search 相关接口

当前任务目标：
实现用户可见范围内的历史消息搜索和文件搜索。

功能要求：

1. MVP 阶段使用 MySQL LIKE 或 FULLTEXT。
2. 搜索逻辑必须封装在 search service 中。
3. 用户只能搜索自己有权限访问的会话。
4. 搜索消息时必须排除：
   - 当前用户已删除的消息
   - 当前用户清空聊天记录之前的消息
   - 已撤回的消息
   - 当前用户不属于的会话消息
5. 文件搜索按文件名搜索。
6. 文件搜索也必须遵守会话可见权限。
7. 后续预留 Elasticsearch / Meilisearch 替换空间。

接口要求：

1. GET /api/search/messages
2. GET /api/search/files

严格限制：

1. 不要引入 Elasticsearch。
2. 不要搜索用户无权限的消息。
3. 不要让被删除或撤回消息出现在搜索结果中。
4. 不要把复杂搜索 SQL 写在 handler 中。

请先输出实现计划、搜索过滤规则、预计修改文件，等我确认后再写代码。

## 第十二轮提示词：Docker 和上线部署

你现在只完成 Step 12：Docker 化和基础上线部署配置。

请先阅读：

1. AGENTS.md
2. docs/09_DEPLOYMENT_DOCKER.md
3. docs/06_BACKEND_ARCHITECTURE.md
4. docs/07_FRONTEND_ARCHITECTURE.md
5. docs/08_SECURITY_AND_CONCURRENCY.md

当前任务目标：
完善 Dockerfile、docker-compose.yml、Nginx 配置和部署说明。

功能要求：

1. 后端提供 Dockerfile。
2. 前端提供 Dockerfile。
3. docker-compose.yml 包含：
   - backend
   - frontend
   - mysql
   - redis
   - nginx
4. MySQL 数据目录需要 volume。
5. Redis 数据目录需要 volume。
6. uploads 文件目录需要 volume。
7. 配置文件通过环境变量注入。
8. Nginx 反向代理：
   - /api 转发到后端
   - /ws 转发到后端 WebSocket
   - / 转发到前端
9. 提供部署说明文档。
10. 提供本地启动命令。

严格限制：

1. 不要修改业务逻辑。
2. 不要引入 Kubernetes。
3. 不要引入复杂 CI/CD。
4. 不要把密钥写死在配置文件中。

请先输出实现计划、部署结构说明、预计修改文件，等我确认后再写代码。

# codex自查

请基于本次修改做一次代码审查。

重点检查：

1. 是否符合 AGENTS.md。
2. 是否只完成了本次任务，没有开发越界功能。
3. 是否有分层错误，例如 handler 写了复杂业务逻辑。
4. 是否有权限漏洞。
5. 是否有 SQL 注入风险。
6. 是否有 Token 校验遗漏。
7. 是否有用户越权访问会话、消息、文件的问题。
8. 是否有未处理错误。
9. 是否有重复代码。
10. 是否需要补充测试。

请按以下格式输出：

- 发现的问题
- 风险等级
- 建议修改方式
- 是否建议现在修复

# codex修复

请只修复你刚才代码审查中标记为“高风险”和“必须修复”的问题。

限制：

1. 不要重构无关代码。
2. 不要新增新功能。
3. 不要修改已通过的接口路径。
4. 修复完成后说明修改文件和测试方法。

# 启动

## docker启动

$env:MYSQL_ROOT_PASSWORD="rootpass"
$env:MYSQL_PASSWORD="goimpass"
docker compose up -d mysql redis

## 后端启动

cd backend
$env:MYSQL_DSN="goim:goimpass@tcp(127.0.0.1:3307)/go_im?charset=utf8mb4&parseTime=true&loc=Local"
$env:REDIS_ADDR="127.0.0.1:6379"
$env:JWT_ACCESS_SECRET="local-access-secret"
$env:JWT_REFRESH_SECRET="local-refresh-secret"
go run ./cmd/server

## 前端启动

cd frontend
npm.cmd install
npm.cmd run dev

## 浏览器打开

cd frontend
npm.cmd install
npm.cmd run dev
