CREATE TABLE IF NOT EXISTS messages (
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

CREATE TABLE IF NOT EXISTS message_user_states (
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
