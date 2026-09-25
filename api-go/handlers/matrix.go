package handlers

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"reto-tecnico/api-go/config"
	"reto-tecnico/api-go/models"
	"reto-tecnico/api-go/services"
)

// Health responde 200 para probes de Docker/Render.
//
// @Summary Salud del servicio
// @Tags ops
// @Produce json
// @Success 200 {object} map[string]string "ej: {status: ok, service: api-go}"
// @Router /health [get]
func Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok", "service": "api-go"})
}

// IssueToken emite un JWT de prueba (firmado con el secreto compartido,
// incluyendo issuer y audience para interoperar con la API de Node).
// Si ENABLE_DEMO_TOKEN=false responde 404: en despliegues públicos nadie
// debe poder autofirmarse tokens. En producción, reemplazar por login real.
//
// @Summary Emite un JWT demo
// @Description Paso 1 para probar en Swagger: copia el token, pulsa Authorize e introdúcelo como "Bearer <token>".
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.TokenRequest false "Usuario (opcional, default demo)"
// @Success 200 {object} models.TokenResponse
// @Failure 404 {object} models.ErrorResponse "Emisor demo desactivado"
// @Router /api/auth/token [post]
func IssueToken(cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !cfg.EnableDemoToken {
			return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{Error: "not found", Code: 404})
		}
		var req models.TokenRequest
		_ = c.BodyParser(&req) // user opcional
		if req.User == "" {
			req.User = "demo"
		}
		now := time.Now()
		claims := jwt.MapClaims{
			"sub": req.User,
			"iss": cfg.JWTIssuer,
			"aud": cfg.JWTAudience,
			"iat": now.Unix(),
			"exp": now.Add(cfg.JWTExpiry).Unix(),
		}
		t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := t.SignedString([]byte(cfg.JWTSecret))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{Error: "could not sign token", Code: 500})
		}
		return c.JSON(models.TokenResponse{Token: signed})
	}
}

// NewQRHandler construye el handler POST /api/matrix/qr (protegido).
// Recibe ABSTRACCIONES (DIP): el handler no conoce las implementaciones
// concretas de QR ni del forwarder, así que en tests se inyectan dobles.
//
// Flujo: 1. Valida matriz. 2. Calcula QR. 3. Reenvía a Node. 4. Responde Q+R+stats.
//
// @Summary Factoriza la matriz (QR) y devuelve Q, R y estadísticas de Node
// @Description Flujo completo: valida la matriz rectangular, calcula QR con Gram-Schmidt y reenvía {q, r} a la API de Node, que responde las estadísticas.
// @Tags matrix
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body models.MatrixRequest true "Matriz rectangular (array de arrays)"
// @Success 200 {object} models.QRResponse
// @Failure 400 {object} models.ErrorResponse "JSON o matriz inválidos"
// @Failure 401 {object} models.ErrorResponse "Sin token o inválido"
// @Failure 413 {object} models.ErrorResponse "Excede MAX_MATRIX_ELEMENTS"
// @Failure 502 {object} models.QRGatewayError "QR calculado pero Node inalcanzable (incluye q y r)"
// @Router /api/matrix/qr [post]
func NewQRHandler(cfg config.Config, qr services.QRService, fw services.StatsForwarder) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req models.MatrixRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Error: "invalid JSON body, expected {\"matrix\": [[..],[..]]}", Code: 400})
		}
		if err := services.ValidateMatrix(req.Matrix, cfg.MaxMatrixElements); err != nil {
			// 413 para exceso de tamaño (anti-DoS), 400 para el resto.
			var tooLarge *services.TooLargeError
			if errors.As(err, &tooLarge) {
				return c.Status(fiber.StatusRequestEntityTooLarge).JSON(models.ErrorResponse{Error: err.Error(), Code: 413})
			}
			return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Error: err.Error(), Code: 400})
		}

		q, r, err := qr.Decompose(req.Matrix)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Error: err.Error(), Code: 400})
		}

		// Reenviar a Node con el mismo JWT entrante.
		rawToken, _ := c.Locals("rawToken").(string)
		forwarded, err := fw.Forward(cfg.NodeAPIURL, rawToken, models.NodePayload{Q: q, R: r})
		if err != nil {
			// 502: QR se calculó bien pero Node falló; se informa sin ocultar Q y R.
			return c.Status(fiber.StatusBadGateway).JSON(models.QRGatewayError{
				Q:     q,
				R:     r,
				Error: "qr computed but forwarding to node failed: " + err.Error(),
			})
		}

		return c.JSON(models.QRResponse{Q: q, R: r, Forwarded: forwarded})
	}
}
