CREATE TABLE `like` (
    `moment_id` BIGINT NOT NULL,
    `user_id` BIGINT NOT NULL,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`moment_id`, `user_id`), -- 联合主键防重复
    FOREIGN KEY (`moment_id`) REFERENCES `moment`(`id`)
) ENGINE=InnoDB;