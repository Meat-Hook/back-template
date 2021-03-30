job "user" {
  namespace = "default"
  type = "system"
  region = "global"

  datacenters = [
    "dc1",
  ]

  update {
    max_parallel = 1
    health_check = "checks"
    min_healthy_time = "60s"
    healthy_deadline = "5m"
    progress_deadline = "10m"
    auto_revert = true
    auto_promote = true
    canary = 1
    stagger = "30s"
  }

  group "server" {
    count = 1

    restart {
      interval = "30m"
      attempts = 5
      delay = "15s"
      mode = "fail"
    }

    network {
      port "http" {
        to = 8080
      }

      port "grpc" {
        to = 8090
      }

      port "metric" {
        to = 8100
      }
    }

    service {
      name = "caddy"

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
https://meathook-consul.ru {
  {{ range service "consul" }}
    reverse_proxy {{ .Address }}:8500
  {{ end }}

  basicauth {
     Edgar JDJhJDE0JDBxVTlkMENWUUZSZEVyemtSeURhaGVoLmRKb0FOZUtqY2dGMHVpTGs0cDlXbVg3RVRLeVE2
  }
}

https://meathook-nomad.ru {
  {{ range service "nomad-client" }}
    reverse_proxy {{ .Address }}:4646
  {{ end }}

  basicauth {
     Edgar JDJhJDE0JDBxVTlkMENWUUZSZEVyemtSeURhaGVoLmRKb0FOZUtqY2dGMHVpTGs0cDlXbVg3RVRLeVE2
  }
}
EOF
        destination = "configs/Caddyfile"
        change_mode = "restart"
      }
    }
  }
}
