package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVLESSClient_basic(t *testing.T) {
	testAccPreCheck(t)
	port := nextPort()
	inboundRemark := fmt.Sprintf("tf-acc-vless-%d", port)
	email := fmt.Sprintf("tf-acc-user-%d", port)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactories,
		CheckDestroy:             checkInboundDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccVLESSClientConfig(inboundRemark, port, email, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("xui_vless_client.test", "id"),
					resource.TestCheckResourceAttrSet("xui_vless_client.test", "uuid"),
					resource.TestCheckResourceAttr("xui_vless_client.test", "email", email),
					resource.TestCheckResourceAttr("xui_vless_client.test", "enable", "true"),
					resource.TestCheckResourceAttr("xui_vless_client.test", "limit_ip", "0"),
					resource.TestCheckResourceAttrPair(
						"xui_vless_client.test", "inbound_id",
						"xui_inbound.test", "id",
					),
				),
			},
			{
				Config: testAccVLESSClientConfig(inboundRemark, port, email, 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("xui_vless_client.test", "limit_ip", "3"),
					resource.TestCheckResourceAttr("xui_vless_client.test", "email", email),
				),
			},
			{
				ResourceName:      "xui_vless_client.test",
				ImportState:       true,
				ImportStateIdFunc: importVLESSClientIDFunc("xui_vless_client.test"),
				ImportStateVerify: true,
				// Optional panel fields round-trip from the panel as null
				// when they were never set; the user sees them as null in
				// imported state even though they passed empty-string
				// defaults at create time.
				ImportStateVerifyIgnore: []string{"flow", "sub_id", "comment"},
			},
		},
	})
}

func testAccVLESSClientConfig(remark string, port int, email string, limitIP int) string {
	return fmt.Sprintf(`%s

resource "xui_inbound" "test" {
  protocol = "vless"
  remark   = %q
  port     = %d
  settings = jsonencode({
    clients    = []
    decryption = "none"
  })
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

resource "xui_vless_client" "test" {
  inbound_id = xui_inbound.test.id
  email      = %q
  limit_ip   = %d
}
`, providerConfig(), remark, port, email, limitIP)
}
