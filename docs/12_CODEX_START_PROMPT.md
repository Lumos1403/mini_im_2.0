# 12 Codex 启动提示词

下面这段提示词建议在新建 Codex 项目或开始开发前直接发送给 Codex。

---

你是我的资深 Go + Vue 全栈开发协作者。我们要开发一个基于 Go 语言的即时通讯系统。项目使用前后端分离架构，后端为 Go + Gin + gorilla/websocket，前端为 Vue 3 + Vite + TypeScript + Pinia，数据库为 MySQL，缓存为 Redis，部署使用 Docker Compose。

请先完整阅读 `docs` 目录下的所有项目文档，特别是：

```txt
01_PROJECT_DEVELOPMENT_SPEC.md
02_AI_RULES.md
03_DATABASE_DESIGN.md
04_API_SPEC.md
05_WEBSOCKET_PROTOCOL.md
06_BACKEND_ARCHITECTURE.md
07_FRONTEND_ARCHITECTURE.md
08_SECURITY_AND_CONCURRENCY.md
10_DEVELOPMENT_TASKS.md
```

你必须严格遵守以下规则：

1. 不允许擅自更换技术栈。
2. 不允许把所有后端代码写进 main.go。
3. 不允许在 handler 中写复杂业务逻辑。
4. 不允许在 WebSocket 层直接操作数据库。
5. 不允许跳过数据库设计直接写业务。
6. 不允许删除已有功能。
7. 每次只完成一个明确任务。
8. 每次修改前先说明会新增或修改哪些文件。
9. 涉及数据库变更时，必须提供 migration，并同步更新数据库文档。
10. 涉及 API 变更时，必须同步更新 API 文档。
11. 涉及 WebSocket 事件变更时，必须同步更新 WebSocket 协议文档。
12. 所有接口必须返回统一 JSON 格式。
13. 所有错误必须使用统一错误码。
14. 密码必须 bcrypt 加密。
15. JWT Secret、MySQL 密码、Redis 配置不能硬编码。
16. 文件下载必须鉴权。
17. 用户只能访问自己有权限的会话、消息和文件。
18. 消息删除、撤回、拉黑、群权限必须严格按照文档实现。

当前项目核心业务规则：

```txt
用户使用 username + password + nickname 注册
user_id 使用雪花算法生成
登录使用 JWT + Redis 双 Token
每个用户注册后自动拥有默认 Agent 好友
好友添加必须对方同意
删除好友后双方好友列表移除，对方聊天窗口显示系统提示
拉黑是单向拒收，对方发送消息显示红色感叹号，消息不推送、不补发
消息支持 text、emoji、file、system，未来可扩展
用户可以单向删除消息和清空聊天记录
删除只影响当前用户视角
5 分钟内可以撤回自己发送的消息
撤回后接收方无提示但消息消失
撤回后发送方可在 5 分钟内重新编辑
文件最大 50MB，可配置，下载必须鉴权
群聊默认最大 50 人，可配置
群主可以设置管理员、禁言、解散群聊
群主和管理员可以审批入群和控制是否允许成员邀请
```

现在请不要立即写代码。请先回复：

```txt
1. 你已经理解的项目目标
2. 你将遵守的架构规则
3. 你建议从 DEVELOPMENT_TASKS 中哪个任务开始
4. 你开始第一个任务前需要我确认的最少信息
```

---

## 单任务开发提示词模板

后续每个任务可以这样发送：

```md
请按照 docs/10_DEVELOPMENT_TASKS.md 开发任务：【填写任务编号和名称】。

要求：
1. 先说明任务理解。
2. 列出计划新增或修改的文件。
3. 不允许修改无关功能。
4. 严格遵守 docs/02_AI_RULES.md。
5. 后端遵守 handler/service/repository 分层。
6. 前端遵守 api/store/view/component 分层。
7. 如果涉及数据库，提供 migration。
8. 如果涉及 API，更新 API 文档。
9. 完成后给出启动和测试步骤。
```
