# 04 HTTP API 设计文档

## 1. 通用规则

基础路径：

```txt
/api
```

认证方式：

```txt
Authorization: Bearer <access_token>
```

统一成功响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

统一失败响应：

```json
{
  "code": 20001,
  "message": "token expired",
  "data": null
}
```

分页响应建议：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [],
    "page": 1,
    "page_size": 20,
    "total": 100
  }
}
```

聊天记录建议支持 cursor 分页：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [],
    "next_cursor": "message_id_or_timestamp",
    "has_more": true
  }
}
```

## 1.1 健康检查

```txt
GET /api/health
```

响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "status": "ok"
  }
}
```

## 2. 错误码

```txt
0       成功
10000   通用错误
20000   认证错误
30000   用户错误
40000   好友错误
50000   消息错误
60000   群聊错误
70000   文件错误
80000   系统错误
```

常用错误：

```txt
20001 Token 无效
20002 Token 已过期
20003 Refresh Token 无效
30001 用户不存在
30002 用户名已存在
30003 密码错误
40001 好友申请不存在
40002 已经是好友
40003 已被对方拉黑
40004 好友申请待处理
40005 不能添加自己
40006 好友关系不存在
40007 不能拉黑自己
50001 消息不存在
50002 消息不可撤回
50003 消息已撤回
60001 群不存在
60002 用户不在群内
60003 用户已被禁言
70001 文件不存在
70002 无文件访问权限
70003 文件过大
70004 文件无效
```

## 3. 认证接口

### 3.1 注册

```txt
POST /api/auth/register
```

请求：

```json
{
  "username": "test001",
  "password": "123456",
  "nickname": "小明"
}
```

响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": "123456789",
    "username": "test001",
    "nickname": "小明"
  }
}
```

注册成功后必须自动创建默认 Agent 好友和 Agent 单聊会话。

当前 Agent MVP 已实现：注册成功后会创建默认 Agent 用户（如不存在）、好友关系、private conversation、conversation_members 和 conversation_user_states。该能力是本地 IM 数据能力，不依赖 FastAPI Agent 服务在线。

### 3.2 登录

```txt
POST /api/auth/login
```

请求：

```json
{
  "username": "test001",
  "password": "123456"
}
```

响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "access_token": "xxx",
    "refresh_token": "yyy",
    "expires_in": 900,
    "user": {
      "user_id": "123456789",
      "username": "test001",
      "nickname": "小明",
      "avatar_url": ""
    }
  }
}
```

### 3.3 刷新 Token

```txt
POST /api/auth/refresh
```

请求：

```json
{
  "refresh_token": "yyy"
}
```

响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "access_token": "new_xxx",
    "refresh_token": "new_yyy",
    "expires_in": 900
  }
}
```

### 3.4 退出登录

```txt
POST /api/auth/logout
```

需要登录。

请求头：

```txt
Authorization: Bearer <access_token>
```

响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

## 4. 用户接口

### 4.1 获取当前用户信息

```txt
GET /api/users/me
```

需要登录。

响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": "123456789",
    "username": "test001",
    "nickname": "小明",
    "avatar_url": ""
  }
}
```

### 4.2 获取当前用户资料

```txt
GET /api/users/me/profile
```

需要登录。

响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": "123456789",
    "username": "test001",
    "nickname": "小明",
    "avatar_url": "https://example.com/avatar.png",
    "gender": "male",
    "bio": "个性签名",
    "profile_status": "normal",
    "profile_review_reason": ""
  }
}
```

### 4.3 修改当前用户资料

```txt
PUT /api/users/me/profile
```

需要登录。

请求：

```json
{
  "nickname": "新昵称",
  "avatar_url": "https://example.com/avatar.png",
  "gender": "male",
  "bio": "个性签名"
}
```

响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": "123456789",
    "username": "test001",
    "nickname": "新昵称",
    "avatar_url": "https://example.com/avatar.png",
    "gender": "male",
    "bio": "个性签名",
    "profile_status": "normal",
    "profile_review_reason": ""
  }
}
```

说明：

```txt
当前版本不做内容审核，不限制修改频率。
客户端只能修改 avatar_url、nickname、gender、bio。
profile_status、profile_review_reason 字段保留并由服务端返回，客户端不能修改。
```

### 4.4 搜索用户

```txt
GET /api/users/search?keyword=小明&page=1&page_size=20
```

响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "user_id": "123456789",
        "username": "test001",
        "nickname": "小明",
        "avatar_url": "https://example.com/avatar.png",
        "bio": "个性签名"
      }
    ],
    "page": 1,
    "page_size": 20,
    "total": 1
  }
}
```

说明：

```txt
keyword 如果是纯数字，可优先按 user_id 精确匹配
否则按 nickname 模糊匹配
bio 用于前端搜索结果展示个性签名
```

## 5. 好友接口

### 5.1 发起好友申请

```txt
POST /api/friends/requests
```

请求：

```json
{
  "to_user_id": "123456789",
  "message": "我是小明"
}
```

### 5.2 获取好友申请

```txt
GET /api/friends/requests?direction=received&page=1&page_size=20
```

说明：

`direction` 可选值为 `received`、`sent`，不传时默认为 `received`。

### 5.3 同意好友申请

```txt
POST /api/friends/requests/{request_id}/accept
```

### 5.4 拒绝好友申请

```txt
POST /api/friends/requests/{request_id}/reject
```

### 5.5 好友列表

```txt
GET /api/friends
```

响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "friend_user_id": "123456789",
        "nickname": "小明",
        "avatar_url": "https://example.com/avatar.png",
        "bio": "个性签名",
        "conversation_id": "111111",
        "is_blocked_by_me": false,
        "created_at": "2026-04-24T12:00:00+08:00",
        "updated_at": "2026-04-24T12:00:00+08:00"
      }
    ],
    "page": 1,
    "page_size": 20,
    "total": 1
  }
}
```

说明：

```txt
friend_user_id 是好友的 user_id。
conversation_id 是双方私聊会话 ID，前端打开聊天时应优先使用该字段。
is_blocked_by_me 表示当前登录用户是否已单向拉黑该好友。
```

### 5.6 删除好友

```txt
DELETE /api/friends/{user_id}
```

删除后双方好友关系失效，两人之间的旧 private conversation 从双方会话列表移除。

规则：

```txt
删除好友必须在同一事务内处理好友关系和旧 private conversation 生命周期。
删除好友后不向旧 private conversation 写入 system message。
删除好友后双方不能读取、搜索或继续发送到旧 private conversation。
重新添加好友后创建新的空白 private conversation，旧消息不能恢复。
前端操作成功后通过 toast / 状态刷新提示用户，不依赖旧会话 system message。
```

### 5.7 拉黑好友

```txt
POST /api/friends/{user_id}/block
```

拉黑是单向关系，仅创建 `block_relations(blocker_id, blocked_id)`。

### 5.8 解除拉黑

```txt
DELETE /api/friends/{user_id}/block
```

## 6. 会话接口

### 6.1 获取会话列表

```txt
GET /api/conversations?page=1&page_size=20
```

响应字段建议：

```json
{
  "conversation_id": "111",
  "conversation_type": "private",
  "title": "小明",
  "avatar_url": "",
  "peer_user_id": "123456789",
  "peer_nickname": "小明",
  "peer_avatar_url": "",
  "last_message": {
    "content": "你好",
    "message_type": "text",
    "created_at": "2026-04-24 12:00:00"
  },
  "unread_count": 2,
  "is_pinned": false,
  "is_muted": false
}
```

当前实现范围：返回当前登录用户可见的会话列表；private 会话标题和头像取对方用户资料。private 会话额外返回 `peer_user_id`、`peer_nickname`、`peer_avatar_url`，用于前端在缺少好友列表 `conversation_id` 时按对方 user_id 兜底定位会话，禁止按昵称或头像匹配。消息模块尚未实现时，`last_message` 返回 `null`。

### 6.2 获取会话消息

```txt
GET /api/conversations/{conversation_id}/messages?cursor=&limit=30
```

说明：

```txt
cursor 基于雪花 message_id
首次不传 cursor，返回最新 limit 条消息
加载更早消息时传当前页最旧的 message_id，服务端查询 message_id < cursor
服务端倒序查询，返回前按时间正序排列
历史消息以当前用户的 message_user_states 作为可见性依据
允许返回 sent 和发送方可见的 failed_blocked
必须过滤当前用户已删除、已清空、已撤回消息
必须过滤全局删除消息
必须过滤当前用户没有 message_user_states 的消息
private 会话必须要求当前用户仍是 active 成员
group 会话必须要求当前用户仍是 active 群成员
group 会话重新入群后只返回当前 membership joined_at 之后的消息
limit 默认 30，最大 100
```

响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "client_msg_id": "client-uuid",
        "message_id": "999999",
        "conversation_id": "111111",
        "sender_id": "123456",
        "message_type": "text",
        "content": "你好",
        "extra_json": {},
        "send_status": "sent",
        "created_at": "2026-04-24 12:00:00"
      }
    ],
    "next_cursor": "999999",
    "has_more": true,
    "limit": 30
  }
}
```

### 6.3 清空聊天记录

```txt
DELETE /api/conversations/{conversation_id}/messages
```

说明：只清空当前用户视角。

实现要求：

```txt
更新当前用户自己的 conversation_user_states.cleared_at
同时将当前用户该会话的 unread_count 置为 0
不影响其他用户视角
```

### 6.4 置顶会话

```txt
POST /api/conversations/{conversation_id}/pin
```

### 6.5 取消置顶

```txt
DELETE /api/conversations/{conversation_id}/pin
```

### 6.6 免打扰

```txt
POST /api/conversations/{conversation_id}/mute
```

### 6.7 取消免打扰

```txt
DELETE /api/conversations/{conversation_id}/mute
```

## 7. 消息接口

主要消息发送通过 WebSocket。HTTP 接口用于删除、撤回、搜索、历史记录。

### 7.1 删除单条消息

```txt
DELETE /api/conversations/{conversation_id}/messages/{message_id}
```

只删除当前用户视角。

```txt
删除必须只更新当前用户自己的 message_user_states.is_deleted
当前用户没有 message_user_states 时不能删除该消息
单条删除必须幂等，重复删除同一条当前用户可访问的消息仍返回成功
failed_blocked 只有发送方拥有 message_user_states，因此只允许发送方在自己视角删除
```

### 7.2 撤回消息

```txt
POST /api/messages/{message_id}/recall
```

规则：

```txt
只能撤回自己发送的消息
只有 send_status = sent 的消息可以撤回
send_status = failed_blocked 的消息不能撤回，只能由发送方删除
发送后 5 分钟内可撤回
撤回时必须先读取并缓存原始 content，再清空 messages.content
Redis 重新编辑缓存写入失败时，撤回操作不能返回成功
撤回后双方消息消失
撤回成功后再推送 WebSocket 事件，不能在事务提交前推送
发送方可重新编辑
如果被撤回消息是 conversations.last_message_id，需要回退到上一条未撤回、未全局删除、send_status = sent 的消息；没有则置空
```

响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message_id": "123",
    "editable_until": "2026-04-24 12:05:00"
  }
}
```

### 7.3 获取撤回消息的重新编辑内容

```txt
GET /api/messages/{message_id}/recall-edit-cache
```

说明：

```txt
仅发送者本人可访问
仅在 Redis 缓存 5 分钟内可访问
```

## 8. 文件接口

### 8.1 上传文件

```txt
POST /api/files/upload
Content-Type: multipart/form-data
```

字段：

```txt
file：文件
```

限制：

```txt
单文件最大 50MB，可配置
文件保存到后端本地 uploads 目录
不提供在线预览
```

响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "file_id": "123456",
    "original_name": "资料.zip",
    "file_size": 102400,
    "mime_type": "application/zip"
  }
}
```

### 8.2 下载文件

```txt
GET /api/files/{file_id}/download
```

必须登录，且用户必须是文件所在会话成员或文件上传者。

说明：

```txt
下载接口返回附件流
不开放 uploads 静态目录
服务端必须防止路径穿越
```

## 9. 搜索接口

### 9.1 搜索消息

```txt
GET /api/search/messages?keyword=你好&page=1&page_size=20
```

要求登录。

查询参数：

```txt
keyword   必填，去除首尾空白后不能为空，空值返回参数错误
page      可选，默认 1，小于等于 0 时按 1 处理
page_size 可选，默认 20，最大 100
```

响应使用统一分页结构：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "message_id": "123",
        "conversation_id": "456",
        "conversation_type": "private",
        "sender_id": "789",
        "sender_nickname": "小明",
        "sender_avatar_url": "",
        "message_type": "text",
        "content": "你好",
        "created_at": "2026-05-01 12:00:00"
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 20
  }
}
```

必须过滤：

```txt
已删除消息
清空时间前消息
已撤回消息
不属于当前用户会话的消息
删除好友后已清理的旧 private 会话消息
退出群聊后当前用户不可见的群消息
重新入群后只能搜索本次 group_members.joined_at 之后产生的群消息
```

### 9.2 搜索文件

```txt
GET /api/search/files?keyword=资料&page=1&page_size=20
```

要求登录。

查询参数同搜索消息：`keyword` 必填且不能为空；`page` 默认 1；`page_size` 默认 20，最大 100。

响应使用统一分页结构：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "file_id": "123",
        "original_name": "资料.pdf",
        "file_size": 1024,
        "mime_type": "application/pdf",
        "uploader_id": "789",
        "uploader_nickname": "小明",
        "message_id": "111",
        "conversation_id": "456",
        "conversation_type": "group",
        "created_at": "2026-05-01 12:00:00"
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 20
  }
}
```

说明：

```txt
文件搜索基于当前用户可见的 file 类型消息，不直接暴露 files 表。
删除好友后的旧私聊文件消息不可搜索。
退群后的旧群文件消息不可搜索。
重新入群后只能搜索本次 group_members.joined_at 之后产生的群文件消息。
已删除、已清空、已撤回的文件消息不可搜索。
```

## 10. 群聊接口

### 10.1 创建群聊

```txt
POST /api/groups
```

请求：

```json
{
  "name": "测试群",
  "avatar_url": ""
}
```

响应包含 group_id、group_no、conversation_id。本阶段 `member_ids` 不处理初始成员加入，创建者自动成为群主。

### 10.2 搜索群

```txt
GET /api/groups/search?keyword=10001
```

优先按 group_no 精确搜索。响应项包含 `group_id`、`group_no`、`conversation_id`、`name`、`avatar_url`、`max_members`、`allow_member_invite`、`status`、`is_member`。

### 10.3 申请加入群

```txt
POST /api/groups/{group_id}/join-requests
```

### 10.4 获取入群申请

```txt
GET /api/groups/{group_id}/join-requests?page=1&page_size=20
```

仅群主或管理员可访问。

### 10.5 同意入群申请

```txt
POST /api/groups/join-requests/{request_id}/accept
```

### 10.6 拒绝入群申请

```txt
POST /api/groups/join-requests/{request_id}/reject
```

### 10.7 获取群成员

### 10.7 获取群成员

```txt
GET /api/groups/{group_id}/members
```

需要登录，且当前用户必须是该群成员。

响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [
      {
        "user_id": "123456789",
        "nickname": "张三",
        "avatar_url": "",
        "bio": "个性签名",
        "role": "owner",
        "mute_until": null,
        "joined_at": "2026-04-25 12:00:00",
        "status": "active",
        "friendship_status": "self"
      }
    ],
    "total": 1
  }
}
```

字段说明：

```txt
role：owner / admin / member
friendship_status：self / friend / not_friend / pending_sent / pending_received
mute_until：为空表示未禁言；不为空且晚于当前时间表示禁言中
```

说明：

```txt
self：当前登录用户自己
friend：已经是好友
not_friend：不是好友，可以发起好友申请
pending_sent：我已经向对方发过好友申请
pending_received：对方已经向我发过好友申请
```

### 10.8 设置管理员

```txt
POST /api/groups/{group_id}/admins/{user_id}
```

仅群主。

### 10.9 取消管理员

```txt
DELETE /api/groups/{group_id}/admins/{user_id}
```

仅群主。

### 10.10 禁言成员

```txt
POST /api/groups/{group_id}/members/{user_id}/mute
```

请求：

```json
{
  "mute_until": "2026-04-25 12:00:00"
}
```

### 10.11 解除禁言

```txt
DELETE /api/groups/{group_id}/members/{user_id}/mute
```

### 10.12 修改群设置

```txt
PUT /api/groups/{group_id}/settings
```

请求：

```json
{
  "allow_member_invite": true,
  "max_members": 50
}
```

规则：

```txt
max_members 只能由群主修改，不能小于当前成员数，不能超过服务端配置上限。
allow_member_invite 可由群主或管理员修改。
两个字段都支持按需传入。
```

### 10.13 解散群聊

```txt
DELETE /api/groups/{group_id}
```

仅群主。

### 10.14 退出群聊

```txt
POST /api/groups/{group_id}/leave
```

规则：

```txt
当前用户必须是该群 active 成员。
群主不能通过普通退出接口退出群聊；当前没有群主转让能力时，群主只能解散群聊。
管理员和普通成员可以退出群聊。
退出群聊只影响当前用户，不删除 groups、不删除其他 group_members、不删除 messages。
退出后当前用户的群会话从会话列表移除，不能读取该群历史消息，不能继续发送该群消息。
重新入群时必须更新 group_members.joined_at 和 conversation_members.joined_at 为本次入群时间。
重新入群后只读取本次 joined_at 之后的新消息。
```

### 10.15 群消息历史字段补充

获取会话消息接口 `GET /api/conversations/{conversation_id}/messages` 在 conversation_type = group 时，消息项需要额外返回：

```json
{
  "message_id": "999999",
  "conversation_id": "111111",
  "sender_id": "123456789",
  "sender_nickname": "张三",
  "sender_avatar_url": "",
  "sender_group_role": "owner",
  "message_type": "text",
  "content": "大家好",
  "send_status": "sent",
  "created_at": "2026-04-25 12:00:00"
}
```

说明：

```txt
sender_group_role 必须来自 group_members.role
owner 用于前端展示群主标识
admin 用于前端展示管理员标识
member 不展示特殊标识
群消息实时推送和历史消息均返回 sender_nickname、sender_avatar_url、sender_group_role。
```

### 10.16 群内添加好友

群成员资料弹窗中的添加好友功能必须复用已有好友申请接口：

```txt
POST /api/friends/requests
```

请求：

```json
{
  "to_user_id": "123456789",
  "message": "我是群里的成员"
}
```

不新增单独的群内添加好友接口。
