resource "xui_xray_template" "example" {
  json = jsonencode({
    log = {
      access      = "none"
      dnsLog      = false
      error       = ""
      loglevel    = "warning"
      maskAddress = ""
    }
    api = {
      tag      = "api"
      services = ["HandlerService", "LoggerService", "StatsService"]
    }
    inbounds = [{
      tag      = "api"
      listen   = "127.0.0.1"
      port     = 62789
      protocol = "dokodemo-door"
      settings = {
        address = "127.0.0.1"
      }
    }]
    outbounds = [
      {
        tag      = "direct"
        protocol = "freedom"
        settings = {
          domainStrategy = "AsIs"
          redirect       = ""
          noises         = []
        }
      },
      {
        tag      = "blocked"
        protocol = "blackhole"
        settings = {}
      },
    ]
    policy = {
      levels = {
        "0" = {
          statsUserDownlink = true
          statsUserUplink   = true
        }
      }
      system = {
        statsInboundDownlink  = true
        statsInboundUplink    = true
        statsOutboundDownlink = false
        statsOutboundUplink   = false
      }
    }
    routing = {
      domainStrategy = "AsIs"
      rules = [
        {
          type        = "field"
          inboundTag  = ["api"]
          outboundTag = "api"
        },
      ]
    }
    stats = {}
  })

  restart_xray = true
}
