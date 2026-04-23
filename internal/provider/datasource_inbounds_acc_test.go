package provider

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccInboundsDataSource_basic(t *testing.T) {
	testAccPreCheck(t)
	port := nextPort()
	remark := fmt.Sprintf("tf-acc-ds-all-%d", port)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactories,
		CheckDestroy:             checkInboundDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccInboundsDSConfig(remark, port),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.xui_inbounds.all", "json"),
					resource.TestCheckResourceAttrSet("data.xui_inbounds.vless", "json"),
					checkInboundInDataSourceJSON("data.xui_inbounds.all", remark),
					checkInboundInDataSourceJSON("data.xui_inbounds.vless", remark),
					checkAllInboundsHaveProtocol("data.xui_inbounds.vless", "vless"),
				),
			},
		},
	})
}

func TestAccInboundDataSource_basic(t *testing.T) {
	testAccPreCheck(t)
	port := nextPort()
	remark := fmt.Sprintf("tf-acc-ds-one-%d", port)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactories,
		CheckDestroy:             checkInboundDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccInboundDSConfig(remark, port),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.xui_inbound.by_id", "remark", remark),
					resource.TestCheckResourceAttr("data.xui_inbound.by_id", "protocol", "vless"),
					resource.TestCheckResourceAttr("data.xui_inbound.by_id", "port", fmt.Sprintf("%d", port)),
					resource.TestCheckResourceAttrPair(
						"data.xui_inbound.by_id", "id",
						"xui_inbound.test", "id",
					),
				),
			},
		},
	})
}

func testAccInboundsDSConfig(remark string, port int) string {
	return fmt.Sprintf(`%s

resource "xui_inbound" "test" {
  protocol = "vless"
  remark   = %q
  port     = %d
  settings = jsonencode({ clients = [], decryption = "none" })
  stream_settings = jsonencode({
    network  = "tcp"
    security = "none"
    tcpSettings = {
      acceptProxyProtocol = false
      header              = { type = "none" }
    }
  })
  sniffing = "{}"
}

data "xui_inbounds" "all" {
  depends_on = [xui_inbound.test]
}

data "xui_inbounds" "vless" {
  protocol   = "vless"
  depends_on = [xui_inbound.test]
}
`, providerConfig(), remark, port)
}

func testAccInboundDSConfig(remark string, port int) string {
	return fmt.Sprintf(`%s

resource "xui_inbound" "test" {
  protocol = "vless"
  remark   = %q
  port     = %d
  settings = jsonencode({ clients = [], decryption = "none" })
  stream_settings = jsonencode({
    network  = "tcp"
    security = "none"
    tcpSettings = {
      acceptProxyProtocol = false
      header              = { type = "none" }
    }
  })
  sniffing = "{}"
}

data "xui_inbound" "by_id" {
  id = xui_inbound.test.id
}
`, providerConfig(), remark, port)
}

// checkInboundInDataSourceJSON asserts the `json` attribute of a list-style
// data source contains an inbound whose remark matches `wantRemark`.
func checkInboundInDataSourceJSON(resourceName, wantRemark string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("data source %q not in state", resourceName)
		}
		raw := rs.Primary.Attributes["json"]
		if raw == "" {
			return fmt.Errorf("data source %q has empty json attribute", resourceName)
		}
		var list []map[string]any
		if err := json.Unmarshal([]byte(raw), &list); err != nil {
			return fmt.Errorf("decode %s.json: %w", resourceName, err)
		}
		for _, in := range list {
			if r, _ := in["remark"].(string); r == wantRemark {
				return nil
			}
		}
		return fmt.Errorf("inbound with remark %q not found in %s (got %d inbounds)", wantRemark, resourceName, len(list))
	}
}

// checkAllInboundsHaveProtocol asserts every inbound returned by the list
// data source has the given protocol, confirming the server-side filter
// applied.
func checkAllInboundsHaveProtocol(resourceName, wantProto string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("data source %q not in state", resourceName)
		}
		raw := rs.Primary.Attributes["json"]
		var list []map[string]any
		if err := json.Unmarshal([]byte(raw), &list); err != nil {
			return fmt.Errorf("decode %s.json: %w", resourceName, err)
		}
		for i, in := range list {
			proto, _ := in["protocol"].(string)
			if proto != wantProto {
				return fmt.Errorf("%s: inbound %d has protocol %q, want %q", resourceName, i, proto, wantProto)
			}
		}
		return nil
	}
}
