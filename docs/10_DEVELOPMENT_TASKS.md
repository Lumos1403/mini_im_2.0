# 10 Codex 分阶段开发任务

## 使用规则

每次只把一个任务交给 Codex。不要一次性让 Codex 开发完整系统。

每个任务都要求 Codex：

```txt
先说明理解
列出修改文件
再写代码
最后给测试方法
```

## 阶段 0：项目初始化

### Task 0.1 初始化后端项目

目标：创建 Go 后端基础结构。

要求：

```txt
创建 backend 目录
初始化 go.mod
创建 cmd/server/main.go
创建 internal 目录结构
实现 /api/health
配置 Gin
配置统一响应结构
```

验收：

```txt
go run ./cmd/server 可启动
GET /api/health 返回 success
```

### Task 0.2 初始化前端项目

目标：创建 Vue 3 + Vite 前端基础结构。

要求：

```txt
创建 frontend
配置 Vue Router
配置 Pinia
创建 Login/Register/Chat/Profile 页面占位
配置 Axios 封装
```

### Task 0.3 配置 Docker Compose

目标：启动 MySQL、Redis、后端、前端。

## 阶段 1：用户与认证

### Task 1.1 数据库 migration

目标：创建 users、user_profiles 基础表。

要求：按照 `03_DATABASE_DESIGN.md`。

### Task 1.2 雪花 ID 工具

实现 user_id、message_id 等业务 ID 生成器。

### Task 1.3 注册接口

实现：

```txt
POST /api/auth/register
bcrypt 加密
username 唯一
创建 user_profiles
初始化 Agent 好友预留，可先创建 Agent 用户
```

### Task 1.4 登录接口

实现：

```txt
POST /api/auth/login
bcrypt 校验
生成 access_token / refresh_token
refresh_token 存 Redis
```

### Task 1.5 鉴权中间件

实现：

```txt
AuthMiddleware
解析 JWT
注入 user_id 到 context
```

### Task 1.6 刷新和退出登录

实现 refresh/logout。

## 阶段 2：用户资料

### Task 2.1 当前用户资料接口

```txt
GET /api/users/me
PUT /api/users/me/profile
```

### Task 2.2 用户搜索接口

```txt
GET /api/users/search
```

支持 user_id 精确搜索和 nickname 模糊搜索。

## 阶段 3：好友系统

### Task 3.1 好友申请表和模型

创建 friend_requests、friendships、block_relations。

### Task 3.2 发起好友申请

```txt
POST /api/friends/requests
```

### Task 3.3 好友申请列表与处理

```txt
GET received/sent
POST accept/reject
```

同意后创建好友关系和单聊会话。

### Task 3.4 好友列表

```txt
GET /api/friends
```

### Task 3.5 删除好友

删除后双方好友列表移除，对方收到系统提示事件。

### Task 3.6 拉黑与解除拉黑

实现 block_relations。

## Step 4.6：好友接口字段补齐与前端 GUI 修复

补齐好友前端 GUI 依赖的 HTTP 响应字段，并修复打开聊天的可靠定位逻辑。

```txt
GET /api/users/search 返回 bio
GET /api/friends 返回 friend_user_id、nickname、avatar_url、bio、conversation_id、is_blocked_by_me
GET /api/conversations 的 private 会话返回 peer_user_id、peer_nickname、peer_avatar_url
好友列表拉黑按钮状态以后端 is_blocked_by_me 为准
打开好友聊天优先使用 conversation_id，缺失时使用 peer_user_id 兜底
不按 nickname 或 avatar_url 匹配会话
```

## 阶段 4：会话系统

### Task 4.1 会话表和状态表

创建：

```txt
conversations
conversation_members
conversation_user_states
```

### Task 4.2 获取会话列表

```txt
GET /api/conversations
```

### Task 4.3 清空聊天记录

```txt
DELETE /api/conversations/{conversation_id}/messages
```

## 阶段 5：WebSocket 基础

### Task 5.1 WebSocket 鉴权连接

实现 `/ws?token=`。

### Task 5.2 Hub 和 Client

实现连接注册、注销、发送队列。

### Task 5.3 心跳机制

实现 ping/pong。

### Task 5.4 统一事件协议

实现 envelope 格式解析。

## 阶段 6：单聊文本消息

### Task 6.1 messages 表和模型

创建 messages、message_user_states。

### Task 6.2 发送文本消息

通过 WebSocket 实现 chat.message.send。

规则：

```txt
校验会话成员
校验拉黑
写入消息
推送对方
返回 ack
```

### Task 6.3 历史消息分页

```txt
GET /api/conversations/{id}/messages
```

必须过滤删除、清空、撤回。

## 阶段 7：消息管理

### Task 7.1 单条消息删除

```txt
DELETE /api/messages/{message_id}
```

### Task 7.2 撤回消息

```txt
POST /api/messages/{message_id}/recall
```

规则：发送后 5 分钟内。

### Task 7.3 重新编辑缓存

Redis 缓存原消息内容 5 分钟。

### Step 7.5 单聊实时接收、会话列表同步和未读提醒

目标：修复在线接收方收到单聊文本消息后，当前聊天窗口、会话列表和未读数不能实时同步的问题。

范围：

```txt
只处理 private text 消息的实时接收和发送状态同步
不实现群聊
不实现文件消息
不修改撤回、删除、清空逻辑
不新增 migration
```

要求：

```txt
前端 WebSocket 收到 chat.message.receive 后必须由 ws store 集中处理
chat store 负责维护当前消息列表、会话列表、last_message、unread_count
当前打开该 conversation_id 时，立即追加消息、按 message_id / client_msg_id 去重、触发滚动到底部，未读数清零
未打开该 conversation_id 时，更新 last_message、unread_count，并将会话移动到顶部
本地没有该 conversation_id 时，按 conversation_id 创建本地会话项并拉取会话列表补齐资料
禁止按 nickname / avatar_url 匹配会话，只能使用 conversation_id 或 peer_user_id
收到 chat.message.ack 后，将 sending 改为 sent，并用服务端 message_id 替换本地临时 message_id
收到 chat.message.failed 后，将本地临时消息改为 failed / failed_blocked，failed_blocked 显示红色感叹号
后端 chat.message.receive 必须包含 message_id、conversation_id、sender_id、message_type、content、send_status、created_at
```

## 阶段 8：文件消息

### Task 8.1 文件表和本地存储

创建 files 表，实现 LocalFileStorage。

### Task 8.2 文件上传

```txt
POST /api/files/upload
```

### Task 8.3 文件下载鉴权

```txt
GET /api/files/{file_id}/download
```

### Task 8.4 文件消息发送

通过 WebSocket 发送 file 类型消息。

## 阶段 9：搜索

### Task 9.1 消息搜索

```txt
GET /api/search/messages
```

### Task 9.2 文件搜索

```txt
GET /api/search/files
```

必须遵守删除、清空、撤回过滤规则。

## 阶段 10：群聊

### Task 10.1 群聊表

创建 groups、group_members、group_join_requests。

### Task 10.2 创建群聊

```txt
POST /api/groups
```

### Task 10.3 搜索群和申请入群

```txt
GET /api/groups/search
POST /api/groups/{id}/join-requests
```

### Task 10.4 审批入群

群主和管理员可处理。

### Task 10.5 群消息

通过同一个 chat.message.send 发送群消息。

### Task 10.6 群权限

实现：

```txt
设置管理员
禁言
允许成员邀请
解散群聊
```

## 阶段 11：前端完整界面

### Task 11.1 登录注册页面

### Task 11.2 聊天主页面布局

### Task 11.3 会话列表

### Task 11.4 消息列表和输入框

### Task 11.5 好友搜索和申请

### Task 11.6 文件上传和文件消息

### Task 11.7 群聊管理界面

## 阶段 12：工程化上线

### Task 12.1 统一日志

### Task 12.2 限流中间件

### Task 12.3 Dockerfile

### Task 12.4 docker-compose

### Task 12.5 Nginx 配置

### Task 12.6 部署文档和环境变量
