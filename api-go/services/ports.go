package services

// Puertos (interfaces) del servicio — principio de Inversión de Dependencias (DIP):
// los handlers dependen de ESTAS abstracciones, no de las funciones concretas.
// Esto permite inyectar dobles (mocks) en tests y cambiar implementaciones
// sin tocar la capa HTTP (OCP: abierto a extensión, cerrado a modificación).

// QRService abstrae la factorización de matrices.
type QRService interface {
	Decompose(a [][]float64) (q, r [][]float64, err error)
}

// QRServiceFunc adapta una función ordinaria al interfaz QRService.
// Útil en main.go y en tests: QRServiceFunc(QRDecomposition).
type QRServiceFunc func(a [][]float64) (q, r [][]float64, err error)

// Decompose implementa QRService.
func (f QRServiceFunc) Decompose(a [][]float64) ([][]float64, [][]float64, error) {
	return f(a)
}

// StatsForwarder abstrae el reenvío de resultados a la API de Node.
type StatsForwarder interface {
	Forward(nodeURL, rawToken string, payload interface{}) (interface{}, error)
}

// StatsForwarderFunc adapta una función ordinaria al interfaz StatsForwarder.
type StatsForwarderFunc func(nodeURL, rawToken string, payload interface{}) (interface{}, error)

// Forward implementa StatsForwarder.
func (f StatsForwarderFunc) Forward(nodeURL, rawToken string, payload interface{}) (interface{}, error) {
	return f(nodeURL, rawToken, payload)
}
