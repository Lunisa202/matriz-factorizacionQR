package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"reto-tecnico/api-go/config"
	"reto-tecnico/api-go/models"
)

// JWTMiddleware valida el header "Authorization: Bearer <token>" con el secreto
// compartido. Debe usarse en ambas APIs con el mismo JWT_SECRET, JWT_ISSUER y
// JWT_AUDIENCE para que los tokens sean interoperables.
// Exigir issuer/audience evita que tokens de otros sistemas (aunque usen el
// mismo secreto) sean aceptados aquí.
func JWTMiddleware(cfg config.Config) fiber.Handler {
	parser := jwt.NewParser(
		jwt.WithIssuer(cfg.JWTIssuer),
		jwt.WithAudience(cfg.JWTAudience),
	)
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{Error: "missing Authorization header", Code: 401})
		}
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") || parts[1] == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{Error: "invalid Authorization format, expected 'Bearer <token>'", Code: 401})
		}
		tokenStr := parts[1]

		token, err := parser.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.NewError(fiber.StatusUnauthorized, "unexpected signing method")
			}
			return []byte(cfg.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(models.ErrorResponse{Error: "invalid or expired token", Code: 401})
		}
		// Guarda los claims para uso posterior (auditoría, forward, etc.)
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Locals("claims", claims)
		}
		// Guarda el token crudo para reenviarlo a la API de Node sin re-firmar.
		c.Locals("rawToken", tokenStr)
		return c.Next()
	}
}
