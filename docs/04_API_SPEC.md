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

当前 Step 2 只预留 Agent 好友创建的 service 方法，不创建好友关系和会话；好友系统与 Agent 聊天在后续阶段实现。

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

说明：

```txt
keyword 如果是纯数字，可优先按 user_id 精确匹配
否则按 nickname 模糊匹配
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

### 5.6 删除好友

```txt
DELETE /api/friends/{user_id}
```

删除后双方好友关系失效，对方会话中显示系统提示。

当前实现预留 system message 创建入口，消息表和 WebSocket 推送在后续阶段实现。

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

当前实现范围：返回当前登录用户可见的会话列表；private 会话标题和头像取对方用户资料。消息模块尚未实现时，`last_message` 返回 `null`。

### 6.2 获取会话消息

```txt
GET /api/conversations/{conversation_id}/messages?cursor=&limit=30
```

说明：

```txt
按时间倒序或正序由前端需求确定，但必须支持分页
必须过滤当前用户已删除、已清空、已撤回消息
```

### 6.3 清空聊天记录

```txt
DELETE /api/conversations/{conversation_id}/messages
```

说明：只清空当前用户视角。

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
DELETE /api/messages/{message_id}
```

只删除当前用户视角。

### 7.2 撤回消息

```txt
POST /api/messages/{message_id}/recall
```

规则：

```txt
只能撤回自己发送的消息
发送后 5 分钟内可撤回
撤回后双方消息消失
发送方可重新编辑
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
GET /api/messages/{message_id}/recall-edit
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

## 9. 搜索接口

### 9.1 搜索消息

```txt
GET /api/search/messages?keyword=你好&page=1&page_size=20
```

必须过滤：

```txt
已删除消息
清空时间前消息
已撤回消息
不属于当前用户会话的消息
```

### 9.2 搜索文件

```txt
GET /api/search/files?keyword=资料&page=1&page_size=20
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
  "avatar_url": "",
  "member_ids": ["123", "456"]
}
```

响应包含 group_id、group_no、conversation_id。

### 10.2 搜索群

```txt
GET /api/groups/search?keyword=10001
```

优先按 group_no 精确搜索。

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

```txt
GET /api/groups/{group_id}/members
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

### 10.13 解散群聊

```txt
DELETE /api/groups/{group_id}
```

仅群主。
