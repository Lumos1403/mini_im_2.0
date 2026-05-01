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

删除后双方好友列表移除，对应 private conversation 从双方会话列表移除，旧单聊历史不可读取、不可搜索，重新添加好友后从空白会话开始。

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

当前拆分状态：

```txt
Step 11-A 后端搜索接口已完成：GET /api/search/messages、GET /api/search/files。
Step 11-B 前端搜索功能待开发。
```

## 阶段 10：群聊基础功能

## 阶段 10：群聊基础功能

阶段 10 拆分为：

```txt
Step 10    群聊基础功能
Step 10.5  群聊成员 GUI 和身份标识
Step 10.6  关系变更后的会话生命周期修复
```

### Step 10：群聊基础功能

目标：实现群聊后端基础能力和最小可用前端入口。

范围：

```txt
创建群聊
群号搜索
申请入群
审批入群
群成员管理
设置管理员
取消管理员
成员禁言
修改群设置
解散群聊
群 text 消息
群消息实时推送
群消息历史分页
```

要求：

```txt
用户可以创建群聊
创建者自动成为群主 owner
群默认最大人数 50，必须可配置
系统生成 group_id 和 group_no
用户可以通过 group_no 搜索群聊
用户可以申请加入群聊
群主或管理员可以同意 / 拒绝入群申请
群角色包括 owner、admin、member
群主可以设置管理员
群主可以取消管理员
群主和管理员可以设置成员禁言
被禁言用户不能发送群消息
群主和管理员可以修改 allow_member_invite
群主可以解散群聊
群解散后不能继续发送消息
群消息复用 conversations 表和 messages 表
群聊会话 conversation_type = group
群消息本阶段只实现 text
```

当前实现记录：

```txt
已新增 backend/migrations/006_create_group_system.sql
POST /api/groups 本阶段只创建群主，不处理 member_ids 初始成员加入
group_no 为 8～10 位数字字符串，依赖唯一索引保证唯一，冲突重试
群消息复用 sender_id + conversation_id + client_msg_id 幂等规则
发送方成功后只收到 chat.message.ack，不额外收到自己的 chat.message.receive
```

群消息发送校验：

```txt
群存在
群未解散
当前用户是群成员
当前用户未被禁言
message_type = text
content 非空且不超过配置长度
sender_id 必须来自服务端鉴权上下文
```

群消息 WebSocket 要求：

```txt
群消息复用 chat.message.send
服务端根据 conversation_type = group 进入群聊发送逻辑
发送成功后给发送方返回 chat.message.ack
群内在线成员收到 chat.message.receive
离线成员上线后通过历史消息接口分页拉取
```

群消息返回字段必须包含：

```txt
message_id
conversation_id
sender_id
sender_nickname
sender_avatar_url
sender_group_role
message_type
content
send_status
created_at
```

权限规则：

```txt
群主 owner 可以设置管理员、取消管理员、禁言成员、修改群设置、解散群聊
管理员 admin 可以审批入群、禁言普通成员、修改部分群设置
普通成员 member 可以发送消息、查看群成员
群主不能被管理员禁言
管理员不能禁言群主
管理员之间默认不能互相禁言
所有群权限必须服务端校验
```

接口范围：

```txt
POST   /api/groups
GET    /api/groups/search
POST   /api/groups/{group_id}/join-requests
GET    /api/groups/{group_id}/join-requests
POST   /api/groups/join-requests/{request_id}/accept
POST   /api/groups/join-requests/{request_id}/reject
GET    /api/groups/{group_id}/members
POST   /api/groups/{group_id}/admins/{user_id}
DELETE /api/groups/{group_id}/admins/{user_id}
POST   /api/groups/{group_id}/members/{user_id}/mute
DELETE /api/groups/{group_id}/members/{user_id}/mute
PUT    /api/groups/{group_id}/settings
DELETE /api/groups/{group_id}
```

最小前端要求：

```txt
可以创建群聊
可以搜索群号
可以申请入群
群主或管理员可以处理入群申请
可以进入群聊会话
可以发送群 text 消息
被禁言时发送失败并显示错误
群解散后不能继续发送
可以查看基础群成员列表
```

严格限制：

```txt
不实现复杂群公告
不实现群文件空间
不实现语音通话
不实现复杂邀请流程
不重写好友系统
不破坏单聊、文件消息、撤回、删除、清空逻辑
不只靠前端隐藏按钮做权限控制
```

### Step 10.5：群聊成员 GUI 和身份标识

目标：在 Step 10 群聊基础能力完成后，补齐群聊前端体验。

范围：

```txt
群消息中显示群主 / 管理员身份标识
查看群成员列表
点击群成员查看资料
从群成员资料弹窗中添加好友
根据群角色和好友状态显示按钮
```

群消息身份标识：

```txt
owner 显示“群主”标识，金色
admin 显示“管理员”标识，绿色
member 不显示身份标识
前端必须使用 sender_group_role，不允许根据昵称判断身份
历史消息和实时消息都要显示身份标识
```

群成员列表展示：

```txt
头像 avatar_url
昵称 nickname
user_id
个性签名 bio
群角色 role
禁言状态 mute_until
好友状态 friendship_status
```

群成员排序建议：

```txt
群主 owner 在最上方
管理员 admin 其次
普通成员 member 再后
同角色内按 joined_at 或 nickname 排序
```

群成员资料弹窗：

```txt
展示头像、昵称、user_id、bio、role、friendship_status
点击自己时不显示添加好友按钮
friendship_status = friend 时显示已是好友或发消息
friendship_status = not_friend 时显示添加好友
friendship_status = pending_sent 时显示申请中
friendship_status = pending_received 时提示对方已申请添加你
添加好友必须复用 POST /api/friends/requests
```

建议新增前端组件：

```txt
GroupRoleBadge
GroupMemberList
GroupMemberDrawer 或 GroupMemberModal
GroupMemberProfileModal
```

当前实现记录：

```txt
已新增 GroupRoleBadge、GroupMemberList、GroupMemberDrawer、GroupMemberProfileModal
Chat.vue 群消息昵称旁使用 sender_group_role 展示群主 / 管理员 Badge
sender_group_role 缺失时按 member 处理，不展示 Badge
群成员入口位于群聊会话头部
stores/group.ts 维护成员列表所属 group_id、抽屉状态、资料弹窗状态和好友申请 loading
群成员列表加载按 group_id 覆盖，不追加
群内添加好友复用 POST /api/friends/requests
添加好友成功后本地更新 friendship_status = pending_sent，并刷新群成员列表同步后端状态
friend 状态本轮只显示“已是好友”
未修改数据库、群聊后端主链路或 WebSocket 协议
```

严格限制：

```txt
不重写 Step 10 群聊主链路
不重写好友系统
不重写单聊消息逻辑
不新增重复好友接口
不通过 nickname 或 avatar 匹配用户
必须使用 user_id、group_id、conversation_id
```

本轮不新增完整群管理 GUI：

```txt
设置管理员
禁言
允许成员邀请
解散群聊
```

### Step 10.6：关系变更后的会话生命周期修复

目标：修复删除好友、退出群聊后旧会话和旧历史仍可见的数据一致性问题。

范围：

```txt
删除好友时同步处理旧 private conversation 生命周期
好友重新添加时创建新的空白 private conversation
新增普通成员退出群聊接口 POST /api/groups/{group_id}/leave
退出群聊后仅影响当前用户 membership 和会话状态
历史消息和文件下载补充 active membership / joined_at 可见性过滤
前端只做删除好友、退出群聊后的最小状态刷新
```

规则：

```txt
删除好友不再向旧 private conversation 写入 system message
删除好友后旧 private conversation 相关 messages / message_user_states / conversation_user_states / conversation_members / conversations 可物理删除
退出群聊不能删除 groups、其他 group_members 或群 messages
重新入群必须更新 group_members.joined_at 和 conversation_members.joined_at 为本次入群时间
重新入群后只能读取本次 joined_at 之后产生的新消息
当前操作者前端必须立即移除对应会话，并重新拉取会话列表兜底
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
