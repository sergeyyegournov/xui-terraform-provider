package provider

import (
	"strings"
	"testing"
)

func TestMergeInboundSettingsPreservingClients(t *testing.T) {
	t.Parallel()

	server := `{"clients":[{"id":"existing","email":"a"}],"decryption":"none","fallbacks":[]}`
	user := `{"clients":[{"id":"new","email":"b"}],"decryption":"none","udp":true}`

	merged, err := mergeInboundSettingsPreservingClients(server, user)
	if err != nil {
		t.Fatalf("mergeInboundSettingsPreservingClients() error = %v", err)
	}

	if !strings.Contains(merged, `"udp":true`) {
		t.Fatalf("expected user key to be merged, got %s", merged)
	}
	if !strings.Contains(merged, `"id":"existing"`) {
		t.Fatalf("expected server clients to be preserved, got %s", merged)
	}
	if strings.Contains(merged, `"id":"new"`) {
		t.Fatalf("expected user clients to be ignored, got %s", merged)
	}
}

func TestEnsureDummyInboundClient(t *testing.T) {
	t.Parallel()

	settings := `{"clients":[],"decryption":"none"}`
	updated, dummyID, err := ensureDummyInboundClient(settings, "")
	if err != nil {
		t.Fatalf("ensureDummyInboundClient() error = %v", err)
	}
	if dummyID == "" {
		t.Fatalf("expected generated dummy UUID")
	}
	if !strings.Contains(updated, inboundDummyClientEmail) {
		t.Fatalf("expected dummy client email in settings: %s", updated)
	}
}

func TestFindDummyClientUUID(t *testing.T) {
	t.Parallel()

	settings := `{"clients":[{"id":"11111111-1111-1111-1111-111111111111","email":"__xui_tf_do_not_delete__"}]}`
	id, err := findDummyClientUUID(settings)
	if err != nil {
		t.Fatalf("findDummyClientUUID() error = %v", err)
	}
	if id != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("unexpected dummy client id: %s", id)
	}
}
