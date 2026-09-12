# EventGo（活动报名通）

活动报名通是一个支持活动发布、在线报名、名额管理、签到统计、活动日历视图以及评论收藏与消息通知的活动管理系统，适用于讲座、培训、聚会、比赛等各类线下或线上活动场景。

## 快速启动（Docker Compose 一键部署）

```bash
cp .env.example .env
docker compose up -d --build
```

启动完成后访问：

- 前端：http://localhost:18506
- 后端 API：http://localhost:19506
- 后端健康检查：http://localhost:19506/healthz
- MySQL：localhost:57506

预置账号（密码见 database/init.sql 与 backend/internal/service/seed.go）：

| 用户名 | 密码 | 角色 |
| --- | --- | --- |
| admin | Admin@123 | 管理员 |
| organizer | User@123 | 组织者 |
| user | User@123 | 普通用户 |

## 本地开发

后端：

```bash
cd backend && go mod tidy && go run ./cmd/server
```

构建：`cd backend && go build ./...`

前端：

```bash
cd frontend && npm install && npm run dev
```

前端开发服务器通过 Vite 代理将 `/api` 转发到 `http://localhost:19506`。

## 技术栈

| 层 | 技术 |
| --- | --- |
| 前端 | Vue 3 + TypeScript + Element Plus + Vite + Pinia |
| 后端 | Go 1.22 + Gin + GORM |
| 数据库 | MySQL 8.0 |
| 认证 | JWT（github.com/golang-jwt/jwt/v5）+ RBAC |
| 其他依赖 | gin-contrib/cors、golang.org/x/crypto/bcrypt、go-playground/validator/v10 |

## 项目目录结构

```
cy-306/
├── docker-compose.yml
├── .env.example
├── database/init.sql              # MySQL 初始化脚本（建表 + 种子数据）
├── backend/
│   ├── cmd/server/main.go
│   └── internal/
│       ├── config/                # 环境变量配置
│       ├── model/                 # user/activity/registration/check_in_record/comment/favorite/notification/audit_log
│       ├── repository/            # 按实体分文件
│       ├── service/               # 业务逻辑 + 种子数据 + 单元测试
│       ├── handler/               # HTTP 处理器（含 upload_handler）
│       ├── router/                # router.go + 按实体分文件
│       ├── middleware/            # auth/rbac/rate_limiter/error_handler/audit_log/cors/request_logger
│       ├── dto/                   # 请求/响应结构体
│       ├── constants/             # 枚举、错误码、日志模板、文案
│       └── util/                  # jwt/logger/formatters/app_error/voucher_generator/file_upload
└── frontend/
    └── src/
        ├── api/                   # user/activity/registration/checkIn/comment/favorite/notification
        ├── stores/                # authStore/userStore/activityStore/registrationStore/commentStore/favoriteStore/notificationStore
        ├── types/
        ├── components/common/     # ActivityCard/ActivityCalendar/ActivityFilter/SignupForm/CommentList/RatingStars/ActivityForm/RegistrationTable/CheckInQrCode/CheckInPanel/MyRegistrations/FavoriteList/NotificationList/ImageUploader/EmptyState/RoleGuard
        ├── hooks/                 # useAuth/useActivityStats/useCheckIn
        ├── pages/                 # Calendar/Activities/ActivityDetail/OrganizerActivities/OrganizerRegistrations/Profile/Login/Register
        ├── router/                # index.ts + guards.ts
        ├── utils/                 # dateFormat/voucherGenerator/request
        └── constants/             # activity/registration/notification/errorCodes
```

## 环境变量

| 变量 | 说明 | 默认值 |
| --- | --- | --- |
| COMPOSE_PROJECT_NAME | Docker Compose 项目名/容器前缀 | gbevent |
| DB_NAME | 数据库名 | gbevent_db |
| DB_USER | 数据库用户 | gbevent_user |
| DB_PASSWORD | 数据库密码 | gbevent_pwd |
| DB_ROOT_PASSWORD | 数据库 root 密码 | gbevent_root |
| JWT_SECRET | JWT 签名密钥 | change_me_to_a_long_random_string |
| JWT_EXPIRE_HOURS | JWT 过期小时数 | 72 |
| APP_CORS_ORIGINS | 允许的跨域来源（逗号分隔） | http://localhost:18506 |
| FRONTEND_PORT | 前端端口 | 18506 |
| BACKEND_PORT | 后端端口 | 19506 |
| DB_PORT | 数据库端口 | 57506 |

## Docker 部署说明

- 端口映射：前端 18506:80、后端 19506:8080、MySQL 57506:3306。
- 数据卷：`db-data` 持久化 MySQL 数据；`upload-data` 持久化上传图片。
- 服务依赖：backend `depends_on` db（service_healthy），frontend `depends_on` backend（service_healthy）。
- 前端 Nginx 将 `/api/` 反代到 `http://backend:8080/`，支持 SPA 路由 `try_files`。
- 常见问题：
  - 端口冲突：修改 `.env` 中 `FRONTEND_PORT/BACKEND_PORT/DB_PORT` 后重新 `docker compose up -d`。
  - 数据库重置：`docker compose down -v` 后重新启动。

## 枚举出现位置清单

### ActivityStatus（draft/published/ended）
- 后端：`backend/internal/constants/activity.go`、`backend/internal/model/activity.go`、`backend/internal/service/activity_service.go`、`backend/internal/util/formatters.go`、`backend/internal/constants/log_templates.go`、`backend/internal/constants/error_codes.go`、`backend/internal/dto/dto_activity.go`、`database/init.sql`
- 前端：`frontend/src/constants/activity.ts`、`frontend/src/components/common/ActivityCard.vue`、`frontend/src/components/common/ActivityFilter.vue`、`frontend/src/pages/ActivityDetail.vue`、`frontend/src/pages/OrganizerActivities.vue`、`frontend/src/pages/Activities.vue`

### RegistrationStatus（registered/cancelled/checked_in）
- 后端：`backend/internal/constants/registration.go`、`backend/internal/model/registration.go`、`backend/internal/service/registration_service.go`、`backend/internal/service/check_in_record_service.go`、`backend/internal/util/formatters.go`、`backend/internal/constants/log_templates.go`、`backend/internal/constants/error_codes.go`、`database/init.sql`
- 前端：`frontend/src/constants/registration.ts`、`frontend/src/components/common/RegistrationTable.vue`、`frontend/src/components/common/MyRegistrations.vue`、`frontend/src/pages/OrganizerRegistrations.vue`、`frontend/src/pages/Profile.vue`

### ActivityType（lecture/training/party/competition）
- 后端：`backend/internal/constants/activity.go`、`backend/internal/model/activity.go`、`backend/internal/service/activity_service.go`、`backend/internal/util/formatters.go`、`backend/internal/constants/log_templates.go`、`backend/internal/constants/error_codes.go`、`backend/internal/dto/dto_activity.go`、`database/init.sql`
- 前端：`frontend/src/constants/activity.ts`、`frontend/src/components/common/ActivityCard.vue`、`frontend/src/components/common/ActivityFilter.vue`、`frontend/src/components/common/ActivityForm.vue`、`frontend/src/pages/ActivityDetail.vue`、`frontend/src/pages/OrganizerActivities.vue`

## API 接口清单

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | /healthz | 服务健康检查 |
| GET | /api/healthz | Nginx 反代健康检查 |
| GET | /api/v1/healthz | API 版本健康检查 |
| POST | /api/v1/auth/register | 用户注册 |
| POST | /api/v1/auth/login | 用户登录，返回 JWT |
| GET | /api/v1/users/me | 当前登录用户信息 |
| PUT | /api/v1/users/me | 修改个人资料 |
| GET | /api/v1/users | 用户列表（仅管理员） |
| GET | /api/v1/activities | 活动分页列表 |
| GET | /api/v1/activities/calendar | 按月活动日历 |
| GET | /api/v1/activities/mine | 我的活动（组织者/管理员） |
| GET | /api/v1/activities/:id | 活动详情 |
| GET | /api/v1/activities/:id/stats | 报名/签到统计 |
| POST | /api/v1/activities | 创建活动（组织者/管理员） |
| PUT | /api/v1/activities/:id | 更新活动（组织者/管理员） |
| POST | /api/v1/activities/:id/publish | 发布活动 |
| POST | /api/v1/activities/:id/end | 结束活动 |
| DELETE | /api/v1/activities/:id | 删除活动 |
| POST | /api/v1/registrations | 在线报名 |
| GET | /api/v1/registrations | 报名列表（组织者/管理员） |
| GET | /api/v1/registrations/mine | 我的报名 |
| GET | /api/v1/registrations/export | 导出报名 CSV |
| POST | /api/v1/registrations/offline | 线下补录报名 |
| POST | /api/v1/registrations/:id/cancel | 取消报名 |
| POST | /api/v1/registrations/:id/review | 审核报名 |
| POST | /api/v1/check-ins | 凭证/扫码签到 |
| GET | /api/v1/check-ins | 签到记录 |
| GET | /api/v1/activities/:id/comments | 活动评论列表 |
| POST | /api/v1/activities/:id/comments | 发表评论 |
| GET | /api/v1/comments/mine | 我的评论 |
| DELETE | /api/v1/comments/:id | 删除评论 |
| POST | /api/v1/activities/:id/favorite | 收藏活动 |
| DELETE | /api/v1/activities/:id/favorite | 取消收藏 |
| GET | /api/v1/favorites/mine | 我的收藏 |
| GET | /api/v1/notifications/mine | 我的通知 |
| POST | /api/v1/notifications/:id/read | 标记单条已读 |
| POST | /api/v1/notifications/read-all | 全部标记已读 |

## 主要功能

- 活动发布：创建、编辑、发布、结束、下架活动，活动封面图上传。
- 在线报名：名额校验、报名截止校验、防重复报名、凭证号生成、审核与取消。
- 签到管理：凭证号签到、扫码签到、签到率统计、报名名单导出 CSV。
- 活动日历：月历视图展示活动分布，日期格子显示活动数量，点击日期展开当天活动。
- 评论收藏：评分评论、平均分展示、收藏与取消收藏。
- 消息通知：报名成功、审核结果、签到成功自动通知。
- 角色权限：JWT + RBAC（user/organizer/admin），操作审计日志。

## License

MIT License
