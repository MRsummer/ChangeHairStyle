-- 创建发型生成记录表
CREATE TABLE IF NOT EXISTS hair_style_records (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL,
    image_url TEXT NOT NULL,
    prompt TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci; 


-- 用户信息表
CREATE TABLE IF NOT EXISTS user_info (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id VARCHAR(64) NOT NULL UNIQUE,
    nickname VARCHAR(64),
    avatar_url VARCHAR(255),
    coin INT DEFAULT 60,
    invite_code VARCHAR(6) UNIQUE,
    used_invite_code VARCHAR(6),
    last_sign_in_date DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_invite_code (invite_code),
    INDEX idx_used_invite_code (used_invite_code)
);

-- 广场内容表
CREATE TABLE IF NOT EXISTS square_content (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id VARCHAR(64) NOT NULL,
    record_id BIGINT NOT NULL,
    like_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_record_id (record_id),
    INDEX idx_created_at (created_at)
);

-- 点赞记录表
CREATE TABLE IF NOT EXISTS like_record (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id VARCHAR(64) NOT NULL,
    content_id BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_user_content (user_id, content_id),
    INDEX idx_content_id (content_id)
);

-- 修改用户信息表，添加coin和邀请码相关字段
ALTER TABLE user_info
    ADD COLUMN coin INT DEFAULT 60,
    ADD COLUMN invite_code VARCHAR(6) UNIQUE,
    ADD COLUMN used_invite_code VARCHAR(6),
    ADD COLUMN last_sign_in_date DATE,
    ADD INDEX idx_invite_code (invite_code),
    ADD INDEX idx_used_invite_code (used_invite_code); 