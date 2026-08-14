package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"

	"reto-tecnico/go-api/matrix"
)

// QRRequest es el payload de entrada: una matriz rectangular como array
// de arrays de números.
type QRRequest struct {
	Matriz [][]float64 `json:"matriz"`
}

// nodeStatsPayload es lo que se envía a la API de Node.js: las dos
// matrices resultantes de la factorización QR (Q y R). Se envían ambas
// -en plural- porque el enunciado pide estadísticas sobre "las matrices
// devueltas" y verificar "si alguna matriz es diagonal", lo cual solo
// tiene sentido pleno cuando hay más de una matriz de salida (como en
// QR), a diferencia de una rotación que produce una única matriz.
type nodeStatsPayload struct {
	Q [][]float64 `json:"q"`
	R [][]float64 `json:"r"`
}

// QRHandler crea el handler HTTP para POST /api/qr. Calcula la
// factorización QR de la matriz recibida y reenvía Q y R a la API de
// Node.js para que calcule las estadísticas, devolviendo al cliente la
// respuesta combinada.
func QRHandler(nodeAPIURL string) fiber.Handler {
	client := &http.Client{Timeout: 5 * time.Second}

	return func(c *fiber.Ctx) error {
		var req QRRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "body inválido: se espera {\"matriz\": [[...], [...]]}",
			})
		}

		a := matrix.Matrix(req.Matriz)
		result, err := matrix.Decompose(a)
		if err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		stats, statusCode, err := forwardToNode(client, nodeAPIURL, c.Get("Authorization"), nodeStatsPayload{
			Q: [][]float64(result.Q),
			R: [][]float64(result.R),
		})
		if err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
				"error": "no se pudo contactar a la API de estadísticas (Node.js): " + err.Error(),
			})
		}
		if statusCode >= 400 {
			return c.Status(statusCode).JSON(fiber.Map{
				"error":            "la API de estadísticas devolvió un error",
				"detalleNode":      stats,
				"codigoHTTPOrigen": statusCode,
			})
		}

		return c.JSON(fiber.Map{
			"matrizOriginal": req.Matriz,
			"factorizacionQR": fiber.Map{
				"q": result.Q,
				"r": result.R,
			},
			"estadisticas": stats,
		})
	}
}

func forwardToNode(client *http.Client, nodeAPIURL, authHeader string, payload nodeStatsPayload) (map[string]interface{}, int, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, err
	}

	req, err := http.NewRequest(http.MethodPost, nodeAPIURL, bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	var parsed map[string]interface{}
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &parsed); err != nil {
			return nil, resp.StatusCode, err
		}
	}
	return parsed, resp.StatusCode, nil
}
