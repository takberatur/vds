package middleware

import (
	"encoding/json"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
)

const settingsScopeLocalKey = "settings_scope"

var (
	settingsScopeMapOnce sync.Once
	settingsScopeMap     map[string]string
)

func SettingsScopeMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Locals(settingsScopeLocalKey) == nil {
			scope := ResolveSettingsScope(c)
			if scope != "" {
				c.Locals(settingsScopeLocalKey, scope)
			}
		}
		return c.Next()
	}
}

func GetSettingsScope(c *fiber.Ctx) string {
	v := c.Locals(settingsScopeLocalKey)
	if v == nil {
		return "default"
	}
	if s, ok := v.(string); ok {
		if s == "" {
			return "default"
		}
		return s
	}
	return "default"
}

func ResolveSettingsScope(c *fiber.Ctx) string {
	loadSettingsScopeMap()

	domain := resolveClientDomain(c)
	if domain == "" {
		return "default"
	}

	if len(settingsScopeMap) == 0 {
		return "default"
	}

	if scope, ok := settingsScopeMap[domain]; ok {
		return scope
	}

	for k, scope := range settingsScopeMap {
		if k == "" {
			continue
		}
		if strings.HasPrefix(k, "*.") {
			suffix := strings.TrimPrefix(k, "*.")
			if suffix != "" && (domain == suffix || strings.HasSuffix(domain, "."+suffix)) {
				return scope
			}
			continue
		}
		if strings.HasPrefix(k, ".") {
			suffix := strings.TrimPrefix(k, ".")
			if suffix != "" && (domain == suffix || strings.HasSuffix(domain, "."+suffix)) {
				return scope
			}
			continue
		}
	}

	return "default"
}

func loadSettingsScopeMap() {
	settingsScopeMapOnce.Do(func() {
		raw := strings.TrimSpace(os.Getenv("SETTINGS_SCOPE_MAP"))
		if raw == "" {
			settingsScopeMap = map[string]string{}
			return
		}

		var m map[string]string
		if err := json.Unmarshal([]byte(raw), &m); err != nil {
			settingsScopeMap = map[string]string{}
			return
		}

		n := make(map[string]string, len(m))
		for k, v := range m {
			key := normalizeDomain(k)
			val := strings.TrimSpace(v)
			if key == "" || val == "" {
				continue
			}
			n[key] = val
		}
		settingsScopeMap = n
	})
}

func resolveClientDomain(c *fiber.Ctx) string {
	if origin := strings.TrimSpace(c.Get("Origin")); origin != "" {
		if u, err := url.Parse(origin); err == nil {
			if d := normalizeDomain(u.Host); d != "" {
				return d
			}
		}
	}

	if referer := strings.TrimSpace(c.Get("Referer")); referer != "" {
		if u, err := url.Parse(referer); err == nil {
			if d := normalizeDomain(u.Host); d != "" {
				return d
			}
		}
	}

	if xfHost := strings.TrimSpace(c.Get("X-Forwarded-Host")); xfHost != "" {
		if d := normalizeDomain(xfHost); d != "" {
			return d
		}
	}

	if host := strings.TrimSpace(c.Get("Host")); host != "" {
		if d := normalizeDomain(host); d != "" {
			return d
		}
	}

	return ""
}

func normalizeDomain(host string) string {
	host = strings.TrimSpace(host)
	host = strings.Trim(host, "\"'")
	host = strings.ToLower(host)
	if host == "" {
		return ""
	}
	if h, _, ok := strings.Cut(host, ":"); ok {
		host = h
	}

	host = strings.TrimPrefix(host, "www.")

	return host
}
