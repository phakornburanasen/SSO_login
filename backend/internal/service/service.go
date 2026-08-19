package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/netip"
	"strings"
	"time"

	"sso-login/backend/internal/auth"
	"sso-login/backend/internal/model"
	"sso-login/backend/internal/store"
)

type Service struct {
	repo     store.Repository
	jwt      *auth.JWT
	ttl      time.Duration
	employee *EmployeeAPIClient
}

func New(repo store.Repository, jwt *auth.JWT, jwtTTL time.Duration, emp *EmployeeAPIClient) *Service {
	return &Service{repo: repo, jwt: jwt, ttl: jwtTTL, employee: emp}
}

// GetEmployee คืนข้อมูล employee สำหรับแสดงใน navbar
// empID คือ ADUser (เช่น T9058) — จะถูกส่งเป็น emp_id ให้ external API
// คืน (employee, fromCache, error)
func (s *Service) GetEmployee(ctx context.Context, empID string) (Employee, bool, error) {
	if s.employee == nil {
		return Employee{}, false, errors.New("employee api client not configured")
	}
	return s.employee.GetEmployee(ctx, empID)
}

// CheckAccess ตรวจสิทธิ์ตาม (base_url, client_ip, ad_username)
// คืนผลลัพธ์ + บันทึก audit log ทุกครั้ง
func (s *Service) CheckAccess(ctx context.Context, in model.CheckAccessRequest) (model.CheckAccessResult, error) {
	res := model.CheckAccessResult{
		BaseURL:    in.BaseURL,
		ADUsername: strings.TrimSpace(in.ADUsername),
		ClientIP:   strings.TrimSpace(in.ClientIP),
	}

	// validate
	if res.BaseURL == "" {
		res.Result, res.Reason = "DENY_APP", "base_url is required"
		s.audit(ctx, res, in, "missing base_url")
		return res, nil
	}
	ip, ok := model.ParseClientIP(res.ClientIP)
	if !ok {
		res.Result, res.Reason = "DENY_IP", "invalid client_ip"
		s.audit(ctx, res, in, "invalid client_ip")
		return res, nil
	}
	res.ClientIP = ip.String()
	if res.ADUsername == "" {
		res.Result, res.Reason = "DENY_USER", "ad_username is required"
		s.audit(ctx, res, in, "missing ad_username")
		return res, nil
	}

	// 1) resolve env
	env, err := s.repo.ResolveEnvByBaseURL(ctx, res.BaseURL)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			res.Result, res.Reason = "DENY_APP", "application or base_url not found"
			s.audit(ctx, res, in, "env not found")
			return res, nil
		}
		res.Result, res.Reason = "ERROR", err.Error()
		s.audit(ctx, res, in, "resolve env error")
		return res, err
	}
	res.AppCode, res.EnvCode = env.AppCode, env.EnvCode

	// 2) check IP
	ipOK, err := s.repo.IPAllowed(ctx, env.ID, ip)
	if err != nil {
		res.Result, res.Reason = "ERROR", err.Error()
		s.audit(ctx, res, in, "ip check error")
		return res, err
	}
	if !ipOK {
		res.Result, res.Reason = "DENY_IP", "client_ip is not in allowed list"
		s.audit(ctx, res, in, "ip denied")
		return res, nil
	}

	// 3) check user
	userOK, reason, err := s.repo.UserAllowed(ctx, env.ID, res.ADUsername)
	if err != nil {
		res.Result, res.Reason = "ERROR", err.Error()
		s.audit(ctx, res, in, "user check error")
		return res, err
	}
	if !userOK {
		msg := "ad_username is not allowed"
		if reason == "expired" {
			msg = "ad_username access expired"
		}
		res.Result, res.Reason = "DENY_USER", msg
		s.audit(ctx, res, in, "user denied")
		return res, nil
	}

	res.Allow = true
	res.Result = "ALLOW"
	res.Reason = ""
	s.audit(ctx, res, in, "")
	return res, nil
}

func (s *Service) audit(ctx context.Context, res model.CheckAccessResult, in model.CheckAccessRequest, extra string) {
	ev := model.LoginAudit{
		AppCode:     res.AppCode,
		EnvCode:     res.EnvCode,
		BaseURL:     res.BaseURL,
		ClientIP:    res.ClientIP,
		ADUsername:  res.ADUsername,
		EmployeeID:  in.EmployeeID,
		DisplayName: in.DisplayName,
		Result:      res.Result,
		DenyReason:  joinReason(res.Reason, extra),
		UserAgent:   in.UserAgent,
		RequestID:   newRequestID(),
	}
	if e := s.repo.InsertAudit(ctx, ev); e != nil {
		log.Printf("audit insert failed: %v", e)
	}
}

func joinReason(reason, extra string) string {
	if reason == "" {
		return extra
	}
	if extra == "" {
		return reason
	}
	return reason + " | " + extra
}

func newRequestID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

// ---------- Pass-through to repo (for HTTP handlers) ----------

func (s *Service) ListApps(ctx context.Context) ([]model.Application, error) {
	return s.repo.ListApplications(ctx)
}
func (s *Service) CreateApp(ctx context.Context, a model.Application) (model.Application, error) {
	if strings.TrimSpace(a.Code) == "" || strings.TrimSpace(a.Name) == "" {
		return a, errors.New("code and name are required")
	}
	return s.repo.CreateApplication(ctx, a)
}
func (s *Service) UpdateApp(ctx context.Context, a model.Application) (model.Application, error) {
	if a.ID <= 0 {
		return a, errors.New("id is required")
	}
	return s.repo.UpdateApplication(ctx, a)
}
func (s *Service) DeleteApp(ctx context.Context, id int64) error {
	return s.repo.DeleteApplication(ctx, id)
}

func (s *Service) ListEnvs(ctx context.Context, appID int64) ([]model.Environment, error) {
	return s.repo.ListEnvironments(ctx, appID)
}
func (s *Service) CreateEnv(ctx context.Context, e model.Environment) (model.Environment, error) {
	if e.AppID <= 0 || strings.TrimSpace(e.EnvCode) == "" || strings.TrimSpace(e.EnvName) == "" || strings.TrimSpace(e.BaseURL) == "" {
		return e, errors.New("appId, envCode, envName, baseUrl are required")
	}
	// ตรวจ host_ip ถ้ากรอกมา
	if e.HostIP != "" {
		if _, err := netip.ParseAddr(e.HostIP); err != nil {
			return e, errors.New("hostIp is invalid")
		}
	}
	return s.repo.CreateEnvironment(ctx, e)
}
func (s *Service) UpdateEnv(ctx context.Context, e model.Environment) (model.Environment, error) {
	if e.ID <= 0 {
		return e, errors.New("id is required")
	}
	if e.HostIP != "" {
		if _, err := netip.ParseAddr(e.HostIP); err != nil {
			return e, errors.New("hostIp is invalid")
		}
	}
	return s.repo.UpdateEnvironment(ctx, e)
}
func (s *Service) DeleteEnv(ctx context.Context, id int64) error {
	return s.repo.DeleteEnvironment(ctx, id)
}

func (s *Service) ListAllowedIPs(ctx context.Context, envID int64) ([]model.AllowedIP, error) {
	return s.repo.ListAllowedIPs(ctx, envID)
}
func (s *Service) CreateAllowedIP(ctx context.Context, v model.AllowedIP) (model.AllowedIP, error) {
	if v.EnvID <= 0 || strings.TrimSpace(v.IPCIDR) == "" {
		return v, errors.New("envId and ipCidr are required")
	}
	if _, err := netip.ParsePrefix(v.IPCIDR); err != nil {
		if _, err2 := netip.ParseAddr(v.IPCIDR); err2 != nil {
			return v, errors.New("ipCidr is invalid (ต้องเป็น IP เช่น 10.0.32.71 หรือ CIDR เช่น 10.0.32.0/24)")
		}
	}
	return s.repo.CreateAllowedIP(ctx, v)
}
func (s *Service) DeleteAllowedIP(ctx context.Context, id int64) error {
	return s.repo.DeleteAllowedIP(ctx, id)
}

func (s *Service) ListAllowedUsers(ctx context.Context, envID int64) ([]model.AllowedUser, error) {
	return s.repo.ListAllowedUsers(ctx, envID)
}
func (s *Service) CreateAllowedUser(ctx context.Context, v model.AllowedUser) (model.AllowedUser, error) {
	if v.EnvID <= 0 || strings.TrimSpace(v.ADUsername) == "" {
		return v, errors.New("envId and adUsername are required")
	}
	return s.repo.CreateAllowedUser(ctx, v)
}
func (s *Service) DeleteAllowedUser(ctx context.Context, id int64) error {
	return s.repo.DeleteAllowedUser(ctx, id)
}

// BulkCreateAllowedUsers creates allowed-user entries for the same AD user across
// multiple environments in one call.  Skips environments where the user already
// has a grant (ErrConflict) and returns the list of successfully created entries.
func (s *Service) BulkCreateAllowedUsers(ctx context.Context, adUsername string, envIDs []int64, expiresAt *model.FlexibleTime, grantedBy string) ([]model.AllowedUser, []string) {
	var created []model.AllowedUser
	var skipped []string
	for _, envID := range envIDs {
		v := model.AllowedUser{
			EnvID:      envID,
			ADUsername: strings.TrimSpace(adUsername),
			GrantedBy:  grantedBy,
		}
		if expiresAt != nil && !expiresAt.Time.IsZero() {
			v.ExpiresAt = expiresAt
		}
		out, err := s.repo.CreateAllowedUser(ctx, v)
		if err != nil {
			skipped = append(skipped, "env#"+fmt.Sprint(envID))
			continue
		}
		created = append(created, out)
	}
	return created, skipped
}

// ListAllowedUsersByADUsername returns all environment grants for a given AD user.
// Used by the "copy permissions from another user" feature.
func (s *Service) ListAllowedUsersByADUsername(ctx context.Context, adUsername string) ([]model.AllowedUser, error) {
	return s.repo.ListAllowedUsersByADUsername(ctx, adUsername)
}

func (s *Service) ListAudit(ctx context.Context, limit int) ([]model.LoginAudit, error) {
	return s.repo.ListAudit(ctx, limit)
}

// ---------- Admin Auth (JWT, stateless) ----------

// ErrNoPermission ใช้ตอน login สำเร็จแต่ไม่มี env ให้ดูแล
var ErrNoPermission = errors.New("คุณยังไม่ได้รับสิทธิ์เข้าใช้งานระบบ กรุณาติดต่อผู้ดูแลเพื่อขอสิทธิ์")

// ErrInvalidCredentials ใช้ตอน username/password ผิด
var ErrInvalidCredentials = errors.New("ชื่อผู้ใช้หรือรหัสผ่านไม่ถูกต้อง")

// ErrSystemNotReady ใช้ตอน schema/table ไม่พร้อม
var ErrSystemNotReady = errors.New("ระบบยังไม่พร้อมใช้งาน กรุณาติดต่อผู้ดูแลระบบ")

// friendlyDBError แปลง error จาก DB ให้เป็นข้อความที่ user เข้าใจ
// - ถ้าเป็น "table not found" (SQLSTATE 42P01) → แจ้งให้รัน migration
// - อื่นๆ → คืน error เดิม
func friendlyDBError(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	// ตรวจ SQLSTATE 42P01 (undefined_table) — pgx ส่งผ่าน wrapped error
	if strings.Contains(msg, "SQLSTATE 42P01") ||
		strings.Contains(msg, "does not exist") && strings.Contains(msg, "relation") {
		return fmt.Errorf("%w — %s", ErrSystemNotReady,
			"ตาราง sso_admins ยังไม่ถูกสร้าง กรุณารัน migration: psql -f migrations/003_sso_admins_table.sql")
	}
	return err
}

// Login ตรวจ AD username + password แบบง่าย (ยังไม่ผูก AD จริง ตามที่ผู้ใช้ระบุ)
// แล้วออก JWT ให้ client — ไม่ต้องเก็บ session ในฐานข้อมูล
//
// สำคัญ: หลัง auth สำเร็จ ต้อง query sso_environments.ADUser เพื่อดูว่า user
// มีสิทธิ์จัดการ env ใดบ้าง ถ้าไม่มีเลย → ErrNoPermission
// แล้วแนบ list ของ envs ที่ดูแลได้ไปกับ LoginResult
func (s *Service) Login(ctx context.Context, username, password, clientIP string) (model.LoginResult, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return model.LoginResult{}, ErrInvalidCredentials
	}
	if len(username) > 15 {
		return model.LoginResult{}, fmt.Errorf("ชื่อผู้ใช้ยาวเกิน 15 ตัวอักษร")
	}
	// ตอนนี้: ยอมรับ user ที่ไม่ว่าง และ password ที่ไม่ว่าง
	// (ในอนาคตสามารถต่อ AD/LDAP ได้ที่นี่)
	if len(password) < 3 {
		return model.LoginResult{}, fmt.Errorf("รหัสผ่านต้องมีอย่างน้อย 3 ตัวอักษร")
	}

	// เช็คว่า user เป็น admin หรือไม่
	isAdmin, err := s.repo.IsAdminUser(ctx, username)
	if err != nil {
		return model.LoginResult{}, friendlyDBError(err)
	}

	var (
		accessibleEnvs []model.AccessibleEnv
		role           string
	)

	if isAdmin {
		// admin: ดู env ทั้งหมด
		accessibleEnvs, err = s.listAllActiveEnvs(ctx)
		if err != nil {
			return model.LoginResult{}, friendlyDBError(err)
		}
		role = "admin"
	} else {
		// user ทั่วไป: เฉพาะ env ที่ตัวเองเป็น ADUser
		accessibleEnvs, err = s.repo.ListAccessibleEnvsByADUser(ctx, username)
		if err != nil {
			return model.LoginResult{}, friendlyDBError(err)
		}
		// ถ้าไม่มี env เลย → ไม่มีสิทธิ์ → DENY
		if len(accessibleEnvs) == 0 {
			return model.LoginResult{}, ErrNoPermission
		}
		role = "user"
	}

	token, exp, err := s.jwt.Sign(username, username, role, s.ttl)
	if err != nil {
		return model.LoginResult{}, fmt.Errorf("ไม่สามารถสร้าง token ได้: %w", err)
	}

	_ = clientIP // ใช้สำหรับ audit ในอนาคต

	return model.LoginResult{
		Token:          token,
		Username:       username,
		DisplayName:    username,
		ExpiresAt:      exp,
		AccessibleEnvs: accessibleEnvs,
		Role:           role,
	}, nil
}

// listAllActiveEnvs คืน env ทั้งหมดที่ active (ใช้สำหรับ admin)
func (s *Service) listAllActiveEnvs(ctx context.Context) ([]model.AccessibleEnv, error) {
	all, err := s.repo.ListEnvironments(ctx, 0)
	if err != nil {
		return nil, err
	}
	out := make([]model.AccessibleEnv, 0, len(all))
	for _, e := range all {
		if !e.Active {
			continue
		}
		out = append(out, model.AccessibleEnv{
			ID:       e.ID,
			AppCode:  e.AppCode,
			EnvCode:  e.EnvCode,
			EnvName:  e.EnvName,
			BaseURL:  e.BaseURL,
			HostIP:   e.HostIP,
			BasePath: e.BasePath,
			Active:   e.Active,
		})
	}
	return out, nil
}

// MyEnvs คืน env ทั้งหมดที่ user มีสิทธิ์จัดการ (ใช้กับ /api/my-envs)
// - admin: env ทั้งหมดที่ active
// - user: เฉพาะ env ที่ตัวเองเป็น ADUser
func (s *Service) MyEnvs(ctx context.Context, username string) ([]model.AccessibleEnv, error) {
	isAdmin, err := s.repo.IsAdminUser(ctx, username)
	if err != nil {
		return nil, err
	}
	if isAdmin {
		return s.listAllActiveEnvs(ctx)
	}
	return s.repo.ListAccessibleEnvsByADUser(ctx, username)
}

// Logout สำหรับ JWT: เป็น no-op เพราะ token เป็น stateless
// Client ควรลบ token ทิ้งจาก localStorage เอง
func (s *Service) Logout(ctx context.Context, token string) error {
	return nil
}

// VerifyToken ตรวจ JWT signature + expiration (ไม่ต้อง query DB)
func (s *Service) VerifyToken(ctx context.Context, token string) (string, error) {
	uname, _, err := s.VerifyTokenWithRole(ctx, token)
	return uname, err
}

// VerifyTokenWithRole ตรวจ JWT แล้วคืน username + role
// role ฝังอยู่ใน JWT claims ตอน Login — ไม่ต้อง query DB
func (s *Service) VerifyTokenWithRole(ctx context.Context, token string) (string, string, error) {
	if token == "" {
		return "", "", store.ErrUnauthorized
	}
	claims, err := s.jwt.Verify(token)
	if err != nil {
		return "", "", store.ErrUnauthorized
	}
	return claims.Sub, claims.Role, nil
}

// ---------- Admin management ----------

func (s *Service) ListAdmins(ctx context.Context) ([]model.Admin, error) {
	return s.repo.ListAdmins(ctx)
}

func (s *Service) AddAdmin(ctx context.Context, a model.Admin) (int64, error) {
	return s.repo.AddAdmin(ctx, a)
}

func (s *Service) RemoveAdmin(ctx context.Context, id int64) error {
	return s.repo.RemoveAdmin(ctx, id)
}
