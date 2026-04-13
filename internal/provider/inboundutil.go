package provider

import (
	"encoding/json"
	"fmt"
)

func inboundMapFromJSON(raw []byte) (map[string]any, error) {
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func intFromMap(m map[string]any, key string) (int, error) {
	v, ok := m[key]
	if !ok {
		return 0, fmt.Errorf("missing %q", key)
	}
	switch n := v.(type) {
	case float64:
		return int(n), nil
	case int:
		return n, nil
	case int64:
		return int(n), nil
	default:
		return 0, fmt.Errorf("invalid type for %q", key)
	}
}

func stringFromMap(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}

func int64FromMap(m map[string]any, key string) int64 {
	v, ok := m[key]
	if !ok || v == nil {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return int64(n)
	case int64:
		return n
	case int:
		return int64(n)
	default:
		return 0
	}
}

func boolFromMap(m map[string]any, key string) bool {
	v, ok := m[key]
	if !ok || v == nil {
		return false
	}
	b, ok := v.(bool)
	if ok {
		return b
	}
	return false
}

// mergeInboundSettingsPreservingClients applies non-`clients` keys from userJSON onto serverJSON.
// If the server JSON has a `clients` key, its value is kept (so clients managed via API / xui_vless_client stay).
// If the server has no `clients` key (e.g. some protocols), the merged object has no `clients` key unless user-only merge added nothing there — user `clients` are never applied on update.
func mergeInboundSettingsPreservingClients(serverJSON, userJSON string) (string, error) {
	var server map[string]any
	if err := json.Unmarshal([]byte(serverJSON), &server); err != nil {
		return "", fmt.Errorf("parse server settings: %w", err)
	}
	if server == nil {
		server = map[string]any{}
	}
	existingClients, hadClients := server["clients"]

	var user map[string]any
	if err := json.Unmarshal([]byte(userJSON), &user); err != nil {
		return "", fmt.Errorf("parse settings: %w", err)
	}
	if user == nil {
		user = map[string]any{}
	}
	for k, v := range user {
		if k == "clients" {
			continue
		}
		server[k] = v
	}
	if hadClients {
		server["clients"] = existingClients
	} else {
		delete(server, "clients")
	}
	out, err := json.MarshalIndent(server, "", "  ")
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func findVLESSClientByEmail(settingsJSON, email string) (map[string]any, error) {
	var root map[string]any
	if err := json.Unmarshal([]byte(settingsJSON), &root); err != nil {
		return nil, err
	}
	raw, ok := root["clients"].([]any)
	if !ok {
		return nil, fmt.Errorf("no clients in inbound settings")
	}
	for _, c := range raw {
		cm, ok := c.(map[string]any)
		if !ok {
			continue
		}
		if em, _ := cm["email"].(string); em == email {
			return cm, nil
		}
	}
	return nil, fmt.Errorf("client with email %q not found", email)
}

func clientUUID(cm map[string]any) string {
	id, _ := cm["id"].(string)
	return id
}
