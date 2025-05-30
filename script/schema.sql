-- 创建数据库（如果不存在）
CREATE DATABASE IF NOT EXISTS `uportal` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 使用数据库
USE `uportal`;

-- 1. 用户表，存储基础用户信息
CREATE TABLE IF NOT EXISTS `users` (
                         `user_id` BIGINT       NOT NULL AUTO_INCREMENT COMMENT '用户ID，主键，自增',
                         `phone`   VARCHAR(20)  DEFAULT NULL             COMMENT '手机号，用户使用手机号注册/登录时的号码，唯一',
                         `email`   VARCHAR(100) DEFAULT NULL             COMMENT '邮箱，用户邮箱地址，唯一',
                         `password_hash` VARCHAR(255) DEFAULT NULL       COMMENT '密码哈希，用于手机号/邮箱注册的情况，第三方登录用户此字段为空',
                         `nickname` VARCHAR(50)  DEFAULT NULL            COMMENT '用户昵称，显示名称',
                         `avatar_url` VARCHAR(255) DEFAULT NULL          COMMENT '头像URL，用户头像图片链接',
                         `language` VARCHAR(10)  NOT NULL DEFAULT 'zh-CN' COMMENT '界面语言偏好，如 zh-CN、en-US 等',
                         `status`  TINYINT       NOT NULL DEFAULT 1      COMMENT '账号状态：1=正常，0=禁用',
                         `token_balance` INT     NOT NULL DEFAULT 0      COMMENT '代币余额',
                         `inviter_id` BIGINT    DEFAULT NULL            COMMENT '邀请人ID',
                         `created_at` DATETIME   NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '注册时间',
                         `updated_at` DATETIME   NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录更新时间',
                         `last_login_at` DATETIME DEFAULT NULL           COMMENT '最后登录时间',
                         PRIMARY KEY (`user_id`),
                         UNIQUE KEY `uk_users_phone` (`phone`),
                         UNIQUE KEY `uk_users_email` (`email`),
                         KEY `idx_users_status` (`status`),
                         KEY `idx_users_inviter` (`inviter_id`),
                         CONSTRAINT `fk_users_inviter` FOREIGN KEY (`inviter_id`) REFERENCES `users` (`user_id`) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
  COMMENT='用户表，存储基础用户信息';

-- 2. 管理员用户表，存储后台管理员账号
CREATE TABLE IF NOT EXISTS `admin_users` (
                               `admin_id`     INT          NOT NULL AUTO_INCREMENT COMMENT '管理员ID，主键，自增',
                               `username`     VARCHAR(50)  NOT NULL               COMMENT '登录用户名，唯一',
                               `password_hash` VARCHAR(255) NOT NULL              COMMENT '密码哈希',
                               `role`         VARCHAR(20)  NOT NULL DEFAULT 'admin' COMMENT '角色，如 superadmin、admin',
                               `status`       TINYINT      NOT NULL DEFAULT 1     COMMENT '账号状态：1=正常，0=停用',
                               `created_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                               `last_login_at` DATETIME    DEFAULT NULL           COMMENT '最后登录时间',
                               PRIMARY KEY (`admin_id`),
                               UNIQUE KEY `uk_admin_username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
  COMMENT='管理员用户表，存储后台管理员账号信息';

-- 3. 系统配置表，存储全局系统参数
CREATE TABLE IF NOT EXISTS `system_config` (
                                 `config_key`   VARCHAR(50)  NOT NULL                COMMENT '配置键，主键，如 TOKEN_EXCHANGE_RATE',
                                 `config_value` VARCHAR(100) NOT NULL                COMMENT '配置值，以文本形式存储',
                                 `description`  VARCHAR(100) DEFAULT NULL            COMMENT '配置描述',
                                 PRIMARY KEY (`config_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
  COMMENT='系统配置表，存储全局配置项';

-- 4. 代币任务配置表，配置可奖励任务
CREATE TABLE IF NOT EXISTS `reward_tasks` (
                                `task_id`       INT          NOT NULL AUTO_INCREMENT COMMENT '任务ID，主键，自增',
                                `task_name`     VARCHAR(100) NOT NULL               COMMENT '任务名称，如 注册奖励、邀请好友、观看广告 等',
                                `task_desc`     VARCHAR(255) DEFAULT NULL           COMMENT '任务描述，详细说明',
                                `token_reward`  INT          NOT NULL               COMMENT '完成一次任务获得的代币数',
                                `daily_limit`   INT          NOT NULL DEFAULT 0     COMMENT '每日奖励上限，0表示不限制',
                                `interval_seconds` INT       NOT NULL DEFAULT 0     COMMENT '两次完成任务的最小间隔秒数，0表示不限制',
                                `valid_from`    DATETIME     DEFAULT NULL           COMMENT '任务生效时间，NULL表示即时生效',
                                `valid_to`      DATETIME     DEFAULT NULL           COMMENT '任务截止时间，NULL表示永久有效',
                                `repeatable`    TINYINT      NOT NULL DEFAULT 1     COMMENT '是否可重复完成：1=是，0=否',
                                `status`        TINYINT      NOT NULL DEFAULT 1     COMMENT '任务状态：1=启用，0=停用',
                                PRIMARY KEY (`task_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
  COMMENT='代币任务配置表';

-- 5. 代币消耗功能表，配置功能消耗规则
CREATE TABLE IF NOT EXISTS `token_consume_rules` (
                                       `feature_id`   INT          NOT NULL AUTO_INCREMENT COMMENT '功能ID，主键，自增',
                                       `feature_name` VARCHAR(100) NOT NULL               COMMENT '功能名称，如 高级过滤器解锁 等',
                                       `feature_desc` VARCHAR(255) DEFAULT NULL           COMMENT '功能描述',
                                       `token_cost`   INT          NOT NULL               COMMENT '使用一次该功能消耗的代币数',
                                       `feature_code` VARCHAR(50)  DEFAULT NULL           COMMENT '功能代码，用于程序内部识别',
                                       `status`       TINYINT      NOT NULL DEFAULT 1     COMMENT '功能状态：1=启用，0=停用',
                                       PRIMARY KEY (`feature_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
  COMMENT='代币消耗功能配置表';

-- 6. 充值方案表，预设代币充值套餐
CREATE TABLE IF NOT EXISTS `recharge_plans` (
                                  `plan_id`     INT           NOT NULL AUTO_INCREMENT COMMENT '方案ID，主键，自增',
                                  `token_amount` INT          NOT NULL               COMMENT '方案提供的代币数量',
                                  `price`       DECIMAL(10,2) NOT NULL               COMMENT '售价(元)',
                                  `currency`    CHAR(3)       NOT NULL DEFAULT 'CNY' COMMENT '货币类型代码',
                                  `description` VARCHAR(100)  DEFAULT NULL           COMMENT '方案描述，如 赠送20%代币 等',
                                  `status`      TINYINT       NOT NULL DEFAULT 1     COMMENT '方案状态：1=可用，0=下架',
                                  `created_at`  DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                  `updated_at`  DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                                  PRIMARY KEY (`plan_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
  COMMENT='充值方案表';

-- 7. 用户第三方认证表，关联用户与第三方平台
CREATE TABLE IF NOT EXISTS `user_auth` (
                             `auth_id`          BIGINT     NOT NULL AUTO_INCREMENT COMMENT '认证记录ID，主键，自增',
                             `user_id`          BIGINT     NOT NULL               COMMENT '用户ID，外键关联 users.user_id',
                             `provider`         VARCHAR(20) NOT NULL              COMMENT '登录平台类型，如 wechat、apple、google、twitter',
                             `provider_user_id` VARCHAR(100) NOT NULL             COMMENT '第三方平台内用户唯一ID，如 openid、OAuth ID 等',
                             `created_at`       DATETIME   NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '绑定时间',
                             PRIMARY KEY (`auth_id`),
                             KEY `idx_user_auth_user` (`user_id`),
                             CONSTRAINT `fk_user_auth_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`user_id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
  COMMENT='用户第三方认证关联表';

-- 8. 用户登录日志表，记录登录行为
CREATE TABLE IF NOT EXISTS `user_login_log` (
                                  `log_id`       BIGINT     NOT NULL AUTO_INCREMENT COMMENT '日志ID，主键，自增',
                                  `user_id`      BIGINT     NOT NULL               COMMENT '用户ID，外键关联 users.user_id',
                                  `login_time`   DATETIME   NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '登录时间',
                                  `login_method` VARCHAR(20) NOT NULL              COMMENT '登录方式，如 password、wechat、phone',
                                  `login_platform` VARCHAR(20) DEFAULT NULL         COMMENT '登录平台，如 iOSApp、Web、WeChatMiniProg',
                                  `ip_address`   VARCHAR(45) DEFAULT NULL          COMMENT '登录IP地址',
                                  `device_info`  VARCHAR(100) DEFAULT NULL         COMMENT '设备信息或User-Agent简述',
                                  PRIMARY KEY (`log_id`),
                                  KEY `idx_login_log_user` (`user_id`),
                                  CONSTRAINT `fk_login_log_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`user_id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
  COMMENT='用户登录日志表';

-- 9. 充值订单表，记录每笔充值交易
CREATE TABLE IF NOT EXISTS `recharge_orders` (
                                   `order_id`      BIGINT        NOT NULL AUTO_INCREMENT COMMENT '订单ID，主键，自增',
                                   `user_id`       BIGINT        NOT NULL               COMMENT '用户ID，外键关联 users.user_id',
                                   `plan_id`       INT           DEFAULT NULL           COMMENT '方案ID，外键关联 recharge_plans.plan_id',
                                   `token_amount`  INT           NOT NULL               COMMENT '本次订单获得的代币数量',
                                   `amount_paid`   DECIMAL(10,2) NOT NULL               COMMENT '支付金额(元)',
                                   `payment_method` VARCHAR(20) NOT NULL               COMMENT '支付方式，如 Alipay、WeChat',
                                   `status`        TINYINT       NOT NULL DEFAULT 0     COMMENT '订单状态：0=待支付，1=支付成功，2=支付失败，3=已退款',
                                   `transaction_id` VARCHAR(100) DEFAULT NULL          COMMENT '第三方交易号，如支付宝交易号、微信订单号',
                                   `created_at`    DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '订单创建时间',
                                   `paid_at`       DATETIME      DEFAULT NULL           COMMENT '支付完成时间',
                                   `updated_at`    DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
                                   PRIMARY KEY (`order_id`),
                                   KEY `idx_recharge_orders_user` (`user_id`),
                                   KEY `idx_recharge_orders_plan` (`plan_id`),
                                   CONSTRAINT `fk_recharge_orders_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`user_id`) ON DELETE CASCADE ON UPDATE CASCADE,
                                   CONSTRAINT `fk_recharge_orders_plan` FOREIGN KEY (`plan_id`) REFERENCES `recharge_plans`(`plan_id`) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
  COMMENT='充值订单表';

-- 10. 退款记录表，记录充值退款详情
CREATE TABLE IF NOT EXISTS `refunds` (
                           `refund_id`    BIGINT        NOT NULL AUTO_INCREMENT COMMENT '退款ID，主键，自增',
                           `order_id`     BIGINT        NOT NULL               COMMENT '原订单ID，外键关联 recharge_orders.order_id',
                           `user_id`      BIGINT        NOT NULL               COMMENT '用户ID，外键关联 users.user_id',
                           `refund_amount` DECIMAL(10,2) NOT NULL               COMMENT '退款金额(元)',
                           `refund_tokens` INT           NOT NULL               COMMENT '收回代币数',
                           `refund_method` VARCHAR(20)   NOT NULL               COMMENT '退款方式，如 Alipay、WeChat',
                           `status`       TINYINT       NOT NULL DEFAULT 0     COMMENT '退款状态：0=处理中，1=成功，2=失败',
                           `admin_id`     INT           DEFAULT NULL           COMMENT '操作管理员ID，外键关联 admin_users.admin_id',
                           `reason`       VARCHAR(255)  DEFAULT NULL           COMMENT '退款原因说明',
                           `refund_time`  DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '退款完成时间',
                           PRIMARY KEY (`refund_id`),
                           KEY `idx_refunds_order` (`order_id`),
                           KEY `idx_refunds_user` (`user_id`),
                           KEY `idx_refunds_admin` (`admin_id`),
                           CONSTRAINT `fk_refunds_order` FOREIGN KEY (`order_id`) REFERENCES `recharge_orders`(`order_id`) ON DELETE CASCADE ON UPDATE CASCADE,
                           CONSTRAINT `fk_refunds_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`user_id`) ON DELETE CASCADE ON UPDATE CASCADE,
                           CONSTRAINT `fk_refunds_admin` FOREIGN KEY (`admin_id`) REFERENCES `admin_users`(`admin_id`) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
  COMMENT='退款记录表';

-- 11. 用户代币记录表，记录所有代币变动流水
CREATE TABLE IF NOT EXISTS `token_records` (
                                 `record_id`    BIGINT     NOT NULL AUTO_INCREMENT COMMENT '记录ID，主键，自增',
                                 `user_id`      BIGINT     NOT NULL               COMMENT '用户ID，外键关联 users.user_id',
                                 `change_amount` INT       NOT NULL               COMMENT '代币变动数，正为增加，负为扣除',
                                 `balance_after` INT       NOT NULL               COMMENT '变动后余额',
                                 `change_type`  VARCHAR(20) NOT NULL              COMMENT '变动类型，如 TASK_REWARD、FEATURE_COST、PURCHASE、REFUND、ADMIN_ADJUST',
                                 `task_id`      INT        DEFAULT NULL           COMMENT '任务ID来源，外键关联 reward_tasks.task_id',
                                 `feature_id`   INT        DEFAULT NULL           COMMENT '功能ID来源，外键关联 token_consume_rules.feature_id',
                                 `order_id`     BIGINT     DEFAULT NULL           COMMENT '订单ID来源，外键关联 recharge_orders.order_id',
                                 `admin_id`     INT        DEFAULT NULL           COMMENT '管理员ID来源，外键关联 admin_users.admin_id',
                                 `remark`       VARCHAR(255) DEFAULT NULL         COMMENT '备注说明，如 新用户注册奖励、功能消费等',
                                 `change_time`  DATETIME   NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '变动时间',
                                 PRIMARY KEY (`record_id`),
                                 KEY `idx_token_records_user` (`user_id`),
                                 KEY `idx_token_records_task` (`task_id`),
                                 KEY `idx_token_records_feature` (`feature_id`),
                                 KEY `idx_token_records_order` (`order_id`),
                                 KEY `idx_token_records_admin` (`admin_id`),
                                 CONSTRAINT `fk_token_records_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`user_id`) ON DELETE CASCADE ON UPDATE CASCADE,
                                 CONSTRAINT `fk_token_records_task` FOREIGN KEY (`task_id`) REFERENCES `reward_tasks`(`task_id`) ON DELETE SET NULL ON UPDATE CASCADE,
                                 CONSTRAINT `fk_token_records_feature` FOREIGN KEY (`feature_id`) REFERENCES `token_consume_rules`(`feature_id`) ON DELETE SET NULL ON UPDATE CASCADE,
                                 CONSTRAINT `fk_token_records_order` FOREIGN KEY (`order_id`) REFERENCES `recharge_orders`(`order_id`) ON DELETE SET NULL ON UPDATE CASCADE,
                                 CONSTRAINT `fk_token_records_admin` FOREIGN KEY (`admin_id`) REFERENCES `admin_users`(`admin_id`) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
  COMMENT='用户代币记录表，记录每笔代币增减流水';

-- 12. 支付回调通知记录表
CREATE TABLE IF NOT EXISTS `payment_notify_records` (
    `record_id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '记录ID',
    `order_id` BIGINT NOT NULL COMMENT '订单ID',
    `transaction_id` VARCHAR(64) NOT NULL COMMENT '微信支付交易号',
    `notify_type` VARCHAR(32) NOT NULL COMMENT '通知类型',
    `notify_time` DATETIME NOT NULL COMMENT '通知时间',
    `process_status` TINYINT NOT NULL DEFAULT 0 COMMENT '处理状态：0=待处理，1=处理成功，2=处理失败',
    `retry_count` INT NOT NULL DEFAULT 0 COMMENT '重试次数',
    `error_message` VARCHAR(255) DEFAULT NULL COMMENT '错误信息',
    `process_time` DATETIME DEFAULT NULL COMMENT '处理时间',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`record_id`),
    UNIQUE KEY `uk_order_transaction` (`order_id`, `transaction_id`),
    KEY `idx_notify_time` (`notify_time`),
    KEY `idx_process_status` (`process_status`),
    CONSTRAINT `fk_payment_notify_order` FOREIGN KEY (`order_id`) REFERENCES `recharge_orders` (`order_id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付回调通知记录表';

-- 任务完成记录表
CREATE TABLE IF NOT EXISTS task_completion_records (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL COMMENT '用户ID',
    task_id INT NOT NULL COMMENT '任务ID',
    token_reward INT NOT NULL COMMENT '获得的代币奖励',
    completed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '完成时间',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_task (user_id, task_id),
    INDEX idx_completed_at (completed_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='任务完成记录表';

-- 通知表
CREATE TABLE IF NOT EXISTS notifications (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL COMMENT '用户ID',
    type VARCHAR(32) NOT NULL COMMENT '通知类型',
    title VARCHAR(128) NOT NULL COMMENT '通知标题',
    content TEXT NOT NULL COMMENT '通知内容',
    status TINYINT NOT NULL DEFAULT 0 COMMENT '状态：0-未读，1-已读',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_status (user_id, status),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='通知表';

-- 邀请记录表
CREATE TABLE IF NOT EXISTS `invite_records` (
    `record_id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '记录ID，主键，自增',
    `inviter_id` BIGINT NOT NULL COMMENT '邀请人ID',
    `invitee_id` BIGINT NOT NULL COMMENT '被邀请人ID',
    `token_reward` INT NOT NULL COMMENT '邀请奖励代币数',
    `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态：0=待发放，1=已发放，2=发放失败',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`record_id`),
    UNIQUE KEY `uk_invite_invitee` (`invitee_id`),
    KEY `idx_invite_inviter` (`inviter_id`),
    CONSTRAINT `fk_invite_inviter` FOREIGN KEY (`inviter_id`) REFERENCES `users` (`user_id`) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT `fk_invite_invitee` FOREIGN KEY (`invitee_id`) REFERENCES `users` (`user_id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
  COMMENT='邀请记录表，记录用户邀请关系和奖励发放状态';
