package odoo

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strings"
	"time"
)

type Client struct {
	baseURL  string
	db       string
	user     string
	password string
	hc       *http.Client
}

type authRequest struct {
	JSONRPC string     `json:"jsonrpc"`
	Method  string     `json:"method"`
	Params  authParams `json:"params"`
	ID      int        `json:"id"`
}

type authParams struct {
	DB       string `json:"db"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

type jsonRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      int         `json:"id"`
}

type callKwParams struct {
	Model  string                 `json:"model"`
	Method string                 `json:"method"`
	Args   []interface{}          `json:"args"`
	Kwargs map[string]interface{} `json:"kwargs"`
}

type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result"`
	Error   *rpcError       `json:"error"`
}

type rpcError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type posSession struct {
	BalanceEndReal float64       `json:"cash_register_balance_end_real"`
	State          string        `json:"state"`
	ConfigID       []interface{} `json:"config_id"`
}

func NewFromEnv() (*Client, error) {
	base := os.Getenv("ODOO_URL")
	db := os.Getenv("ODOO_DB")
	user := os.Getenv("ODOO_USER")
	pass := os.Getenv("ODOO_PASSWORD")
	if base == "" || db == "" || user == "" || pass == "" {
		return nil, errors.New("variables ODOO_URL, ODOO_DB, ODOO_USER, ODOO_PASSWORD requeridas")
	}
	base = strings.TrimRight(base, "/")
	jar, _ := cookiejar.New(nil)
	return &Client{baseURL: base, db: db, user: user, password: pass, hc: &http.Client{Timeout: 20 * time.Second, Jar: jar}}, nil
}

func (c *Client) Authenticate() error {
	payload := authRequest{
		JSONRPC: "2.0",
		Method:  "call",
		Params: authParams{
			DB:       c.db,
			Login:    c.user,
			Password: c.password,
		},
		ID: 1,
	}
	b, _ := json.Marshal(payload)
	resp, err := c.hc.Post(c.baseURL+"/web/session/authenticate", "application/json", strings.NewReader(string(b)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("auth fallo status %d", resp.StatusCode)
	}
	var r jsonRPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return err
	}
	if r.Error != nil {
		return fmt.Errorf("auth error: %s", r.Error.Message)
	}
	return nil
}

// Helper genérico para llamar a Odoo (call_kw)
func (c *Client) callOdoo(model, method string, args []interface{}, kwargs map[string]interface{}) (json.RawMessage, error) {
	req := jsonRPCRequest{JSONRPC: "2.0", Method: "call", Params: callKwParams{Model: model, Method: method, Args: args, Kwargs: kwargs}, ID: 99}
	b, _ := json.Marshal(req)
	resp, err := c.hc.Post(c.baseURL+"/web/dataset/call_kw", "application/json", strings.NewReader(string(b)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("http %d", resp.StatusCode)
	}
	var r jsonRPCResponse
	if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	if r.Error != nil {
		dataJSON, _ := json.Marshal(r.Error.Data)
		return nil, fmt.Errorf("rpc code=%d msg=%s data=%s", r.Error.Code, r.Error.Message, string(dataJSON))
	}
	return r.Result, nil
}

// Construye un dominio OR para nombres que comienzan por prefijos (name ilike 'Prefijo%')
func buildNamePrefixDomain(prefixes []string) []interface{} {
	var domain []interface{}
	// Para n prefijos se necesitan n-1 '|' al inicio encadenando OR
	if len(prefixes) == 0 {
		return []interface{}{}
	}
	if len(prefixes) == 1 {
		return []interface{}{[]interface{}{"name", "ilike", prefixes[0] + "%"}}
	}
	// Ejemplo para 3 prefijos: '|','|', cond1, cond2, cond3
	for i := 0; i < len(prefixes)-1; i++ {
		domain = append(domain, "|")
	}
	for _, p := range prefixes {
		domain = append(domain, []interface{}{"name", "ilike", p + "%"})
	}
	return domain
}

// FetchPOSBalances optimizado:
// 1. Busca pos.config cuyos nombres comiencen por prefijos.
// 2. Obtiene sesiones abiertas/abriendo por esos config_id (orden desc) y toma la más reciente por local.
// 3. Para locales sin sesión abierta, obtiene la sesión cerrada más reciente.
// 4. Calcula balance según estado: opened/opening_control -> cash_register_balance_end; closed -> cash_register_balance_end_real.
func (c *Client) FetchPOSBalances() (map[string]float64, float64, error) {
	prefixes := []string{"Gran San", "Visto", "Lo Nuestro", "Burbuja", "San Jose", "Medellin"}

	cfgDomain := buildNamePrefixDomain(prefixes)
	cfgFields := []string{"name"}
	cfgKw := map[string]interface{}{"domain": cfgDomain, "fields": cfgFields, "limit": 200}
	rawCfg, err := c.callOdoo("pos.config", "search_read", []interface{}{}, cfgKw)
	if err != nil {
		return nil, 0, fmt.Errorf("config search: %w", err)
	}
	var cfgs []map[string]interface{}
	if err = json.Unmarshal(rawCfg, &cfgs); err != nil {
		return nil, 0, fmt.Errorf("config decode: %w", err)
	}
	if len(cfgs) == 0 {
		return map[string]float64{}, 0, nil
	}

	configIDs := make([]int64, 0, len(cfgs))
	nameByID := make(map[int64]string)
	for _, cobj := range cfgs {
		if idf, ok := cobj["id"].(float64); ok {
			id := int64(idf)
			configIDs = append(configIDs, id)
			if n, okn := cobj["name"].(string); okn {
				nameByID[id] = n
			}
		}
	}
	if len(configIDs) == 0 {
		return map[string]float64{}, 0, nil
	}

	// Helper dominio config_ids IN
	makeInDomain := func(ids []int64) []interface{} {
		val := make([]interface{}, 0, len(ids))
		for _, id := range ids {
			val = append(val, id)
		}
		return []interface{}{[]interface{}{"config_id", "in", val}}
	}

	// Paso 2: sesiones abiertas / opening_control
	openDomain := makeInDomain(configIDs)
	openDomain = append(openDomain, []interface{}{"state", "in", []interface{}{"opened", "opening_control"}})
	sessFields := []string{"config_id", "state", "cash_register_balance_end_real", "cash_register_balance_end"}
	openKw := map[string]interface{}{"domain": openDomain, "fields": sessFields, "limit": len(configIDs) * 3, "order": "id desc"}
	rawOpen, err := c.callOdoo("pos.session", "search_read", []interface{}{}, openKw)
	if err != nil {
		return nil, 0, fmt.Errorf("open sessions: %w", err)
	}
	var openSessions []map[string]interface{}
	if err = json.Unmarshal(rawOpen, &openSessions); err != nil {
		return nil, 0, fmt.Errorf("open decode: %w", err)
	}

	chosen := make(map[int64]map[string]interface{})
	for _, s := range openSessions {
		cfgID := extractConfigID(s)
		if cfgID == 0 {
			continue
		}
		if _, exists := chosen[cfgID]; !exists { // primera (orden desc => más reciente)
			chosen[cfgID] = s
		}
	}

	// Paso 3: sesiones cerradas para los locales sin abierta
	missing := make([]int64, 0)
	for _, id := range configIDs {
		if _, ok := chosen[id]; !ok {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		closedDomain := makeInDomain(missing) // quitar filtro de estado para capturar cualquier última sesión
		closedKw := map[string]interface{}{"domain": closedDomain, "fields": sessFields, "limit": len(missing) * 5, "order": "id desc"}
		rawClosed, err2 := c.callOdoo("pos.session", "search_read", []interface{}{}, closedKw)
		if err2 == nil {
			var closedSessions []map[string]interface{}
			if err3 := json.Unmarshal(rawClosed, &closedSessions); err3 == nil {
				for _, s := range closedSessions {
					cfgID := extractConfigID(s)
					if cfgID == 0 {
						continue
					}
					if _, exists := chosen[cfgID]; !exists {
						chosen[cfgID] = s
					}
				}
			}
		}
	}

	// Paso 4: calcular balances
	locales := make(map[string]float64)
	var total float64
	for cfgID, sess := range chosen {
		name := nameByID[cfgID]
		key := normalizeLocalKey(name)
		state := fmt.Sprintf("%v", sess["state"])
		var balance float64
		if state == "opened" || state == "opening_control" {
			balance = numberAsFloat(sess["cash_register_balance_end"])
		} else if state == "closed" {
			balance = numberAsFloat(sess["cash_register_balance_end_real"])
			if balance == 0 {
				balance = numberAsFloat(sess["cash_register_balance_end"])
			}
		} else { // cualquier otro estado
			balance = numberAsFloat(sess["cash_register_balance_end_real"])
			if balance == 0 {
				balance = numberAsFloat(sess["cash_register_balance_end"])
			}
		}
		locales[key] += balance
		total += balance
	}
	// Agregar explícitamente configs sin sesión (balance 0) para que aparezcan en respuesta
	for _, cfgID := range configIDs {
		if _, ok := chosen[cfgID]; !ok {
			key := normalizeLocalKey(nameByID[cfgID])
			if _, exists := locales[key]; !exists {
				locales[key] = 0
			}
		}
	}
	return locales, total, nil
}

func extractConfigID(sess map[string]interface{}) int64 {
	if cfg, ok := sess["config_id"].([]interface{}); ok && len(cfg) >= 1 {
		if idf, ok2 := cfg[0].(float64); ok2 {
			return int64(idf)
		}
	}
	return 0
}

func normalizeLocalKey(name string) string {
	lower := strings.ToLower(name)
	switch {
	case strings.HasPrefix(lower, "visto"):
		return "visto"
	case strings.HasPrefix(lower, "burbuja"):
		return "burbuja"
	case strings.HasPrefix(lower, "lo nuestro"):
		return "lo_nuestro"
	case strings.HasPrefix(lower, "gran san"):
		return "gran_san"
	case strings.HasPrefix(lower, "san jose"):
		return "san_jose"
	case strings.HasPrefix(lower, "medellin"):
		return "medellin"
	default:
		// tomar primera palabra segura
		parts := strings.Fields(lower)
		if len(parts) > 0 {
			return parts[0]
		}
		return "desconocido"
	}
}

// Utilidades
func numberAsFloat(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case json.Number:
		f, _ := n.Float64()
		return f
	default:
		return 0
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max < 10 {
		return s[:max]
	}
	return s[:max-7] + "[...]"
}
