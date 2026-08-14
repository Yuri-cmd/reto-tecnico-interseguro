package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"reto-tecnico/go-api/handlers"
	authmw "reto-tecnico/go-api/middleware"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	port := getEnv("PORT", "3000")
	// NODE_API_HOST (solo host, p.ej. "reto-node-api.onrender.com") tiene
	// prioridad para despliegues en la nube, donde Render solo puede
	// inyectar el hostname del otro servicio, no una URL completa.
	// NODE_API_URL (URL completa) se usa para docker-compose local.
	nodeAPIURL := getEnv("NODE_API_URL", "http://localhost:4000/api/estadisticas")
	if host := os.Getenv("NODE_API_HOST"); host != "" {
		nodeAPIURL = "https://" + host + "/api/estadisticas"
	}
	jwtSecret := getEnv("JWT_SECRET", "dev-secret-cambiar-en-produccion")
	demoUser := getEnv("DEMO_USER", "admin")
	demoPass := getEnv("DEMO_PASS", "admin123")

	app := fiber.New(fiber.Config{
		AppName: "QR Matrix API (Go + Fiber)",
	})

	app.Use(logger.New())
	app.Use(cors.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	api := app.Group("/api")
	api.Post("/auth/login", handlers.Login(jwtSecret, demoUser, demoPass))
	api.Post("/qr", authmw.Protect(jwtSecret), handlers.QRHandler(nodeAPIURL))

	log.Printf("Go/Fiber API escuchando en :%s (Node API en %s)", port, nodeAPIURL)
	if err := app.Listen(":" + port); err != nil {
		log.Fatal(err)
	}
}
