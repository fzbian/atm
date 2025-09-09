package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Payload representa el cuerpo JSON que exige el endpoint externo
type Payload struct {
	Number string `json:"number"`
	Text   string `json:"text"`
}

// SendMovement envía un mensaje de confirmación de movimiento (create/update/delete)
// action: CREATE | UPDATE | DELETE
// entity: nombre de la entidad (ej: TRANSACCION, CATEGORIA)
// detail: texto adicional (ej: id, montos)
func SendMovement(action, entity, detail string) {
	apiURL := os.Getenv("MESSAGE_API_URL")
	apiKey := os.Getenv("MESSAGE_API_KEY")
	defaultNumber := os.Getenv("MESSAGE_DEFAULT_NUMBER")
	if apiURL == "" || apiKey == "" || defaultNumber == "" {
		// Falta configuración, no interrumpe el flujo principal
		return
	}

	msg := fmt.Sprintf("%s %s: %s", action, entity, detail)
	payload := Payload{Number: defaultNumber, Text: msg}
	b, _ := json.Marshal(payload)

	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewReader(b))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	// No procesamos respuesta; propósito solo notificar
}

// SendText envía un mensaje de texto arbitrario usando las credenciales del .env
func SendText(text string) {
	apiURL := os.Getenv("MESSAGE_API_URL")
	apiKey := os.Getenv("MESSAGE_API_KEY")
	defaultNumber := os.Getenv("MESSAGE_DEFAULT_NUMBER")
	if apiURL == "" || apiKey == "" || defaultNumber == "" {
		return
	}
	payload := Payload{Number: defaultNumber, Text: text}
	b, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewReader(b))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", apiKey)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}
