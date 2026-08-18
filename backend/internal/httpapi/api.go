package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"sso-login/backend/internal/config"
	"sso-login/backend/internal/model"
	"sso-login/backend/internal/service"
	"sso-login/backend/internal/store"
)

type API struct {
	s   *service.Service
	cfg config.Config
}

func New(s *service.Service, cfg config.Config) http.Handler {
	a := &API{s: s, cfg: cfg}
	mux := http.NewServeMux()

	// health
	mux.HandleFunc("GET /health", a.health)
	mux.HandleFunc("GET /healthz", a.health)

	// admin auth
	mux.HandleFunc("POST /api/auth/login", a.login)
	mux.HandleFunc("POST /api/auth/logout", a.logout)
	mux.Handle("GET /api/auth/me", a.session(http.HandlerFunc(a.me)))

	// check access
	mux.HandleFunc("POST /api/check-access", a.checkAccess)

	// applications
	mux.HandleFunc("GET /api/apps", a.listApps)
	mux.HandleFunc("POST /api/apps", a.createApp)
	mux.HandleFunc("PUT /api/apps/{id}", a.updateApp)
	mux.HandleFunc("DELETE /api/apps/{id}", a.deleteApp)

	// environments
	mux.HandleFunc("GET /api/envs", a.listEnvs)
	mux.HandleFunc("POST /api/envs", a.createEnv)
	mux.HandleFunc("PUT /api/envs/{id}", a.updateEnv)
	mux.HandleFunc("DELETE /api/envs/{id}", a.deleteEnv)

	// allowed IPs
	mux.HandleFunc("GET /api/allowed-ips", a.listAllowedIPs)
	mux.HandleFunc("POST /api/allowed-ips", a.createAllowedIP)
	mux.HandleFunc("DELETE /api/allowed-ips/{id}", a.deleteAllowedIP)

	// allowed users
	mux.HandleFunc("GET /api/allowed-users", a.listAllowedUsers)
	mux.HandleFunc("POST /api/allowed-users", a.createAllowedUser)
	mux.HandleFunc("DELETE /api/allowed-users/{id}", a.deleteAllowedUser)

	// audit
	mux.HandleFunc("GET /api/audit", a.listAudit)

	return a.recover(a.cors(a.log(mux)))
}

func (a *API) health(w http.ResponseWriter, _ *http.Request) {
	write(w, http.StatusOK, map[string]any{"status": "ok", "service": "sso-login"})
}

// ----- admin auth -----

func (a *API) login(w http.ResponseWriter, r *http.Request) {
	var in model.LoginRequest
	if !decode(w, r, &in) {
		return
	}
	res, err := a.s.Login(r.Context(), in.Username, in.Password, clientIP(r))
	if err != nil {
		write(w, http.StatusUnauthorized, map[string]any{"error": err.Error()})
		return
	}
	write(w, http.StatusOK, map[string]any{
		"token":       res.Token,
		"username":    res.Username,
		"displayName": res.DisplayName,
		"expiresAt":   res.ExpiresAt,
	})
}

func (a *API) logout(w http.ResponseWriter, r *http.Request) {
	token := tokenFromRequest(r)
	if token == "" {
		write(w, http.StatusOK, map[string]any{"ok": true})
		return
	}
	if e := a.s.Logout(r.Context(), token); e != nil {
		serverError(w, e)
		return
	}
	write(w, http.StatusOK, map[string]any{"ok": true})
}

func (a *API) me(w http.ResponseWriter, r *http.Request) {
	uname, _ := r.Context().Value(ctxKeyUser).(string)
	write(w, http.StatusOK, map[string]any{"username": uname})
}

type ctxKey string

const ctxKeyUser ctxKey = "user"

// session middleware — ตรวจ token จาก header Authorization: Bearer <token>
// หรือ query ?token=<token>
func (a *API) session(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := tokenFromRequest(r)
		if token == "" {
			write(w, http.StatusUnauthorized, map[string]any{"error": "missing token"})
			return
		}
		uname, err := a.s.VerifyToken(r.Context(), token)
		if err != nil {
			write(w, http.StatusUnauthorized, map[string]any{"error": "invalid or expired session"})
			return
		}
		ctx := context.WithValue(r.Context(), ctxKeyUser, uname)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func tokenFromRequest(r *http.Request) string {
	// 1) Authorization: Bearer <token>
	if h := r.Header.Get("Authorization"); h != "" {
		parts := strings.SplitN(h, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return strings.TrimSpace(parts[1])
		}
	}
	// 2) X-Auth-Token header
	if h := r.Header.Get("X-Auth-Token"); h != "" {
		return strings.TrimSpace(h)
	}
	// 3) query string
	return strings.TrimSpace(r.URL.Query().Get("token"))
}

// ----- check access -----

func (a *API) checkAccess(w http.ResponseWriter, r *http.Request) {
	var in model.CheckAccessRequest
	if !decode(w, r, &in) {
		return
	}
	// เติม client IP จาก header ถ้าไม่ได้ส่งมาใน body
	if strings.TrimSpace(in.ClientIP) == "" {
		in.ClientIP = clientIP(r)
	}
	if strings.TrimSpace(in.UserAgent) == "" {
		in.UserAgent = r.UserAgent()
	}
	res, err := a.s.CheckAccess(r.Context(), in)
	if err != nil {
		write(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	status := http.StatusOK
	if !res.Allow {
		status = http.StatusForbidden
	}
	write(w, status, res)
}

// ----- apps -----

func (a *API) listApps(w http.ResponseWriter, r *http.Request) {
	items, e := a.s.ListApps(r.Context())
	if e != nil {
		serverError(w, e)
		return
	}
	write(w, http.StatusOK, map[string]any{"apps": items})
}

func (a *API) createApp(w http.ResponseWriter, r *http.Request) {
	var in model.Application
	if !decode(w, r, &in) {
		return
	}
	out, e := a.s.CreateApp(r.Context(), in)
	if e != nil {
		if errors.Is(e, store.ErrConflict) {
			write(w, http.StatusConflict, map[string]any{"error": "app_code already exists"})
			return
		}
		write(w, http.StatusBadRequest, map[string]any{"error": e.Error()})
		return
	}
	write(w, http.StatusCreated, map[string]any{"app": out})
}

func (a *API) updateApp(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	var in model.Application
	if !decode(w, r, &in) {
		return
	}
	in.ID = id
	out, e := a.s.UpdateApp(r.Context(), in)
	if e != nil {
		if errors.Is(e, store.ErrNotFound) {
			write(w, http.StatusNotFound, map[string]any{"error": "app not found"})
			return
		}
		write(w, http.StatusBadRequest, map[string]any{"error": e.Error()})
		return
	}
	write(w, http.StatusOK, map[string]any{"app": out})
}

func (a *API) deleteApp(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	if e := a.s.DeleteApp(r.Context(), id); e != nil {
		if errors.Is(e, store.ErrNotFound) {
			write(w, http.StatusNotFound, map[string]any{"error": "app not found"})
			return
		}
		serverError(w, e)
	}
}

// ----- envs -----

func (a *API) listEnvs(w http.ResponseWriter, r *http.Request) {
	appID, _ := strconv.ParseInt(r.URL.Query().Get("appId"), 10, 64)
	items, e := a.s.ListEnvs(r.Context(), appID)
	if e != nil {
		serverError(w, e)
		return
	}
	write(w, http.StatusOK, map[string]any{"envs": items})
}

func (a *API) createEnv(w http.ResponseWriter, r *http.Request) {
	var in model.Environment
	if !decode(w, r, &in) {
		return
	}
	out, e := a.s.CreateEnv(r.Context(), in)
	if e != nil {
		if errors.Is(e, store.ErrConflict) {
			write(w, http.StatusConflict, map[string]any{"error": "envCode already exists for this app"})
			return
		}
		write(w, http.StatusBadRequest, map[string]any{"error": e.Error()})
		return
	}
	write(w, http.StatusCreated, map[string]any{"env": out})
}

func (a *API) updateEnv(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	var in model.Environment
	if !decode(w, r, &in) {
		return
	}
	in.ID = id
	out, e := a.s.UpdateEnv(r.Context(), in)
	if e != nil {
		if errors.Is(e, store.ErrNotFound) {
			write(w, http.StatusNotFound, map[string]any{"error": "env not found"})
			return
		}
		write(w, http.StatusBadRequest, map[string]any{"error": e.Error()})
		return
	}
	write(w, http.StatusOK, map[string]any{"env": out})
}

func (a *API) deleteEnv(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	if e := a.s.DeleteEnv(r.Context(), id); e != nil {
		if errors.Is(e, store.ErrNotFound) {
			write(w, http.StatusNotFound, map[string]any{"error": "env not found"})
			return
		}
		serverError(w, e)
	}
}

// ----- allowed IPs -----

func (a *API) listAllowedIPs(w http.ResponseWriter, r *http.Request) {
	envID, _ := strconv.ParseInt(r.URL.Query().Get("envId"), 10, 64)
	items, e := a.s.ListAllowedIPs(r.Context(), envID)
	if e != nil {
		serverError(w, e)
		return
	}
	write(w, http.StatusOK, map[string]any{"allowedIps": items})
}

func (a *API) createAllowedIP(w http.ResponseWriter, r *http.Request) {
	var in model.AllowedIP
	if !decode(w, r, &in) {
		return
	}
	out, e := a.s.CreateAllowedIP(r.Context(), in)
	if e != nil {
		if errors.Is(e, store.ErrConflict) {
			write(w, http.StatusConflict, map[string]any{"error": "ipCidr already exists for this env"})
			return
		}
		write(w, http.StatusBadRequest, map[string]any{"error": e.Error()})
		return
	}
	write(w, http.StatusCreated, map[string]any{"allowedIp": out})
}

func (a *API) deleteAllowedIP(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	if e := a.s.DeleteAllowedIP(r.Context(), id); e != nil {
		if errors.Is(e, store.ErrNotFound) {
			write(w, http.StatusNotFound, map[string]any{"error": "allowed ip not found"})
			return
		}
		serverError(w, e)
	}
}

// ----- allowed users -----

func (a *API) listAllowedUsers(w http.ResponseWriter, r *http.Request) {
	envID, _ := strconv.ParseInt(r.URL.Query().Get("envId"), 10, 64)
	items, e := a.s.ListAllowedUsers(r.Context(), envID)
	if e != nil {
		serverError(w, e)
		return
	}
	write(w, http.StatusOK, map[string]any{"allowedUsers": items})
}

func (a *API) createAllowedUser(w http.ResponseWriter, r *http.Request) {
	var in model.AllowedUser
	if !decode(w, r, &in) {
		return
	}
	out, e := a.s.CreateAllowedUser(r.Context(), in)
	if e != nil {
		if errors.Is(e, store.ErrConflict) {
			write(w, http.StatusConflict, map[string]any{"error": "adUsername already exists for this env"})
			return
		}
		write(w, http.StatusBadRequest, map[string]any{"error": e.Error()})
		return
	}
	write(w, http.StatusCreated, map[string]any{"allowedUser": out})
}

func (a *API) deleteAllowedUser(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	if e := a.s.DeleteAllowedUser(r.Context(), id); e != nil {
		if errors.Is(e, store.ErrNotFound) {
			write(w, http.StatusNotFound, map[string]any{"error": "allowed user not found"})
			return
		}
		serverError(w, e)
	}
}

// ----- audit -----

func (a *API) listAudit(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, e := a.s.ListAudit(r.Context(), limit)
	if e != nil {
		serverError(w, e)
		return
	}
	write(w, http.StatusOK, map[string]any{"audit": items})
}

// ===================== middleware & helpers =====================

func write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	if r.Body == nil {
		write(w, http.StatusBadRequest, map[string]any{"error": "missing body"})
		return false
	}
	if e := json.NewDecoder(r.Body).Decode(v); e != nil {
		write(w, http.StatusBadRequest, map[string]any{"error": "invalid json: " + e.Error()})
		return false
	}
	return true
}

func pathID(w http.ResponseWriter, r *http.Request, name string) (int64, bool) {
	raw := r.PathValue(name)
	id, e := strconv.ParseInt(raw, 10, 64)
	if e != nil || id <= 0 {
		write(w, http.StatusBadRequest, map[string]any{"error": "invalid " + name})
		return 0, false
	}
	return id, true
}

func serverError(w http.ResponseWriter, e error) {
	log.Printf("server error: %v", e)
	write(w, http.StatusInternalServerError, map[string]any{"error": e.Error()})
}

// CORS — เปิดกว้างเพื่อให้ frontend เรียกข้าม origin ได้
// ถ้า FRONTEND_ORIGIN=* (default)  จะ allow ทุก origin
// ถ้าระบุ origin มา เช่น http://localhost:3000 จะ echo origin กลับให้
func (a *API) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Allow-Origin
		if a.cfg.FrontendOrigin == "*" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else if origin != "" && strings.Contains(a.cfg.FrontendOrigin, origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}

		// Allow-Credentials (browser spec: ห้ามใช้ร่วมกับ Allow-Origin=* เลยไม่ตั้งตอน *)
		if a.cfg.FrontendOrigin != "*" {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// Allow-Methods
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		// Allow-Headers — ครอบคลุมทั้ง Authorization, X-Auth-Token, X-Requested-With
		w.Header().Set("Access-Control-Allow-Headers",
			"Content-Type, Authorization, X-Auth-Token, X-Requested-With, Accept, Origin")

		// Expose-Headers — ให้ browser อ่าน header ที่ backend ตอบกลับได้
		w.Header().Set("Access-Control-Expose-Headers",
			"Content-Type, Authorization, X-Auth-Token")

		// Cache preflight 24h
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Preflight
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *API) recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic: %v", rec)
				write(w, http.StatusInternalServerError, map[string]any{"error": "internal error"})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (a *API) log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s ip=%s ua=%q dur=%s",
			r.Method, r.URL.Path, clientIP(r), clientIP(r), r.UserAgent(), time.Since(start))
	})
}

// clientIP ดึง IP จาก X-Forwarded-For ก่อน ถ้าไม่มีใช้ RemoteAddr
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	if rip := r.Header.Get("X-Real-IP"); rip != "" {
		return strings.TrimSpace(rip)
	}
	host, _, _ := strings.Cut(r.RemoteAddr, ":")
	return host
}
