package provider

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync/atomic"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/syegournov/xkeen-gen/terraform-provider-xui/internal/provider/acctest"
)

// protoV6ProviderFactories is wired up by TestMain when TF_ACC=1 so that
// acceptance tests can attach the in-process provider to terraform.
var protoV6ProviderFactories map[string]func() (tfprotov6.ProviderServer, error)

// accPanel points at the shared 3x-ui container started by TestMain. Nil when
// TF_ACC is unset (unit-only runs).
var accPanel *acctest.Panel

// randPortCounter generates unique, stable ports per test within one run.
// Acceptance tests share a single panel, so every inbound must pick a free
// port; collisions would make tests order-dependent.
var randPortCounter atomic.Int64

func init() {
	// 24000..24999 is a generous, unprivileged range unlikely to clash with
	// anything the panel allocates on its own.
	randPortCounter.Store(24000)
}

// nextPort returns a monotonically increasing int port in [24000, 25000).
// Wraps silently if a single test binary somehow runs >1000 inbound steps.
func nextPort() int {
	return int(24000 + (randPortCounter.Add(1)-24001)%1000)
}

func TestMain(m *testing.M) {
	if os.Getenv("TF_ACC") == "" {
		os.Exit(m.Run())
	}

	ctx := context.Background()
	panel, stop, err := acctest.StartPanel(ctx)
	if err != nil {
		log.Fatalf("acctest: start 3x-ui panel: %v", err)
	}
	accPanel = panel
	protoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"xui": providerserver.NewProtocol6WithError(New("acc")()),
	}

	code := m.Run()
	stop()
	os.Exit(code)
}

// testAccPreCheck skips the calling test when TF_ACC is not set. Mirrors the
// gate that resource.Test applies internally so that any pre-test setup
// (fetching panel state, building HCL with providerConfig(), etc.) doesn't
// run during plain `go test ./...`.
func testAccPreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set; skipping acceptance test")
	}
}

// providerConfig returns an HCL block that wires the provider at the shared
// test panel. Must only be called from inside an acceptance test that has
// already passed testAccPreCheck.
func providerConfig() string {
	if accPanel == nil {
		// Defensive: callers should always go through testAccPreCheck.
		return `provider "xui" { base_url = "http://invalid", username = "x", password = "x", insecure_skip_verify = true }`
	}
	return fmt.Sprintf(`
provider "xui" {
  base_url             = %q
  username             = %q
  password             = %q
  insecure_skip_verify = true
}
`, accPanel.BaseURL, accPanel.Username, accPanel.Password)
}
