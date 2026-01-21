package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/linkyfish/kxl_backend_go/internal/config"
)

type Manager struct {
	Client          *redis.Client
	Prefix          string
	UserTTL         time.Duration
	AdminTTL        time.Duration
	UserCookieName  string
	AdminCookieName string
	CookieSecure    bool
}

type UserSession struct {
	UserID             string `json:"user_id"`
	UserSessionVersion int    `json:"user_session_version"`
}

type AdminSession struct {
	AdminID string `json:"admin_id"`
}

func NewManager(client *redis.Client, cfg *config.Config) *Manager {
	m := &Manager{Client: client}
	if cfg == nil {
		m.Prefix = "kxl_session:"
		m.UserTTL = 7 * 24 * time.Hour
		m.AdminTTL = 2 * time.Hour
		m.UserCookieName = "kxl_user_session"
		m.AdminCookieName = "kxl_admin_session"
		m.CookieSecure = false
		return m
	}
	m.Prefix = cfg.Session.Prefix
	if m.Prefix == "" {
		m.Prefix = "kxl_session:"
	}
	m.UserTTL = time.Duration(cfg.Session.UserTTLSeconds) * time.Second
	if m.UserTTL <= 0 {
		m.UserTTL = 7 * 24 * time.Hour
	}
	m.AdminTTL = time.Duration(cfg.Session.AdminTTLSeconds) * time.Second
	if m.AdminTTL <= 0 {
		m.AdminTTL = 2 * time.Hour
	}
	m.UserCookieName = cfg.Session.UserCookieName
	if m.UserCookieName == "" {
		m.UserCookieName = "kxl_user_session"
	}
	m.AdminCookieName = cfg.Session.AdminCookieName
	if m.AdminCookieName == "" {
		m.AdminCookieName = "kxl_admin_session"
	}
	m.CookieSecure = cfg.Session.CookieSecure
	return m
}

func (m *Manager) CreateUserSession(ctx context.Context, userID string, sessionVersion int) (string, error) {
	sid, err := newSID()
	if err != nil {
		return "", err
	}
	key := m.Prefix + "user:" + sid
	payload, _ := json.Marshal(UserSession{UserID: userID, UserSessionVersion: sessionVersion})
	if err := m.Client.SetEX(ctx, key, payload, m.UserTTL).Err(); err != nil {
		return "", err
	}
	return sid, nil
}

func (m *Manager) DeleteUserSession(ctx context.Context, sid string) error {
	return m.Client.Del(ctx, m.Prefix+"user:"+sid).Err()
}

func (m *Manager) GetUserSession(ctx context.Context, sid string) (*UserSession, error) {
	raw, err := m.Client.Get(ctx, m.Prefix+"user:"+sid).Bytes()
	if err != nil {
		return nil, err
	}
	var s UserSession
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func (m *Manager) CreateAdminSession(ctx context.Context, adminID string) (string, error) {
	sid, err := newSID()
	if err != nil {
		return "", err
	}
	key := m.Prefix + "admin:" + sid
	payload, _ := json.Marshal(AdminSession{AdminID: adminID})
	if err := m.Client.SetEX(ctx, key, payload, m.AdminTTL).Err(); err != nil {
		return "", err
	}
	return sid, nil
}

func (m *Manager) DeleteAdminSession(ctx context.Context, sid string) error {
	return m.Client.Del(ctx, m.Prefix+"admin:"+sid).Err()
}

func (m *Manager) GetAdminSession(ctx context.Context, sid string) (*AdminSession, error) {
	raw, err := m.Client.Get(ctx, m.Prefix+"admin:"+sid).Bytes()
	if err != nil {
		return nil, err
	}
	var s AdminSession
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func newSID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (m *Manager) CookieInfo() string {
	return fmt.Sprintf("user_cookie=%s admin_cookie=%s secure=%v", m.UserCookieName, m.AdminCookieName, m.CookieSecure)
}

