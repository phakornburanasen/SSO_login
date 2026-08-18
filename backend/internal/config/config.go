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
	HTTPAddr        string
	FrontendOrigin  string
	DatabaseURL     string
	DBMaxOpen       int
	DBMaxLifetime   time.Duration
	JWTSecret       string
	JWTTTLMinutes   int
}

func Load() (Config, error) {
	loadEnv(".env")
	maxOpen, e1 := strconv.Atoi(env("DB_MAX_OPEN", "20"))
	lifeMin, e2 := strconv.Atoi(env("DB_MAX_LIFETIME_MIN", "30"))
	jwtTTL, e3 := strconv.Atoi(env("JWT_TTL_MINUTES", "480"))
	if e1 != nil || e2 != nil || e3 != nil || maxOpen <= 0 || lifeMin <= 0 || jwtTTL <= 0 {
		return Config{}, errors.New("invalid DB_MAX_OPEN, DB_MAX_LIFETIME_MIN, or JWT_TTL_MINUTES")
	}
	c := Config{
		HTTPAddr:       env("HTTP_ADDR", ":12010"),
		FrontendOrigin: env("FRONTEND_ORIGIN", "*"),
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		DBMaxOpen:      maxOpen,
		DBMaxLifetime:  time.Duration(lifeMin) * time.Minute,
		JWTSecret:      env("JWT_SECRET", "dev-only-change-me-please-32-bytes-min!!"),
		JWTTTLMinutes:  jwtTTL,
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
