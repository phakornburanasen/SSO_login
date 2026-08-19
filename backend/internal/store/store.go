package store

import (
	"context"
	"errors"
	"net/netip"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"sso-login/backend/internal/model"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("conflict")
	ErrUnauthorized = errors.New("unauthorized")
)

type Repository interface {
	// Check access
	ResolveEnvByBaseURL(ctx context.Context, baseURL string) (model.Environment, error)
	IPAllowed(ctx context.Context, envID int64, clientIP netip.Addr) (bool, error)
	UserAllowed(ctx context.Context, envID int64, adUsername string) (bool, string, error)
	InsertAudit(ctx context.Context, a model.LoginAudit) error

	// Application
	ListApplications(ctx context.Context) ([]model.Application, error)
	CreateApplication(ctx context.Context, a model.Application) (model.Application, error)
	UpdateApplication(ctx context.Context, a model.Application) (model.Application, error)
	DeleteApplication(ctx context.Context, id int64) error

	// Environment
	ListEnvironments(ctx context.Context, appID int64) ([]model.Environment, error)
	CreateEnvironment(ctx context.Context, e model.Environment) (model.Environment, error)
	UpdateEnvironment(ctx context.Context, e model.Environment) (model.Environment, error)
	DeleteEnvironment(ctx context.Context, id int64) error

	// Allowed IPs
	ListAllowedIPs(ctx context.Context, envID int64) ([]model.AllowedIP, error)
	CreateAllowedIP(ctx context.Context, ip model.AllowedIP) (model.AllowedIP, error)
	DeleteAllowedIP(ctx context.Context, id int64) error

	// Allowed Users
	ListAllowedUsers(ctx context.Context, envID int64) ([]model.AllowedUser, error)
	ListAllowedUsersByADUsername(ctx context.Context, adUsername string) ([]model.AllowedUser, error)
	CreateAllowedUser(ctx context.Context, u model.AllowedUser) (model.AllowedUser, error)
	DeleteAllowedUser(ctx context.Context, id int64) error

	// Audit
	ListAudit(ctx context.Context, limit int) ([]model.LoginAudit, error)

	// Auth: envs ที่ user เป็นเจ้าของ (กรองจาก sso_environments.ADUser)
	ListAccessibleEnvsByADUser(ctx context.Context, adUsername string) ([]model.AccessibleEnv, error)
	CountAccessibleEnvsByADUser(ctx context.Context, adUsername string) (int, error)
	IsAdminUser(ctx context.Context, adUsername string) (bool, error)

	// Admin management (sso_admins)
	ListAdmins(ctx context.Context) ([]model.Admin, error)
	AddAdmin(ctx context.Context, a model.Admin) (int64, error)
	RemoveAdmin(ctx context.Context, id int64) error

	// Note: admin auth is JWT-based (stateless) and does not use the repository.
	// The sso_sessions table is reserved for app-level access sessions.
}

type SQLStore struct{ pool *pgxpool.Pool }

func Open(ctx context.Context, dsn string, maxOpen int, maxLifetime time.Duration) (*SQLStore, error) {
	cfg, e := pgxpool.ParseConfig(dsn)
	if e != nil {
		return nil, e
	}
	if maxOpen > 0 {
		cfg.MaxConns = int32(maxOpen)
	}
	if maxLifetime > 0 {
		cfg.MaxConnLifetime = maxLifetime
	}
	pool, e := pgxpool.NewWithConfig(ctx, cfg)
	if e != nil {
		return nil, e
	}
	pingCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	if e = pool.Ping(pingCtx); e != nil {
		pool.Close()
		return nil, e
	}
	return &SQLStore{pool: pool}, nil
}

func (s *SQLStore) Close() { s.pool.Close() }

// ---------- Check Access ----------

func (s *SQLStore) ResolveEnvByBaseURL(ctx context.Context, baseURL string) (model.Environment, error) {
	var e model.Environment
	row := s.pool.QueryRow(ctx, `
		SELECT e.env_id, e.app_id, a.app_code, e.env_code, e.env_name,
		       e.base_url, COALESCE(host(e.host_ip),''), COALESCE(e.base_path,''),
		       COALESCE(e."ADUser",''),
		       e.is_active, e.created_at, e.updated_at
		FROM sso_environments e
		JOIN sso_applications a ON a.app_id = e.app_id
		WHERE e.base_url = $1 AND e.is_active = TRUE AND a.is_active = TRUE`,
		baseURL)
	err := row.Scan(&e.ID, &e.AppID, &e.AppCode, &e.EnvCode, &e.EnvName,
		&e.BaseURL, &e.HostIP, &e.BasePath, &e.ADUser,
		&e.Active, &e.CreatedAt, &e.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return e, ErrNotFound
	}
	return e, err
}

func (s *SQLStore) IPAllowed(ctx context.Context, envID int64, clientIP netip.Addr) (bool, error) {
	var ok bool
	row := s.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM sso_allowed_ips
			WHERE env_id = $1
			  AND is_active = TRUE
			  AND $2::inet <<= ip_cidr
		)`, envID, clientIP.String())
	err := row.Scan(&ok)
	return ok, err
}

func (s *SQLStore) UserAllowed(ctx context.Context, envID int64, adUsername string) (bool, string, error) {
	var (
		ok     bool
		reason string
	)
	row := s.pool.QueryRow(ctx, `
		SELECT
			EXISTS(
				SELECT 1 FROM sso_allowed_users
				WHERE env_id = $1
				  AND is_active = TRUE
				  AND LOWER(ad_username) = LOWER($2)
				  AND (expires_at IS NULL OR expires_at > NOW())
			) AS allowed,
			COALESCE((
				SELECT 'expired' FROM sso_allowed_users
				WHERE env_id = $1 AND LOWER(ad_username) = LOWER($2)
				LIMIT 1
			), '') AS status`,
		envID, adUsername)
	err := row.Scan(&ok, &reason)
	return ok, reason, err
}

func (s *SQLStore) InsertAudit(ctx context.Context, a model.LoginAudit) error {
	_, e := s.pool.Exec(ctx, `
		INSERT INTO sso_login_audit
		    (app_code, env_code, base_url, client_ip, ad_username, employee_id,
		     display_name, result, deny_reason, user_agent, request_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		a.AppCode, a.EnvCode, nullStr(a.BaseURL),
		nullStr(a.ClientIP), nullStr(a.ADUsername), nullStr(a.EmployeeID),
		nullStr(a.DisplayName), a.Result, nullStr(a.DenyReason),
		nullStr(a.UserAgent), nullStr(a.RequestID))
	return e
}

// ---------- Application CRUD ----------

func (s *SQLStore) ListApplications(ctx context.Context) ([]model.Application, error) {
	rows, e := s.pool.Query(ctx, `
		SELECT app_id, app_code, app_name, COALESCE(description,''), is_active, created_at, updated_at
		FROM sso_applications ORDER BY app_code`)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := make([]model.Application, 0)
	for rows.Next() {
		var a model.Application
		if e = rows.Scan(&a.ID, &a.Code, &a.Name, &a.Description, &a.Active, &a.CreatedAt, &a.UpdatedAt); e != nil {
			return nil, e
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *SQLStore) CreateApplication(ctx context.Context, a model.Application) (model.Application, error) {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO sso_applications (app_code, app_name, description)
		VALUES ($1,$2,$3)
		RETURNING app_id, app_code, app_name, COALESCE(description,''), is_active, created_at, updated_at`,
		a.Code, a.Name, nullStr(a.Description))
	if e := row.Scan(&a.ID, &a.Code, &a.Name, &a.Description, &a.Active, &a.CreatedAt, &a.UpdatedAt); e != nil {
		if isUniqueViolation(e) {
			return a, ErrConflict
		}
		return a, e
	}
	return a, nil
}

func (s *SQLStore) UpdateApplication(ctx context.Context, a model.Application) (model.Application, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE sso_applications
		SET app_name = $2, description = $3, is_active = $4, updated_at = NOW()
		WHERE app_id = $1
		RETURNING app_id, app_code, app_name, COALESCE(description,''), is_active, created_at, updated_at`,
		a.ID, a.Name, nullStr(a.Description), a.Active)
	if e := row.Scan(&a.ID, &a.Code, &a.Name, &a.Description, &a.Active, &a.CreatedAt, &a.UpdatedAt); e != nil {
		if errors.Is(e, pgx.ErrNoRows) {
			return a, ErrNotFound
		}
		return a, e
	}
	return a, nil
}

func (s *SQLStore) DeleteApplication(ctx context.Context, id int64) error {
	tag, e := s.pool.Exec(ctx, `DELETE FROM sso_applications WHERE app_id = $1`, id)
	if e != nil {
		return e
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ---------- Environment CRUD ----------

func (s *SQLStore) ListEnvironments(ctx context.Context, appID int64) ([]model.Environment, error) {
	rows, e := s.pool.Query(ctx, `
		SELECT e.env_id, e.app_id, a.app_code, e.env_code, e.env_name,
		       e.base_url, COALESCE(host(e.host_ip),''), COALESCE(e.base_path,''),
		       COALESCE(e."ADUser",''),
		       e.is_active, e.created_at, e.updated_at
		FROM sso_environments e
		JOIN sso_applications a ON a.app_id = e.app_id
		WHERE ($1 = 0 OR e.app_id = $1)
		ORDER BY a.app_code, e.env_code`, appID)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := make([]model.Environment, 0)
	for rows.Next() {
		var v model.Environment
		if e = rows.Scan(&v.ID, &v.AppID, &v.AppCode, &v.EnvCode, &v.EnvName,
			&v.BaseURL, &v.HostIP, &v.BasePath, &v.ADUser,
			&v.Active, &v.CreatedAt, &v.UpdatedAt); e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *SQLStore) CreateEnvironment(ctx context.Context, v model.Environment) (model.Environment, error) {
	hostIP := interface{}(nil)
	if v.HostIP != "" {
		hostIP = v.HostIP
	}
	row := s.pool.QueryRow(ctx, `
		INSERT INTO sso_environments (app_id, env_code, env_name, base_url, host_ip, base_path, "ADUser")
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING env_id, app_id, env_code, env_name, base_url,
		          COALESCE(host(host_ip),''), COALESCE(base_path,''),
		          COALESCE("ADUser",''),
		          is_active, created_at, updated_at`,
		v.AppID, v.EnvCode, v.EnvName, v.BaseURL, hostIP, nullStr(v.BasePath), nullStr(v.ADUser))
	if e := row.Scan(&v.ID, &v.AppID, &v.EnvCode, &v.EnvName,
		&v.BaseURL, &v.HostIP, &v.BasePath, &v.ADUser,
		&v.Active, &v.CreatedAt, &v.UpdatedAt); e != nil {
		if isUniqueViolation(e) {
			return v, ErrConflict
		}
		return v, e
	}
	return v, nil
}

func (s *SQLStore) UpdateEnvironment(ctx context.Context, v model.Environment) (model.Environment, error) {
	hostIP := interface{}(nil)
	if v.HostIP != "" {
		hostIP = v.HostIP
	}
	row := s.pool.QueryRow(ctx, `
		UPDATE sso_environments
		SET env_code = $2, env_name = $3, base_url = $4, host_ip = $5,
		    base_path = $6, is_active = $7, "ADUser" = $8, updated_at = NOW()
		WHERE env_id = $1
		RETURNING env_id, app_id, env_code, env_name, base_url,
		          COALESCE(host(host_ip),''), COALESCE(base_path,''),
		          COALESCE("ADUser",''),
		          is_active, created_at, updated_at`,
		v.ID, v.EnvCode, v.EnvName, v.BaseURL, hostIP, nullStr(v.BasePath), v.Active, nullStr(v.ADUser))
	if e := row.Scan(&v.ID, &v.AppID, &v.EnvCode, &v.EnvName,
		&v.BaseURL, &v.HostIP, &v.BasePath, &v.ADUser,
		&v.Active, &v.CreatedAt, &v.UpdatedAt); e != nil {
		if errors.Is(e, pgx.ErrNoRows) {
			return v, ErrNotFound
		}
		return v, e
	}
	return v, nil
}

func (s *SQLStore) DeleteEnvironment(ctx context.Context, id int64) error {
	tag, e := s.pool.Exec(ctx, `DELETE FROM sso_environments WHERE env_id = $1`, id)
	if e != nil {
		return e
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ---------- Allowed IP CRUD ----------

func (s *SQLStore) ListAllowedIPs(ctx context.Context, envID int64) ([]model.AllowedIP, error) {
	rows, e := s.pool.Query(ctx, `
		SELECT allowed_ip_id, env_id, host(ip_cidr) || COALESCE('/' || masklen(ip_cidr)::text,''),
		       COALESCE(description,''), is_active, created_at, COALESCE(created_by,'')
		FROM sso_allowed_ips
		WHERE ($1 = 0 OR env_id = $1)
		ORDER BY env_id, ip_cidr`, envID)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := make([]model.AllowedIP, 0)
	for rows.Next() {
		var v model.AllowedIP
		if e = rows.Scan(&v.ID, &v.EnvID, &v.IPCIDR, &v.Description, &v.Active, &v.CreatedAt, &v.CreatedBy); e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *SQLStore) CreateAllowedIP(ctx context.Context, v model.AllowedIP) (model.AllowedIP, error) {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO sso_allowed_ips (env_id, ip_cidr, description, created_by)
		VALUES ($1, $2::inet, $3, $4)
		RETURNING allowed_ip_id, env_id,
		          host(ip_cidr) || COALESCE('/' || masklen(ip_cidr)::text,''),
		          COALESCE(description,''), is_active, created_at, COALESCE(created_by,'')`,
		v.EnvID, v.IPCIDR, nullStr(v.Description), nullStr(v.CreatedBy))
	if e := row.Scan(&v.ID, &v.EnvID, &v.IPCIDR, &v.Description, &v.Active, &v.CreatedAt, &v.CreatedBy); e != nil {
		if isUniqueViolation(e) {
			return v, ErrConflict
		}
		return v, e
	}
	return v, nil
}

func (s *SQLStore) DeleteAllowedIP(ctx context.Context, id int64) error {
	tag, e := s.pool.Exec(ctx, `DELETE FROM sso_allowed_ips WHERE allowed_ip_id = $1`, id)
	if e != nil {
		return e
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ---------- Allowed User CRUD ----------

func (s *SQLStore) ListAllowedUsers(ctx context.Context, envID int64) ([]model.AllowedUser, error) {
	rows, e := s.pool.Query(ctx, `
		SELECT allowed_user_id, env_id, ad_username, COALESCE(employee_id,''), COALESCE(display_name,''),
		       COALESCE(email,''), COALESCE(department,''), is_active, granted_at, COALESCE(granted_by,''),
		       expires_at, last_sync_at
		FROM sso_allowed_users
		WHERE ($1 = 0 OR env_id = $1)
		ORDER BY env_id, ad_username`, envID)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := make([]model.AllowedUser, 0)
	for rows.Next() {
		var v model.AllowedUser
		if e = rows.Scan(&v.ID, &v.EnvID, &v.ADUsername, &v.EmployeeID, &v.DisplayName,
			&v.Email, &v.Department, &v.Active, &v.GrantedAt, &v.GrantedBy,
			&v.ExpiresAt, &v.LastSyncAt); e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *SQLStore) ListAllowedUsersByADUsername(ctx context.Context, adUsername string) ([]model.AllowedUser, error) {
	rows, e := s.pool.Query(ctx, `
		SELECT allowed_user_id, env_id, ad_username, COALESCE(employee_id,''), COALESCE(display_name,''),
		       COALESCE(email,''), COALESCE(department,''), is_active, granted_at, COALESCE(granted_by,''),
		       expires_at, last_sync_at
		FROM sso_allowed_users
		WHERE ad_username = $1
		ORDER BY env_id`, adUsername)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := make([]model.AllowedUser, 0)
	for rows.Next() {
		var v model.AllowedUser
		if e = rows.Scan(&v.ID, &v.EnvID, &v.ADUsername, &v.EmployeeID, &v.DisplayName,
			&v.Email, &v.Department, &v.Active, &v.GrantedAt, &v.GrantedBy,
			&v.ExpiresAt, &v.LastSyncAt); e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *SQLStore) CreateAllowedUser(ctx context.Context, v model.AllowedUser) (model.AllowedUser, error) {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO sso_allowed_users
		    (env_id, ad_username, employee_id, display_name, email, department, granted_by, expires_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING allowed_user_id, env_id, ad_username, COALESCE(employee_id,''), COALESCE(display_name,''),
		          COALESCE(email,''), COALESCE(department,''), is_active, granted_at, COALESCE(granted_by,''),
		          expires_at, last_sync_at`,
		v.EnvID, v.ADUsername, nullStr(v.EmployeeID), nullStr(v.DisplayName),
		nullStr(v.Email), nullStr(v.Department), nullStr(v.GrantedBy), v.ExpiresAt)
	if e := row.Scan(&v.ID, &v.EnvID, &v.ADUsername, &v.EmployeeID, &v.DisplayName,
		&v.Email, &v.Department, &v.Active, &v.GrantedAt, &v.GrantedBy,
		&v.ExpiresAt, &v.LastSyncAt); e != nil {
		if isUniqueViolation(e) {
			return v, ErrConflict
		}
		return v, e
	}
	return v, nil
}

func (s *SQLStore) DeleteAllowedUser(ctx context.Context, id int64) error {
	tag, e := s.pool.Exec(ctx, `DELETE FROM sso_allowed_users WHERE allowed_user_id = $1`, id)
	if e != nil {
		return e
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ---------- Audit ----------

func (s *SQLStore) ListAudit(ctx context.Context, limit int) ([]model.LoginAudit, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, e := s.pool.Query(ctx, `
		SELECT audit_id, COALESCE(app_code,''), COALESCE(env_code,''), COALESCE(base_url,''),
		       COALESCE(host(client_ip),''), COALESCE(ad_username,''), COALESCE(employee_id,''),
		       COALESCE(display_name,''), result, COALESCE(deny_reason,''), COALESCE(user_agent,''),
		       COALESCE(request_id,''), created_at
		FROM sso_login_audit
		ORDER BY created_at DESC
		LIMIT $1`, limit)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := make([]model.LoginAudit, 0, limit)
	for rows.Next() {
		var v model.LoginAudit
		if e = rows.Scan(&v.ID, &v.AppCode, &v.EnvCode, &v.BaseURL, &v.ClientIP,
			&v.ADUsername, &v.EmployeeID, &v.DisplayName, &v.Result, &v.DenyReason,
			&v.UserAgent, &v.RequestID, &v.CreatedAt); e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// ---------- Auth: envs ที่ user ดูแล ----------

// ListAccessibleEnvsByADUser คืน env ทั้งหมดที่ user เป็นเจ้าของ
// (กรองจาก sso_environments.ADUser = <username> และ active)
func (s *SQLStore) ListAccessibleEnvsByADUser(ctx context.Context, adUsername string) ([]model.AccessibleEnv, error) {
	rows, e := s.pool.Query(ctx, `
		SELECT e.env_id, a.app_code, e.env_code, e.env_name,
		       e.base_url, COALESCE(host(e.host_ip),''), COALESCE(e.base_path,''),
		       e.is_active
		FROM sso_environments e
		JOIN sso_applications a ON a.app_id = e.app_id
		WHERE LOWER(COALESCE(e."ADUser",'')) = LOWER($1)
		  AND e.is_active = TRUE
		  AND a.is_active = TRUE
		ORDER BY a.app_code, e.env_code`, adUsername)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := make([]model.AccessibleEnv, 0)
	for rows.Next() {
		var v model.AccessibleEnv
		if e = rows.Scan(&v.ID, &v.AppCode, &v.EnvCode, &v.EnvName,
			&v.BaseURL, &v.HostIP, &v.BasePath, &v.Active); e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// CountAccessibleEnvsByADUser นับจำนวน env ที่ user ดูแล
func (s *SQLStore) CountAccessibleEnvsByADUser(ctx context.Context, adUsername string) (int, error) {
	var n int
	row := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM sso_environments
		WHERE LOWER(COALESCE("ADUser",'')) = LOWER($1)
		  AND is_active = TRUE`, adUsername)
	err := row.Scan(&n)
	return n, err
}

// IsAdminUser เช็คว่า user เป็น admin หรือไม่
// ใช้ตาราง sso_admins เท่านั้น — ไม่ปนกับการเช็ค env ownership
// (admin = ผู้ดูแลระบบ, เห็น env ทั้งหมด, จัดการทุกอย่างได้)
func (s *SQLStore) IsAdminUser(ctx context.Context, adUsername string) (bool, error) {
	var n int
	row := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM sso_admins
		WHERE LOWER(ad_username) = LOWER($1)
		  AND is_active = TRUE`, adUsername)
	if err := row.Scan(&n); err != nil {
		return false, err
	}
	return n > 0, nil
}

// ListAdmins คืนรายชื่อ admin ทั้งหมด
func (s *SQLStore) ListAdmins(ctx context.Context) ([]model.Admin, error) {
	rows, e := s.pool.Query(ctx, `
		SELECT admin_id, ad_username, COALESCE(display_name,''),
		       COALESCE(note,''), is_active, created_at, updated_at
		FROM sso_admins
		ORDER BY ad_username`)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := make([]model.Admin, 0)
	for rows.Next() {
		var a model.Admin
		if e = rows.Scan(&a.ID, &a.ADUsername, &a.DisplayName, &a.Note,
			&a.Active, &a.CreatedAt, &a.UpdatedAt); e != nil {
			return nil, e
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// AddAdmin เพิ่ม admin ใหม่ (ถ้ามีอยู่แล้วคืน id เดิม)
func (s *SQLStore) AddAdmin(ctx context.Context, a model.Admin) (int64, error) {
	if len(a.ADUsername) == 0 || len(a.ADUsername) > 15 {
		return 0, errors.New("ad_username must be 1-15 chars")
	}
	var id int64
	row := s.pool.QueryRow(ctx, `
		INSERT INTO sso_admins (ad_username, display_name, note, is_active)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (ad_username) DO UPDATE
			SET display_name = EXCLUDED.display_name,
			    note         = EXCLUDED.note,
			    is_active    = EXCLUDED.is_active,
			    updated_at   = NOW()
		RETURNING admin_id`,
		a.ADUsername, nullStr(a.DisplayName), nullStr(a.Note), a.Active)
	if err := row.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

// RemoveAdmin ลบ admin (soft delete: set is_active=false)
func (s *SQLStore) RemoveAdmin(ctx context.Context, id int64) error {
	tag, e := s.pool.Exec(ctx, `
		UPDATE sso_admins SET is_active = FALSE, updated_at = NOW()
		WHERE admin_id = $1`, id)
	if e != nil {
		return e
	}
	if tag.RowsAffected() == 0 {
		return errors.New("admin not found")
	}
	return nil
}

// ---------- helpers ----------

func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func isUniqueViolation(e error) bool {
	var pgErr *pgconn.PgError
	if errors.As(e, &pgErr) {
		return pgErr.Code == "23505" // unique_violation
	}
	return false
}
