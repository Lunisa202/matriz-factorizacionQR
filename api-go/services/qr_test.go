package services

import (
	"math"
	"testing"
)

// casiIguales compara con tolerancia (los flotantes de Gram-Schmidt no son exactos).
func casiIguales(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

// multMatrices multiplica A (m×k) por B (k×n). Solo para verificar Q·R ≈ A.
func multMatrices(a, b [][]float64) [][]float64 {
	m, k, n := len(a), len(b), len(b[0])
	out := make([][]float64, m)
	for i := range out {
		out[i] = make([]float64, n)
		for j := 0; j < n; j++ {
			for p := 0; p < k; p++ {
				out[i][j] += a[i][p] * b[p][j]
			}
		}
	}
	return out
}

func assertReconstruccion(t *testing.T, a, q, r [][]float64) {
	t.Helper()
	qr := multMatrices(q, r)
	for i := range a {
		for j := range a[i] {
			if !casiIguales(qr[i][j], a[i][j]) {
				t.Fatalf("Q·R[%d][%d] = %v, esperado %v (A original)", i, j, qr[i][j], a[i][j])
			}
		}
	}
}

func TestQRDescomposicionIdentidad(t *testing.T) {
	a := [][]float64{{1, 0}, {0, 1}}
	q, r, err := QRDecomposition(a)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	// Q y R deben ser la identidad.
	for i := range a {
		for j := range a[i] {
			esperado := 0.0
			if i == j {
				esperado = 1.0
			}
			if !casiIguales(q[i][j], esperado) || !casiIguales(r[i][j], esperado) {
				t.Fatalf("Q,R incorrectos en [%d][%d]: q=%v r=%v", i, j, q[i][j], r[i][j])
			}
		}
	}
}

func TestQRDescomposicionCuadrada(t *testing.T) {
	a := [][]float64{{1, 2}, {3, 4}}
	q, r, err := QRDecomposition(a)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	assertReconstruccion(t, a, q, r)
	// R debe ser triangular superior: r[1][0] ≈ 0.
	if !casiIguales(r[1][0], 0) {
		t.Fatalf("R no es triangular superior: r[1][0] = %v", r[1][0])
	}
}

func TestQRDescomposicionRectangular(t *testing.T) {
	// Matriz 3×2 (más filas que columnas).
	a := [][]float64{{1, 2}, {3, 4}, {5, 6}}
	q, r, err := QRDecomposition(a)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if len(q) != 3 || len(q[0]) != 2 {
		t.Fatalf("Q debe ser 3×2, fue %dx%d", len(q), len(q[0]))
	}
	if len(r) != 2 || len(r[0]) != 2 {
		t.Fatalf("R debe ser 2×2, fue %dx%d", len(r), len(r[0]))
	}
	assertReconstruccion(t, a, q, r)
}

func TestQRColumnaDependiente(t *testing.T) {
	// Segunda columna = 2× primera: el algoritmo NO debe fallar,
	// deja q_1 en ceros y r[1][1] = 0 (robustez documentada).
	a := [][]float64{{1, 2}, {2, 4}}
	q, r, err := QRDecomposition(a)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if !casiIguales(r[1][1], 0) {
		t.Fatalf("r[1][1] debe ser 0 en columna dependiente, fue %v", r[1][1])
	}
	_ = q
}

func TestQRNoRectangularFalla(t *testing.T) {
	a := [][]float64{{1, 2}, {3}}
	if _, _, err := QRDecomposition(a); err == nil {
		t.Fatal("se esperaba error con matriz no rectangular")
	}
}

func TestQRVaciaFalla(t *testing.T) {
	if _, _, err := QRDecomposition(nil); err == nil {
		t.Fatal("se esperaba error con matriz vacía")
	}
}
