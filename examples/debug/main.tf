# Dev workflow: ./dev-setup.sh  then  export TF_CLI_CONFIG_FILE="$(pwd)/dev.terraform.rc"
# Skip terraform init when using dev_overrides. Use validate / plan / apply.

terraform {
  required_providers {
    xui = {
      source  = "example.com/xui/xui"
      version = "0.0.1"
    }
  }
}

variable "xui_base_url" {
  type        = string
  description = "Panel root URL including the random path prefix and trailing slash."
}

variable "xui_username" {
  type = string
}

variable "xui_password" {
  type      = string
  sensitive = true
}

variable "debug_inbound_port" {
  type        = number
  description = "Listen port for the example inbound (must be free on the server)."
  default     = 24444
}

variable "debug_inbound_remark" {
  type        = string
  description = "Remark for the example inbound."
  default     = "terraform-debug-inbound"
}

provider "xui" {
  base_url             = var.xui_base_url
  username             = var.xui_username
  password             = var.xui_password
  insecure_skip_verify = true
}

data "xui_inbounds" "all" {}

data "xui_inbounds" "vless_only" {
  protocol = "vless"
}

resource "xui_inbound" "debug" {
  protocol = "vless"
  remark   = var.debug_inbound_remark
  listen   = ""
  port     = var.debug_inbound_port
  settings = jsonencode({
    clients    = []
    decryption = "none"
  })
  stream_settings = jsonencode({
    network  = "tcp"
    security = "none"
    tcpSettings = {
      acceptProxyProtocol = false
      header = {
        type = "none"
      }
    }
  })
  sniffing = "{}"
}

resource "xui_vless_client" "debug" {
  inbound_id = xui_inbound.debug.id
  email      = "terraform-debug-xui"
}

output "all_inbounds_json" {
  description = "Full list from the panel; use jsondecode() locally if you need other ids."
  value       = data.xui_inbounds.all.json
}

output "vless_inbounds_json" {
  value = data.xui_inbounds.vless_only.json
}

output "debug_inbound" {
  value = {
    id              = xui_inbound.debug.id
    tag             = xui_inbound.debug.tag
    remark          = xui_inbound.debug.remark
    port            = xui_inbound.debug.port
    protocol        = xui_inbound.debug.protocol
    stream_settings = xui_inbound.debug.stream_settings
  }
}

output "debug_client_id" {
  value = xui_vless_client.debug.id
}
