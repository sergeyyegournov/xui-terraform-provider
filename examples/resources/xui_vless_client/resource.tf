resource "xui_vless_client" "example" {
  inbound_id = xui_inbound.example.id
  email      = "client@example.com"
}
