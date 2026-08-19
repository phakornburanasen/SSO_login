package gateway

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"api-gateway-go/internal/config"
)

type Options struct {
	Transport http.RoundTripper
	Logger    *log.Logger
	CORS      CORSOptions
}

type CORSOptions struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	MaxAge           time.Duration
}

type Gateway struct {
	services  *config.Store
	transport http.RoundTripper
	logger    *log.Logger
	cors      CORSOptions
	buffers   httputil.BufferPool
}

func New(services *config.Store, options Options) *Gateway {
	transport := options.Transport
	if transport == nil {
		transport = NewTransport(300 * time.Second)
	}

	logger := options.Logger
	if logger == nil {
		logger = log.Default()
	}

	cors := options.CORS
	if len(cors.AllowedOrigins) == 0 {
		cors.AllowedOrigins = []string{"*"}
	}
	if len(cors.AllowedMethods) == 0 {
		cors.AllowedMethods = []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodHead,
			http.MethodOptions,
		}
	}
	if len(cors.AllowedHeaders) == 0 {
		cors.AllowedHeaders = []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"Origin",
			"X-Requested-With",
		}
	}
	if cors.MaxAge == 0 {
		cors.MaxAge = 24 * time.Hour
	}

	return &Gateway{
		services:  services,
		transport: transport,
		logger:    logger,
		cors:      cors,
		buffers:   newBufferPool(32 * 1024),
	}
}

func NewTransport(responseHeaderTimeout time.Duration) http.RoundTripper {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          4096,
		MaxIdleConnsPerHost:   512,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: time.Second,
		ResponseHeaderTimeout: responseHeaderTimeout,
	}
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.applyCORS(w, r)

	if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	switch {
	case r.URL.Path == "/api/health":
		g.handleHealth(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/"):
		g.handleProxy(w, r)
	case r.URL.Path == "/openapi.json" || strings.HasPrefix(r.URL.Path, "/docs") || strings.HasPrefix(r.URL.Path, "/redoc"):
		g.handleDocsProxy(w, r)
	default:
		writeJSONError(w, http.StatusNotFound, "route not found")
	}
}

func (g *Gateway) handleDocsProxy(w http.ResponseWriter, r *http.Request) {
	referer := r.Header.Get("Referer")
	var serviceName string
	if referer != "" {
		if u, err := url.Parse(referer); err == nil {
			serviceName, _, _ = splitServicePath(u.Path)
		}
	}
	
	if serviceName == "" {
		snapshot, _ := g.services.Snapshot()
		names := snapshot.Names()
		if len(names) > 0 {
			serviceName = names[0]
		}
	}

	if serviceName == "" {
		writeJSONError(w, http.StatusNotFound, "service name is required for docs")
		return
	}

	snapshot, err := g.services.Snapshot()
	if err != nil {
		g.logger.Printf("config refresh failed: %v", err)
	}

	service, found := snapshot.Get(serviceName)
	if !found {
		writeJSONError(w, http.StatusNotFound, fmt.Sprintf("service %q not found", serviceName))
		return
	}

	target, err := parseTarget(service.URL)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, err.Error())
		return
	}

	proxy := &httputil.ReverseProxy{
		Transport:     g.transport,
		FlushInterval: -1,
		BufferPool:    g.buffers,
		Rewrite: func(proxyRequest *httputil.ProxyRequest) {
			proxyRequest.SetURL(target)
			proxyRequest.SetXForwarded()

			out := proxyRequest.Out
			out.URL.Path = joinURLPath(target.Path, r.URL.Path)
			out.URL.RawPath = ""
			out.URL.RawQuery = r.URL.RawQuery
			out.Host = target.Host
		},
		ErrorHandler: func(w http.ResponseWriter, req *http.Request, err error) {
			if errors.Is(err, http.ErrAbortHandler) {
				return
			}
			g.logger.Printf("proxy %s %s -> %s failed: %v", req.Method, req.URL.Path, target.String(), err)
			writeJSONError(w, http.StatusBadGateway, "gateway upstream error")
		},
	}

	proxy.ServeHTTP(w, r)
}

func (g *Gateway) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", http.MethodGet+", "+http.MethodHead+", "+http.MethodOptions)
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	snapshot, err := g.services.Snapshot()
	if err != nil {
		g.logger.Printf("config refresh failed: %v", err)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":   "ok",
		"services": snapshot.Names(),
	})
}

func (g *Gateway) handleProxy(w http.ResponseWriter, r *http.Request) {
	serviceName, endpoint, ok := splitServicePath(r.URL.Path)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "service name is required")
		return
	}

	snapshot, err := g.services.Snapshot()
	if err != nil {
		g.logger.Printf("config refresh failed: %v", err)
	}

	service, found := snapshot.Get(serviceName)
	if !found {
		writeJSONError(w, http.StatusNotFound, fmt.Sprintf("service %q not found", serviceName))
		return
	}

	target, err := parseTarget(service.URL)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, err.Error())
		return
	}

	prefix := "/api/" + serviceName
	proxy := &httputil.ReverseProxy{
		Transport:     g.transport,
		FlushInterval: -1,
		BufferPool:    g.buffers,
		Rewrite: func(proxyRequest *httputil.ProxyRequest) {
			proxyRequest.SetURL(target)
			proxyRequest.SetXForwarded()

			out := proxyRequest.Out
			out.URL.Path = joinURLPath(target.Path, endpoint)
			out.URL.RawPath = ""
			out.URL.RawQuery = joinRawQuery(target.RawQuery, proxyRequest.In.URL.RawQuery)
			out.Host = target.Host
			out.Header.Set("X-Forwarded-Prefix", prefix)
			out.Header.Set("X-Gateway-Service", serviceName)
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			if errors.Is(err, http.ErrAbortHandler) {
				return
			}
			g.logger.Printf("proxy %s %s -> %s failed: %v", r.Method, r.URL.Path, target.String(), err)
			writeJSONError(w, http.StatusBadGateway, "gateway upstream error")
		},
	}

	proxy.ServeHTTP(w, r)
}

func (g *Gateway) applyCORS(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return
	}

	allowedOrigin := g.allowedOrigin(origin)
	if allowedOrigin == "" {
		return
	}

	header := w.Header()
	header.Add("Vary", "Origin")
	header.Set("Access-Control-Allow-Origin", allowedOrigin)
	if g.cors.AllowCredentials {
		header.Set("Access-Control-Allow-Credentials", "true")
	}
	header.Set("Access-Control-Allow-Methods", strings.Join(g.cors.AllowedMethods, ", "))
	header.Set("Access-Control-Expose-Headers", "*")

	requestedHeaders := r.Header.Get("Access-Control-Request-Headers")
	if requestedHeaders != "" {
		header.Set("Access-Control-Allow-Headers", requestedHeaders)
	} else {
		header.Set("Access-Control-Allow-Headers", strings.Join(g.cors.AllowedHeaders, ", "))
	}

	if g.cors.MaxAge > 0 {
		header.Set("Access-Control-Max-Age", fmt.Sprintf("%.0f", g.cors.MaxAge.Seconds()))
	}
}

func (g *Gateway) allowedOrigin(origin string) string {
	for _, allowed := range g.cors.AllowedOrigins {
		allowed = strings.TrimSpace(allowed)
		if allowed == "*" {
			if g.cors.AllowCredentials {
				return origin
			}
			return "*"
		}
		if allowed == origin {
			return origin
		}
	}
	return ""
}

func splitServicePath(path string) (string, string, bool) {
	tail := strings.TrimPrefix(path, "/api/")
	if tail == "" {
		return "", "", false
	}

	serviceName, endpoint, hasEndpoint := strings.Cut(tail, "/")
	serviceName = strings.TrimSpace(serviceName)
	if serviceName == "" {
		return "", "", false
	}
	if !hasEndpoint {
		return serviceName, "", true
	}
	return serviceName, "/" + endpoint, true
}

func parseTarget(rawURL string) (*url.URL, error) {
	target, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid upstream url: %w", err)
	}

	switch target.Scheme {
	case "ws":
		target.Scheme = "http"
	case "wss":
		target.Scheme = "https"
	case "http", "https":
	default:
		return nil, fmt.Errorf("unsupported upstream scheme %q", target.Scheme)
	}

	return target, nil
}

func joinURLPath(basePath, endpoint string) string {
	if basePath == "" {
		if endpoint == "" {
			return "/"
		}
		return endpoint
	}
	if endpoint == "" {
		return basePath
	}

	baseHasSlash := strings.HasSuffix(basePath, "/")
	endpointHasSlash := strings.HasPrefix(endpoint, "/")

	switch {
	case baseHasSlash && endpointHasSlash:
		return basePath + endpoint[1:]
	case !baseHasSlash && !endpointHasSlash:
		return basePath + "/" + endpoint
	default:
		return basePath + endpoint
	}
}

func joinRawQuery(baseQuery, requestQuery string) string {
	switch {
	case baseQuery == "":
		return requestQuery
	case requestQuery == "":
		return baseQuery
	default:
		return baseQuery + "&" + requestQuery
	}
}

func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	if statusCode == http.StatusNoContent {
		return
	}
	_ = json.NewEncoder(w).Encode(payload)
}

type bufferPool struct {
	pool sync.Pool
}

func newBufferPool(size int) *bufferPool {
	buffers := &bufferPool{}
	buffers.pool.New = func() any {
		return make([]byte, size)
	}
	return buffers
}

func (p *bufferPool) Get() []byte {
	return p.pool.Get().([]byte)
}

func (p *bufferPool) Put(buffer []byte) {
	if buffer == nil {
		return
	}
	p.pool.Put(buffer)
}
