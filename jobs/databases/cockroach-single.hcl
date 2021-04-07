job "cockroach-single-node" {
  namespace = "default"
  type = "service"
  region = "global"

  datacenters = [
    "DC",
  ]

  constraint {
    attribute = "${attr.unique.hostname}"
    value = "ADDR"
  }

  update {
    max_parallel = 1
    health_check = "checks"
    min_healthy_time = "90s"
    healthy_deadline = "5m"
    progress_deadline = "10m"
    auto_revert = true
    stagger = "30s"
  }

  group "databases" {
    volume "certs" {
      type = "host"
      read_only = true
      source = "cockroach-certs"
    }

    volume "data" {
      type = "host"
      read_only = false
      source = "cockroach-data"
    }

    restart {
      interval = "30m"
      attempts = 5
      delay = "15s"
      mode = "fail"
    }

    network {
      port "http" {
        static = 3500
      }
      port "tcp" {
        static = 26257
      }
    }

    task "serve" {
      driver = "docker"

      resources {
        cpu = 1000
        memory = 1024
      }

      service {
        name = "cockroach-single"

        port = "tcp"

        tags = [
          "database",
          "single",
          "docker",
        ]

        check {
          name = "alive-tcp"
          type = "tcp"
          port = "tcp"
          interval = "10s"
          timeout = "2s"

          check_restart {
            limit = 3
            grace = "60s"
            ignore_warnings = false
          }
        }

        check {
          name = "alive-http"
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

      config {
        image = "cockroachdb/cockroach:v20.2.7"

        ports = [
          "http",
          "tcp",
        ]

        hostname = "cockroach-single.service.consul"

        args = [
          "start-single-node",
          "--certs-dir",
          "/opt/cockroach/certs",
          "--store",
          "/opt/cockroach/data",
          "--host",
          "cockroach-single.service.consul",
          "--port",
          "${NOMAD_PORT_tcp}",
          "--http-port",
          "${NOMAD_PORT_http}"
        ]
      }

      volume_mount {
        volume = "certs"
        destination = "/opt/cockroach/certs"
        read_only = true
      }

      volume_mount {
        volume = "data"
        destination = "/opt/cockroach/data/"
        read_only = false
      }

      logs {
        max_files = 10
        max_file_size = 2
      }
    }
  }
}
