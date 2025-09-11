package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Payload representa el cuerpo JSON que exige el endpoint externo
type Payload struct {
	Number string `json:"number"`
	Text   string `json:"text"`
}

// getEnvAny devuelve el primer valor no vacío para una lista de claves
func getEnvAny(keys ...string) string {
	for _, k := range keys {
		if v := strings.TrimSpace(os.Getenv(k)); v != "" {
			return v
		}
	}
	return ""
}

// SendMovement envía un mensaje de confirmación de movimiento (create/update/delete)
// action: CREATE | UPDATE | DELETE
// entity: nombre de la entidad (ej: TRANSACCION, CATEGORIA)
// detail: texto adicional (ej: id, montos)
func SendMovement(action, entity, detail string) {
	apiURL := getEnvAny("NOTIFY_URL", "MESSAGE_API_URL")
	apiKey := getEnvAny("NOTIFY_APIKEY", "MESSAGE_API_KEY")
	defaultNumber := getEnvAny("NOTIFY_NUMBER", "MESSAGE_DEFAULT_NUMBER")
	if apiURL == "" || apiKey == "" || defaultNumber == "" {
		// Falta configuración; registrar y salir sin interrumpir el flujo principal
		fmt.Printf("[NOTIFY] configuración incompleta (url=%t, apikey=%t, number=%t)\n", apiURL != "", apiKey != "", defaultNumber != "")
		return
	}

	msg := fmt.Sprintf("%s %s: %s", action, entity, detail)
	payload := Payload{Number: defaultNumber, Text: msg}
	b, _ := json.Marshal(payload)

	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewReader(b))
	if err != nil {
		fmt.Printf("[NOTIFY] error creando request: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	// Algunos servicios usan distintas variantes del header
	req.Header.Set("apikey", apiKey)
	req.Header.Set("X-Api-Key", apiKey)
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("User-Agent", "atm-notify/1.0")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("[NOTIFY] error enviando request: %v\n", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		fmt.Printf("[NOTIFY] respuesta no exitosa: status=%d body=%s\n", resp.StatusCode, string(body))
	}
}

// SendText envía un mensaje de texto arbitrario usando las credenciales del .env
func SendText(text string) {
	apiURL := getEnvAny("NOTIFY_URL", "MESSAGE_API_URL")
	apiKey := getEnvAny("NOTIFY_APIKEY", "MESSAGE_API_KEY")
	defaultNumber := getEnvAny("NOTIFY_NUMBER", "MESSAGE_DEFAULT_NUMBER")
	if apiURL == "" || apiKey == "" || defaultNumber == "" {
		fmt.Printf("[NOTIFY] configuración incompleta (url=%t, apikey=%t, number=%t)\n", apiURL != "", apiKey != "", defaultNumber != "")
		return
	}
	payload := Payload{Number: defaultNumber, Text: text}
	b, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewReader(b))
	if err != nil {
		fmt.Printf("[NOTIFY] error creando request: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", apiKey)
	req.Header.Set("X-Api-Key", apiKey)
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("User-Agent", "atm-notify/1.0")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("[NOTIFY] error enviando request: %v\n", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		fmt.Printf("[NOTIFY] respuesta no exitosa: status=%d body=%s\n", resp.StatusCode, string(body))
	}
}
