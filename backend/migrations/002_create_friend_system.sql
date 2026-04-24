CREATE TABLE IF NOT EXISTS friend_requests (
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

CREATE TABLE IF NOT EXISTS friendships (
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

CREATE TABLE IF NOT EXISTS block_relations (
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
