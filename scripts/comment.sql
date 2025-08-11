CREATE TABLE `comment` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
    `moment_id` BIGINT NOT NULL,
    `user_id` BIGINT NOT NULL,
    `content` VARCHAR(500) NOT NULL,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (`moment_id`) REFERENCES `moment`(`id`),
    INDEX `idx_moment` (`moment_id`)
) ENGINE=InnoDB;