package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	swagger "github.com/gofiber/swagger"
	"reto-tecnico/api-go/config"
	_ "reto-tecnico/api-go/docs" // spec generada por `swag init` (ver README §5)
	"reto-tecnico/api-go/handlers"
	"reto-tecnico/api-go/middleware"
	"reto-tecnico/api-go/models"
	"reto-tecnico/api-go/services"
)

// @title Reto técnico — API Go (Fiber)
// @version 1.0.0
// @description Recibe una matriz rectangular, calcula su factorización QR (Gram-Schmidt) y reenvía {q, r} a la API de Node. "Try it out" apunta a este mismo servidor: pide un token en POST /api/auth/token, pulsa Authorize (Bearer) y prueba.
// @BasePath /
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description JWT con formato "Bearer <token>".
func main() {
	cfg := config.Load()

	// Fail-fast: no arrancar con configuración insegura en producción.
	if err := config.Validate(cfg); err != nil {
		log.Fatal(err)
	}

	app := fiber.New(fiber.Config{
		// Límite de body configurable (defensa anti-DoS).
		BodyLimit: cfg.MaxBodyBytes,
		// Manejo centralizado de errores: siempre JSON {error, code}.
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(models.ErrorResponse{Error: err.Error(), Code: code})
		},
	})

	app.Use(recover.New()) // evita caídas por pánicos
	app.Use(logger.New())  // log de requests
	// CORS restringible por env (en local '*' para Postman/browser).
	app.Use(cors.New(cors.Config{AllowOrigins: cfg.CORSOrigin}))

	// Rutas públicas (health + emisión de token demo, con guard).
	app.Get("/health", handlers.Health)
	app.Post("/api/auth/token", handlers.IssueToken(cfg))

	// Documentación interactiva: Swagger UI en /docs (spec en /docs/doc.json).
	// Pública y sin rate-limit para que los evaluadores prueben sin fricción.
	app.Get("/docs/*", swagger.HandlerDefault)

	// Inyección de dependencias (DIP): el handler recibe interfaces,
	// las implementaciones concretas se conectan aquí, en el composition root.
	qrSvc := services.QRServiceFunc(services.QRDecomposition)
	forwarder := services.StatsForwarderFunc(func(nodeURL, rawToken string, payload interface{}) (interface{}, error) {
		return services.ForwardToNodeWithTimeout(nodeURL, rawToken, payload, cfg.ForwardTimeout)
	})

	// Rutas protegidas con JWT compartido (secreto + issuer + audience).
	// Rate limit anti fuerza-bruta/DoS solo en /api (el /health queda libre).
	protected := app.Group("/api",
		limiter.New(limiter.Config{
			Max:        cfg.RateLimitMax,
			Expiration: cfg.RateLimitWindow,
		}),
		middleware.JWTMiddleware(cfg),
	)
	protected.Post("/matrix/qr", handlers.NewQRHandler(cfg, qrSvc, forwarder))

	log.Printf("api-go listening on :%s | forwarding to %s", cfg.Port, cfg.NodeAPIURL)
	log.Fatal(app.Listen(":" + cfg.Port))
}
