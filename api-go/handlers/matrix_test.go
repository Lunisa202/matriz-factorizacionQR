package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"reto-tecnico/api-go/config"
	"reto-tecnico/api-go/middleware"
	"reto-tecnico/api-go/services"
)

// testConfig usa secreto propio para no depender de env vars.
func testConfig() config.Config {
	return config.Config{
		Env:               "test",
		JWTSecret:         "test-secret-only-for-tests",
		JWTIssuer:         "test-issuer",
		JWTAudience:       "test-audience",
		JWTExpiry:         time.Hour,
		NodeAPIURL:        "http://node-test/api/stats",
		MaxMatrixElements: 10000,
		EnableDemoToken:   true,
	}
}

// stubQR devuelve Q y R fijas sin calcular nada (doble de test).
func stubQR() services.QRService {
	return services.QRServiceFunc(func(a [][]float64) ([][]float64, [][]float64, error) {
		q := [][]float64{{1, 0}, {0, 1}}
		r := [][]float64{{5, 0}, {0, 3}}
		return q, r, nil
	})
}

// stubForwarder simula a la API de Node sin red (doble de test).
func stubForwarder() services.StatsForwarder {
	return services.StatsForwarderFunc(func(nodeURL, rawToken string, payload interface{}) (interface{}, error) {
		return map[string]interface{}{"max": 5.0, "isDiagonal": true}, nil
	})
}

// firmarToken genera un Bearer válido para el cfg de test.
func firmarToken(t *testing.T, cfg config.Config) string {
	t.Helper()
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": "tester",
		"iss": cfg.JWTIssuer,
		"aud": cfg.JWTAudience,
		"iat": now.Unix(),
		"exp": now.Add(time.Hour).Unix(),
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		t.Fatalf("no se pudo firmar token: %v", err)
	}
	return signed
}

// nuevaApp integra middleware JWT real + handler con dobles (test de integración
// de la capa HTTP de Go; solo Node queda simulado por stubForwarder).
func nuevaApp(cfg config.Config) *fiber.App {
	app := fiber.New()
	protected := app.Group("/api", middleware.JWTMiddleware(cfg))
	protected.Post("/matrix/qr", NewQRHandler(cfg, stubQR(), stubForwarder()))
	return app
}

func hacerPost(t *testing.T, app *fiber.App, body, token string) *http.Response {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/matrix/qr", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test falló: %v", err)
	}
	return resp
}

func leerBody(t *testing.T, resp *http.Response) map[string]interface{} {
	t.Helper()
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var decoded map[string]interface{}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("respuesta no es JSON: %s", string(raw))
	}
	return decoded
}

func TestQRHandlerIntegracionOK(t *testing.T) {
	cfg := testConfig()
	resp := hacerPost(t, nuevaApp(cfg), `{"matrix": [[1,2],[3,4]]}`, firmarToken(t, cfg))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("esperado 200, fue %d", resp.StatusCode)
	}
	body := leerBody(t, resp)
	if _, ok := body["q"]; !ok {
		t.Fatal("respuesta sin campo q")
	}
	if _, ok := body["r"]; !ok {
		t.Fatal("respuesta sin campo r")
	}
	stats, ok := body["stats"].(map[string]interface{})
	if !ok {
		t.Fatal("respuesta sin campo stats (forwarder no integrado)")
	}
	if stats["max"] != 5.0 {
		t.Fatalf("stats.max esperado 5, fue %v", stats["max"])
	}
}

func TestQRHandlerSinToken401(t *testing.T) {
	cfg := testConfig()
	resp := hacerPost(t, nuevaApp(cfg), `{"matrix": [[1,2],[3,4]]}`, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("esperado 401, fue %d", resp.StatusCode)
	}
}

func TestQRHandlerMatrizInvalida400(t *testing.T) {
	cfg := testConfig()
	resp := hacerPost(t, nuevaApp(cfg), `{"matrix": [[1,2],[3]]}`, firmarToken(t, cfg))
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("esperado 400, fue %d", resp.StatusCode)
	}
}

func TestQRHandlerExcesoTamano413(t *testing.T) {
	cfg := testConfig()
	cfg.MaxMatrixElements = 3 // 2×2 = 4 elementos > 3
	resp := hacerPost(t, nuevaApp(cfg), `{"matrix": [[1,2],[3,4]]}`, firmarToken(t, cfg))
	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("esperado 413, fue %d", resp.StatusCode)
	}
}

func TestIssueTokenGuard404(t *testing.T) {
	cfg := testConfig()
	cfg.EnableDemoToken = false
	app := fiber.New()
	app.Post("/api/auth/token", IssueToken(cfg))

	req := httptest.NewRequest(http.MethodPost, "/api/auth/token", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test falló: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("esperado 404 con demo desactivado, fue %d", resp.StatusCode)
	}
}

func TestIssueTokenOK(t *testing.T) {
	cfg := testConfig()
	app := fiber.New()
	app.Post("/api/auth/token", IssueToken(cfg))

	req := httptest.NewRequest(http.MethodPost, "/api/auth/token", strings.NewReader(`{"user":"tester"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test falló: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("esperado 200, fue %d", resp.StatusCode)
	}
	body := leerBody(t, resp)
	if _, ok := body["token"]; !ok {
		t.Fatal("respuesta sin campo token")
	}
}
