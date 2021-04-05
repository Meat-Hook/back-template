job "traefik" {
  namespace = "default"
  type = "system"
  region = "global"

  datacenters = [
    "dc1",
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

  group "load-balancer" {
    restart {
      interval = "30m"
      attempts = 5
      delay = "15s"
      mode = "fail"
    }

    network {
      port "http" {
        static = 8080
      }
    }

    service {
      name = "load-balancer"

      port = "http"

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
        image = "traefik:v2.4"

        network_mode = "host"

        ports = [
          "http",
        ]

        volumes = [
          "local/traefik.yml:/etc/traefik/traefik.yml",
        ]
      }

      logs {
        max_files = 10
        max_file_size = 2
      }

      template {
        data = <<EOF
entryPoints:
  web:
   address: ":8080"

providers:
  consulCatalog:
    prefix: "load-balancer"
    refreshInterval: 30s
    requireConsistent: true
    endpoint:
      datacenter: dc1
      address: consul.service.consul:8500

EOF

        destination = "local/traefik.yml"
      }
    }
  }
}
