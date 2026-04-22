package xui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestRealPanelListInboundsMatchesEnvelopeShape(t *testing.T) {
	if os.Getenv("XUI_REAL_PANEL_TEST") != "1" {
		t.Skip("set XUI_REAL_PANEL_TEST=1 to run real panel test")
	}

	baseURL, user, pass, ok := readXUIConfigFromDotFile()
	if !ok {
		t.Skip(".xui not found or missing required fields")
	}

	c, err := NewClient(baseURL, user, pass, true)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	gotObj, err := c.ListInbounds()
	if err != nil {
		t.Fatalf("ListInbounds() error = %v", err)
	}

	fixture := mustReadJSONFile(t, "testdata/real_api_list_sanitized.json")
	fixtureObjRaw, ok := fixture["obj"]
	if !ok {
		t.Fatalf("fixture missing obj")
	}

	var gotList []map[string]any
	if err := json.Unmarshal(gotObj, &gotList); err != nil {
		t.Fatalf("unmarshal real list obj: %v", err)
	}
	fixtureObjBytes, _ := json.Marshal(fixtureObjRaw)
	var fixtureList []map[string]any
	if err := json.Unmarshal(fixtureObjBytes, &fixtureList); err != nil {
		t.Fatalf("unmarshal fixture obj: %v", err)
	}

	if len(gotList) == 0 {
		t.Fatalf("real panel returned zero inbounds")
	}
	if len(fixtureList) == 0 {
		t.Fatalf("fixture list is empty")
	}

	for _, requiredKey := range []string{"id", "port", "protocol", "settings"} {
		if _, ok := gotList[0][requiredKey]; !ok {
			t.Fatalf("real list first item missing key %q", requiredKey)
		}
		if _, ok := fixtureList[0][requiredKey]; !ok {
			t.Fatalf("fixture first item missing key %q", requiredKey)
		}
	}
}

func mustReadJSONFile(t *testing.T, relPath string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(relPath))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	return out
}

func readXUIConfigFromDotFile() (baseURL, user, pass string, ok bool) {
	raw, err := os.ReadFile(filepath.Join("..", "..", ".xui"))
	if err != nil {
		return "", "", "", false
	}
	text := string(raw)
	get := func(re string) string {
		m := regexp.MustCompile(re).FindStringSubmatch(text)
		if len(m) < 2 {
			return ""
		}
		return strings.TrimSpace(m[1])
	}
	baseURL = get(`Access URL:\s*(\S+)`)
	user = get(`Username:\s*(\S+)`)
	pass = get(`Password:\s*(\S+)`)
	if baseURL == "" || user == "" || pass == "" {
		return "", "", "", false
	}
	return baseURL, user, pass, true
}
