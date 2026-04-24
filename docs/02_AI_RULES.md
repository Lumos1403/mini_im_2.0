# 02 Codex / AI 开发规则

## 1. 总原则

本项目使用 Codex 辅助开发。AI 必须严格按照已有文档、目录结构、接口规范和业务规则执行，不能自行发挥、不能随意更换技术栈、不能一次性生成超大范围代码。

AI 的角色是：

```txt
资深 Go + Vue 全栈开发协作者
```

AI 必须做到：

```txt
先读文档
先说明计划
再改代码
每次只改一个明确任务
改完说明测试方法
```

## 2. 禁止行为

Codex 不允许：

```txt
不允许擅自更换技术栈
不允许把后端全部代码写进 main.go
不允许把业务逻辑写在 handler 中
不允许在 WebSocket 层直接操作数据库
不允许删除已有功能
不允许大范围重构无关代码
不允许跳过数据库设计直接写业务
不允许硬编码数据库密码、JWT Secret、Redis 地址
不允许跳过错误处理
不允许忽略鉴权和权限校验
不允许让前端决定核心权限
不允许直接公开文件 URL 让任何人下载
不允许让用户伪造 sender_id
不允许一次性实现多个大模块
```

## 3. 每次开发前必须输出

每次让 Codex 开发时，必须先让它输出：

```txt
1. 当前任务理解
2. 涉及的业务规则
3. 计划新增或修改的文件
4. 数据库是否需要变更
5. API 是否需要变更
6. WebSocket 协议是否需要变更
7. 测试方法
```

只有确认后再生成代码。

## 4. 修改范围规则

每次任务必须限制修改范围。

示例：

```txt
本次只实现用户注册接口，不实现登录、不实现前端页面、不实现好友系统。
```

如果 Codex 发现需要修改额外文件，必须先说明原因。

## 5. 后端分层规则

后端必须遵守：

```txt
handler：参数绑定、调用 service、返回响应
service：业务逻辑
repository：数据库读写
model：数据结构
middleware：中间件
ws：WebSocket 连接管理和消息转发
pkg：通用工具
```

禁止：

```txt
handler 直接写复杂 SQL
handler 直接操作 Redis 业务 Key
WebSocket client 直接操作 MySQL
repository 写业务判断
```

## 6. 前端分层规则

前端必须遵守：

```txt
api：HTTP 请求封装
stores：Pinia 状态
views：页面
components：组件
utils：通用工具
router：路由
```

禁止：

```txt
页面组件里到处直接写 axios
页面组件里到处直接写 localStorage token 逻辑
多个地方重复创建 WebSocket 连接
组件里硬编码后端地址
```

## 7. 数据库规则

涉及数据库时必须：

```txt
先更新 docs/03_DATABASE_DESIGN.md
再写 migration
字段必须有明确类型
关键字段必须建索引
状态字段必须列出枚举值
时间字段统一使用 created_at / updated_at / deleted_at
业务删除优先逻辑删除
```

不得随意物理删除核心业务数据，除非文档明确允许。

## 8. API 规则

所有 HTTP API 必须：

```txt
使用 /api 前缀
需要登录的接口必须经过 AuthMiddleware
返回统一 JSON 格式
错误必须使用统一错误码
分页接口必须支持 page/page_size 或 cursor/limit
不能返回 password_hash
不能返回 Redis 内部 Key
```

统一响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

## 9. WebSocket 规则

WebSocket 必须：

```txt
连接时校验 Access Token
连接建立后注册 user_id 与连接
断开时清理连接和在线状态
实现 ping/pong 心跳
所有事件使用统一 envelope 格式
消息发送必须有 ack
失败必须返回明确 failed event
```

禁止客户端直接相信本地发送成功，必须等待服务端 ack 后确认。

## 10. 安全规则

必须遵守：

```txt
密码必须 bcrypt
JWT Secret 必须来自环境变量
文件下载必须鉴权
用户只能访问自己所在会话
用户只能操作自己的消息状态
撤回只能撤回自己 5 分钟内的消息
群权限必须服务端校验
文件大小必须限制
上传路径必须防止路径穿越
```

## 11. 高并发预留规则

当前版本可以同步处理消息，但必须预留接口：

```txt
MessageDispatcher
MessageQueue
FileStorage
SearchService
OnlineStatusStore
```

后续可以把同步发送替换为：

```txt
Redis Stream / Kafka 异步消息队列
MinIO / OSS 文件存储
Elasticsearch / Meilisearch 搜索
Redis Pub/Sub 多节点消息推送
```

## 12. 每次完成后必须输出

Codex 每次完成任务后，必须说明：

```txt
1. 完成了什么
2. 修改了哪些文件
3. 新增了哪些接口
4. 数据库是否变更
5. 如何启动
6. 如何测试
7. 还有哪些未完成或风险点
```

## 13. 推荐给 Codex 的单任务提示词模板

```md
你是本项目的 Go + Vue 全栈开发助手。请先阅读 docs 文件夹内的项目文档，严格遵守 AI 开发规则。

当前任务：
【只写一个明确任务】

要求：
1. 先说明你对任务的理解。
2. 列出会新增或修改的文件。
3. 不允许修改无关模块。
4. 不允许更换技术栈。
5. 后端必须遵守 handler/service/repository 分层。
6. 前端必须遵守 api/store/view/component 分层。
7. 涉及数据库变更必须提供 migration。
8. 涉及 API 变更必须更新 API 文档。
9. 完成后给出测试步骤。

请开始。
```
