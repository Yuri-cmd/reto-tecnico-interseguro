package matrix

import (
	"math"
	"testing"
)

func multiply(a, b Matrix) Matrix {
	rows, inner, cols := a.Rows(), b.Rows(), b.Cols()
	out := make(Matrix, rows)
	for i := 0; i < rows; i++ {
		out[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			sum := 0.0
			for k := 0; k < inner; k++ {
				sum += a[i][k] * b[k][j]
			}
			out[i][j] = sum
		}
	}
	return out
}

func approxEqual(a, b Matrix, eps float64) bool {
	if a.Rows() != b.Rows() || a.Cols() != b.Cols() {
		return false
	}
	for i := range a {
		for j := range a[i] {
			if math.Abs(a[i][j]-b[i][j]) > eps {
				return false
			}
		}
	}
	return true
}

func TestDecompose_ReconstructsOriginalMatrix(t *testing.T) {
	a := Matrix{
		{12, -51, 4},
		{6, 167, -68},
		{-4, 24, -41},
	}

	result, err := Decompose(a)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	reconstructed := multiply(result.Q, result.R)
	if !approxEqual(a, reconstructed, 1e-6) {
		t.Fatalf("Q*R no reconstruye la matriz original.\nQ=%v\nR=%v\nQ*R=%v\nA=%v", result.Q, result.R, reconstructed, a)
	}

	if result.R.Rows() != 3 || result.R.Cols() != 3 {
		t.Fatalf("R debería ser 3x3, es %dx%d", result.R.Rows(), result.R.Cols())
	}
	for i := 0; i < result.R.Rows(); i++ {
		for j := 0; j < i; j++ {
			if result.R[i][j] != 0 {
				t.Fatalf("R no es triangular superior en [%d][%d] = %v", i, j, result.R[i][j])
			}
		}
	}
}

func TestDecompose_RectangularMoreRowsThanCols(t *testing.T) {
	a := Matrix{
		{1, 2},
		{3, 4},
		{5, 6},
	}

	result, err := Decompose(a)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	reconstructed := multiply(result.Q, result.R)
	if !approxEqual(a, reconstructed, 1e-6) {
		t.Fatalf("Q*R no reconstruye la matriz original rectangular.\nQ*R=%v\nA=%v", reconstructed, a)
	}
}

func TestDecompose_FewerRowsThanColsReturnsError(t *testing.T) {
	a := Matrix{
		{1, 2, 3},
		{4, 5, 6},
	}
	if _, err := Decompose(a); err == nil {
		t.Fatal("se esperaba error para m < n (más columnas que filas)")
	}
}

func TestValidate_NonRectangularReturnsError(t *testing.T) {
	a := Matrix{
		{1, 2},
		{3},
	}
	if err := Validate(a); err == nil {
		t.Fatal("se esperaba error para matriz no rectangular")
	}
}
