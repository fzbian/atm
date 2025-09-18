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
	Chat    string `json:"chat"`
	Message string `json:"message"`
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
	base := strings.TrimRight(getEnvAny("NOTIFY_URL"), "/")
	if base == "" {
		fmt.Printf("[NOTIFY] configuración incompleta (url=%t)\n", base != "")
		return
	}
	endpoint := buildSendURL(base)
	msg := fmt.Sprintf("%s %s: %s", action, entity, detail)
	send(endpoint, msg)
}

// SendText envía un mensaje de texto arbitrario usando las credenciales del .env
func SendText(text string) {
	base := strings.TrimRight(getEnvAny("NOTIFY_URL"), "/")
	if base == "" {
		fmt.Printf("[NOTIFY] configuración incompleta (url=%t)\n", base != "")
		return
	}
	endpoint := buildSendURL(base)
	send(endpoint, text)
}

// buildSendURL arma la URL final para enviar texto
func buildSendURL(base string) string {
	// base ya viene sin slash final; agregamos uno fijo + path
	return base + "/whatsapp/send-text"
}

// send ejecuta el POST con el nuevo formato {chat:"atm", message:"..."}
func send(endpoint, message string) {
	payload := Payload{Chat: "atm", Message: message}
	b, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(b))
	if err != nil {
		fmt.Printf("[NOTIFY] error creando request: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "atm-notify/2.0")
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
