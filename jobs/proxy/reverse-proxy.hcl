job "caddy" {
  namespace = "default"
  type = "system"
  region = "global"

  datacenters = [
    "DC",
  ]

  update {
    max_parallel = 1
    health_check = "checks"
    min_healthy_time = "90s"
    healthy_deadline = "5m"
    progress_deadline = "10m"
    auto_revert = true
    auto_promote = true
    canary = 1
    stagger = "30s"
  }

  group "proxy" {
    restart {
      interval = "30m"
      attempts = 5
      delay = "15s"
      mode = "fail"
    }

    network {
      port "http" {
        static = 80
      }
      port "https" {
        static = 443
      }
    }

    service {
      name = "reverse-proxy"

      port = "https"

      tags = [
        "https",
        "docker",
        "reverse-proxy",
      ]

      check {
        name = "alive"
        type = "tcp"
        port = "https"
        interval = "10s"
        timeout = "2s"

        check_restart {
          limit = 3
          grace = "60s"
          ignore_warnings = false
        }
      }
    }

    service {
      name = "reverse-proxy"

      port = "http"

      tags = [
        "http",
        "docker",
        "reverse-proxy",
      ]

      check {
        name = "alive"
        type = "tcp"
        port = "http"
        interval = "10s"
        timeout = "2s"

        check_restart {
          limit = 3
          grace = "60s"
          ignore_warnings = false
        }
      }
    }

    task "serve" {
      driver = "docker"

      resources {
        cpu = 100
        memory = 128
      }

      config {
        image = "caddy:2-alpine"

        ports = [
          "http",
          "https",
        ]

        volumes = [
          "configs/Caddyfile:/etc/caddy/Caddyfile",
        ]
      }

      logs {
        max_files = 10
        max_file_size = 2
      }

      template {
        data = <<EOF
https://domain.com {
  reverse_proxy http://load-balancer.service.consul:8080
}

https://domain-consul.ru {
  reverse_proxy http://consul.service.consul:8500
  basicauth {
     Edgar JDJhJDE0JDBxVTlkMENWUUZSZEVyemtSeURhaGVoLmRKb0FOZUtqY2dGMHVpTGs0cDlXbVg3RVRLeVE2
  }
}

https://domain-nomad.ru {
  reverse_proxy http://nomad-client.service.consul:4646
  basicauth {
     admin JDJhJDE0JGZXSFh0L3lKL0x3M2RDTWNvMUhoWk9yQlQ2TTVveEFKZ2x6anh2MHZwLlYySnNDeTBWU0oy
  }
}

https://domain-database.ru {
  reverse_proxy http://cockroach-single.service.consul:3500
  basicauth {
     admin JDJhJDE0JGZXSFh0L3lKL0x3M2RDTWNvMUhoWk9yQlQ2TTVveEFKZ2x6anh2MHZwLlYySnNDeTBWU0oy
  }
}
EOF
        destination = "configs/Caddyfile"
        change_mode = "restart"
      }
    }
  }
}
