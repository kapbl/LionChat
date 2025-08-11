CREATE TABLE `moment` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '动态ID',
    `user_id` BIGINT NOT NULL COMMENT '发布者ID',
    `content` TEXT COMMENT '文本内容',
    `visibility` TINYINT DEFAULT 0 COMMENT '可见性：0-好友 1-公开 2-私密 3-部分可见',
    `visible_user_ids` JSON COMMENT '指定可见的用户ID列表',
    `expire_time` DATETIME COMMENT '三天可见过期时间',
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (`user_id`) REFERENCES `user`(`id`),
    INDEX `idx_user_created` (`user_id`, `created_at`)
) ENGINE=InnoDB;