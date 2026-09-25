package models

// MatrixRequest es el payload que recibe POST /api/matrix/qr
type MatrixRequest struct {
	Matrix [][]float64 `json:"matrix"`
}

// QRResponse es la respuesta del servicio Go (incluye lo que devuelva Node).
type QRResponse struct {
	Q         [][]float64 `json:"q"`
	R         [][]float64 `json:"r"`
	Forwarded interface{} `json:"stats,omitempty"`
}

// NodePayload es lo que Go reenvía a la API de Node.
type NodePayload struct {
	Q [][]float64 `json:"q"`
	R [][]float64 `json:"r"`
}

// QRGatewayError es la respuesta 502: el QR se calculó pero Node falló.
// Se devuelven Q y R igualmente (fail-soft) junto al detalle del error.
type QRGatewayError struct {
	Q     [][]float64 `json:"q"`
	R     [][]float64 `json:"r"`
	Error string      `json:"error"`
}
type TokenRequest struct {
	User string `json:"user" example:"demo"`
}

// TokenResponse devuelve el JWT firmado.
type TokenResponse struct {
	Token string `json:"token"`
}

// ErrorResponse estandariza los errores de la API.
type ErrorResponse struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}
