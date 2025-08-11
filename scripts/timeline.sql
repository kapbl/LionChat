CREATE TABLE `timeline` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
    `user_id` BIGINT NOT NULL COMMENT '接收者ID',
    `moment_id` BIGINT NOT NULL COMMENT '动态ID',
    `is_own` BOOLEAN DEFAULT 0 COMMENT '是否自己发布',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY `uk_user_moment` (`user_id`, `moment_id`),
    INDEX `idx_user_created` (`user_id`, `created_at`)
) ENGINE=InnoDB;