package services

import (
	"errors"
	"math"
)

// QRDecomposition calcula la descomposición QR reducida de una matriz
// rectangular A (m x n) usando el proceso de Gram-Schmidt clásico.
//
// Retorna:
//   - Q (m x n): columnas ortonormales.
//   - R (n x n): triangular superior.
//
// Si una columna es linealmente dependiente de las anteriores (norma ~ 0),
// se deja la columna de Q en ceros y la fila correspondiente de R en ceros,
// en lugar de fallar: esto hace la función robusta para cualquier matriz
// rectangular, incluyendo m < n.
func QRDecomposition(a [][]float64) (q, r [][]float64, err error) {
	m := len(a)
	if m == 0 {
		return nil, nil, errors.New("matrix must have at least one row")
	}
	n := len(a[0])
	if n == 0 {
		return nil, nil, errors.New("matrix must have at least one column")
	}
	// Validar que sea rectangular.
	for _, row := range a {
		if len(row) != n {
			return nil, nil, errors.New("matrix must be rectangular (all rows same length)")
		}
	}

	// Inicializar Q (m x n) y R (n x n) en ceros.
	q = make([][]float64, m)
	for i := range q {
		q[i] = make([]float64, n)
	}
	r = make([][]float64, n)
	for i := range r {
		r[i] = make([]float64, n)
	}

	const eps = 1e-12

	// Gram-Schmidt por columnas: para cada columna j de A.
	for j := 0; j < n; j++ {
		// v = columna j de A.
		v := make([]float64, m)
		for i := 0; i < m; i++ {
			v[i] = a[i][j]
		}
		// Restar proyecciones sobre q_0..q_{j-1}.
		for i := 0; i < j; i++ {
			// r[i][j] = q_i . v
			dot := 0.0
			for k := 0; k < m; k++ {
				dot += q[k][i] * v[k]
			}
			r[i][j] = dot
			for k := 0; k < m; k++ {
				v[k] -= dot * q[k][i]
			}
		}
		// r[j][j] = ||v||
		norm := 0.0
		for k := 0; k < m; k++ {
			norm += v[k] * v[k]
		}
		norm = math.Sqrt(norm)
		r[j][j] = norm

		if norm < eps {
			// Columna dependiente: dejar q_j en ceros (robusto, no falla).
			continue
		}
		for k := 0; k < m; k++ {
			q[k][j] = v[k] / norm
		}
	}
	return q, r, nil
}
