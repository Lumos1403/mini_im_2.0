CREATE TABLE IF NOT EXISTS files (
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
