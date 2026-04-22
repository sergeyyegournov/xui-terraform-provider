package provider

import "testing"

func TestFillInboundDSModelFromRaw(t *testing.T) {
	t.Parallel()

	raw := []byte(`{
		"id": 42,
		"remark": "test",
		"listen": "",
		"port": 443,
		"protocol": "vless",
		"enable": true,
		"settings": "{\"clients\":[]}",
		"streamSettings": "{}",
		"sniffing": "{}"
	}`)

	cfg := inboundDSModel{}
	if err := fillInboundDSModelFromRaw(raw, &cfg); err != nil {
		t.Fatalf("fillInboundDSModelFromRaw() error = %v", err)
	}
	if cfg.Port.ValueInt64() != 443 {
		t.Fatalf("expected port 443, got %d", cfg.Port.ValueInt64())
	}
	if cfg.Protocol.ValueString() != "vless" {
		t.Fatalf("expected protocol vless, got %s", cfg.Protocol.ValueString())
	}
	if cfg.JSON.ValueString() == "" {
		t.Fatalf("expected json to be populated")
	}
}
