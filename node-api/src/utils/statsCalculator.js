/**
 * Aplana una matriz (array de arrays) en un array plano de números.
 */
function flatten(matriz) {
  return matriz.reduce((acc, fila) => acc.concat(fila), []);
}

/**
 * Verifica si una matriz es diagonal: debe ser cuadrada y todos los
 * elementos fuera de la diagonal principal deben ser 0.
 */
function esDiagonal(matriz) {
  const filas = matriz.length;
  const columnas = filas > 0 ? matriz[0].length : 0;

  if (filas !== columnas) {
    return { esDiagonal: false, razon: 'la matriz no es cuadrada' };
  }

  for (let i = 0; i < filas; i++) {
    for (let j = 0; j < columnas; j++) {
      if (i !== j && Math.abs(matriz[i][j]) > 1e-9) {
        return { esDiagonal: false, razon: `elemento no nulo fuera de la diagonal en [${i}][${j}]` };
      }
    }
  }
  return { esDiagonal: true, razon: null };
}

/**
 * Calcula máximo, mínimo, promedio y suma sobre el conjunto combinado de
 * valores de todas las matrices recibidas, y verifica individualmente
 * si cada una es diagonal.
 *
 * @param {Record<string, number[][]>} matrices - mapa nombre -> matriz (p.ej. { q: [...], r: [...] })
 */
function calcularEstadisticas(matrices) {
  const nombres = Object.keys(matrices);
  const todosLosValores = nombres.reduce(
    (acc, nombre) => acc.concat(flatten(matrices[nombre])),
    []
  );

  if (todosLosValores.length === 0) {
    throw new Error('no hay valores numéricos para calcular estadísticas');
  }

  const suma = todosLosValores.reduce((acc, v) => acc + v, 0);
  const maximo = Math.max(...todosLosValores);
  const minimo = Math.min(...todosLosValores);
  const promedio = suma / todosLosValores.length;

  const diagonalPorMatriz = {};
  let algunaEsDiagonal = false;
  for (const nombre of nombres) {
    const resultado = esDiagonal(matrices[nombre]);
    diagonalPorMatriz[nombre] = resultado;
    if (resultado.esDiagonal) algunaEsDiagonal = true;
  }

  return {
    maximo,
    minimo,
    promedio,
    sumaTotal: suma,
    matrizDiagonal: {
      algunaEsDiagonal,
      detalle: diagonalPorMatriz,
    },
  };
}

module.exports = { calcularEstadisticas, esDiagonal, flatten };
