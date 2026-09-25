package services

import (
	"errors"
	"math"
	"testing"
)

func TestValidateMatrixOK(t *testing.T) {
	m := [][]float64{{1, 2}, {3, 4}}
	if err := ValidateMatrix(m, 10000); err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
}

func TestValidateMatrixVacia(t *testing.T) {
	if err := ValidateMatrix(nil, 10000); err == nil {
		t.Fatal("se esperaba error con matriz vacía")
	}
}

func TestValidateMatrixNoRectangular(t *testing.T) {
	m := [][]float64{{1, 2}, {3}}
	if err := ValidateMatrix(m, 10000); err == nil {
		t.Fatal("se esperaba error con matriz no rectangular")
	}
}

func TestValidateMatrixNoFinita(t *testing.T) {
	for _, v := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		m := [][]float64{{v}}
		if err := ValidateMatrix(m, 10000); err == nil {
			t.Fatalf("se esperaba error con valor no finito %v", v)
		}
	}
}

func TestValidateMatrixExcedeTope(t *testing.T) {
	// 101×101 = 10201 elementos > tope de 10000.
	m := make([][]float64, 101)
	for i := range m {
		m[i] = make([]float64, 101)
		for j := range m[i] {
			m[i][j] = 1
		}
	}
	err := ValidateMatrix(m, 10000)
	if err == nil {
		t.Fatal("se esperaba TooLargeError")
	}
	var tle *TooLargeError
	if !errors.As(err, &tle) {
		t.Fatalf("se esperaba *TooLargeError, fue %T", err)
	}
}
