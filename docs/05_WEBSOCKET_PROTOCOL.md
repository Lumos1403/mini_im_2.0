# 05 WebSocket 协议设计文档

## 1. 连接地址

本地开发：

```txt
ws://localhost:8080/ws?token=<access_token>
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
写入 Redis 在线状态
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
超过 60 秒无响应则断开
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

## 4. 客户端发送消息

### 4.1 发送单聊 / 群聊消息

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
    "message_type": "text",
    "content": "你好",
    "extra_json": {}
  },
  "timestamp": 1710000000000
}
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

## 5. 服务端确认 ACK

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
    "message_id": "999999",
    "conversation_id": "111111",
    "send_status": "sent",
    "created_at": "2026-04-24 12:00:00"
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
    "conversation_id": "111111",
    "send_status": "failed_blocked",
    "reason": "对方已拒收你的消息"
  },
  "timestamp": 1710000001000
}
```

## 6. 服务端推送新消息

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
    "message_id": "999999",
    "conversation_id": "111111",
    "sender_id": "123456",
    "message_type": "text",
    "content": "你好",
    "extra_json": {},
    "created_at": "2026-04-24 12:00:00"
  },
  "timestamp": 1710000001000
}
```

## 7. 撤回消息事件

撤回通过 HTTP API 发起，WebSocket 用于通知相关在线用户移除消息。

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
接收方：直接移除该消息，不显示任何提示
发送方：移除该消息，显示“你撤回了一条消息”和“重新编辑”按钮
```

## 8. 删除消息事件

单条删除只影响当前用户，不需要推送给对方。

可选事件类型：

```txt
chat.message.deleted
```

用于多端同步当前用户自己的其他设备。

## 9. 好友相关事件

### 9.1 收到好友申请

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

### 9.2 好友申请被接受

```txt
friend.request.accepted
```

### 9.3 好友被删除

```txt
friend.deleted
```

```json
{
  "type": "friend.deleted",
  "data": {
    "from_user_id": "123",
    "conversation_id": "111",
    "notice": "对方已将你删除"
  }
}
```

## 10. 群聊事件

### 10.1 群消息

群消息也使用：

```txt
chat.message.receive
```

通过 conversation_id 区分。

### 10.2 群成员变更

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

### 10.3 群禁言变更

```txt
group.member.muted
```

## 11. 系统通知

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

## 12. 前端发送状态规则

前端发送消息时：

```txt
1. 生成本地 seq
2. 消息先显示为 sending
3. 收到 chat.message.ack 后替换为 sent，并使用服务端 message_id
4. 收到 chat.message.failed 后显示红色感叹号
5. failed_blocked 显示“对方已拒收你的消息”
```

## 13. 服务端处理规则

服务端收到 `chat.message.send` 后必须：

```txt
校验用户登录
校验 conversation 是否存在
校验用户是否在 conversation 内
如果是群聊，校验是否禁言
如果是单聊，校验是否被接收方拉黑
生成 message_id
写入 MySQL
更新会话 last_message
推送给在线接收方
给发送方返回 ack
```

## 14. 多节点预留

MVP 单节点即可。后续多节点时：

```txt
Redis 记录 user_id 所在 server_id
Redis Pub/Sub 或 Redis Stream 转发跨节点消息
Nginx 负责负载均衡
WebSocket 可使用 sticky session 或集中式路由
```
