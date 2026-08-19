package model

import (
	"database/sql/driver"
	"net/netip"
	"strings"
	"time"
)

// FlexibleTime is a time.Time that accepts multiple JSON input formats:
//   - RFC3339:         "2006-01-02T15:04:05Z07:00"
//   - datetime-local:  "2006-01-02T15:04"  (no seconds, no tz)
//   - date only:       "2006-01-02"
//
// This lets browser <input type="datetime-local"> work without manual formatting.
type FlexibleTime struct {
	time.Time
}

func (t *FlexibleTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		t.Time = time.Time{}
		return nil
	}
	for _, layout := range []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006-01-02",
	} {
		if parsed, err := time.Parse(layout, s); err == nil {
			t.Time = parsed
			return nil
		}
	}
	// fallback: try RFC3339 with local tz
	parsed, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}
	t.Time = parsed
	return nil
}

func (t FlexibleTime) MarshalJSON() ([]byte, error) {
	if t.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + t.Time.Format(time.RFC3339) + `"`), nil
}

// sql.Scanner support — pgx returns time.Time from timestamptz columns.
func (t *FlexibleTime) Scan(value interface{}) error {
	if value == nil {
		t.Time = time.Time{}
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		t.Time = v
		return nil
	default:
		return nil
	}
}

// driver.Valuer support — store as time.Time.
func (t FlexibleTime) Value() (driver.Value, error) {
	if t.Time.IsZero() {
		return nil, nil
	}
	return t.Time, nil
}

// --- Application / Environment ---

type Application struct {
	ID          int64      `json:"id"`
	Code        string     `json:"code"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Active      bool       `json:"active"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty"`
}

type Environment struct {
	ID        int64      `json:"id"`
	AppID     int64      `json:"appId"`
	AppCode   string     `json:"appCode,omitempty"`
	EnvCode   string     `json:"envCode"`
	EnvName   string     `json:"envName"`
	BaseURL   string     `json:"baseUrl"`
	HostIP    string     `json:"hostIp"`
	BasePath  string     `json:"basePath"`
	ADUser    string     `json:"adUser,omitempty"`
	Active    bool       `json:"active"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
}

// --- Allowed IP / Allowed User ---

type AllowedIP struct {
	ID          int64     `json:"id"`
	EnvID       int64     `json:"envId"`
	IPCIDR      string    `json:"ipCidr"`
	Description string    `json:"description"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"createdAt"`
	CreatedBy   string    `json:"createdBy"`
}

type AllowedUser struct {
	ID          int64         `json:"id"`
	EnvID       int64         `json:"envId"`
	ADUsername  string        `json:"adUsername"`
	EmployeeID  string        `json:"employeeId"`
	DisplayName string        `json:"displayName"`
	Email       string        `json:"email"`
	Department  string        `json:"department"`
	Active      bool          `json:"active"`
	GrantedAt   time.Time     `json:"grantedAt"`
	GrantedBy   string        `json:"grantedBy"`
	ExpiresAt   *FlexibleTime `json:"expiresAt,omitempty"`
	LastSyncAt  *time.Time    `json:"lastSyncAt,omitempty"`
}

// --- Access Policy (optional) ---

type AccessPolicy struct {
	ID         int64     `json:"id"`
	EnvID      int64     `json:"envId"`
	PolicyName string    `json:"policyName"`
	IPCIDR     string    `json:"ipCidr,omitempty"`
	ADUsername string    `json:"adUsername,omitempty"`
	Active     bool      `json:"active"`
	CreatedAt  time.Time `json:"createdAt"`
}

// --- Audit & Session ---

type LoginAudit struct {
	ID          int64     `json:"id"`
	AppCode     string    `json:"appCode"`
	EnvCode     string    `json:"envCode"`
	BaseURL     string    `json:"baseUrl"`
	ClientIP    string    `json:"clientIp"`
	ADUsername  string    `json:"adUsername"`
	EmployeeID  string    `json:"employeeId"`
	DisplayName string    `json:"displayName"`
	Result      string    `json:"result"`
	DenyReason  string    `json:"denyReason"`
	UserAgent   string    `json:"userAgent"`
	RequestID   string    `json:"requestId"`
	CreatedAt   time.Time `json:"createdAt"`
}

type Session struct {
	ID          int64      `json:"id"`
	AppID       int64      `json:"appId"`
	EnvID       int64      `json:"envId"`
	ADUsername  string     `json:"adUsername"`
	EmployeeID  string     `json:"employeeId"`
	ClientIP    string     `json:"clientIp"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"createdAt"`
	ExpiresAt   time.Time  `json:"expiresAt"`
	LoggedOutAt *time.Time `json:"loggedOutAt,omitempty"`
}

// --- Auth ---

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AccessibleEnv — ข้อมูลย่อของ env ที่ user มีสิทธิ์จัดการ
// (กรองจาก sso_environments.ADUser = <username>)
type AccessibleEnv struct {
	ID       int64  `json:"id"`
	AppCode  string `json:"appCode"`
	EnvCode  string `json:"envCode"`
	EnvName  string `json:"envName"`
	BaseURL  string `json:"baseUrl"`
	HostIP   string `json:"hostIp"`
	BasePath string `json:"basePath"`
	Active   bool   `json:"active"`
}

// Admin — รายชื่อ admin จากตาราง sso_admins
type Admin struct {
	ID          int64     `json:"id"`
	ADUsername  string    `json:"adUsername"`
	DisplayName string    `json:"displayName"`
	Note        string    `json:"note"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type LoginResult struct {
	Token          string          `json:"token"`
	Username       string          `json:"username"`
	DisplayName    string          `json:"displayName"`
	ExpiresAt      time.Time       `json:"expiresAt"`
	AccessibleEnvs []AccessibleEnv `json:"accessibleEnvs"`
	Role           string          `json:"role"` // "admin" | "user" — admin เห็น env ทั้งหมด
}

// --- Check Access Request / Result ---

type CheckAccessRequest struct {
	BaseURL     string `json:"baseUrl"`
	ClientIP    string `json:"clientIp"`
	ADUsername  string `json:"adUsername"`
	EmployeeID  string `json:"employeeId"`
	DisplayName string `json:"displayName"`
	UserAgent   string `json:"userAgent"`
}

type CheckAccessResult struct {
	Allow      bool   `json:"allow"`
	Result     string `json:"result"` // ALLOW | DENY_APP | DENY_IP | DENY_USER | ERROR
	Reason     string `json:"reason,omitempty"`
	AppCode    string `json:"appCode,omitempty"`
	EnvCode    string `json:"envCode,omitempty"`
	BaseURL    string `json:"baseUrl,omitempty"`
	ClientIP   string `json:"clientIp,omitempty"`
	ADUsername string `json:"adUsername,omitempty"`
}

// --- helpers ---

// ParseClientIP รับ IP string แล้วแปลงเป็น netip.Addr (รองรับ IPv4 / IPv6)
func ParseClientIP(s string) (netip.Addr, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return netip.Addr{}, false
	}
	a, err := netip.ParseAddr(s)
	if err != nil {
		return netip.Addr{}, false
	}
	return a, true
}
