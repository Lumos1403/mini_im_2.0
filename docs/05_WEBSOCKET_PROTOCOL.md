# 05 WebSocket 协议设计文档

## 1. 连接地址

本地开发：

```txt
ws://localhost:8081/ws?token=<access_token>
```

生产环境：

```txt
wss://your-domain.com/ws?token=<access_token>
```

连接建立时，后端必须校验 Access Token。Token 无效则拒绝连接。

## 2. 连接生命周期

### 2.1 建立连接

后端流程：

```txt
解析 token
校验 token
获取 user_id
升级为 WebSocket
注册连接到 Hub
写入 Redis 在线状态，默认 TTL 60 秒
启动读协程
启动写协程
启动心跳
```

### 2.2 断开连接

断开时：

```txt
从 Hub 删除连接
关闭 send channel
删除或更新 Redis 在线状态
记录日志
```

### 2.3 心跳

建议：

```txt
服务端每 30 秒 ping
客户端收到后 pong
服务端收到协议级 pong 后刷新 im:online:{user_id} TTL
超过 60 秒无响应则断开
```

### 2.4 安全限制

```txt
WebSocket 读取消息大小默认限制为 64KB
发送队列必须有容量上限，默认 256
发送队列满时关闭该慢连接
开发环境允许 localhost / 127.0.0.1 Origin
生产环境必须通过 WS_ALLOWED_ORIGINS 配置允许域名
Redis 黑名单检查失败时拒绝连接
```

## 3. 统一消息 Envelope

客户端和服务端统一使用：

```json
{
  "seq": "client_message_id",
  "type": "chat.message.send",
  "data": {},
  "timestamp": 1710000000000
}
```

字段说明：

```txt
seq：客户端生成的临时 ID，用于匹配 ack
type：事件类型
data：事件数据
timestamp：客户端或服务端时间戳，毫秒
```

未知事件类型统一返回 error envelope：

```json
{
  "seq": "tmp-unknown-001",
  "type": "error",
  "data": {
    "code": "unsupported_event",
    "message": "unsupported websocket event",
    "event_type": "unknown.event"
  },
  "timestamp": 1710000001000
}
```

## 4. 基础测试事件

当前 WebSocket 基础阶段只实现连接、在线状态、心跳和测试事件，暂不实现真实聊天消息发送。

客户端可发送：

```json
{
  "seq": "tmp-ping-001",
  "type": "ping",
  "data": {},
  "timestamp": 1710000000000
}
```

服务端返回：

```json
{
  "seq": "tmp-ping-001",
  "type": "pong",
  "data": {},
  "timestamp": 1710000001000
}
```

说明：

```txt
ping / pong 是业务层测试事件。
服务端仍会使用 WebSocket 协议级 ping/pong 做心跳检测。
```

## 5. 客户端发送消息

### 5.1 发送单聊 / 群聊消息

以下聊天消息事件在后续单聊消息阶段实现，当前 WebSocket 基础阶段暂不处理。

事件类型：

```txt
chat.message.send
```

请求：

```json
{
  "seq": "tmp-001",
  "type": "chat.message.send",
  "data": {
    "conversation_id": "111111",
    "client_msg_id": "550e8400-e29b-41d4-a716-446655440000",
    "message_type": "text",
    "content": "你好",
    "extra_json": {}
  },
  "timestamp": 1710000000000
}
```

说明：

```txt
seq 只用于 WebSocket 请求响应匹配
client_msg_id 必填，由客户端生成，建议 UUID，最大长度 64
client_msg_id 在同一 sender_id + conversation_id 范围内唯一，用于消息发送幂等
sender_id 必须来自 WebSocket 鉴权结果，禁止信任客户端传入
text 消息 content trim 后不能为空，默认最大 2000，可通过配置调整
```

文件消息：

```json
{
  "seq": "tmp-002",
  "type": "chat.message.send",
  "data": {
    "conversation_id": "111111",
    "message_type": "file",
    "content": "123456789",
    "extra_json": {
      "file_id": "123456789",
      "file_name": "资料.zip",
      "file_size": 102400,
      "mime_type": "application/zip"
    }
  },
  "timestamp": 1710000000000
}
```

说明：

```txt
file 消息复用 chat.message.send 发送流程
content 必须为 file_id 字符串
服务端必须校验 file_id 存在且发送者是文件上传者
服务端根据 files 表生成并保存 extra_json，不能信任客户端传入的文件名、大小和 MIME 类型
```

## 6. 服务端确认 ACK

事件类型：

```txt
chat.message.ack
```

成功响应：

```json
{
  "seq": "tmp-001",
  "type": "chat.message.ack",
  "data": {
    "client_msg_id": "550e8400-e29b-41d4-a716-446655440000",
    "message_id": "999999",
    "conversation_id": "111111",
    "send_status": "sent",
    "server_time": "2026-04-24 12:00:00"
  },
  "timestamp": 1710000001000
}
```

失败响应：

```json
{
  "seq": "tmp-001",
  "type": "chat.message.failed",
  "data": {
    "client_msg_id": "550e8400-e29b-41d4-a716-446655440000",
    "message_id": "999999",
    "conversation_id": "111111",
    "send_status": "failed_blocked",
    "code": "failed_blocked",
    "message": "对方已拒收你的消息",
    "server_time": "2026-04-24 12:00:00"
  },
  "timestamp": 1710000001000
}
```

## 7. 服务端推送新消息

事件类型：

```txt
chat.message.receive
```

推送：

```json
{
  "seq": "server-999999",
  "type": "chat.message.receive",
  "data": {
    "client_msg_id": "550e8400-e29b-41d4-a716-446655440000",
    "message_id": "999999",
    "conversation_id": "111111",
    "sender_id": "123456",
    "message_type": "text",
    "content": "你好",
    "extra_json": {},
    "send_status": "sent",
    "created_at": "2026-04-24 12:00:00"
  },
  "timestamp": 1710000001000
}
```

字段说明：

```txt
send_status 当前单聊正常推送固定为 sent，用于前端复用统一消息模型并更新当前聊天窗口。
前端必须使用 conversation_id 定位会话，不能使用 nickname / avatar_url 匹配会话。
```

## 8. 撤回消息事件

撤回通过 HTTP API 发起，WebSocket 用于通知相关在线用户移除消息。事件必须在撤回事务提交成功后推送，不能在事务提交前推送。

事件类型：

```txt
chat.message.recalled
```

推送：

```json
{
  "seq": "server-recall-1",
  "type": "chat.message.recalled",
  "data": {
    "message_id": "999999",
    "conversation_id": "111111",
    "recalled_by": "123456",
    "recalled_at": "2026-04-24 12:03:00"
  },
  "timestamp": 1710000180000
}
```

前端规则：

```txt
chat.message.recalled 只用于让前端移除原消息
接收方：直接移除该消息，不显示任何提示
发送方：移除该消息，显示“你撤回了一条消息”和“重新编辑”按钮
```

## 9. 删除消息事件

单条删除只影响当前用户，不需要推送给对方。

可选事件类型：

```txt
chat.message.deleted
```

用于多端同步当前用户自己的其他设备。

## 10. 好友相关事件

### 10.1 收到好友申请

```txt
friend.request.receive
```

```json
{
  "type": "friend.request.receive",
  "data": {
    "request_id": "111",
    "from_user_id": "123",
    "nickname": "小明",
    "avatar_url": "",
    "message": "我是小明"
  }
}
```

### 10.2 好友申请被接受

```txt
friend.request.accepted
```

### 10.3 好友被删除

```txt
friend.deleted
```

当前 Step 10.6 不依赖该事件刷新页面，也不再向旧 private conversation 写入 system message。
删除好友后的当前操作者提示由 HTTP 操作成功后的前端 toast 和状态刷新完成。
如后续实现该事件，只能用于通知客户端刷新好友列表 / 会话列表，不能用于恢复旧 conversation 或写入旧会话消息。

```json
{
  "type": "friend.deleted",
  "data": {
    "from_user_id": "123"
  }
}
```

## 11. 群聊事件

### 11.1 群消息

群消息也使用：

```txt
chat.message.send
chat.message.ack
chat.message.receive
```

通过 `conversation_id` 区分单聊和群聊。

当 `conversation_type = group` 时，服务端必须走群聊消息逻辑。

客户端发送：

```json
{
  "seq": "tmp-group-001",
  "type": "chat.message.send",
  "data": {
    "conversation_id": "group-conversation-id",
    "client_msg_id": "client-generated-id",
    "message_type": "text",
    "content": "大家好",
    "extra_json": {}
  },
  "timestamp": 1710000000000
}
```

服务端推送：

```json
{
  "seq": "server-999999",
  "type": "chat.message.receive",
  "data": {
    "client_msg_id": "client-generated-id",
    "message_id": "999999",
    "conversation_id": "group-conversation-id",
    "sender_id": "123456789",
    "sender_nickname": "张三",
    "sender_avatar_url": "",
    "sender_group_role": "owner",
    "message_type": "text",
    "content": "大家好",
    "extra_json": {},
    "send_status": "sent",
    "created_at": "2026-04-25 12:00:00"
  },
  "timestamp": 1710000001000
}
```

字段说明：

```txt
sender_group_role：owner / admin / member
sender_group_role 必须由服务端根据 group_members.role 查询
前端不能根据昵称、头像或本地缓存判断群角色
```

群消息发送校验：

```txt
用户必须是群成员
群必须是 normal 状态
用户不能处于禁言期
消息类型当前只支持 text
sender_id 必须来自 WebSocket 鉴权结果
```

失败情况：

```txt
非群成员发送：chat.message.failed
被禁言发送：chat.message.failed，code = group_member_muted
群已解散发送：chat.message.failed，code = group_dissolved
发送方发送成功后只收到 chat.message.ack，不额外收到自己的 chat.message.receive
```

### 11.2 群成员变更

```txt
group.member.changed
```

```json
{
  "type": "group.member.changed",
  "data": {
    "group_id": "888",
    "conversation_id": "111",
    "action": "join",
    "user_id": "123"
  }
}
```

### 11.3 群禁言变更

```txt
group.member.muted
```

## 12. 系统通知

事件类型：

```txt
system.notice
```

```json
{
  "type": "system.notice",
  "data": {
    "level": "info",
    "message": "系统通知内容"
  }
}
```

## 13. 前端发送状态规则

前端发送消息时：

```txt
1. 生成本地 seq
2. 生成 client_msg_id
3. 消息先显示为 sending
4. 收到 chat.message.ack 后替换为 sent，并使用服务端 message_id
5. 收到 chat.message.failed 后显示红色感叹号
6. failed_blocked 显示“对方已拒收你的消息”
```

## 14. 服务端处理规则

服务端收到 `chat.message.send` 后必须：

```txt
校验用户登录
校验 conversation 是否存在
校验用户是否在 conversation 内
校验 client_msg_id 必填且长度不超过 64
同一 sender_id + conversation_id + client_msg_id 重复且内容一致时返回已有 ack
同一 sender_id + conversation_id + client_msg_id 重复且已有消息是 failed_blocked 时返回已有 failed
同一 sender_id + conversation_id + client_msg_id 重复但内容不一致时返回 duplicate_client_msg_id_conflict
如果是群聊，校验是否禁言
如果是单聊，校验是否被接收方拉黑
生成 message_id
写入 MySQL
正常消息更新会话 last_message 并推送给在线接收方
failed_blocked 只为发送方持久化，不更新 last_message，不推送给接收方
给发送方返回 ack 或 failed
如果接收方是默认 Agent 且消息类型为 text，发送方 ack 成功后由后端异步调用 Agent 服务
Agent 回复或失败提示作为新的 text 消息入库，并通过 chat.message.receive 推送给用户
```

被拉黑导致的 `failed_blocked` 写入 messages，只创建发送方 message_user_states，不创建接收方 message_user_states。发送方刷新后可通过历史消息接口看到，接收方永远不可见；解除拉黑后不补发、不转正。

## 15. 多节点预留

MVP 单节点即可。后续多节点时：

```txt
Redis 记录 user_id 所在 server_id
Redis Pub/Sub 或 Redis Stream 转发跨节点消息
Nginx 负责负载均衡
WebSocket 可使用 sticky session 或集中式路由
```
## Agent stream events addendum

Agent replies can use step-level streaming snapshots. This is not token-level streaming.

New events:

```txt
agent.message.start
agent.message.chunk
agent.message.done
agent.message.error
```

`agent.message.chunk` uses replace snapshot semantics:

```json
{
  "stream_id": "agent-stream-1",
  "conversation_id": "111111",
  "client_msg_id": "agent-222222",
  "content": "current display snapshot",
  "chunk_index": 1,
  "mode": "replace",
  "mermaid_pending": false
}
```

`agent.message.done` carries the formal message. For compatibility, the backend then pushes `chat.message.receive` with the same `message_id` and `client_msg_id`; new frontends must dedupe it.
