# 06 Go 后端架构文档

## 1. 后端目录结构

推荐结构：

```txt
backend/
  cmd/
    server/
      main.go
  internal/
    api/
      handler/
      router/
    middleware/
    service/
      auth/
      user/
      friend/
      conversation/
      message/
      group/
      file/
      search/
      agent/
    repository/
      mysql/
      redis/
    model/
    ws/
      hub.go
      client.go
      message.go
      dispatcher.go
    config/
    pkg/
      jwt/
      snowflake/
      response/
      errors/
      logger/
      password/
      validator/
  migrations/
  configs/
    config.example.yaml
  uploads/
  Dockerfile
  go.mod
```

## 2. 分层职责

### handler

职责：

```txt
参数绑定
参数基础校验
调用 service
返回统一响应
```

不允许：

```txt
写复杂业务逻辑
直接写 SQL
直接操作 Redis 业务 key
直接处理 WebSocket 连接
```

### service

职责：

```txt
业务规则判断
调用 repository
事务编排
权限校验
消息流程控制
```

### repository

职责：

```txt
MySQL 读写
Redis 读写
数据库查询封装
```

不应包含复杂业务判断。

### ws

职责：

```txt
WebSocket 连接管理
在线用户连接映射
心跳
读写协程
事件分发
消息推送
```

不允许在 ws 层直接写业务数据库逻辑，必须调用 service 或 dispatcher。

## 3. 关键模块

### 3.1 auth 模块

实现：

```txt
注册
登录
刷新 Token
退出登录
JWT 生成和解析
Redis Refresh Token
bcrypt 密码校验
```

### 3.2 user 模块

实现：

```txt
获取当前用户资料
修改资料
搜索用户
初始化 Agent 用户
```

### 3.3 friend 模块

实现：

```txt
好友申请
同意 / 拒绝
好友列表
删除好友
拉黑 / 解除拉黑
好友关系判断
```

### 3.4 conversation 模块

实现：

```txt
创建单聊会话
创建群聊会话
获取会话列表
用户会话状态
清空聊天记录
未读数
置顶和免打扰
```

### 3.5 message 模块

实现：

```txt
发送消息
写入消息
查询历史消息
删除消息
撤回消息
重新编辑缓存
消息搜索过滤
```

### 3.6 group 模块

实现：

```txt
创建群聊
申请入群
审批入群
群成员管理
设置管理员
禁言
修改群设置
解散群聊
```

### 3.7 file 模块

实现：

```txt
文件上传
文件元信息保存
文件下载鉴权
文件搜索
本地存储
后续对象存储扩展
```

### 3.8 search 模块

实现：

```txt
用户搜索
消息搜索
文件搜索
```

搜索逻辑必须封装，后续可替换 Elasticsearch。

### 3.9 agent 模块

实现：

```txt
默认 Agent 用户 ensure
注册和登录补偿时默认 Agent 好友 / private conversation ensure
封装 FastAPI Agent 同步 /api/chat 调用
Agent 回复和失败提示落库后通过 WebSocket 推送
后续可替换为 /api/chat/stream SSE
```

## 4. 配置设计

配置示例：

```yaml
server:
  port: 8081
  mode: debug

mysql:
  dsn: "root:password@tcp(mysql:3306)/go_im?charset=utf8mb4&parseTime=True&loc=Local"

redis:
  addr: "redis:6379"
  password: ""
  db: 0

jwt:
  access_secret: "change-me"
  refresh_secret: "change-me"
  access_expire_minutes: 15
  refresh_expire_days: 7

file:
  storage_type: "local"
  local_path: "./uploads"
  max_size_mb: 50

im:
  group_max_members: 50
  recall_minutes: 5

agent:
  enabled: true
  api_base_url: "http://127.0.0.1:8100"
  api_timeout_seconds: 30
  default_username: "default_agent"
  default_nickname: "IM Agent"
  default_avatar_url: ""
```

生产环境必须通过环境变量覆盖敏感配置。

## 5. 中间件

必须实现：

```txt
RecoveryMiddleware
RequestIDMiddleware
LoggerMiddleware
AuthMiddleware
RateLimitMiddleware
CORS Middleware
```

## 6. 统一响应

pkg/response：

```go
type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data"`
}
```

## 7. 错误处理

pkg/errors 定义业务错误码。

service 返回业务错误，handler 统一转换为 JSON。

## 8. 事务要求

以下操作必须使用事务：

```txt
注册用户 + 创建用户资料 + 创建 Agent 好友 + 创建会话
同意好友申请 + 创建好友关系 + 创建会话
创建群聊 + 创建群会话 + 创建群成员
发送消息 + 更新会话 last_message
撤回消息 + 更新消息状态 + Redis 缓存重新编辑内容
```

## 9. WebSocket Hub 设计

Hub 维护：

```txt
map[user_id][]*Client
register channel
unregister channel
sendToUser 方法
sendToConversation 方法
```

MVP 可先支持单设备连接，后续扩展多设备。

## 10. 文件存储抽象

定义接口：

```go
type FileStorage interface {
    Save(ctx context.Context, file multipart.File, filename string) (*StoredFile, error)
    Open(ctx context.Context, storagePath string) (io.ReadCloser, error)
    Delete(ctx context.Context, storagePath string) error
}
```

MVP 实现 LocalFileStorage，后续实现 MinIOStorage。

## 11. 消息队列抽象

定义接口：

```go
type MessageDispatcher interface {
    Dispatch(ctx context.Context, msg *MessageEvent) error
}
```

MVP 实现 SyncMessageDispatcher。

后续可实现 RedisStreamDispatcher、KafkaDispatcher。

## 12. 日志要求

必须记录：

```txt
用户注册登录
Token 刷新失败
WebSocket 连接和断开
消息发送失败
文件上传失败
权限校验失败
群权限操作
系统异常
```

日志不能记录明文密码、完整 Token。
## Agent stream architecture addendum

Agent stream replies are handled in the existing layered flow:

```txt
ws dispatcher -> message service -> agent service -> agentclient
agent service -> agent notifier -> websocket hub
```

`agentclient` owns the tolerant SSE reader for `/api/chat/stream`. The stream payload is plain text `data:` content, not JSON and not token deltas.

`AgentService` owns business flow: create `stream_id`, push start/chunk/error/done events, filter prompt echoes and empty chunks, persist one final Agent text message, then push a compatibility `chat.message.receive`.

`ws.AgentNotifier` only marshals WebSocket envelopes and sends them through `Hub`; it does not own Agent filtering or database logic.
