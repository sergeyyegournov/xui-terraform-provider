resource "xui_inbound" "example" {
  protocol = "vless"
  remark   = "example"
  listen   = ""
  port     = 443
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
