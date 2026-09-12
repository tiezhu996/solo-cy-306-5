SET NAMES utf8mb4;
-- 活动报名通 (gbevent) 初始化脚本：首次启动容器时自动执行
CREATE DATABASE IF NOT EXISTS gbevent_db DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE gbevent_db;

CREATE TABLE IF NOT EXISTS users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  username VARCHAR(50) NOT NULL,
  password_hash VARCHAR(100) NOT NULL,
  nickname VARCHAR(50) NOT NULL DEFAULT '',
  avatar VARCHAR(255) NOT NULL DEFAULT '',
  role VARCHAR(20) NOT NULL DEFAULT 'user',
  email VARCHAR(100) NOT NULL DEFAULT '',
  phone VARCHAR(20) NOT NULL DEFAULT '',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  UNIQUE KEY uk_users_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS activities (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  title VARCHAR(200) NOT NULL,
  description TEXT,
  cover_image VARCHAR(255) NOT NULL DEFAULT '',
  activity_type VARCHAR(20) NOT NULL DEFAULT 'lecture',
  start_time DATETIME NOT NULL,
  end_time DATETIME NOT NULL,
  location VARCHAR(255) NOT NULL DEFAULT '',
  capacity INT NOT NULL DEFAULT 0,
  signup_deadline DATETIME NOT NULL,
  status VARCHAR(20) NOT NULL DEFAULT 'draft',
  organizer_id BIGINT UNSIGNED NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_activities_status (status),
  KEY idx_activities_type (activity_type),
  KEY idx_activities_organizer (organizer_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS registrations (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  activity_id BIGINT UNSIGNED NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  name VARCHAR(50) NOT NULL,
  phone VARCHAR(20) NOT NULL,
  remark VARCHAR(255) NOT NULL DEFAULT '',
  voucher_no VARCHAR(50) NOT NULL,
  status VARCHAR(20) NOT NULL DEFAULT 'registered',
  review_status VARCHAR(20) NOT NULL DEFAULT 'pending',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  UNIQUE KEY uk_registrations_voucher (voucher_no),
  UNIQUE KEY uk_registrations_activity_user (activity_id, user_id),
  KEY idx_registrations_activity (activity_id),
  KEY idx_registrations_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS check_in_records (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  registration_id BIGINT UNSIGNED NOT NULL,
  activity_id BIGINT UNSIGNED NOT NULL,
  check_in_method VARCHAR(20) NOT NULL DEFAULT 'voucher',
  check_in_time DATETIME NOT NULL,
  operator_id BIGINT UNSIGNED NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_checkin_activity (activity_id),
  KEY idx_checkin_registration (registration_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS comments (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  activity_id BIGINT UNSIGNED NOT NULL,
  user_id BIGINT UNSIGNED NOT NULL,
  rating INT NOT NULL DEFAULT 5,
  content TEXT,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_comments_activity (activity_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS favorites (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  activity_id BIGINT UNSIGNED NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  UNIQUE KEY uk_favorites_user_activity (user_id, activity_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS notifications (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id BIGINT UNSIGNED NOT NULL,
  notification_type VARCHAR(30) NOT NULL DEFAULT 'signup_success',
  title VARCHAR(200) NOT NULL,
  content TEXT,
  is_read TINYINT(1) NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id),
  KEY idx_notifications_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS audit_logs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  operator_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
  operator_name VARCHAR(50) NOT NULL DEFAULT '',
  action VARCHAR(50) NOT NULL,
  entity_type VARCHAR(50) NOT NULL,
  entity_id VARCHAR(50) NOT NULL DEFAULT '',
  detail TEXT,
  ip VARCHAR(50) NOT NULL DEFAULT '',
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 预置种子数据（密码：admin/Admin@123，organizer 与 user/User@123）
INSERT INTO users (id, username, password_hash, nickname, avatar, role, email, phone, created_at) VALUES
(1, 'admin', '$2a$10$bFfMuQAuKWflKxpuDYdFpeGJPVgD83q/.278LHYLL5S0DDmEfChX2', '系统管理员', '', 'admin', 'admin@gbevent.dev', '13800000001', NOW(3)),
(2, 'organizer', '$2a$10$TMTpnDbEwRbtcbF9VJxAxe5IswQjmo7pboKI9zVtU.BYnhzdJpX9a', '活动组织者', '', 'organizer', 'org@gbevent.dev', '13800000002', NOW(3)),
(3, 'user', '$2a$10$TMTpnDbEwRbtcbF9VJxAxe5IswQjmo7pboKI9zVtU.BYnhzdJpX9a', '普通用户', '', 'user', 'user@gbevent.dev', '13800000003', NOW(3));

INSERT INTO activities (id, title, description, cover_image, activity_type, start_time, end_time, location, capacity, signup_deadline, status, organizer_id, created_at) VALUES
(1, 'Go 语言企业级开发实战讲座', '深入讲解 Go 1.22 + Gin + GORM 的企业级工程实践。', '', 'lecture', DATE_ADD(NOW(), INTERVAL 7 DAY), DATE_ADD(NOW(), INTERVAL 7 DAY), '线上直播', 200, DATE_ADD(NOW(), INTERVAL 6 DAY), 'published', 2, NOW(3)),
(2, '新员工安全培训', '面向新入职员工的安全意识与应急处理培训。', '', 'training', DATE_ADD(NOW(), INTERVAL 14 DAY), DATE_ADD(NOW(), INTERVAL 14 DAY), 'A 座 3 楼培训室', 50, DATE_ADD(NOW(), INTERVAL 13 DAY), 'published', 2, NOW(3)),
(3, '秋季团队趣味运动会', '团队协作趣味运动会，包含拔河、接力、跳绳等项目。', '', 'party', DATE_ADD(NOW(), INTERVAL 30 DAY), DATE_ADD(NOW(), INTERVAL 30 DAY), '城市体育公园', 120, DATE_ADD(NOW(), INTERVAL 28 DAY), 'published', 2, NOW(3)),
(4, '黑客松编程竞赛（草稿）', '24 小时黑客松编程竞赛，暂未发布。', '', 'competition', DATE_ADD(NOW(), INTERVAL 60 DAY), DATE_ADD(NOW(), INTERVAL 62 DAY), '创新中心', 80, DATE_ADD(NOW(), INTERVAL 55 DAY), 'draft', 2, NOW(3)),
(5, '上季度读书分享会（已结束）', '已结束的读书分享会。', '', 'lecture', DATE_SUB(NOW(), INTERVAL 10 DAY), DATE_SUB(NOW(), INTERVAL 10 DAY), '咖啡厅', 30, DATE_SUB(NOW(), INTERVAL 11 DAY), 'ended', 2, NOW(3));

INSERT INTO registrations (id, activity_id, user_id, name, phone, remark, voucher_no, status, review_status, created_at) VALUES
(1, 1, 3, '张三', '13900000001', '希望了解工程实践', 'GB20260816000001', 'registered', 'approved', NOW(3)),
(2, 2, 3, '张三', '13900000001', '', 'GB20260816000002', 'checked_in', 'approved', NOW(3)),
(3, 3, 3, '张三', '13900000001', '', 'GB20260816000003', 'registered', 'pending', NOW(3));

INSERT INTO check_in_records (id, registration_id, activity_id, check_in_method, check_in_time, operator_id, created_at) VALUES
(1, 2, 2, 'voucher', NOW(), 2, NOW(3));

INSERT INTO comments (id, activity_id, user_id, rating, content, created_at) VALUES
(1, 1, 3, 5, '内容充实，收获很大！', NOW(3));

INSERT INTO favorites (id, user_id, activity_id, created_at) VALUES
(1, 3, 1, NOW(3));

INSERT INTO notifications (id, user_id, notification_type, title, content, is_read, created_at) VALUES
(1, 3, 'signup_success', '报名成功', '您已成功报名《Go 语言企业级开发实战讲座》', 0, NOW(3)),
(2, 3, 'checkin_success', '签到成功', '您已在新员工安全培训中完成签到', 0, NOW(3));
