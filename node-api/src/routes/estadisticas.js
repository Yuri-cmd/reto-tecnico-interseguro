const express = require('express');
const { calcularEstadisticas } = require('../utils/statsCalculator');

const router = express.Router();

function esMatrizValida(m) {
  if (!Array.isArray(m) || m.length === 0) return false;
  const columnas = m[0].length;
  return m.every(
    (fila) => Array.isArray(fila) && fila.length === columnas && fila.every((v) => typeof v === 'number' && Number.isFinite(v))
  );
}

/**
 * POST /api/estadisticas
 * Recibe las matrices resultantes de la factorización QR (Q y R)
 * calculada por la API en Go, y devuelve máximo, mínimo, promedio, suma
 * total y verificación de matriz diagonal sobre ellas.
 *
 * Body esperado: { "q": number[][], "r": number[][] }
 */
router.post('/', (req, res) => {
  const { q, r } = req.body || {};

  if (!esMatrizValida(q) || !esMatrizValida(r)) {
    return res.status(422).json({
      error: 'se esperan las matrices "q" y "r" como arrays de arrays de números, rectangulares y no vacíos',
    });
  }

  try {
    const estadisticas = calcularEstadisticas({ q, r });
    return res.json(estadisticas);
  } catch (err) {
    return res.status(422).json({ error: err.message });
  }
});

module.exports = router;
