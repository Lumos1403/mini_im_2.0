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

## 第九轮提示词：文件消息ing

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

## 第十轮提示词：群聊

你现在只完成 Step 10：群聊基础功能。

请先阅读：

1. AGENTS.md
2. docs/01_PROJECT_DEVELOPMENT_SPEC.md 中“群聊系统”部分
3. docs/03_DATABASE_DESIGN.md 中 groups、group_members、group_join_requests 表
4. docs/04_API_SPEC.md 中 groups 相关接口
5. docs/05_WEBSOCKET_PROTOCOL.md 中群消息事件

当前任务目标：
实现创建群聊、群号搜索、申请入群、审批入群、群成员管理、群消息、禁言、解散群聊。

功能要求：

1. 用户可以创建群聊。
2. 创建者自动成为群主 owner。
3. 群默认最大人数 50，可配置。
4. 系统生成 group_id 和 group_no。
5. 用户可以通过群号搜索群聊。
6. 用户可以申请加入群聊。
7. 群主或管理员可以同意/拒绝入群申请。
8. 群角色：
   - owner
   - admin
   - member
9. 群主可以设置管理员。
10. 群主和管理员可以设置成员禁言。
11. 群主和管理员可以设置是否允许普通成员邀请。
12. 群主可以解散群聊。
13. 被禁言用户不能发送群消息。
14. 群消息复用 messages 表和 conversations 表。

接口要求：

1. POST /api/groups
2. GET /api/groups/search
3. POST /api/groups/{group_id}/join-requests
4. GET /api/groups/{group_id}/join-requests
5. POST /api/groups/join-requests/{id}/accept
6. POST /api/groups/join-requests/{id}/reject
7. GET /api/groups/{group_id}/members
8. POST /api/groups/{group_id}/admins
9. DELETE /api/groups/{group_id}/admins/{user_id}
10. POST /api/groups/{group_id}/mute
11. PUT /api/groups/{group_id}/settings
12. DELETE /api/groups/{group_id}

严格限制：

1. 不要实现复杂群公告。
2. 不要实现群文件空间。
3. 不要实现语音通话。
4. 群权限必须服务端校验，不能只靠前端隐藏按钮。
5. 不要破坏单聊消息逻辑。

请先输出实现计划、群权限校验方案、预计修改文件，等我确认后再写代码。

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
