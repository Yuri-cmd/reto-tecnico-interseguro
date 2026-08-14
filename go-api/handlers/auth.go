package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// LoginRequest son las credenciales de demo para obtener un token.
// No existe una base de usuarios en el enunciado, así que se valida
// contra un usuario/clave de demostración configurable por entorno.
type LoginRequest struct {
	Usuario string `json:"usuario"`
	Clave   string `json:"clave"`
}

// Login emite un JWT válido por 1 hora si las credenciales coinciden con
// las de demostración. Este token se usa tanto para llamar a /api/qr
// como, internamente, para la llamada Go -> Node.
func Login(secret, demoUser, demoPass string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "body inválido"})
		}
		if req.Usuario != demoUser || req.Clave != demoPass {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "credenciales inválidas"})
		}

		claims := jwt.MapClaims{
			"sub": req.Usuario,
			"exp": time.Now().Add(time.Hour).Unix(),
			"iat": time.Now().Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := token.SignedString([]byte(secret))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "no se pudo generar el token"})
		}

		return c.JSON(fiber.Map{"token": signed, "expiraEn": "1h"})
	}
}
