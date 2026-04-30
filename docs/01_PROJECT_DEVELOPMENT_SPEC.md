# 01 项目总开发文档

## 1. 项目定位

本项目是一个基于 Go 语言开发的网页端即时通讯系统。系统采用前后端分离架构，当前目标是支持少量真实用户进行实时聊天测试，但在技术选型、代码结构、数据库设计、缓存设计、WebSocket 连接管理和消息流程上，需要按照高并发即时通讯系统的思路进行设计，方便后续扩展到多节点部署、消息队列、对象存储、搜索引擎和监控体系。

项目核心功能是聊天。用户注册登录后进入聊天界面，可以搜索用户、添加好友、创建群聊、申请加入群聊，并与好友、群聊成员以及系统默认分配的特殊 Agent 好友进行聊天。

## 2. 技术栈

### 后端

```txt
语言：Go
HTTP 框架：Gin
WebSocket：gorilla/websocket
数据库：MySQL
缓存：Redis
认证：JWT + Redis 双 Token
密码加密：bcrypt
日志：zap 或 logrus
配置：yaml + 环境变量
部署：Docker / Docker Compose
```

### 前端

```txt
框架：Vue 3
构建工具：Vite
语言：TypeScript
状态管理：Pinia
路由：Vue Router
HTTP 请求：Axios
WebSocket：浏览器原生 WebSocket 封装
UI 组件库：Element Plus 或 Naive UI，二选一即可
```

### 存储

```txt
MySQL：用户、好友、会话、消息、群聊、文件元信息
Redis：Token、在线状态、WebSocket 节点信息、撤回重新编辑缓存、限流计数
本地磁盘：MVP 文件存储
后续扩展：MinIO / 阿里云 OSS / 腾讯云 COS / AWS S3
```

## 3. 核心功能范围

### 必须实现

```txt
用户注册登录
JWT + Redis 双 Token
用户资料：头像、性别、昵称、个性签名
默认 Agent 好友
用户搜索：昵称 + user_id
好友申请、同意、拒绝
好友列表
删除好友
拉黑和解除拉黑
单聊文字消息
表情消息
文件消息
文件上传和鉴权下载
聊天记录分页加载
单条消息删除
清空聊天记录
5 分钟内撤回消息
撤回后重新编辑，缓存 5 分钟
消息搜索
文件搜索
创建群聊
群号搜索
申请加入群聊
群主 / 管理员审批入群
群主设置管理员
群禁言
是否允许成员邀请
群主解散群聊
WebSocket 实时通信
Redis 在线状态
Docker Compose 本地部署
```

### 当前不实现，但必须预留

```txt
邮箱验证码
短信验证码
第三方登录
语音消息
语音通话
视频通话
文件在线预览
管理后台
内容审核
多端完整同步
已读回执
Kafka / RabbitMQ
Elasticsearch / Meilisearch
多节点 WebSocket
```

## 4. 用户注册登录规则

注册字段：

```txt
username：登录账号，全局唯一
password：登录密码
nickname：展示昵称，可重复
```

系统生成：

```txt
user_id：雪花算法生成的全局唯一用户 ID
password_hash：bcrypt 加密后的密码
user_type：normal / agent / system
status：normal / disabled / deleted
```

登录方式：

```txt
username + password
```

不做验证码，不做邮箱手机号绑定。

## 5. Token 规则

采用 JWT + Redis 双 Token。

```txt
Access Token：15 分钟有效，用于 API 和 WebSocket 鉴权
Refresh Token：7 天有效，存入 Redis，用于刷新 Access Token
```

建议 Redis Key：

```txt
im:auth:refresh:{user_id}:{device_id}
im:auth:blacklist:{jti}
im:online:{user_id}
im:message:recall_edit:{message_id}:{user_id}
```

## 6. 默认 Agent 好友

每个用户注册成功后，系统自动为该用户分配一个默认特殊好友 Agent。

当前版本：

```txt
只实现 Agent 作为特殊用户和聊天对象
不强制接入真实 AI 服务
```

后续扩展：

```txt
用户向 Agent 发消息
后端识别接收方是 Agent
消息入库
后端调用 Agent 服务
Agent 生成回复
回复写入 messages
通过 WebSocket 推送给用户
```

## 7. 好友规则

### 添加好友

```txt
通过 user_id 精确搜索
通过 nickname 模糊搜索
添加好友需要对方同意
同意后创建双向好友关系和单聊会话
拒绝后申请状态变为 rejected
删除后允许重新添加
```

### 删除好友

```txt
删除好友后，双方好友列表都移除对方
删除好友后，两人之间的旧单聊会话从双方会话列表移除
删除好友后，两人不能继续读取、搜索或发送到旧单聊会话
删除好友不再向旧单聊会话写入 system message
删除好友后的提示由前端操作结果提示和状态刷新完成
删除后允许重新添加
重新添加后必须创建新的空白单聊会话，旧历史不能恢复
```

### 拉黑

```txt
拉黑是单向关系
A 拉黑 B 后，B 仍然可以点击发送消息
后端发现 A 拉黑 B 后，不推送给 A，不进入 A 的聊天记录
B 的消息气泡旁显示红色感叹号，状态为 failed_blocked
B 刷新后仍可在自己的历史消息中看到该 failed_blocked 消息
A 对本次消息无感知
解除拉黑后，拉黑期间的消息不补发
```

## 8. 消息类型

当前支持：

```txt
text：文字
emoji：表情
file：文件
system：系统提示
```

预留：

```txt
image：图片
audio：语音消息
video：视频消息
voice_call_signal：语音通话信令
video_call_signal：视频通话信令
custom：自定义消息
```

消息表必须使用 `message_type + content + extra_json` 的扩展结构，不能把文件、表情、语音等类型写死到多个独立字段里。

## 9. 消息删除、清空、撤回

### 单条删除

```txt
用户可以删除某条消息
删除只影响当前用户
对方不受影响
删除只更新当前用户自己的 message_user_states
failed_blocked 只有发送方拥有 message_user_states，因此只能由发送方在自己视角删除
前端立即移除该消息
提示删除成功
当前用户搜索不到已删除消息
双方都删除后，可由后台任务清理
```

### 清空聊天记录

```txt
用户可以清空某个会话的全部聊天记录
清空只影响当前用户
实现时更新 conversation_user_states.cleared_at
查询历史消息时过滤 created_at <= cleared_at 的消息
```

### 撤回

```txt
只有发送者可以撤回
只能在发送后 5 分钟内撤回
撤回后双方聊天窗口中的原消息都消失
接收方无任何提示
发送方看到浅色小气泡提示：你撤回了一条消息
发送方可以点击重新编辑
重新编辑内容缓存在 Redis 5 分钟
5 分钟后缓存清除，不能重新编辑
被撤回消息不参与搜索
```

## 10. 文件规则

```txt
MVP 使用本地磁盘存储
单文件最大 50MB，可配置
文件类型不限制
不提供在线预览
必须通过鉴权接口下载
只有上传者或会话成员可以下载
文件名支持搜索
已被当前用户删除或清空前的文件消息，不出现在搜索结果中
```

## 11. 群聊规则

```txt
群主为创建群聊的人
默认最大成员数 50，可配置
用户可以通过群号搜索并申请加入群聊
群主或管理员可以审批入群申请
群主可以设置管理员
群主和管理员可以设置成员禁言开始和结束
群主和管理员可以控制是否允许普通成员邀请别人
群主可以解散群聊
群聊解散后不能继续发送消息
有效群成员可以查看自己可见范围内的群历史
用户退出群聊后，该群会话从当前用户会话列表移除，且不能继续读取或发送群消息
用户重新入群后，以本次 joined_at 作为消息可见起点，不能自动恢复退出前或退出期间的旧消息
```

## 12. 高并发设计边界

当前不做过度架构，但必须按照以下规则开发：

```txt
HTTP 接口无状态
Token 和在线状态不存进单机内存
WebSocket 连接管理与业务逻辑分离
聊天记录必须分页加载
消息服务必须封装，便于后续接入 Redis Stream / Kafka
文件服务必须封装，便于后续替换 MinIO / OSS
搜索服务必须封装，便于后续替换 Elasticsearch
所有关键表必须加索引
所有接口必须统一错误码
所有配置必须可通过环境变量覆盖
```
