# 07 Vue 前端架构文档

## 1. 前端目录结构

推荐结构：

```txt
frontend/
  src/
    api/
      auth.ts
      user.ts
      friend.ts
      conversation.ts
      message.ts
      group.ts
      file.ts
      search.ts
    assets/
    components/
      common/
      layout/
      chat/
      friend/
      group/
    views/
      Login.vue
      Register.vue
      Chat.vue
      Profile.vue
    router/
      index.ts
    stores/
      auth.ts
      user.ts
      chat.ts
      ws.ts
      friend.ts
      group.ts
    utils/
      request.ts
      token.ts
      websocket.ts
      time.ts
      file.ts
    types/
      api.ts
      user.ts
      chat.ts
      group.ts
    App.vue
    main.ts
  vite.config.ts
  package.json
  Dockerfile
```

## 2. 页面规划

### 登录页

路径：

```txt
/login
```

功能：

```txt
输入 username
输入 password
登录
跳转聊天页
```

### 注册页

路径：

```txt
/register
```

功能：

```txt
输入 username
输入 password
输入 nickname
注册成功后跳转登录或自动登录
```

### 聊天主界面

路径：

```txt
/chat
```

布局：

```txt
左侧：会话列表
中间：聊天窗口
右侧：好友面板 / 用户资料 / 群资料 / 成员列表，可折叠
顶部：当前聊天对象信息
底部：消息输入区
```

当前好友面板：

```txt
搜索用户
好友申请列表
好友列表
打开聊天
删除好友
拉黑 / 解除拉黑
```

### 个人资料页

路径：

```txt
/profile
```

功能：

```txt
修改头像
修改昵称
修改性别
修改个性签名
```

## 3. 状态管理

### auth store

保存：

```txt
access_token
refresh_token
user 信息
是否登录
```

方法：

```txt
login
register
refreshToken
logout
```

### chat store

保存：

```txt
会话列表
当前会话
当前会话消息
消息发送状态
未读数
```

方法：

```txt
loadConversations
loadMessages
appendMessage
updateMessageStatus
removeMessage
clearConversation
recallMessage
```

### ws store

保存：

```txt
WebSocket 实例
连接状态
重连次数
```

方法：

```txt
connect
disconnect
sendMessage
handleEvent
reconnect
```

### friend store

保存：

```txt
好友列表
收到的好友申请
用户搜索结果
好友操作 loading / error / notice 状态
```

方法：

```txt
loadFriends
loadReceivedRequests
search
sendFriendRequest
acceptRequest
rejectRequest
removeFriend
block
unblock
```

说明：

```txt
好友列表拉黑状态以后端 is_blocked_by_me 为准。
打开好友聊天时优先使用好友列表返回的 conversation_id。
缺少 conversation_id 时只能使用会话列表的 peer_user_id 兜底匹配，不允许按 nickname 或 avatar_url 匹配。
```

## 4. HTTP 请求封装

Axios 必须统一封装：

```txt
自动添加 Authorization
Access Token 过期后自动调用 refresh
统一处理错误码
统一弹出错误提示
```

禁止在页面组件中直接写 axios 原始调用。

## 5. WebSocket 封装

前端必须只维护一个 WebSocket 连接。

连接时机：

```txt
登录成功后连接
刷新页面且本地有 token 时重连
退出登录时断开
```

发送消息流程：

```txt
生成 seq
本地插入 sending 消息
通过 WebSocket 发送
收到 ack 后替换 message_id 和状态
收到 failed 后显示红色感叹号
```

## 6. 消息组件规则

消息气泡需要支持：

```txt
自己发送 / 对方发送样式区分
发送中状态
发送失败红色感叹号
文字消息
表情消息
文件消息
系统提示消息
撤回提示
重新编辑按钮
```

系统提示样式：

```txt
居中
浅色小圆角矩形
字体较小
不作为普通消息气泡展示
```

## 7. 删除和撤回交互

### 删除单条消息

```txt
用户右键或长按消息
点击删除
调用 DELETE /api/messages/{message_id}
成功后当前页面立即移除
显示删除成功提示
```

### 清空聊天记录

```txt
点击会话设置中的清空聊天记录
二次确认
调用 DELETE /api/conversations/{id}/messages
成功后当前会话消息列表清空
显示删除成功
```

### 撤回消息

```txt
只有自己发送的消息显示撤回按钮
发送 5 分钟内可点击
调用 POST /api/messages/{id}/recall
成功后消息移除
显示“你撤回了一条消息”提示
提示中显示“重新编辑”按钮
点击重新编辑后请求 GET /api/messages/{id}/recall-edit-cache
把内容填回输入框
```

## 8. 文件消息交互

```txt
用户选择文件
先调用文件上传接口
上传成功获得 file_id
再通过 WebSocket 发送 file 消息
聊天窗口展示文件名、大小、下载按钮
点击下载时调用鉴权下载接口
```

MVP 不做预览。

## 9. 搜索交互

搜索入口：

```txt
顶部搜索框或会话内搜索按钮
```

支持：

```txt
搜索用户
搜索历史消息
搜索文件名
搜索群号
```

搜索结果点击后跳转到对应会话或用户资料。

## 10. 群聊交互

必须支持：

```txt
创建群聊
群号搜索
申请入群
审批入群申请
查看群成员
设置管理员
禁言成员
修改是否允许成员邀请
解散群聊
```

权限按钮必须根据当前用户角色显示，但最终权限仍由服务端校验。

## 11. 路由守卫

```txt
未登录访问 /chat 自动跳转 /login
已登录访问 /login 可跳转 /chat
Token 失效则清空登录状态并跳转 /login
```

## 12. 样式建议

即时通讯界面建议：

```txt
左侧会话列表固定宽度
中间聊天窗口自适应
消息列表虚拟滚动预留
输入框支持 Enter 发送，Shift+Enter 换行
文件上传显示进度
失败消息可点击重试
```

## 13. 前端验收重点

```txt
登录后自动连接 WebSocket
刷新页面后可恢复登录状态
发送消息有 sending -> sent 状态变化
被拉黑时显示红色感叹号
删除消息后立即消失
撤回后对方消息消失且无提示
文件下载未登录不可访问
群禁言后发送按钮禁用或发送失败提示
```
