package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr           string
	FrontendOrigin     string
	DatabaseURL        string
	DBMaxOpen          int
	DBMaxLifetime      time.Duration
	JWTSecret          string
	JWTTTLMinutes      int
	EmployeeAPIURL     string        // base URL ของ employee API
	EmployeeAPITimeout time.Duration // timeout ต่อ request
	EmployeeCacheTTL   time.Duration // TTL ของ cache (กันยิงบ่อย)
}

func Load() (Config, error) {
	loadEnv(".env")
	maxOpen, e1 := strconv.Atoi(env("DB_MAX_OPEN", "20"))
	lifeMin, e2 := strconv.Atoi(env("DB_MAX_LIFETIME_MIN", "30"))
	jwtTTL, e3 := strconv.Atoi(env("JWT_TTL_MINUTES", "480"))
	empTimeoutSec, e4 := strconv.Atoi(env("EMPLOYEE_API_TIMEOUT_SEC", "5"))
	empCacheMin, e5 := strconv.Atoi(env("EMPLOYEE_API_CACHE_MIN", "60"))
	if e1 != nil || e2 != nil || e3 != nil || e4 != nil || e5 != nil ||
		maxOpen <= 0 || lifeMin <= 0 || jwtTTL <= 0 ||
		empTimeoutSec <= 0 || empCacheMin <= 0 {
		return Config{}, errors.New("invalid DB_MAX_OPEN, DB_MAX_LIFETIME_MIN, JWT_TTL_MINUTES, EMPLOYEE_API_TIMEOUT_SEC, or EMPLOYEE_API_CACHE_MIN")
	}
	c := Config{
		HTTPAddr:           env("HTTP_ADDR", ":12080"),
		FrontendOrigin:     env("FRONTEND_ORIGIN", "*"),
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		DBMaxOpen:          maxOpen,
		DBMaxLifetime:      time.Duration(lifeMin) * time.Minute,
		JWTSecret:          env("JWT_SECRET", "dev-only-change-me-please-32-bytes-min!!"),
		JWTTTLMinutes:      jwtTTL,
		EmployeeAPIURL:     env("EMPLOYEE_API_URL", "http://10.0.32.202:3030/api_local"),
		EmployeeAPITimeout: time.Duration(empTimeoutSec) * time.Second,
		EmployeeCacheTTL:   time.Duration(empCacheMin) * time.Minute,
	}
	var missing []string
	for k, v := range map[string]string{"DATABASE_URL": c.DatabaseURL} {
		if strings.TrimSpace(v) == "" {
			missing = append(missing, k)
		}
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}
	return c, nil
}

func env(k, d string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return d
}

func loadEnv(path string) {
	f, e := os.Open(path)
	if e != nil {
		return
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if ok {
			k = strings.TrimSpace(k)
			v = strings.Trim(strings.TrimSpace(v), `"'`)
			if _, exists := os.LookupEnv(k); !exists {
				_ = os.Setenv(k, v)
			}
		}
	}
}
