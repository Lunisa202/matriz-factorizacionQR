package services

import (
	"fmt"
	"math"
)

// TooLargeError indica que la matriz supera el tope anti-DoS.
// Es un tipo propio para que el handler pueda distinguir 413 de 400
// con errors.As en lugar de comparar strings (código limpio).
type TooLargeError struct {
	Elements int
	Limit    int
}

func (e *TooLargeError) Error() string {
	return fmt.Sprintf("payload too large: %d elements exceed the limit of %d", e.Elements, e.Limit)
}

// ValidateMatrix verifica que la matriz sea rectangular, contenga solo
// números finitos (rechaza NaN/±Inf del JSON) y no supere maxElements.
// SRP: toda la validación vive aquí; QRDecomposition solo calcula.
func ValidateMatrix(m [][]float64, maxElements int) error {
	if len(m) == 0 {
		return fmt.Errorf("matrix must be a non-empty array of arrays")
	}
	n := len(m[0])
	if n == 0 {
		return fmt.Errorf("matrix must have at least one column")
	}
	for i, row := range m {
		if len(row) != n {
			return fmt.Errorf("matrix must be rectangular (row %d has %d cols, expected %d)", i, len(row), n)
		}
		for _, v := range row {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return fmt.Errorf("matrix must contain only finite numbers")
			}
		}
	}
	if total := len(m) * n; total > maxElements {
		return &TooLargeError{Elements: total, Limit: maxElements}
	}
	return nil
}
