package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	EnvGatewayConfig = "GATEWAY_CONFIG"
	EnvServiceConfig = "SERVICE_CONFIG"

	DefaultReloadInterval = time.Second
)

type Service struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	Enabled bool   `json:"enabled"`
}

type Snapshot struct {
	Source   string             `json:"source"`
	LoadedAt time.Time          `json:"loadedAt"`
	Services map[string]Service `json:"services"`
}

func (s Snapshot) Get(name string) (Service, bool) {
	service, ok := s.Services[name]
	if !ok || !service.Enabled {
		return Service{}, false
	}
	return service, true
}

func (s Snapshot) Names() []string {
	names := make([]string, 0, len(s.Services))
	for name, service := range s.Services {
		if service.Enabled {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

type Store struct {
	path          string
	reloadEvery   time.Duration
	lastCheckTime time.Time
	lastModTime   time.Time
	mu            sync.Mutex
	current       atomic.Value
}

func NewStore(path string, reloadEvery time.Duration) *Store {
	if reloadEvery <= 0 {
		reloadEvery = DefaultReloadInterval
	}

	store := &Store{
		path:        path,
		reloadEvery: reloadEvery,
	}
	store.current.Store(Snapshot{
		Source:   path,
		LoadedAt: time.Now(),
		Services: map[string]Service{},
	})
	return store
}

func (s *Store) Path() string {
	return s.path
}

func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadLocked()
}

func (s *Store) Snapshot() (Snapshot, error) {
	return s.RefreshIfNeeded()
}

func (s *Store) RefreshIfNeeded() (Snapshot, error) {
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.lastCheckTime.IsZero() && now.Sub(s.lastCheckTime) < s.reloadEvery {
		return s.snapshotLocked(), nil
	}
	s.lastCheckTime = now

	info, err := os.Stat(s.path)
	if err != nil {
		return s.snapshotLocked(), err
	}

	if !s.lastModTime.IsZero() && info.ModTime().Equal(s.lastModTime) {
		return s.snapshotLocked(), nil
	}

	if err := s.loadLocked(); err != nil {
		return s.snapshotLocked(), err
	}
	return s.snapshotLocked(), nil
}

func (s *Store) snapshotLocked() Snapshot {
	snapshot, ok := s.current.Load().(Snapshot)
	if !ok {
		return Snapshot{Source: s.path, LoadedAt: time.Now(), Services: map[string]Service{}}
	}
	return snapshot
}

func (s *Store) loadLocked() error {
	info, err := os.Stat(s.path)
	if err != nil {
		return err
	}

	snapshot, err := LoadFile(s.path)
	if err != nil {
		return err
	}

	s.lastModTime = info.ModTime()
	s.lastCheckTime = time.Now()
	s.current.Store(snapshot)
	return nil
}

func ResolvePath(explicit string) (string, error) {
	if explicit != "" {
		return absolutePath(explicit)
	}

	for _, envName := range []string{EnvGatewayConfig, EnvServiceConfig} {
		if value := strings.TrimSpace(os.Getenv(envName)); value != "" {
			return absolutePath(value)
		}
	}

	for _, candidate := range []string{"service.json", filepath.Join("..", "service.json")} {
		if _, err := os.Stat(candidate); err == nil {
			return absolutePath(candidate)
		}
	}

	return absolutePath("service.json")
}

func LoadFile(path string) (Snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Snapshot{}, err
	}

	snapshot, err := Parse(data)
	if err != nil {
		return Snapshot{}, fmt.Errorf("parse %s: %w", path, err)
	}
	snapshot.Source = path
	snapshot.LoadedAt = time.Now()
	return snapshot, nil
}

func Parse(data []byte) (Snapshot, error) {
	var root map[string]json.RawMessage
	if err := json.Unmarshal(data, &root); err != nil {
		return Snapshot{}, err
	}

	var services map[string]Service
	if rawServices, ok := root["services"]; ok {
		parsed, err := parseServices(rawServices)
		if err != nil {
			return Snapshot{}, err
		}
		services = parsed
	} else {
		parsed, err := parseServices(data)
		if err != nil {
			return Snapshot{}, err
		}
		services = parsed
	}

	if len(services) == 0 {
		return Snapshot{}, errors.New("service list is empty")
	}

	return Snapshot{
		LoadedAt: time.Now(),
		Services: services,
	}, nil
}

func parseServices(data []byte) (map[string]Service, error) {
	var rawMap map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMap); err == nil && rawMap != nil {
		return parseServiceMap(rawMap)
	}

	var rawList []json.RawMessage
	if err := json.Unmarshal(data, &rawList); err != nil {
		return nil, errors.New("services must be an object or array")
	}

	services := make(map[string]Service, len(rawList))
	for _, raw := range rawList {
		service, err := parseServiceDefinition("", raw)
		if err != nil {
			return nil, err
		}
		if service.Name == "" {
			return nil, errors.New("service name is required")
		}
		services[service.Name] = service
	}
	return services, nil
}

func parseServiceMap(rawMap map[string]json.RawMessage) (map[string]Service, error) {
	services := make(map[string]Service, len(rawMap))
	for name, raw := range rawMap {
		service, err := parseServiceDefinition(name, raw)
		if err != nil {
			return nil, err
		}
		services[service.Name] = service
	}
	return services, nil
}

func parseServiceDefinition(name string, raw json.RawMessage) (Service, error) {
	name = strings.TrimSpace(name)

	var target string
	if err := json.Unmarshal(raw, &target); err == nil {
		return normalizeService(Service{
			Name:    name,
			URL:     target,
			Enabled: true,
		})
	}

	var definition struct {
		Name    string `json:"name"`
		URL     string `json:"url"`
		BaseURL string `json:"base_url"`
		Target  string `json:"target"`
		Enabled *bool  `json:"enabled"`
	}
	if err := json.Unmarshal(raw, &definition); err != nil {
		return Service{}, err
	}

	if definition.Name != "" {
		name = strings.TrimSpace(definition.Name)
	}

	target = firstNonEmpty(definition.URL, definition.BaseURL, definition.Target)
	enabled := true
	if definition.Enabled != nil {
		enabled = *definition.Enabled
	}

	return normalizeService(Service{
		Name:    name,
		URL:     target,
		Enabled: enabled,
	})
}

func normalizeService(service Service) (Service, error) {
	service.Name = strings.Trim(strings.TrimSpace(service.Name), "/")
	service.URL = strings.TrimSpace(service.URL)

	if service.Name == "" {
		return Service{}, errors.New("service name is required")
	}
	if service.URL == "" {
		return Service{}, fmt.Errorf("service %q url is required", service.Name)
	}

	parsed, err := url.Parse(service.URL)
	if err != nil {
		return Service{}, fmt.Errorf("service %q has invalid url: %w", service.Name, err)
	}
	if parsed.Host == "" {
		return Service{}, fmt.Errorf("service %q url must include a host", service.Name)
	}

	switch parsed.Scheme {
	case "http", "https", "ws", "wss":
	default:
		return Service{}, fmt.Errorf("service %q url scheme must be http, https, ws, or wss", service.Name)
	}

	return service, nil
}

func absolutePath(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return path, err
	}
	return absolute, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
