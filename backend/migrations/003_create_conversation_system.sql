CREATE TABLE IF NOT EXISTS conversations (
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

CREATE TABLE IF NOT EXISTS conversation_members (
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

CREATE TABLE IF NOT EXISTS conversation_user_states (
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
