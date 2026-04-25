CREATE TABLE IF NOT EXISTS `groups` (
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

CREATE TABLE IF NOT EXISTS group_members (
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
  CONSTRAINT fk_group_member_group FOREIGN KEY (group_id) REFERENCES `groups`(group_id),
  CONSTRAINT fk_group_member_user FOREIGN KEY (user_id) REFERENCES users(user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS group_join_requests (
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
  CONSTRAINT fk_group_join_group FOREIGN KEY (group_id) REFERENCES `groups`(group_id),
  CONSTRAINT fk_group_join_user FOREIGN KEY (user_id) REFERENCES users(user_id),
  CONSTRAINT fk_group_join_handler FOREIGN KEY (handled_by) REFERENCES users(user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
