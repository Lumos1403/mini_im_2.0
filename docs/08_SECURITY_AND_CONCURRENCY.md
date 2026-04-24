# 08 安全与高并发设计文档

## 1. 安全设计

### 1.1 密码安全

```txt
密码必须使用 bcrypt 加密
数据库不能保存明文密码
登录错误不能提示“用户名存在但密码错误”这种过细信息
```

### 1.2 Token 安全

```txt
Access Token 短期有效，建议 15 分钟
Refresh Token 存 Redis，建议 7 天
退出登录删除 Refresh Token
JWT Secret 必须来自环境变量
不能在日志中打印完整 Token
```

当前 Step 2 实现规则：

```txt
Access Token 使用 JWT，默认 15 分钟有效
Refresh Token 使用 JWT 承载 user_id、device_id、jti
Redis 只保存 Refresh Token 的 jti，Key 为 im:auth:refresh:{user_id}:{device_id}
刷新 Token 时轮换 Access Token 和 Refresh Token
退出登录删除 Refresh Token，并将当前 Access Token 的 jti 加入短期黑名单
```

### 1.3 接口鉴权

以下接口必须登录：

```txt
用户资料
好友相关
会话相关
消息相关
文件上传下载
群聊相关
搜索历史消息和文件
WebSocket 连接
```

### 1.4 权限校验

服务端必须校验：

```txt
用户只能查看自己所在会话
用户只能删除自己视角的消息
用户只能撤回自己发送且 5 分钟内的消息
用户只能下载自己有权限访问的文件
群主才能设置管理员
群主和管理员才能禁言
群主才能解散群聊
```

### 1.5 文件安全

```txt
限制单文件大小，默认 50MB
下载必须鉴权
存储路径不能使用用户原始文件名直接拼接
必须防止路径穿越
建议计算 sha256
不提供公开静态目录直接下载
```

### 1.6 输入安全

```txt
所有输入必须做长度限制
nickname、bio、群名需要限制长度
消息内容需要限制最大长度
前端展示用户输入内容时要防 XSS
```

## 2. 当前阶段高并发基础设计

虽然当前真实用户较少，但必须按高并发思维设计。

### 2.1 无状态 HTTP

HTTP API 不依赖单机内存 session，登录态使用 JWT + Redis。

好处：

```txt
后续可以水平扩容多个后端实例
Nginx 可以负载均衡
```

### 2.2 Redis 在线状态

在线状态不能只存在 Go 进程内存中。

Redis Key：

```txt
im:online:{user_id}
```

Value：

```json
{
  "server_id": "ws-1",
  "connected_at": "2026-04-24 12:00:00"
}
```

MVP 可简单记录在线/离线。

### 2.3 WebSocket 连接管理

WebSocket 连接保存在 Hub 中，业务逻辑在 service 中。

```txt
Hub 只负责连接和推送
MessageService 负责发送规则
Repository 负责落库
```

### 2.4 聊天记录分页

禁止一次性加载全部聊天记录。

建议：

```txt
每次加载 20-50 条
上拉加载更多
使用 cursor 分页
```

### 2.5 数据库索引

必须重点优化：

```txt
messages(conversation_id, created_at, message_id)
conversation_user_states(user_id, is_deleted, updated_at)
friendships(user_id_1, user_id_2)
group_members(group_id, user_id)
block_relations(blocker_id, blocked_id)
```

## 3. 消息链路设计

### 3.1 MVP 同步链路

```txt
客户端 WebSocket 发送消息
后端校验权限
写入 MySQL
更新会话 last_message
推送给在线用户
返回 ack
```

优点：简单、容易调试。

### 3.2 后续异步链路

```txt
客户端发送消息
后端校验基础权限
写入 Redis Stream
消费者消费消息
写 MySQL
推送 WebSocket
失败重试
```

优点：抗峰值流量能力更强。

### 3.3 为什么暂时不用 Kafka

当前阶段不建议一开始上 Kafka，因为：

```txt
部署和学习成本高
调试复杂
会拖慢主线功能
当前用户量不需要
```

建议先预留 MessageDispatcher 接口，后续从同步实现替换为 Redis Stream 或 Kafka。

## 4. 多节点 WebSocket 扩展

MVP 单节点。

后续多节点设计：

```txt
Nginx 负载均衡
每个 ws server 有唯一 server_id
Redis 记录 user_id -> server_id
同节点用户直接推送
跨节点用户通过 Redis Pub/Sub 或 Redis Stream 转发
```

## 5. 限流设计

建议实现基础限流：

```txt
登录接口限流
注册接口限流
发送消息限流
文件上传限流
搜索接口限流
```

MVP 可用 Redis 计数器实现。

示例：

```txt
im:rate:login:{ip}
im:rate:send_msg:{user_id}
im:rate:upload:{user_id}
```

## 6. 缓存设计

适合 Redis 的数据：

```txt
Refresh Token
在线状态
撤回重新编辑内容
限流计数
热点用户资料，可选
群成员短期缓存，可选
```

不建议缓存：

```txt
所有聊天记录
大量文件内容
复杂权限长期结果
```

## 7. 搜索扩展

MVP：

```txt
MySQL LIKE 或 FULLTEXT
```

后续：

```txt
Elasticsearch / Meilisearch
```

必须通过 SearchService 封装，避免 controller 直接写 SQL。

## 8. 文件存储扩展

MVP：

```txt
本地 uploads 目录
Docker volume 挂载
```

后续：

```txt
MinIO 自建对象存储
云 OSS / COS / S3
```

必须通过 FileStorage 接口封装。

## 9. 监控预留

后续建议接入：

```txt
Prometheus
Grafana
Loki
Jaeger / OpenTelemetry
```

关键指标：

```txt
在线用户数
WebSocket 连接数
消息发送 QPS
消息失败数
数据库慢查询
Redis 错误数
文件上传失败数
接口响应时间
```
