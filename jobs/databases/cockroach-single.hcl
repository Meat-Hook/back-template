job "cockroach" {
  namespace = "default"
  type = "service"
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
      port "http" {}
      port "tcp" {}
    }

    task "serve" {
      driver = "exec"

      resources {
        cpu = 500
        memory = 256
      }

      service {
        name = "cockroach"

        port = "tcp"

        tags = [
          "database",
          "single",
          "exec",
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
        command = "cockroach"
        args = [
          "start-single-node",
          "--certs-dir",
          "/opt/cockroach/certs",
          "--store",
          "/opt/cockroach/data",
          "--host",
          "X.X.X.X",
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

      artifact {
        source = "https://binaries.cockroachdb.com/cockroach-vX.X.X.linux-amd64.tgz"
      }

      logs {
        max_files = 10
        max_file_size = 2
      }
    }
  }
}
