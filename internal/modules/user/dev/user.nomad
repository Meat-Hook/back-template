job "user" {
  datacenters = [
    "dc1",
  ]

  type = "service"

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

  migrate {
    max_parallel = 1
    health_check = "checks"
    min_healthy_time = "10s"
    healthy_deadline = "5m"
  }

  group "db" {
    count = 1

    restart {
      interval = "5m"
      attempts = 10
      delay = "30s"
      mode = "delay"
    }

    network {
      port "admin" {
        to = 8080
      }
      port "tcp" {
        to = 26257
      }
    }

    service {
      tags = [
        "dev",
        "db",
        "user",
      ]

      check {
        name = "http-check"
        type = "tcp"
        port = "admin"
        interval = "10s"
        timeout = "2s"
      }

      check {
        name = "tcp-check"
        type = "tcp"
        port = "tcp"
        interval = "10s"
        timeout = "2s"
      }
    }

    task "node" {
      driver = "docker"

      config {
        image = "cockroachdb/cockroach:v20.2.0"

        args = [
          "start-single-node",
          "--insecure",
        ]

        ports = [
          "tcp",
          "admin"
        ]

        network_mode = "bridge"

        hostname = "user-db.service.consul"

        dns_servers = [
          "${attr.unique.network.ip-address}"]
      }

      resources {
        memory = 256
        cpu = 100
      }

      logs {
        max_files = 10
        max_file_size = 10
      }
    }
  }

  group "app" {
    count = 1

    restart {
      interval = "5m"
      attempts = 10
      delay = "25s"
      mode = "delay"
    }

    network {
      port "http" {
        to = 8080
      }
    }

    service {
      tags = [
        "dev",
        "app",
        "user",
      ]

      check {
        name = "http-health"
        path = "/health"
        type = "http"
        port = "http"
        interval = "10s"
        timeout = "2s"
      }
    }

    task "service" {
      driver = "docker"

      env {
        DB_NAME = "postgres"
        DB_USER = "root"
        DB_PASS = "root"
        DB_HOST = "user-db.service.consul"
        DB_PORT = 26257

        NATS = "nats.service.consul:4222"
      }

      resources {
        memory = 256
        cpu = 100
      }

      config {
        image = "docker.pkg.github.com/meat-hook/back-template/user:dev"

        command = "/user"

        ports = [
          "http"
        ]

        args = [
          "start",
        ]

        dns_servers = [
          "${attr.unique.network.ip-address}",
        ]

        network_mode = "bridge"
        hostname = "user.service.consul"

        labels {
          group = "user"
        }
      }

      logs {
        max_files = 10
        max_file_size = 10
      }
    }
  }
}
