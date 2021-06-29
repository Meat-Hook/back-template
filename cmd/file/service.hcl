job "session" {
  namespace = "default"
  type = "service"
  region = "global"

  datacenters = [
    "dc",
  ]

  constraint {
    distinct_hosts = true
  }

  update {
    max_parallel = 1
    health_check = "checks"
    min_healthy_time = "10s"
    healthy_deadline = "5m"
    progress_deadline = "10m"
    auto_revert = true
    auto_promote = true
    canary = 1
    stagger = "30s"
  }

  group "server" {

    restart {
      interval = "30m"
      attempts = 5
      delay = "15s"
      mode = "fail"
    }

    network {
      mode = "host"
      port "http" {
        static = 10001
        to = 8080
      }
      port "grpc" {
        static = 15001
        to = 8090
      }
      port "metric" {
        static = 20001
        to = 8100
      }
      dns {
        servers = [
          "192.168.31.116",
        ]
      }
    }

    // http service
    service {
      name = "session-http"

      port = "http"

      tags = [
        "service",
        "docker",
        "http",
      ]

      check {
        name = "alive-http"
        type = "http"
        port = "http"
        path = "/health"
        interval = "10s"
        timeout = "2s"

        check_restart {
          limit = 3
          grace = "60s"
          ignore_warnings = false
        }
      }
    }

    // grpc service
    service {
      name = "session-grpc"

      port = "grpc"

      tags = [
        "service",
        "docker",
        "grpc",
      ]

      //      TODO add grpc check.
      check {
        name = "alive-grpc"
        type = "tcp"
        port = "metric"
        interval = "10s"
        timeout = "2s"

        check_restart {
          limit = 3
          grace = "60s"
          ignore_warnings = false
        }
      }
    }

    // metric service
    service {
      name = "session-metric"

      port = "metric"

      tags = [
        "service",
        "docker",
        "metric",
      ]

      check {
        name = "alive-metric"
        type = "tcp"
        port = "metric"
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
        image = "back-template/session-service:v0.1.0"

        ports = [
          "http",
          "grpc",
          "metric",
        ]

        command = "./session"
      }

      template {
        data = <<EOH
DB_NAME="session_db"
DB_USER="user"
DB_PASS="PASS"
DB_HOST="cockroach-single.service.consul"
DB_PORT="26257"
DB_SSL_MODE="require"
AUTH_KEY=super-duper-secret-key-for-tests
USER_SRV="user-grpc.service.consul:15000"
MIGRATE="true"
EOH

        destination = "secrets/file.env"
        change_mode = "restart"
        env = true
      }

      logs {
        max_files = 10
        max_file_size = 2
      }
    }
  }
}
