/*
 Navicat Premium Data Transfer

 Source Server         : chatlion
 Source Server Type    : MySQL
 Source Server Version : 80042 (8.0.42)
 Source Host           : localhost:3306
 Source Schema         : chat

 Target Server Type    : MySQL
 Target Server Version : 80042 (8.0.42)
 File Encoding         : 65001

 Date: 12/08/2025 20:26:09
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for comment
-- ----------------------------
DROP TABLE IF EXISTS `comment`;
CREATE TABLE `comment`  (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `moment_id` bigint NULL DEFAULT NULL,
  `user_id` bigint NULL DEFAULT NULL,
  `content` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL,
  `created_at` datetime NULL DEFAULT CURRENT_TIMESTAMP,
  `create_at` datetime(3) NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_moment`(`moment_id` ASC) USING BTREE,
  CONSTRAINT `comment_ibfk_1` FOREIGN KEY (`moment_id`) REFERENCES `moment` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for group
-- ----------------------------
DROP TABLE IF EXISTS `group`;
CREATE TABLE `group`  (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '\'id\'',
  `name` varchar(150) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '\'组名\'',
  `desc` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '\'描述\'',
  `member_count` bigint NULL DEFAULT 0 COMMENT '\'成员数量\'',
  `owner_id` bigint NULL DEFAULT NULL COMMENT '\'群主id\'',
  `create_at` datetime(3) NULL DEFAULT NULL COMMENT '\'创建时间\'',
  `update_at` datetime(3) NULL DEFAULT NULL COMMENT '\'更新时间\'',
  `delete_at` datetime(3) NULL DEFAULT NULL COMMENT '\'删除时间\'',
  `uuid` varchar(150) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '\'UUID\'',
  `type` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '\'类型\'',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 2 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for group_member
-- ----------------------------
DROP TABLE IF EXISTS `group_member`;
CREATE TABLE `group_member`  (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '\'id\'',
  `group_id` bigint NULL DEFAULT NULL COMMENT '\'组id\'',
  `user_id` bigint NULL DEFAULT NULL COMMENT '\'用户id\'',
  `create_at` datetime(3) NULL DEFAULT NULL COMMENT '\'创建时间\'',
  `update_at` datetime(3) NULL DEFAULT NULL COMMENT '\'更新时间\'',
  `delete_at` datetime(3) NULL DEFAULT NULL COMMENT '\'删除时间\'',
  `group_uuid` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '\'组UUID\'',
  `user_uuid` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '\'用户UUID\'',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 19 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for like
-- ----------------------------
DROP TABLE IF EXISTS `like`;
CREATE TABLE `like`  (
  `moment_id` bigint NOT NULL,
  `user_id` bigint NOT NULL,
  `created_at` datetime(3) NULL DEFAULT NULL,
  PRIMARY KEY (`moment_id`, `user_id`) USING BTREE,
  CONSTRAINT `like_ibfk_1` FOREIGN KEY (`moment_id`) REFERENCES `moment` (`id`) ON DELETE RESTRICT ON UPDATE RESTRICT
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for message
-- ----------------------------
DROP TABLE IF EXISTS `message`;
CREATE TABLE `message`  (
  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) NULL DEFAULT NULL,
  `updated_at` datetime(3) NULL DEFAULT NULL,
  `deleted_at` datetime(3) NULL DEFAULT NULL,
  `sender_id` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '\'发送者ID\'',
  `receive_id` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '\'接收者ID\'',
  `content` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '\'消息内容\'',
  `status` tinyint(1) NULL DEFAULT 0 COMMENT '\'0未读 1已读 2撤回\'',
  `message_id` bigint UNSIGNED NOT NULL,
  PRIMARY KEY (`id`, `message_id`) USING BTREE,
  INDEX `idx_message_deleted_at`(`deleted_at` ASC) USING BTREE,
  INDEX `idx_sender`(`sender_id` ASC) USING BTREE,
  INDEX `idx_receiver`(`receive_id` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 179 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for moment
-- ----------------------------
DROP TABLE IF EXISTS `moment`;
CREATE TABLE `moment`  (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` bigint NULL DEFAULT NULL,
  `content` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL,
  `create_time` datetime(3) NULL DEFAULT NULL,
  `delete_time` datetime(3) NULL DEFAULT NULL,
  `created_at` datetime(3) NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 9 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for timeline
-- ----------------------------
DROP TABLE IF EXISTS `timeline`;
CREATE TABLE `timeline`  (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `user_id` bigint NULL DEFAULT NULL,
  `moment_id` bigint NULL DEFAULT NULL,
  `is_own` tinyint(1) NULL DEFAULT NULL,
  `created_at` datetime(3) NULL DEFAULT NULL,
  `create_time` datetime(3) NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_user_moment`(`user_id` ASC, `moment_id` ASC) USING BTREE,
  INDEX `idx_user_created`(`user_id` ASC, `created_at` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 11 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for user_friends
-- ----------------------------
DROP TABLE IF EXISTS `user_friends`;
CREATE TABLE `user_friends`  (
  `create_at` datetime(3) NULL DEFAULT NULL,
  `update_at` datetime(3) NULL DEFAULT NULL,
  `delete_at` datetime(3) NULL DEFAULT NULL,
  `user_id` bigint NULL DEFAULT NULL COMMENT '\'用户id\'',
  `friend_id` bigint NULL DEFAULT NULL COMMENT '\'好友id\'',
  `status` bigint NULL DEFAULT NULL COMMENT '\'状态：1或0\'',
  `version` bigint NULL DEFAULT NULL COMMENT '\'版本号\''
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for users
-- ----------------------------
DROP TABLE IF EXISTS `users`;
CREATE TABLE `users`  (
  `id` int NOT NULL AUTO_INCREMENT COMMENT 'id',
  `uuid` varchar(150) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '\'uuid\'',
  `username` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '\'用户名\'',
  `nickname` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL COMMENT '\'昵称\'',
  `email` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '\'邮箱\'',
  `password` varchar(150) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '\'密码\'',
  `avatar` varchar(150) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '\'头像\'',
  `create_at` datetime(3) NULL DEFAULT NULL,
  `update_at` datetime(3) NULL DEFAULT NULL,
  `delete_at` bigint NULL DEFAULT NULL,
  `group_version` bigint NULL DEFAULT NULL COMMENT '\'控制用户的群组版本号，用于防止Redis读取到过期的数据\'',
  `friend_version` bigint NULL DEFAULT NULL COMMENT '\'控制用户的好友版本号，用于防止Redis读取到过期的数据\'',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uni_users_username`(`username` ASC) USING BTREE,
  UNIQUE INDEX `idx_uuid`(`uuid` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 35597 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '用户表' ROW_FORMAT = Dynamic;

SET FOREIGN_KEY_CHECKS = 1;
