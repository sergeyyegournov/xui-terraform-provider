provider "xui" {
  base_url             = "https://panel.example.com/your-random-path/"
  username             = var.xui_username
  password             = var.xui_password
  insecure_skip_verify = false
}
