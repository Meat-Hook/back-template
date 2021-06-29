job "cockroach-single" {
  namespace = "default"
  type = "service"
  region = "global"

  datacenters = [
    "home-server",
  ]

  constraint {
    attribute = "${attr.unique.hostname}"
    value = "home-server"
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

    ephemeral_disk {
      migrate = true
      size = 10240
      sticky = true
    }

    restart {
      interval = "30m"
      attempts = 5
      delay = "15s"
      mode = "fail"
    }

    network {
      mode = "host"
      port "http" {
        static = 3636
        to = 8080
      }
      port "tcp" {
        static = 26257
        to = 26257
      }
    }

    // admin dashboard
    service {
      name = "cockroach-dashboard"

      port = "http"

      tags = [
        "database",
        "single",
        "docker",
        "admin",
      ]

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

    // database
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
    }

    task "serve" {
      driver = "docker"

      resources {
        cpu = 1024
        memory = 2048
      }

      config {
        image = "cockroachdb/cockroach:v21.1.2"

        ports = [
          "http",
          "tcp",
        ]

        args = [
          "start-single-node",
          "--certs-dir",
          "/opt/cockroach/certs",
          "--store",
          "/opt/cockroach/data/single-node/",
          "--host",
          "0.0.0.0",
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
