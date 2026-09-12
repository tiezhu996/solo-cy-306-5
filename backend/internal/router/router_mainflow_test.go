package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gbevent/internal/config"
	"gbevent/internal/handler"
	"gbevent/internal/model"
	"gbevent/internal/repository"
	"gbevent/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	gormlogger "gorm.io/gorm/logger"
)

// mainFlowEnv 按 cmd/server/main.go 的装配方式搭建完整 HTTP 栈（SQLite 内存库）。
type mainFlowEnv struct {
	t      *testing.T
	db     *gorm.DB
	engine *gin.Engine
}

func newMainFlowEnv(t *testing.T) *mainFlowEnv {
	t.Helper()
	dsn := "file:" + strings.NewReplacer("/", "_", " ", "_").Replace(t.Name()) + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
		// SQLite 不支持 FOR UPDATE，测试环境剥离 LOCKING 子句；生产 MySQL 不变。
		ClauseBuilders: map[string]clause.ClauseBuilder{
			"LOCKING": func(c clause.Clause, builder clause.Builder) {},
		},
	})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(
		&model.User{}, &model.Activity{}, &model.Registration{}, &model.CheckInRecord{},
		&model.Comment{}, &model.Favorite{}, &model.Notification{}, &model.AuditLog{},
	); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}

	cfg := &config.Config{
		ServerPort:         "8080",
		JWTSecret:          "test-secret",
		JWTExpireHours:     72,
		RateLimitPerMinute: 100000,
		UploadDir:          t.TempDir(),
		UploadMaxMB:        10,
		CORSOrigins:        []string{"*"},
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	userRepo := repository.NewUserRepository(db)
	activityRepo := repository.NewActivityRepository(db)
	regRepo := repository.NewRegistrationRepository(db)
	checkinRepo := repository.NewCheckInRecordRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	favoriteRepo := repository.NewFavoriteRepository(db)
	notifyRepo := repository.NewNotificationRepository(db)

	userSvc := service.NewUserService(userRepo, logger)
	notifySvc := service.NewNotificationService(notifyRepo, logger)
	accessPolicy := service.NewActivityAccessPolicy()
	signupGuard := service.NewSignupGuard(activityRepo)
	activitySvc := service.NewActivityService(activityRepo, checkinRepo, accessPolicy, logger)
	regSvc := service.NewRegistrationService(db, regRepo, activitySvc, signupGuard, notifySvc, logger)
	checkinSvc := service.NewCheckInRecordService(db, checkinRepo, regRepo, notifySvc, logger)
	commentSvc := service.NewCommentService(commentRepo, activitySvc, logger)
	favoriteSvc := service.NewFavoriteService(favoriteRepo, activitySvc, logger)

	r := New(cfg, db, logger,
		handler.NewUserHandler(userSvc, logger),
		handler.NewActivityHandler(activitySvc, logger),
		handler.NewRegistrationHandler(regSvc, logger),
		handler.NewCheckInRecordHandler(checkinSvc, logger),
		handler.NewCommentHandler(commentSvc, logger),
		handler.NewFavoriteHandler(favoriteSvc, logger),
		handler.NewNotificationHandler(notifySvc, logger),
		handler.NewUploadHandler(cfg, logger),
	)
	return &mainFlowEnv{t: t, db: db, engine: r.Setup()}
}

// do 发起请求并返回状态码与解析后的响应体。
func (e *mainFlowEnv) do(method, path, token string, body any) (int, map[string]any) {
	e.t.Helper()
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			e.t.Fatalf("marshal body: %v", err)
		}
		reader = bytes.NewReader(raw)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	e.engine.ServeHTTP(w, req)

	var parsed map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &parsed); err != nil {
		e.t.Fatalf("%s %s: response not JSON: %s", method, path, w.Body.String())
	}
	return w.Code, parsed
}

// mustCode 断言 HTTP 状态码与业务错误码。
func (e *mainFlowEnv) mustCode(status, gotStatus int, resp map[string]any, wantCode float64) {
	e.t.Helper()
	if gotStatus != status {
		e.t.Fatalf("http status = %d, want %d, resp=%v", gotStatus, status, resp)
	}
	if resp["code"] != wantCode {
		e.t.Fatalf("business code = %v, want %v, resp=%v", resp["code"], wantCode, resp)
	}
}

// registerAndLogin 注册并登录，返回 token 与用户 ID。
func (e *mainFlowEnv) registerAndLogin(username, password, role string) (string, uint64) {
	e.t.Helper()
	status, resp := e.do(http.MethodPost, "/api/v1/auth/register", "", map[string]any{
		"username": username, "password": password, "nickname": username, "role": role,
	})
	e.mustCode(http.StatusOK, status, resp, 0)

	status, resp = e.do(http.MethodPost, "/api/v1/auth/login", "", map[string]any{
		"username": username, "password": password,
	})
	e.mustCode(http.StatusOK, status, resp, 0)
	data := resp["data"].(map[string]any)
	token := data["token"].(string)
	userID := uint64(data["user"].(map[string]any)["id"].(float64))
	return token, userID
}

// TestPageMainFlows 页面主流程端到端回归：
// 注册登录 → 创建/发布活动 → 列表/日历/详情 → 报名 → 我的报名/通知 →
// 审核 → 签到 → 统计/导出 → 取消冲突，以及关键权限负例。
func TestPageMainFlows(t *testing.T) {
	env := newMainFlowEnv(t)

	// 健康检查。
	status, _ := env.do(http.MethodGet, "/api/v1/healthz", "", nil)
	if status != http.StatusOK {
		t.Fatalf("healthz status = %d", status)
	}

	// 账号：组织者、普通用户、管理员（注册接口仅允许 user/organizer，管理员直接改库）。
	organizerToken, organizerID := env.registerAndLogin("org1", "Passw0rd!", "organizer")
	otherOrgToken, _ := env.registerAndLogin("org2", "Passw0rd!", "organizer")
	userToken, _ := env.registerAndLogin("user1", "Passw0rd!", "user")
	adminToken, _ := env.registerAndLogin("admin1", "Passw0rd!", "user")
	if err := env.db.Exec("UPDATE users SET role = 'admin' WHERE username = 'admin1'").Error; err != nil {
		t.Fatalf("promote admin: %v", err)
	}
	// 角色在 token 内，需重新登录获取 admin 角色。
	status, resp := env.do(http.MethodPost, "/api/v1/auth/login", "", map[string]any{
		"username": "admin1", "password": "Passw0rd!",
	})
	env.mustCode(http.StatusOK, status, resp, 0)
	adminToken = resp["data"].(map[string]any)["token"].(string)

	// /organizer/activities：创建活动（草稿）。
	start := time.Now().Add(48 * time.Hour)
	status, resp = env.do(http.MethodPost, "/api/v1/activities", organizerToken, map[string]any{
		"title": "周末讲座", "activity_type": "lecture", "location": "报告厅",
		"start_time":      start.Format(time.RFC3339),
		"end_time":        start.Add(2 * time.Hour).Format(time.RFC3339),
		"signup_deadline": start.Add(-24 * time.Hour).Format(time.RFC3339),
		"capacity":        2,
	})
	env.mustCode(http.StatusOK, status, resp, 0)
	activity := resp["data"].(map[string]any)
	activityID := uint64(activity["id"].(float64))
	if activity["status"] != "draft" {
		t.Fatalf("new activity status = %v, want draft", activity["status"])
	}

	// 权限负例：普通用户不能创建活动（RBAC 40300）。
	status, resp = env.do(http.MethodPost, "/api/v1/activities", userToken, map[string]any{
		"title": "越权活动", "activity_type": "lecture",
		"start_time": start.Format(time.RFC3339), "end_time": start.Format(time.RFC3339),
		"signup_deadline": start.Format(time.RFC3339),
	})
	env.mustCode(http.StatusForbidden, status, resp, 40300)

	// 权限负例：其他组织者不能编辑本活动（业务 40300 + 文案）。
	status, resp = env.do(http.MethodPut, fmt.Sprintf("/api/v1/activities/%d", activityID), otherOrgToken,
		map[string]any{"title": "越权改名", "capacity": 10})
	env.mustCode(http.StatusForbidden, status, resp, 40300)
	if resp["message"] != fmt.Sprintf("Activity[id=%d] update forbidden: organizer not match", activityID) {
		t.Fatalf("forbidden message = %v", resp["message"])
	}

	// 发布活动。
	status, resp = env.do(http.MethodPost, fmt.Sprintf("/api/v1/activities/%d/publish", activityID), organizerToken, nil)
	env.mustCode(http.StatusOK, status, resp, 0)
	if resp["message"] != "活动已发布" {
		t.Fatalf("publish message = %v", resp["message"])
	}
	if resp["data"].(map[string]any)["status"] != "published" {
		t.Fatalf("status after publish = %v", resp["data"])
	}

	// /activities：列表应包含已发布活动。
	status, resp = env.do(http.MethodGet, "/api/v1/activities?status=published&page=1&page_size=10", "", nil)
	env.mustCode(http.StatusOK, status, resp, 0)
	page := resp["data"].(map[string]any)
	if page["total"].(float64) != 1 {
		t.Fatalf("published total = %v, want 1", page["total"])
	}

	// /calendar：当月日历应包含该活动。
	status, resp = env.do(http.MethodGet, "/api/v1/activities/calendar?month="+start.Format("2006-01"), "", nil)
	env.mustCode(http.StatusOK, status, resp, 0)
	if len(resp["data"].([]any)) != 1 {
		t.Fatalf("calendar activities = %v, want 1", resp["data"])
	}

	// /activities/:id：详情带 registered_count。
	status, resp = env.do(http.MethodGet, fmt.Sprintf("/api/v1/activities/%d", activityID), "", nil)
	env.mustCode(http.StatusOK, status, resp, 0)
	detail := resp["data"].(map[string]any)
	if detail["registered_count"].(float64) != 0 {
		t.Fatalf("registered_count = %v, want 0", detail["registered_count"])
	}

	// /activities/:id 报名：用户报名成功并生成凭证。
	status, resp = env.do(http.MethodPost, "/api/v1/registrations", userToken, map[string]any{
		"activity_id": activityID, "name": "张三", "phone": "13800000000",
	})
	env.mustCode(http.StatusOK, status, resp, 0)
	if resp["message"] != "报名成功" {
		t.Fatalf("signup message = %v", resp["message"])
	}
	reg := resp["data"].(map[string]any)
	regID := uint64(reg["id"].(float64))
	voucher := reg["voucher_no"].(string)
	if voucher == "" {
		t.Fatal("voucher_no should be generated")
	}

	// 重复报名：40903。
	status, resp = env.do(http.MethodPost, "/api/v1/registrations", userToken, map[string]any{
		"activity_id": activityID, "name": "张三", "phone": "13800000000",
	})
	env.mustCode(http.StatusConflict, status, resp, 40903)

	// /profile 我的报名。
	status, resp = env.do(http.MethodGet, "/api/v1/registrations/mine?page=1&page_size=10", userToken, nil)
	env.mustCode(http.StatusOK, status, resp, 0)
	if resp["data"].(map[string]any)["total"].(float64) != 1 {
		t.Fatalf("my registrations = %v, want 1", resp["data"])
	}

	// /profile 我的通知：报名成功通知已写入。
	status, resp = env.do(http.MethodGet, "/api/v1/notifications/mine?page=1&page_size=10", userToken, nil)
	env.mustCode(http.StatusOK, status, resp, 0)
	notifPage := resp["data"].(map[string]any)
	if notifPage["total"].(float64) != 1 {
		t.Fatalf("my notifications = %v, want 1", notifPage["total"])
	}
	firstNotif := notifPage["list"].([]any)[0].(map[string]any)
	if firstNotif["notification_type"] != "signup_success" {
		t.Fatalf("notification type = %v", firstNotif["notification_type"])
	}

	// /organizer/activities 我的活动列表。
	status, resp = env.do(http.MethodGet, "/api/v1/activities/mine?page=1&page_size=10", organizerToken, nil)
	env.mustCode(http.StatusOK, status, resp, 0)
	if resp["data"].(map[string]any)["total"].(float64) != 1 {
		t.Fatalf("organizer activities = %v, want 1", resp["data"])
	}

	// /organizer/registrations 统计：1 人报名。
	status, resp = env.do(http.MethodGet, fmt.Sprintf("/api/v1/activities/%d/stats", activityID), organizerToken, nil)
	env.mustCode(http.StatusOK, status, resp, 0)
	stats := resp["data"].(map[string]any)
	if stats["registered_count"].(float64) != 1 || stats["checked_in_count"].(float64) != 0 {
		t.Fatalf("stats = %v", stats)
	}

	// 审核通过。
	status, resp = env.do(http.MethodPost, fmt.Sprintf("/api/v1/registrations/%d/review", regID), organizerToken,
		map[string]any{"review_status": "approved"})
	env.mustCode(http.StatusOK, status, resp, 0)
	if resp["data"].(map[string]any)["review_status"] != "approved" {
		t.Fatalf("review_status = %v", resp["data"])
	}

	// 审核结果通知已写入。
	status, resp = env.do(http.MethodGet, "/api/v1/notifications/mine?page=1&page_size=10", userToken, nil)
	env.mustCode(http.StatusOK, status, resp, 0)
	if resp["data"].(map[string]any)["total"].(float64) != 2 {
		t.Fatalf("notifications after review = %v, want 2", resp["data"])
	}

	// 凭证签到。
	status, resp = env.do(http.MethodPost,
		fmt.Sprintf("/api/v1/check-ins?activity_id=%d", activityID), organizerToken,
		map[string]any{"voucher": voucher})
	env.mustCode(http.StatusOK, status, resp, 0)
	if resp["message"] != "签到成功" {
		t.Fatalf("checkin message = %v", resp["message"])
	}

	// 重复签到：40904。
	status, resp = env.do(http.MethodPost,
		fmt.Sprintf("/api/v1/check-ins?activity_id=%d", activityID), organizerToken,
		map[string]any{"voucher": voucher})
	env.mustCode(http.StatusConflict, status, resp, 40904)

	// 签到记录列表。
	status, resp = env.do(http.MethodGet, fmt.Sprintf("/api/v1/check-ins?activity_id=%d", activityID), organizerToken, nil)
	env.mustCode(http.StatusOK, status, resp, 0)
	if len(resp["data"].([]any)) != 1 {
		t.Fatalf("checkin records = %v, want 1", resp["data"])
	}

	// 签到后统计。
	status, resp = env.do(http.MethodGet, fmt.Sprintf("/api/v1/activities/%d/stats", activityID), organizerToken, nil)
	env.mustCode(http.StatusOK, status, resp, 0)
	stats = resp["data"].(map[string]any)
	if stats["checked_in_count"].(float64) != 1 || stats["checkin_rate"].(float64) != 100 {
		t.Fatalf("stats after checkin = %v", stats)
	}

	// 导出 CSV。
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/registrations/export?activity_id=%d", activityID), nil)
	req.Header.Set("Authorization", "Bearer "+organizerToken)
	w := httptest.NewRecorder()
	env.engine.ServeHTTP(w, req)
	if w.Code != http.StatusOK || !strings.Contains(w.Header().Get("Content-Type"), "text/csv") {
		t.Fatalf("export status=%d content-type=%s", w.Code, w.Header().Get("Content-Type"))
	}
	if !strings.Contains(w.Body.String(), voucher) {
		t.Fatalf("csv should contain voucher %s", voucher)
	}

	// 已签到不能取消：40906。
	status, resp = env.do(http.MethodPost, fmt.Sprintf("/api/v1/registrations/%d/cancel", regID), userToken, nil)
	env.mustCode(http.StatusConflict, status, resp, 40906)

	// 管理员可结束他人活动；结束后报名返回 40902。
	status, resp = env.do(http.MethodPost, fmt.Sprintf("/api/v1/activities/%d/end", activityID), adminToken, nil)
	env.mustCode(http.StatusOK, status, resp, 0)
	if resp["data"].(map[string]any)["status"] != "ended" {
		t.Fatalf("status after end = %v", resp["data"])
	}
	otherUserToken, _ := env.registerAndLogin("user2", "Passw0rd!", "user")
	status, resp = env.do(http.MethodPost, "/api/v1/registrations", otherUserToken, map[string]any{
		"activity_id": activityID, "name": "李四", "phone": "13900000000",
	})
	// 既有行为：CodeActivityEnded 经 appErrorStatus 落入 default 分支，HTTP 400 + 业务码 40902。
	env.mustCode(http.StatusBadRequest, status, resp, 40902)

	// 未登录访问受保护接口：40100。
	status, resp = env.do(http.MethodGet, "/api/v1/activities/mine", "", nil)
	env.mustCode(http.StatusUnauthorized, status, resp, 40100)

	// 组织者 ID 仅供后续断言使用，避免未使用告警。
	_ = organizerID
}
