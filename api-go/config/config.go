package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// InsecureDefaultSecret es el secreto de ejemplo para desarrollo local.
// Validate() impide arrancar en producción con este valor (fail-fast).
const InsecureDefaultSecret = "supersecret-reto-tecnico"

// Config centraliza TODAS las variables de entorno del servicio Go
// (principio DRY: ningún otro paquete lee os.Getenv directamente).
type Config struct {
	Env               string
	Port              string
	JWTSecret         string
	JWTIssuer         string
	JWTAudience       string
	JWTExpiry         time.Duration
	NodeAPIURL        string // Ej: http://api-node:3002/api/stats
	CORSOrigin        string
	MaxBodyBytes      int
	MaxMatrixElements int // Tope anti-DoS: Gram-Schmidt es O(m·n²)
	EnableDemoToken   bool
	ForwardTimeout    time.Duration
	RateLimitMax      int           // peticiones por ventana (solo /api/*)
	RateLimitWindow   time.Duration // ventana del rate limit
}

// Load lee las env vars con valores por defecto pensados para docker-compose.
func Load() Config {
	// Acepta NODE_ENV (convención Node/Docker) o ENV (convención Go); default local.
	env := getenv("NODE_ENV", getenv("ENV", "development"))
	return Config{
		Env:               env,
		Port:              getenv("PORT", "3001"),
		JWTSecret:         getenv("JWT_SECRET", InsecureDefaultSecret),
		JWTIssuer:         getenv("JWT_ISSUER", "reto-tecnico"),
		JWTAudience:       getenv("JWT_AUDIENCE", "reto-tecnico-apis"),
		JWTExpiry:         getDuration("JWT_EXPIRY", 2*time.Hour),
		NodeAPIURL:        getenv("NODE_API_URL", "http://api-node:3002/api/stats"),
		CORSOrigin:        getenv("CORS_ORIGIN", "*"),
		MaxBodyBytes:      getInt("MAX_BODY_BYTES", 1<<20), // 1 MB
		MaxMatrixElements: getInt("MAX_MATRIX_ELEMENTS", 10000),
		EnableDemoToken:   getBool("ENABLE_DEMO_TOKEN", true),
		ForwardTimeout:    getDuration("FORWARD_TIMEOUT", 10*time.Second),
		RateLimitMax:      getInt("RATE_LIMIT_MAX", 100),
		RateLimitWindow:   getDuration("RATE_LIMIT_WINDOW", time.Minute),
	}
}

// Validate falla rápido si la configuración es insegura para producción:
// mejor no arrancar que arrancar con el secreto de ejemplo.
func Validate(cfg Config) error {
	if strings.EqualFold(cfg.Env, "production") && cfg.JWTSecret == InsecureDefaultSecret {
		return fmt.Errorf("refusing to start in production with the default JWT_SECRET; set a strong JWT_SECRET env var")
	}
	return nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return fallback
}

func getBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "1", "true", "yes":
			return true
		case "0", "false", "no":
			return false
		}
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
	}
	return fallback
}
