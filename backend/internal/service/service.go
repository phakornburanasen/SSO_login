package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"net/netip"
	"strings"
	"time"

	"sso-login/backend/internal/model"
	"sso-login/backend/internal/store"
)

type Service struct {
	repo store.Repository
}

func New(repo store.Repository) *Service { return &Service{repo: repo} }

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

func (s *Service) ListAudit(ctx context.Context, limit int) ([]model.LoginAudit, error) {
	return s.repo.ListAudit(ctx, limit)
}

// ---------- Admin Auth ----------

// Login ตรวจ AD username + password แบบง่าย (ยังไม่ผูก AD จริง ตามที่ผู้ใช้ระบุ)
func (s *Service) Login(ctx context.Context, username, password, clientIP string) (model.LoginResult, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return model.LoginResult{}, errors.New("username and password are required")
	}
	if len(username) > 15 {
		return model.LoginResult{}, errors.New("username too long (max 15)")
	}
	// ตอนนี้: ยอมรับ user ที่ไม่ว่าง และ password ที่ไม่ว่าง
	// (ในอนาคตสามารถต่อ AD/LDAP ได้ที่นี่)
	if len(password) < 3 {
		return model.LoginResult{}, errors.New("password too short")
	}

	// สร้าง token
	var b [32]byte
	if _, e := rand.Read(b[:]); e != nil {
		return model.LoginResult{}, e
	}
	token := hex.EncodeToString(b[:])
	expires := time.Now().Add(8 * time.Hour)
	if e := s.repo.CreateAdminSession(ctx, token, username, clientIP, expires); e != nil {
		return model.LoginResult{}, e
	}
	return model.LoginResult{
		Token:       token,
		Username:    username,
		DisplayName: username,
		ExpiresAt:   expires,
	}, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	return s.repo.DeleteAdminSession(ctx, token)
}

func (s *Service) VerifyToken(ctx context.Context, token string) (string, error) {
	if token == "" {
		return "", store.ErrUnauthorized
	}
	return s.repo.GetAdminSession(ctx, token)
}
