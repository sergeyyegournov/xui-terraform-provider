package provider

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestFilterInboundsJSONByProtocol(t *testing.T) {
	t.Parallel()

	raw := []byte(`[{"id":1,"protocol":"vless"},{"id":2,"protocol":"vmess"}]`)
	out, err := filterInboundsJSON(raw, "vless")
	if err != nil {
		t.Fatalf("filterInboundsJSON() error = %v", err)
	}
	var arr []map[string]any
	if err := json.Unmarshal(out, &arr); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if len(arr) != 1 {
		t.Fatalf("expected one inbound, got %d", len(arr))
	}
	if stringFromMap(arr[0], "protocol") != "vless" {
		t.Fatalf("unexpected protocol: %s", stringFromMap(arr[0], "protocol"))
	}
}

func TestFilterInboundsJSONNoFilter(t *testing.T) {
	t.Parallel()

	raw := []byte(`[{"id":1,"protocol":"vless"},{"id":2,"protocol":"vmess"}]`)
	out, err := filterInboundsJSON(raw, "")
	if err != nil {
		t.Fatalf("filterInboundsJSON() error = %v", err)
	}
	compact := strings.ReplaceAll(string(out), " ", "")
	if !strings.Contains(compact, `"id":1`) || !strings.Contains(compact, `"id":2`) {
		t.Fatalf("expected both records, got %s", string(out))
	}
}
