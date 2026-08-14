// Package matrix implementa la factorización QR de una matriz rectangular
// utilizando reflexiones de Householder (numéricamente más estable que
// Gram-Schmidt clásico).
package matrix

import (
	"errors"
	"math"
)

// Matrix es una matriz representada como slice de filas.
type Matrix [][]float64

// Rows devuelve el número de filas.
func (m Matrix) Rows() int { return len(m) }

// Cols devuelve el número de columnas (0 si la matriz está vacía).
func (m Matrix) Cols() int {
	if len(m) == 0 {
		return 0
	}
	return len(m[0])
}

// Validate verifica que la matriz sea rectangular (todas las filas con
// la misma longitud) y no esté vacía.
func Validate(m Matrix) error {
	if len(m) == 0 || len(m[0]) == 0 {
		return errors.New("la matriz no puede estar vacía")
	}
	cols := len(m[0])
	for i, row := range m {
		if len(row) != cols {
			return errors.New("la matriz debe ser rectangular: todas las filas deben tener la misma cantidad de columnas (fila " + itoa(i) + " no coincide)")
		}
	}
	return nil
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

// QRResult contiene las matrices Q (ortogonal, m x n) y R (triangular
// superior, n x n) tales que A = Q * R (forma "reducida" o "económica").
type QRResult struct {
	Q Matrix
	R Matrix
}

// Decompose calcula la factorización QR de A (m x n) mediante reflexiones
// de Householder. Requiere m >= n (más filas que columnas o matriz
// cuadrada), condición estándar para la QR reducida con columnas
// ortonormales completas. Para m < n no existe una QR reducida con Q de
// columnas ortonormales y R cuadrada invertible en el sentido usual, así
// que se retorna un error explícito en lugar de un resultado ambiguo.
func Decompose(a Matrix) (*QRResult, error) {
	if err := Validate(a); err != nil {
		return nil, err
	}

	m := a.Rows()
	n := a.Cols()
	if m < n {
		return nil, errors.New("la factorización QR requiere que el número de filas sea mayor o igual al número de columnas (m >= n)")
	}

	// Copiamos A en una matriz de trabajo R que iremos triangularizando,
	// y acumulamos las reflexiones en Q (inicialmente la identidad m x m).
	r := cloneMatrix(a)
	q := identity(m)

	for k := 0; k < n && k < m-1; k++ {
		// Construir el vector de Householder para anular la subcolumna
		// r[k:m, k].
		alpha := 0.0
		for i := k; i < m; i++ {
			alpha += r[i][k] * r[i][k]
		}
		alpha = math.Sqrt(alpha)
		if alpha == 0 {
			continue // columna ya nula por debajo de la diagonal
		}
		if r[k][k] > 0 {
			alpha = -alpha
		}

		v := make([]float64, m)
		v[k] = r[k][k] - alpha
		for i := k + 1; i < m; i++ {
			v[i] = r[i][k]
		}

		vNorm := 0.0
		for i := k; i < m; i++ {
			vNorm += v[i] * v[i]
		}
		if vNorm == 0 {
			continue
		}

		// Aplicar H = I - 2*v*v^T/(v^T*v) a R (por la izquierda) y
		// acumular H en Q (por la derecha): Q_nuevo = Q_viejo * H.
		applyHouseholderLeft(r, v, vNorm, k, m, n)
		applyHouseholderRight(q, v, vNorm, k, m)
	}

	// Forma reducida: Q son las primeras n columnas, R las primeras n filas.
	qReduced := make(Matrix, m)
	for i := 0; i < m; i++ {
		qReduced[i] = append([]float64{}, q[i][:n]...)
	}
	rReduced := make(Matrix, n)
	for i := 0; i < n; i++ {
		row := make([]float64, n)
		for j := 0; j < n; j++ {
			if j >= i {
				row[j] = roundSmall(r[i][j])
			}
		}
		rReduced[i] = row
	}
	for i := range qReduced {
		for j := range qReduced[i] {
			qReduced[i][j] = roundSmall(qReduced[i][j])
		}
	}

	return &QRResult{Q: qReduced, R: rReduced}, nil
}

func applyHouseholderLeft(r Matrix, v []float64, vNorm float64, k, m, n int) {
	for j := k; j < n; j++ {
		dot := 0.0
		for i := k; i < m; i++ {
			dot += v[i] * r[i][j]
		}
		factor := 2 * dot / vNorm
		for i := k; i < m; i++ {
			r[i][j] -= factor * v[i]
		}
	}
}

func applyHouseholderRight(q Matrix, v []float64, vNorm float64, k, m int) {
	for i := 0; i < m; i++ {
		dot := 0.0
		for j := k; j < m; j++ {
			dot += q[i][j] * v[j]
		}
		factor := 2 * dot / vNorm
		for j := k; j < m; j++ {
			q[i][j] -= factor * v[j]
		}
	}
}

func cloneMatrix(a Matrix) Matrix {
	out := make(Matrix, len(a))
	for i, row := range a {
		out[i] = append([]float64{}, row...)
	}
	return out
}

func identity(n int) Matrix {
	out := make(Matrix, n)
	for i := range out {
		out[i] = make([]float64, n)
		out[i][i] = 1
	}
	return out
}

// roundSmall limpia el ruido de punto flotante (valores del orden de
// 1e-13 que deberían ser 0) para que la salida sea legible y las
// verificaciones de "matriz diagonal" en la API de Node no fallen por
// errores de redondeo.
func roundSmall(x float64) float64 {
	const eps = 1e-9
	if math.Abs(x) < eps {
		return 0
	}
	rounded := math.Round(x*1e9) / 1e9
	return rounded
}
