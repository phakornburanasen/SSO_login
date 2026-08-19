package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"sso-login/backend/internal/config"
)

// Employee — ข้อมูลที่ใช้แสดงใน navbar
type Employee struct {
	EmpID        string `json:"empId"`
	FirstName    string `json:"form_first_name"`
	LastName     string `json:"form_last_name"`
	FullName     string `json:"fullName"` // "firstName lastName" (คำนวณให้)
	Title        string `json:"form_tittle,omitempty"`
	Department   string `json:"form_department,omitempty"`
	Position     string `json:"form_position,omitempty"`
	Email        string `json:"form_email,omitempty"`
	ProfileImage string `json:"profile_image,omitempty"`
	RawJSON      string `json:"rawJson,omitempty"` // debug
}

// EmployeeAPIClient — เรียก external API + cache ในหน่วยความจำ
//
// URL pattern: {base}/_survey_employee.php?Action=GetEmployeeData&emp_id=<id>&token=<base64(emp_id)>
type EmployeeAPIClient struct {
	baseURL string
	timeout time.Duration
	ttl     time.Duration
	http    *http.Client

	mu    sync.RWMutex
	cache map[string]cacheEntry
}

type cacheEntry struct {
	data    Employee
	expires time.Time
}

// NewEmployeeAPIClient สร้าง client ใหม่
func NewEmployeeAPIClient(cfg config.Config) *EmployeeAPIClient {
	return &EmployeeAPIClient{
		baseURL: cfg.EmployeeAPIURL,
		timeout: cfg.EmployeeAPITimeout,
		ttl:     cfg.EmployeeCacheTTL,
		http: &http.Client{
			Timeout: cfg.EmployeeAPITimeout,
		},
		cache: make(map[string]cacheEntry),
	}
}

// GetEmployee ดึงข้อมูล employee จาก emp_id (cache TTL = configurable)
// คืน Employee + (fromCache bool) — frontend ใช้บอกว่าข้อมูลสดหรือ cache
func (c *EmployeeAPIClient) GetEmployee(ctx context.Context, empID string) (Employee, bool, error) {
	if empID == "" {
		return Employee{}, false, fmt.Errorf("emp_id is required")
	}

	// 1) เช็ค cache
	if v, ok := c.fromCache(empID); ok {
		return v, true, nil
	}

	// 2) เรียก external API
	emp, err := c.callExternal(ctx, empID)
	if err != nil {
		return Employee{}, false, err
	}

	// 3) เก็บลง cache
	c.toCache(empID, emp)
	return emp, false, nil
}

func (c *EmployeeAPIClient) fromCache(empID string) (Employee, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.cache[empID]
	if !ok {
		return Employee{}, false
	}
	if time.Now().After(v.expires) {
		return Employee{}, false
	}
	return v.data, true
}

func (c *EmployeeAPIClient) toCache(empID string, emp Employee) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[empID] = cacheEntry{
		data:    emp,
		expires: time.Now().Add(c.ttl),
	}
}

// callExternal ยิง GET ไปที่ {base}/_survey_employee.php?Action=GetEmployeeData&emp_id=<id>&token=<base64(id)>
// แล้ว parse JSON response
func (c *EmployeeAPIClient) callExternal(ctx context.Context, empID string) (Employee, error) {
	// สร้าง URL
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return Employee{}, fmt.Errorf("invalid EMPLOYEE_API_URL: %w", err)
	}
	u = u.JoinPath("_survey_employee.php")
	q := u.Query()
	q.Set("Action", "GetEmployeeData")
	q.Set("emp_id", empID)
	// token = base64(emp_id) — ตามตัวอย่างที่ผู้ใช้ให้
	q.Set("token", base64.StdEncoding.EncodeToString([]byte(empID)))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return Employee{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "sso-login/1.0")

	res, err := c.http.Do(req)
	if err != nil {
		return Employee{}, fmt.Errorf("employee api request failed: %w", err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		return Employee{}, fmt.Errorf("employee api returned %d: %s", res.StatusCode, string(body))
	}

	// response อาจเป็น array [...] หรือ object {...} — handle ทั้งสอง
	raw := string(body)
	emp, err := parseEmployeeResponse([]byte(raw))
	if err != nil {
		return Employee{}, fmt.Errorf("parse employee api: %w (raw: %s)", err, raw)
	}
	emp.EmpID = empID
	if emp.FullName == "" {
		emp.FullName = strings.TrimSpace(emp.FirstName + " " + emp.LastName)
	}
	emp.RawJSON = raw
	return emp, nil
}

// parseEmployeeResponse รองรับทั้ง 2 รูปแบบ:
//
//	[{"form_first_name":"สมชาย","form_last_name":"ใจดี",...}]
//	{"form_first_name":"สมชาย","form_last_name":"ใจดี",...}
//	{"data":[{...}]}  (บาง API ห่อไว้)
func parseEmployeeResponse(body []byte) (Employee, error) {
	body = []byte(strings.TrimSpace(string(body)))
	if len(body) == 0 {
		return Employee{}, fmt.Errorf("empty body")
	}
	// ลอง array ก่อน
	if body[0] == '[' {
		var arr []map[string]any
		if err := json.Unmarshal(body, &arr); err != nil {
			return Employee{}, err
		}
		if len(arr) == 0 {
			return Employee{}, fmt.Errorf("empty array")
		}
		return mapToEmployee(arr[0]), nil
	}
	// object
	var obj map[string]any
	if err := json.Unmarshal(body, &obj); err != nil {
		return Employee{}, err
	}
	// ถ้ามี key "data" ที่เป็น array → ห่อ
	if d, ok := obj["data"]; ok {
		if arr, ok := d.([]any); ok && len(arr) > 0 {
			if m, ok := arr[0].(map[string]any); ok {
				return mapToEmployee(m), nil
			}
		}
	}
	return mapToEmployee(obj), nil
}

func mapToEmployee(m map[string]any) Employee {
	get := func(k string) string {
		if v, ok := m[k]; ok {
			if s, ok := v.(string); ok {
				return s
			}
			return fmt.Sprintf("%v", v)
		}
		return ""
	}
	return Employee{
		FirstName:    get("form_first_name"),
		LastName:     get("form_last_name"),
		Title:        get("form_tittle"), // typo ตามตัวอย่าง API
		Department:   get("form_department"),
		Position:     get("form_position"),
		Email:        get("form_email"),
		ProfileImage: get("profile_image"),
		FullName:     strings.TrimSpace(get("form_first_name") + " " + get("form_last_name")),
	}
}

// InvalidateCache ลบ cache 1 entry (เผื่อต้องการ force refresh)
func (c *EmployeeAPIClient) InvalidateCache(empID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cache, empID)
}
