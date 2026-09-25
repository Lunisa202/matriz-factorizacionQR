package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ForwardToNode reenvía Q y R a la API de Node vía POST JSON.
// Reutiliza el JWT entrante (rawToken) para autenticarse en Node,
// ya que ambas APIs comparten el mismo JWT_SECRET.
// Usa el timeout por defecto (10s); para un timeout configurable ver
// ForwardToNodeWithTimeout.
func ForwardToNode(nodeURL, rawToken string, payload interface{}) (interface{}, error) {
	return ForwardToNodeWithTimeout(nodeURL, rawToken, payload, 10*time.Second)
}

// ForwardToNodeWithTimeout es la variante con timeout inyectable (DIP):
// main.go la adapta con el valor de configuración (FORWARD_TIMEOUT).
func ForwardToNodeWithTimeout(nodeURL, rawToken string, payload interface{}, timeout time.Duration) (interface{}, error) {
	client := &http.Client{Timeout: timeout}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, nodeURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if rawToken != "" {
		req.Header.Set("Authorization", "Bearer "+rawToken)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("node api unreachable at %s: %w", nodeURL, err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("node api responded %d: %s", resp.StatusCode, string(raw))
	}

	var decoded interface{}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		// Si Node devolvió algo no-JSON, devolverlo como string.
		return string(raw), nil
	}
	return decoded, nil
}
