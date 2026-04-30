# 03 数据库设计文档

## 1. 设计原则

数据库使用 MySQL。核心原则：

```txt
使用雪花算法生成业务 ID
MySQL 自增 id 仅作为内部主键
业务查询优先使用 user_id / message_id / conversation_id 等业务 ID
关键查询字段必须建索引
不轻易物理删除核心数据
用户侧删除使用状态表实现
聊天记录分页加载
消息类型使用 type + extra_json 支持扩展
```

## 2. ID 设计

每张核心表可同时拥有：

```txt
id：BIGINT AUTO_INCREMENT，数据库内部主键
业务 ID：BIGINT，例如 user_id、message_id、conversation_id、group_id、file_id
```

业务 ID 使用雪花算法生成，便于后续分布式扩展。

## 3. 状态枚举

### 用户类型 user_type

```txt
normal：普通用户
agent：系统 Agent
system：系统用户
```

### 用户状态 user_status

```txt
normal：正常
disabled：禁用
deleted：注销或删除
```

### 好友申请状态 request_status

```txt
pending：待处理
accepted：已同意
rejected：已拒绝
expired：已过期
```

### 好友关系状态 friendship_status

```txt
normal：正常
deleted：已删除
```

### 会话类型 conversation_type

```txt
private：单聊
group：群聊
agent：Agent 会话，MVP 可归为 private
system：系统通知，预留
```

### 消息类型 message_type

```txt
text
emoji
file
system
image
audio
video
voice_call_signal
video_call_signal
custom
```

### 消息发送状态 send_status

```txt
sent
failed
failed_blocked
recalled
```

### 群成员角色 group_role

```txt
owner
admin
member
```

### 群状态 group_status

```txt
normal
dissolved
disabled
```

## 4. 建表 SQL

> 说明：以下 SQL 作为初始 migration 参考。Codex 开发时可根据实际 ORM 迁移工具拆分。MySQL 建议使用 `utf8mb4`。

```sql
CREATE DATABASE IF NOT EXISTS go_im DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE go_im;
```

### 4.1 users 用户主表

```sql
CREATE TABLE users (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  username VARCHAR(64) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  user_type VARCHAR(20) NOT NULL DEFAULT 'normal',
  status VARCHAR(20) NOT NULL DEFAULT 'normal',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME NULL,
  UNIQUE KEY uk_users_user_id (user_id),
  UNIQUE KEY uk_users_username (username),
  INDEX idx_users_status (status),
  INDEX idx_users_type (user_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 4.2 user_profiles 用户资料表

```sql
CREATE TABLE user_profiles (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  nickname VARCHAR(64) NOT NULL,
  avatar_url VARCHAR(512) NULL,
  gender VARCHAR(20) NULL,
  bio VARCHAR(255) NULL,
  profile_status VARCHAR(30) NOT NULL DEFAULT 'normal',
  profile_review_reason VARCHAR(255) NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_profiles_user_id (user_id),
  INDEX idx_profiles_nickname (nickname),
  CONSTRAINT fk_profiles_user FOREIGN KEY (user_id) REFERENCES users(user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 4.3 friend_requests 好友申请表

```sql
CREATE TABLE friend_requests (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  request_id BIGINT NOT NULL,
  from_user_id BIGINT NOT NULL,
  to_user_id BIGINT NOT NULL,
  message VARCHAR(255) NULL,
  status VARCHAR(20) NOT NULL DEFAULT 'pending',
  pending_pair_key VARCHAR(64)
    GENERATED ALWAYS AS (
      CASE
        WHEN status = 'pending' THEN CONCAT(LEAST(from_user_id, to_user_id), ':', GREATEST(from_user_id, to_user_id))
        ELSE NULL
      END
    ) STORED,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_friend_requests_request_id (request_id),
  UNIQUE KEY uk_friend_requests_pending_pair (pending_pair_key),
  INDEX idx_friend_requests_to_status (to_user_id, status),
  INDEX idx_friend_requests_from_status (from_user_id, status),
  CONSTRAINT fk_friend_req_from FOREIGN KEY (from_user_id) REFERENCES users(user_id),
  CONSTRAINT fk_friend_req_to FOREIGN KEY (to_user_id) REFERENCES users(user_id),
  CONSTRAINT chk_friend_req_not_self CHECK (from_user_id <> to_user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

说明：`pending_pair_key` 使用 MySQL 8 生成列，只对待处理申请建立唯一约束。这样可以防止同一对用户同时存在多条 pending 申请，同时允许历史 accepted/rejected 记录保留。

### 4.4 friendships 好友关系表

```sql
CREATE TABLE friendships (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  user_id_1 BIGINT NOT NULL,
  user_id_2 BIGINT NOT NULL,
  status VARCHAR(20) NOT NULL DEFAULT 'normal',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_friendship_pair (user_id_1, user_id_2),
  INDEX idx_friendships_user1_status (user_id_1, status),
  INDEX idx_friendships_user2_status (user_id_2, status),
  CONSTRAINT fk_friendships_user1 FOREIGN KEY (user_id_1) REFERENCES users(user_id),
  CONSTRAINT fk_friendships_user2 FOREIGN KEY (user_id_2) REFERENCES users(user_id),
  CONSTRAINT chk_friendships_user_order CHECK (user_id_1 < user_id_2)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

要求：`user_id_1 < user_id_2`，由 service 层保证，数据库 CHECK 约束兜底。

### 4.5 block_relations 拉黑关系表

```sql
CREATE TABLE block_relations (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  blocker_id BIGINT NOT NULL,
  blocked_id BIGINT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_block_pair (blocker_id, blocked_id),
  INDEX idx_blocker (blocker_id),
  INDEX idx_blocked (blocked_id),
  CONSTRAINT fk_blocker FOREIGN KEY (blocker_id) REFERENCES users(user_id),
  CONSTRAINT fk_blocked FOREIGN KEY (blocked_id) REFERENCES users(user_id),
  CONSTRAINT chk_block_not_self CHECK (blocker_id <> blocked_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 4.6 conversations 会话表

```sql
CREATE TABLE conversations (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  conversation_id BIGINT NOT NULL UNIQUE,
  conversation_type VARCHAR(20) NOT NULL,
  ref_id BIGINT NULL,
  last_message_id BIGINT NULL,
  last_message_at DATETIME NULL,
  status VARCHAR(20) NOT NULL DEFAULT 'normal',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_conversations_type_ref (conversation_type, ref_id),
  INDEX idx_conversations_last_message_at (last_message_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

说明：

```txt
private 会话 ref_id 可以为空，成员由 conversation_members 表决定
group 会话 ref_id = group_id
```

### 4.7 conversation_members 会话成员表

```sql
CREATE TABLE conversation_members (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  conversation_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  role VARCHAR(20) NOT NULL DEFAULT 'member',
  status VARCHAR(20) NOT NULL DEFAULT 'active',
  joined_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  left_at DATETIME NULL,
  UNIQUE KEY uk_conversation_member (conversation_id, user_id),
  INDEX idx_conv_members_user (user_id, status),
  INDEX idx_conv_members_conv (conversation_id, status),
  CONSTRAINT fk_conv_member_conv FOREIGN KEY (conversation_id) REFERENCES conversations(conversation_id),
  CONSTRAINT fk_conv_member_user FOREIGN KEY (user_id) REFERENCES users(user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 4.8 conversation_user_states 用户会话状态表

```sql
CREATE TABLE conversation_user_states (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  conversation_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  is_deleted TINYINT(1) NOT NULL DEFAULT 0,
  cleared_at DATETIME NULL,
  last_read_message_id BIGINT NULL,
  last_read_at DATETIME NULL,
  unread_count INT NOT NULL DEFAULT 0,
  is_pinned TINYINT(1) NOT NULL DEFAULT 0,
  is_muted TINYINT(1) NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_conv_user_state (conversation_id, user_id),
  INDEX idx_conv_state_user (user_id, is_deleted, updated_at),
  INDEX idx_conv_state_conv (conversation_id),
  CONSTRAINT fk_conv_state_conv FOREIGN KEY (conversation_id) REFERENCES conversations(conversation_id),
  CONSTRAINT fk_conv_state_user FOREIGN KEY (user_id) REFERENCES users(user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

用途：记录某个用户在某个会话中的状态，例如清空时间、未读数、置顶、免打扰。

当前实现：对应 `backend/migrations/003_create_conversation_system.sql`。private 会话用于好友单聊；group 会话用于群聊。`conversation_members.status` 和 `conversation_user_states.is_deleted` 同时参与会话列表可见性过滤。

关系变更规则：

```txt
删除好友时，双方之间的旧 private conversation 及其 messages / message_user_states / conversation_user_states / conversation_members 会在同一事务内物理删除。
退出群聊时，只将当前用户的 group_members 和 conversation_members 标记为 left，并将当前用户的 conversation_user_states.is_deleted 置为 1，不删除群聊和群消息。
```

### 4.9 messages 消息主表

```sql
CREATE TABLE messages (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  message_id BIGINT NOT NULL UNIQUE,
  conversation_id BIGINT NOT NULL,
  sender_id BIGINT NOT NULL,
  client_msg_id VARCHAR(64) NOT NULL,
  message_type VARCHAR(30) NOT NULL,
  content TEXT NULL,
  extra_json JSON NULL,
  send_status VARCHAR(30) NOT NULL DEFAULT 'sent',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  recalled_at DATETIME NULL,
  recalled_by BIGINT NULL,
  is_deleted_all TINYINT(1) NOT NULL DEFAULT 0,
  UNIQUE KEY uk_sender_conversation_client_msg (sender_id, conversation_id, client_msg_id),
  INDEX idx_messages_conv_time (conversation_id, created_at, message_id),
  INDEX idx_messages_conv_msg (conversation_id, message_id),
  INDEX idx_messages_sender_time (sender_id, created_at),
  INDEX idx_messages_type (message_type),
  INDEX idx_messages_recalled (recalled_at),
  CONSTRAINT fk_messages_conv FOREIGN KEY (conversation_id) REFERENCES conversations(conversation_id),
  CONSTRAINT fk_messages_sender FOREIGN KEY (sender_id) REFERENCES users(user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

说明：

```txt
client_msg_id 由客户端生成，用于同一 sender_id + conversation_id 下的发送幂等，最大长度 64
content 保存文本内容或 file_id 字符串
extra_json 保存扩展信息
撤回后 content 可置空，recalled_at 不为空
```

### 4.10 message_user_states 用户消息状态表

```sql
CREATE TABLE message_user_states (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  message_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  is_deleted TINYINT(1) NOT NULL DEFAULT 0,
  deleted_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_message_user_state (message_id, user_id),
  INDEX idx_msg_state_user_deleted (user_id, is_deleted),
  INDEX idx_msg_state_message (message_id),
  CONSTRAINT fk_msg_state_msg FOREIGN KEY (message_id) REFERENCES messages(message_id),
  CONSTRAINT fk_msg_state_user FOREIGN KEY (user_id) REFERENCES users(user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 4.11 groups 群聊表

```sql
CREATE TABLE groups (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  group_id BIGINT NOT NULL UNIQUE,
  group_no VARCHAR(10) NOT NULL UNIQUE,
  conversation_id BIGINT NOT NULL UNIQUE,
  owner_id BIGINT NOT NULL,
  name VARCHAR(100) NOT NULL,
  avatar_url VARCHAR(512) NULL,
  max_members INT NOT NULL DEFAULT 50,
  allow_member_invite TINYINT(1) NOT NULL DEFAULT 1,
  status VARCHAR(20) NOT NULL DEFAULT 'normal',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_groups_group_no (group_no),
  INDEX idx_groups_owner (owner_id),
  CONSTRAINT fk_groups_owner FOREIGN KEY (owner_id) REFERENCES users(user_id),
  CONSTRAINT fk_groups_conv FOREIGN KEY (conversation_id) REFERENCES conversations(conversation_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 4.12 group_members 群成员表

```sql
CREATE TABLE group_members (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  group_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  role VARCHAR(20) NOT NULL DEFAULT 'member',
  mute_until DATETIME NULL,
  status VARCHAR(20) NOT NULL DEFAULT 'active',
  joined_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  left_at DATETIME NULL,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_group_member (group_id, user_id),
  INDEX idx_group_members_user (user_id, status),
  INDEX idx_group_members_group_role (group_id, role, status),
  CONSTRAINT fk_group_member_group FOREIGN KEY (group_id) REFERENCES groups(group_id),
  CONSTRAINT fk_group_member_user FOREIGN KEY (user_id) REFERENCES users(user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

说明：

```txt
status = active 表示当前有效成员。
status = left 表示用户已退出群聊。
用户重新入群时必须更新 group_members.joined_at 为本次入群时间，left_at 清空。
群消息历史查询以当前 active membership 的 joined_at 作为可见起点。
```

### 4.13 group_join_requests 入群申请表

```sql
CREATE TABLE group_join_requests (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  request_id BIGINT NOT NULL,
  group_id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  message VARCHAR(255) NULL,
  status VARCHAR(20) NOT NULL DEFAULT 'pending',
  handled_by BIGINT NULL,
  pending_key VARCHAR(64)
    GENERATED ALWAYS AS (
      CASE
        WHEN status = 'pending' THEN CONCAT(group_id, ':', user_id)
        ELSE NULL
      END
    ) STORED,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_group_join_requests_request_id (request_id),
  UNIQUE KEY uk_group_join_requests_pending_key (pending_key),
  INDEX idx_group_join_group_status (group_id, status),
  INDEX idx_group_join_user_status (user_id, status),
  CONSTRAINT fk_group_join_group FOREIGN KEY (group_id) REFERENCES groups(group_id),
  CONSTRAINT fk_group_join_user FOREIGN KEY (user_id) REFERENCES users(user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

当前实现：对应 `backend/migrations/006_create_group_system.sql`。`group_id` 使用雪花 ID；`group_no` 为 8～10 位数字字符串，通过唯一索引保证唯一，冲突时由服务层重试。`group_join_requests.pending_key` 只限制同一用户对同一群同时存在一条 pending 申请，历史 accepted / rejected 申请允许保留多条。

### 4.14 files 文件表

```sql
CREATE TABLE files (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  file_id BIGINT NOT NULL UNIQUE,
  uploader_id BIGINT NOT NULL,
  original_name VARCHAR(255) NOT NULL,
  storage_path VARCHAR(1024) NOT NULL,
  file_size BIGINT NOT NULL,
  mime_type VARCHAR(128) NULL,
  sha256 VARCHAR(128) NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_files_uploader (uploader_id, created_at),
  INDEX idx_files_name (original_name),
  INDEX idx_files_sha256 (sha256),
  CONSTRAINT fk_files_uploader FOREIGN KEY (uploader_id) REFERENCES users(user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

当前实现：对应 `backend/migrations/005_create_files.sql`，MVP 仅保存本地存储元信息。

## 5. 查询规则

### 5.1 获取会话消息

必须过滤：

```txt
用户不是该会话成员的消息
private 会话已因删除好友被清理或当前用户不再是 active 成员的消息
群聊中当前用户不是 active 群成员的消息
群聊中早于当前用户本次 joined_at 的消息
当前用户没有 message_user_states 的消息
messages.recalled_at IS NOT NULL 的消息
messages.is_deleted_all = 1 的消息
message_user_states.is_deleted = 1 的消息
messages.created_at <= conversation_user_states.cleared_at 的消息
```

允许返回：

```txt
send_status = sent
send_status = failed_blocked，仅发送方可见
```

### 5.2 搜索消息

必须过滤：

```txt
不是当前用户所在会话的消息
删除好友后已清理的旧 private 会话消息
退出群聊后当前用户不可见的群消息
重新入群前产生的群消息
已撤回消息
当前用户已删除消息
当前用户清空时间之前的消息
物理或逻辑清理的消息
```

### 5.3 获取好友列表

`friendships` 中 user_id_1 或 user_id_2 等于当前用户，status = normal。

### 5.4 拉黑判断

发送单聊消息时查询：

```sql
SELECT 1 FROM block_relations WHERE blocker_id = 接收方 AND blocked_id = 发送方;
```

存在则返回 failed_blocked。

`failed_blocked` 持久化规则：

```txt
生成 message_id
插入 messages，send_status = failed_blocked
只为发送方插入 message_user_states
不为接收方插入 message_user_states
不推送给接收方
不更新 conversations.last_message_id
解除拉黑后不补发、不转正
```

### 5.5 消息幂等

发送消息时必须携带 `client_msg_id`。

```txt
seq 只用于 WebSocket 请求响应匹配，不入库
client_msg_id 只在 sender_id + conversation_id 范围内唯一
重复提交相同 client_msg_id 且内容一致、已有消息是 sent 时，返回已有消息 ack
重复提交相同 client_msg_id 且内容一致、已有消息是 failed_blocked 时，返回已有消息 failed
重复提交相同 client_msg_id 但内容不一致时，返回 duplicate_client_msg_id_conflict
不重复插入 messages，不重复插入 message_user_states，不重复更新 conversations.last_message_id
```

## 6. 后续扩展建议

```txt
messages 按时间或 conversation_id 分表
files 迁移对象存储后保留 storage_provider 字段
搜索迁移 Elasticsearch 后增加 message_search_index
多端登录后增加 user_devices 表
内容审核后增加 audit_logs 表
管理后台后增加 admin_users 表
```
